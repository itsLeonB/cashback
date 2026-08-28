package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type friendshipRepositoryGorm struct {
	crud.Repository[users.Friendship]
}

func NewFriendshipRepository(db *gorm.DB) *friendshipRepositoryGorm {
	return &friendshipRepositoryGorm{
		crud.NewRepository[users.Friendship](db),
	}
}

func (fr *friendshipRepositoryGorm) Insert(ctx context.Context, friendship users.Friendship) (users.Friendship, error) {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipRepository.Insert")
	defer span.End()

	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return users.Friendship{}, err
	}

	if err = db.Create(&friendship).Error; err != nil {
		return users.Friendship{}, ungerr.Wrap(err, appconstant.ErrDataInsert)
	}

	return friendship, nil
}

func (fr *friendshipRepositoryGorm) FindFirstBySpec(ctx context.Context, spec users.FriendshipSpecification) (users.Friendship, error) {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipRepository.FindFirstBySpec")
	defer span.End()

	var friendship users.Friendship

	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return users.Friendship{}, err
	}

	query := db.
		Scopes(
			crud.WhereBySpec(spec.Model),
			crud.PreloadRelations(spec.PreloadRelations),
		).
		Joins("JOIN user_profiles AS up1 ON up1.id = friendships.profile_id1").
		Joins("JOIN user_profiles AS up2 ON up2.id = friendships.profile_id2")

	if spec.Name != "" {
		query = query.Where(
			db.Where("up1.name = ? AND friendships.profile_id1 <> ?", spec.Name, spec.Model.ProfileID1).
				Or("up2.name = ? AND friendships.profile_id2 <> ?", spec.Name, spec.Model.ProfileID1),
		)
	}

	err = query.Take(&friendship).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return users.Friendship{}, nil
		}
		return users.Friendship{}, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return friendship, nil
}

func (fr *friendshipRepositoryGorm) FindAllBySpec(ctx context.Context, spec users.FriendshipSpecification) ([]users.Friendship, error) {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipRepository.FindAllBySpec")
	defer span.End()

	var friendships []users.Friendship

	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	err = db.
		Where(users.Friendship{ProfileID1: spec.Model.ProfileID1}).
		Or(users.Friendship{ProfileID2: spec.Model.ProfileID1}).
		Scopes(
			crud.PreloadRelations(spec.PreloadRelations),
			crud.DefaultOrder(),
		).
		Find(&friendships).
		Error

	if err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return friendships, nil
}

func (fr *friendshipRepositoryGorm) FindByProfileIDs(ctx context.Context, profileID1, profileID2 uuid.UUID) (users.Friendship, error) {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipRepository.FindByProfileIDs")
	defer span.End()

	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return users.Friendship{}, err
	}

	return fr.findByProfileIDsQuery(db.Preload("Profile1").Preload("Profile2"), profileID1, profileID2)
}

// FindByProfileIDsForUpdate is FindByProfileIDs with SELECT ... FOR UPDATE, used to serialize
// concurrent balance recalculations for the same pair.
func (fr *friendshipRepositoryGorm) FindByProfileIDsForUpdate(ctx context.Context, profileID1, profileID2 uuid.UUID) (users.Friendship, error) {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipRepository.FindByProfileIDsForUpdate")
	defer span.End()

	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return users.Friendship{}, err
	}

	return fr.findByProfileIDsQuery(db.Clauses(clause.Locking{Strength: "UPDATE"}), profileID1, profileID2)
}

// findByProfileIDsQuery applies the pair lookup shared by FindByProfileIDs and
// FindByProfileIDsForUpdate onto an already-configured query (preloads, locking, ...).
func (fr *friendshipRepositoryGorm) findByProfileIDsQuery(query *gorm.DB, profileID1, profileID2 uuid.UUID) (users.Friendship, error) {
	var friendship users.Friendship
	err := query.
		Where("(profile_id1 = ? AND profile_id2 = ?) OR (profile_id1 = ? AND profile_id2 = ?)", profileID1, profileID2, profileID2, profileID1).
		First(&friendship).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return users.Friendship{}, nil
		}
		return users.Friendship{}, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return friendship, nil
}

// FindAllByProfileIDForUpdate is the bulk equivalent of FindByProfileIDsForUpdate, locking every
// friendship row profileID is party to.
func (fr *friendshipRepositoryGorm) FindAllByProfileIDForUpdate(ctx context.Context, profileID uuid.UUID) ([]users.Friendship, error) {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipRepository.FindAllByProfileIDForUpdate")
	defer span.End()

	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	var friendships []users.Friendship
	err = db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("profile_id1 = ? OR profile_id2 = ?", profileID, profileID).
		Find(&friendships).
		Error

	if err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return friendships, nil
}

// RepointFriendships repoints every friendship row involving anonProfileID onto realProfileID.
// If realProfileID already has a friendship with the same counterparty, the anonymous row is
// simply dropped (deduplicated) instead of violating unique_friendship.
func (fr *friendshipRepositoryGorm) RepointFriendships(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipRepository.RepointFriendships")
	defer span.End()

	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	var friendships []users.Friendship
	if err := db.Where("profile_id1 = ? OR profile_id2 = ?", anonProfileID, anonProfileID).Find(&friendships).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	for _, f := range friendships {
		counterpartyID := f.ProfileID1
		if counterpartyID == anonProfileID {
			counterpartyID = f.ProfileID2
		}

		// The anon profile's counterparty is the very profile being merged in: a
		// friendship with yourself is meaningless (and would violate the profile_order
		// CHECK constraint), so just drop the stale row.
		if counterpartyID == realProfileID {
			if err := db.Delete(&f).Error; err != nil {
				return ungerr.Wrap(err, "error deleting anonymous friendship row")
			}
			continue
		}

		// The new friendship is only ever meaningful between two real profiles: verify
		// the counterparty actually is one instead of assuming it (anon profiles are
		// normally only ever friended by their real creator, but that's an app-level
		// invariant, not a DB constraint).
		var counterpartyProfile users.UserProfile
		found, err := findOptional(db.Select("id", "user_id").Where("id = ?", counterpartyID), &counterpartyProfile)
		if err != nil {
			return err
		}
		if !found || !counterpartyProfile.IsReal() {
			return ungerr.ConflictError("cannot merge: anonymous profile's friendship counterparty is not a real profile")
		}

		existing, err := fr.FindByProfileIDs(ctx, counterpartyID, realProfileID)
		if err != nil {
			return err
		}

		if err := db.Delete(&f).Error; err != nil {
			return ungerr.Wrap(err, "error deleting anonymous friendship row")
		}

		if !existing.IsZero() {
			continue
		}

		id1, id2 := counterpartyID, realProfileID
		if ezutil.CompareUUID(id2, id1) < 0 {
			id1, id2 = id2, id1
		}

		newFriendship := users.Friendship{ProfileID1: id1, ProfileID2: id2, Type: users.Real}
		if _, err := fr.Insert(ctx, newFriendship); err != nil {
			return err
		}
	}

	return nil
}
