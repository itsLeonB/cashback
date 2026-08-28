package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type otherFeeRepositoryGorm struct {
	crud.Repository[expenses.OtherFee]
}

func NewOtherFeeRepository(db *gorm.DB) *otherFeeRepositoryGorm {
	return &otherFeeRepositoryGorm{
		crud.NewRepository[expenses.OtherFee](db),
	}
}

func (ger *otherFeeRepositoryGorm) SyncParticipants(ctx context.Context, feeID uuid.UUID, participants []expenses.FeeParticipant) error {
	ctx, span := otel.Tracer.Start(ctx, "OtherFeeRepository.SyncParticipants")
	defer span.End()

	db, err := ger.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	profileIDs := make([]uuid.UUID, len(participants))
	for i, p := range participants {
		participants[i].OtherFeeID = feeID
		profileIDs[i] = p.ProfileID
	}

	if len(participants) > 0 {
		// For PostgreSQL
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "other_fee_id"}, {Name: "profile_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"share_amount"}),
		}).Create(&participants).Error; err != nil {
			return ungerr.Wrap(err, appconstant.ErrDataUpdate)
		}
	}

	query := db.Where("other_fee_id = ?", feeID)
	if len(profileIDs) > 0 {
		query = query.Where("profile_id NOT IN ?", profileIDs)
	}
	if err := query.Delete(&expenses.FeeParticipant{}).Error; err != nil {
		return err
	}

	return nil
}

// RepointParticipants merges group_expense_other_fee_participants rows referencing
// anonProfileID onto realProfileID. If realProfileID already has a row for the same fee,
// share_amount is summed and the anonymous row is dropped instead of violating
// unique_fee_participant.
func (ger *otherFeeRepositoryGorm) RepointParticipants(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "OtherFeeRepository.RepointParticipants")
	defer span.End()

	db, err := ger.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	var anonRows []expenses.FeeParticipant
	if err := db.Where("profile_id = ?", anonProfileID).Find(&anonRows).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataSelect)
	}
	if len(anonRows) == 0 {
		return nil
	}

	merged := make([]expenses.FeeParticipant, len(anonRows))
	for i, row := range anonRows {
		merged[i] = expenses.FeeParticipant{
			OtherFeeID:  row.OtherFeeID,
			ProfileID:   realProfileID,
			ShareAmount: row.ShareAmount,
		}
	}

	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "other_fee_id"}, {Name: "profile_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"share_amount": gorm.Expr("group_expense_other_fee_participants.share_amount + excluded.share_amount"),
		}),
	}).Create(&merged).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	if err := db.Where("profile_id = ?", anonProfileID).Delete(&expenses.FeeParticipant{}).Error; err != nil {
		return ungerr.Wrap(err, "error deleting merged fee participants")
	}

	return nil
}
