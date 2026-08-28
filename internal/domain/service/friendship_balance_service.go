package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/shopspring/decimal"
)

// friendshipBalanceServiceImpl depends only on repositories - never on DebtService,
// FriendshipService, or ProfileService - so both of those (which already depend on each other
// via ProfileService) can depend on this without a cycle.
type friendshipBalanceServiceImpl struct {
	debtTransactionRepository repository.DebtTransactionRepository
	friendshipRepository      repository.FriendshipRepository
	balanceRepository         repository.FriendshipBalanceRepository
}

func NewFriendshipBalanceService(
	debtTransactionRepository repository.DebtTransactionRepository,
	friendshipRepository repository.FriendshipRepository,
	balanceRepository repository.FriendshipBalanceRepository,
) FriendshipBalanceService {
	return &friendshipBalanceServiceImpl{
		debtTransactionRepository,
		friendshipRepository,
		balanceRepository,
	}
}

func (fbs *friendshipBalanceServiceImpl) RecalculatePair(ctx context.Context, profileID1, profileID2 uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipBalanceService.RecalculatePair")
	defer span.End()

	// Locks the friendship row first so a concurrent recalculation for the same pair blocks
	// until this one commits, instead of both reading a transaction set that's missing the
	// other's just-written row and racing to overwrite the cache with a stale value.
	friendship, err := fbs.friendshipRepository.FindByProfileIDsForUpdate(ctx, profileID1, profileID2)
	if err != nil {
		return err
	}
	if friendship.IsZero() {
		// No Friendship row for this pair - matches today's read-time behavior of silently
		// dropping balances for non-friend counterparties, so just skip caching it.
		return nil
	}

	transactions, err := fbs.debtTransactionRepository.FindAllByMultipleProfileIDs(
		ctx, []uuid.UUID{friendship.ProfileID1}, []uuid.UUID{friendship.ProfileID2},
	)
	if err != nil {
		return err
	}

	netByCounterparty := mapper.NetBalanceByFriend(transactions, []uuid.UUID{friendship.ProfileID1})

	balances := make([]users.FriendshipBalance, 0, len(netByCounterparty[friendship.ProfileID2]))
	for currency, amount := range netByCounterparty[friendship.ProfileID2] {
		balances = append(balances, users.FriendshipBalance{
			FriendshipID: friendship.ID,
			Currency:     currency,
			NetBalance:   amount,
		})
	}

	if err := fbs.balanceRepository.UpsertMany(ctx, balances); err != nil {
		return err
	}

	// A currency that nets to exactly 0 (e.g. a debt fully settled) is still upserted above so
	// any previously-nonzero row gets corrected first, then swept away here so the table only
	// ever holds nonzero balances - readers rely on that (see FindAllByProfileID).
	return fbs.balanceRepository.DeleteZeroBalances(ctx, []uuid.UUID{friendship.ID})
}

func (fbs *friendshipBalanceServiceImpl) RecalculateAllForProfile(ctx context.Context, profileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipBalanceService.RecalculateAllForProfile")
	defer span.End()

	friendships, err := fbs.friendshipRepository.FindAllByProfileIDForUpdate(ctx, profileID)
	if err != nil {
		return err
	}
	if len(friendships) == 0 {
		return nil
	}

	transactions, err := fbs.debtTransactionRepository.FindAllByProfileIDs(ctx, []uuid.UUID{profileID}, -1, false)
	if err != nil {
		return err
	}

	netByCounterparty := mapper.NetBalanceByFriend(transactions, []uuid.UUID{profileID})

	friendshipIDs := make([]uuid.UUID, len(friendships))
	var balances []users.FriendshipBalance
	for i, f := range friendships {
		friendshipIDs[i] = f.ID

		counterpartyID := f.ProfileID1
		if counterpartyID == profileID {
			counterpartyID = f.ProfileID2
		}

		for currency, amount := range netByCounterparty[counterpartyID] {
			netForProfile1 := amount
			if f.ProfileID1 != profileID {
				// amount is signed relative to profileID (the queried profile), but the cache
				// convention stores balances relative to the friendship's own ProfileID1.
				netForProfile1 = amount.Neg()
			}
			balances = append(balances, users.FriendshipBalance{
				FriendshipID: f.ID,
				Currency:     currency,
				NetBalance:   netForProfile1,
			})
		}
	}

	if err := fbs.balanceRepository.UpsertMany(ctx, balances); err != nil {
		return err
	}

	// Same zero-balance sweep as RecalculatePair, batched across every friendship touched by
	// this profile-wide recalculation.
	return fbs.balanceRepository.DeleteZeroBalances(ctx, friendshipIDs)
}

func (fbs *friendshipBalanceServiceImpl) GetNetBalancesForProfile(ctx context.Context, profileID uuid.UUID) (map[uuid.UUID]map[string]decimal.Decimal, error) {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipBalanceService.GetNetBalancesForProfile")
	defer span.End()

	rows, err := fbs.balanceRepository.FindAllByProfileID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]map[string]decimal.Decimal)
	for _, row := range rows {
		if row.NetBalance.IsZero() {
			// Defensive: FindAllByProfileID already filters net_balance <> 0 in SQL, this just
			// guards against that filter ever drifting.
			continue
		}

		counterpartyID := row.ProfileID2
		netBalance := row.NetBalance
		if row.ProfileID1 != profileID {
			counterpartyID = row.ProfileID1
			netBalance = netBalance.Neg()
		}

		if result[counterpartyID] == nil {
			result[counterpartyID] = make(map[string]decimal.Decimal)
		}
		result[counterpartyID][row.Currency] = netBalance
	}

	return result, nil
}
