package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type friendshipBalanceRepositoryGorm struct {
	crud.Repository[users.FriendshipBalance]
}

func NewFriendshipBalanceRepository(db *gorm.DB) *friendshipBalanceRepositoryGorm {
	return &friendshipBalanceRepositoryGorm{
		crud.NewRepository[users.FriendshipBalance](db),
	}
}

func (fbr *friendshipBalanceRepositoryGorm) UpsertMany(ctx context.Context, balances []users.FriendshipBalance) error {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipBalanceRepository.UpsertMany")
	defer span.End()

	if len(balances) == 0 {
		return nil
	}

	db, err := fbr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "friendship_id"}, {Name: "currency"}},
		DoUpdates: clause.AssignmentColumns([]string{"net_balance"}),
	}).Create(&balances).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	return nil
}

func (fbr *friendshipBalanceRepositoryGorm) FindAllByFriendshipID(ctx context.Context, friendshipID uuid.UUID) ([]users.FriendshipBalance, error) {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipBalanceRepository.FindAllByFriendshipID")
	defer span.End()

	db, err := fbr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	var balances []users.FriendshipBalance
	if err := db.Where("friendship_id = ?", friendshipID).Find(&balances).Error; err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return balances, nil
}

func (fbr *friendshipBalanceRepositoryGorm) FindAllByProfileID(ctx context.Context, profileID uuid.UUID) ([]repository.FriendshipBalanceRow, error) {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipBalanceRepository.FindAllByProfileID")
	defer span.End()

	db, err := fbr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	var rows []repository.FriendshipBalanceRow
	err = db.Table("friendship_balances").
		Select("friendships.profile_id1 AS profile_id1, friendships.profile_id2 AS profile_id2, friendship_balances.currency AS currency, friendship_balances.net_balance AS net_balance").
		Joins("JOIN friendships ON friendships.id = friendship_balances.friendship_id").
		Where("(friendships.profile_id1 = ? OR friendships.profile_id2 = ?) AND friendship_balances.net_balance <> 0", profileID, profileID).
		Scan(&rows).
		Error

	if err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return rows, nil
}

func (fbr *friendshipBalanceRepositoryGorm) DeleteZeroBalances(ctx context.Context, friendshipIDs []uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipBalanceRepository.DeleteZeroBalances")
	defer span.End()

	if len(friendshipIDs) == 0 {
		return nil
	}

	db, err := fbr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	if err := db.Where("friendship_id IN ? AND net_balance = 0", friendshipIDs).
		Delete(&users.FriendshipBalance{}).Error; err != nil {
		return ungerr.Wrap(err, "error deleting settled friendship balance rows")
	}

	return nil
}
