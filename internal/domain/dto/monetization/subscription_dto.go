package monetization

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
)

type NewSubscriptionRequest struct {
	ProfileID     uuid.UUID `json:"profileId" binding:"required"`
	PlanVersionID uuid.UUID `json:"planVersionId" binding:"required"`
	EndsAt        time.Time `json:"endsAt"`
	CanceledAt    time.Time `json:"canceledAt"`
	AutoRenew     bool      `json:"autoRenew"`
}

type PurchaseSubscriptionRequest struct {
	ProfileID     uuid.UUID `json:"-"`
	PlanID        uuid.UUID `json:"-"`
	PlanVersionID uuid.UUID `json:"-"`
}

type SubscriptionResponse struct {
	dto.BaseDTO
	ProfileID          uuid.UUID `json:"profileId"`
	ProfileName        string    `json:"profileName"`
	PlanVersionID      uuid.UUID `json:"planVersionId"`
	PlanName           string    `json:"planName"`
	EndsAt             time.Time `json:"endsAt,omitzero"`
	CanceledAt         time.Time `json:"canceledAt,omitzero"`
	AutoRenew          bool      `json:"autoRenew"`
	BillUploadsDaily   int       `json:"billUploadsDaily"`
	BillUploadsMonthly int       `json:"billUploadsMonthly"`
	Status             string    `json:"status"`
}

type UpdateSubscriptionRequest struct {
	ID            uuid.UUID `json:"-"`
	ProfileID     uuid.UUID `json:"profileId" binding:"required"`
	PlanVersionID uuid.UUID `json:"planVersionId" binding:"required"`
	EndsAt        time.Time `json:"endsAt"`
	CanceledAt    time.Time `json:"canceledAt"`
	AutoRenew     bool      `json:"autoRenew"`
}
