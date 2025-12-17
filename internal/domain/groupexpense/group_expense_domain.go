package groupexpense

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/domain"
	"github.com/itsLeonB/orcashtrator/internal/domain/expenseitem"
	"github.com/itsLeonB/orcashtrator/internal/domain/otherfee"
	"github.com/shopspring/decimal"
)

type GroupExpense struct {
	domain.AuditMetadata
	CreatorProfileID uuid.UUID
	PayerProfileID   uuid.UUID
	TotalAmount      decimal.Decimal
	// Deprecated: use ItemsTotal instead
	Subtotal    decimal.Decimal
	ItemsTotal  decimal.Decimal
	FeesTotal   decimal.Decimal
	Description string
	// Deprecated: refer to Status instead
	IsConfirmed bool
	// Deprecated: refer to Status instead
	IsParticipantsConfirmed bool
	Status                  appconstant.ExpenseStatus

	// Relationships
	Items        []expenseitem.ExpenseItem
	OtherFees    []otherfee.OtherFee
	Participants []ExpenseParticipant
}

func (ge GroupExpense) ProfileIDs() []uuid.UUID {
	profileIDs := make([]uuid.UUID, 0)
	profileIDs = append(profileIDs, ge.CreatorProfileID)
	profileIDs = append(profileIDs, ge.PayerProfileID)
	for _, item := range ge.Items {
		profileIDs = append(profileIDs, item.ProfileIDs()...)
	}
	for _, fee := range ge.OtherFees {
		profileIDs = append(profileIDs, fee.ProfileIDs()...)
	}
	for _, participant := range ge.Participants {
		profileIDs = append(profileIDs, participant.ProfileID)
	}

	return profileIDs
}

type ExpenseParticipant struct {
	ProfileID   uuid.UUID
	ShareAmount decimal.Decimal
}
