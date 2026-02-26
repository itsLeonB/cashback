package monetization

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/shopspring/decimal"
)

type NewPlanRequest struct {
	Name     string `json:"name" binding:"required,min=3"`
	Priority int    `json:"priority" binding:"required"`
}

type PlanResponse struct {
	dto.BaseDTO
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
	Priority int    `json:"priority"`
}

type UpdatePlanRequest struct {
	ID       uuid.UUID `json:"-"`
	Name     string    `json:"name" binding:"required,min=3"`
	IsActive bool      `json:"isActive"`
	Priority int       `json:"priority" binding:"required"`
}

type NewPlanVersionRequest struct {
	PlanID             uuid.UUID       `json:"planId" binding:"required"`
	PriceAmount        decimal.Decimal `json:"priceAmount" binding:"required"`
	PriceCurrency      string          `json:"priceCurrency" binding:"required,len=3"`
	BillingInterval    string          `json:"billingInterval" binding:"required,oneof=monthly yearly"`
	BillUploadsDaily   uint            `json:"billUploadsDaily"`
	BillUploadsMonthly uint            `json:"billUploadsMonthly"`
	EffectiveFrom      time.Time       `json:"effectiveFrom" binding:"required"`
	EffectiveTo        time.Time       `json:"effectiveTo"`
	IsDefault          bool            `json:"isDefault"`
}

type PlanVersionResponse struct {
	dto.BaseDTO
	PlanID             uuid.UUID       `json:"planId"`
	PlanName           string          `json:"planName"`
	PriceAmount        decimal.Decimal `json:"priceAmount"`
	PriceCurrency      string          `json:"priceCurrency"`
	BillingInterval    string          `json:"billingInterval"`
	BillUploadsDaily   uint            `json:"billUploadsDaily"`
	BillUploadsMonthly uint            `json:"billUploadsMonthly"`
	EffectiveFrom      time.Time       `json:"effectiveFrom"`
	EffectiveTo        time.Time       `json:"effectiveTo,omitzero"`
	IsDefault          bool            `json:"isDefault"`
}

type UpdatePlanVersionRequest struct {
	ID                 uuid.UUID       `json:"-"`
	PlanID             uuid.UUID       `json:"planId" binding:"required"`
	PriceAmount        decimal.Decimal `json:"priceAmount" binding:"required"`
	PriceCurrency      string          `json:"priceCurrency" binding:"required,len=3"`
	BillingInterval    string          `json:"billingInterval" binding:"required,oneof=monthly yearly"`
	BillUploadsDaily   uint            `json:"billUploadsDaily"`
	BillUploadsMonthly uint            `json:"billUploadsMonthly"`
	EffectiveFrom      time.Time       `json:"effectiveFrom" binding:"required"`
	EffectiveTo        time.Time       `json:"effectiveTo"`
	IsDefault          bool            `json:"isDefault"`
}
