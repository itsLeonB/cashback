package dto

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/shopspring/decimal"
)

type FeeParticipantResponse struct {
	Profile     SimpleProfile   `json:"profile"`
	ShareAmount decimal.Decimal `json:"shareAmount"`
}

type FeeCalculationMethodInfo struct {
	Name        expenses.FeeCalculationMethod `json:"name"`
	Display     string                        `json:"display"`
	Description string                        `json:"description"`
}

type OtherFeeResponse struct {
	BaseDTO
	Name              string                        `json:"name"`
	Amount            decimal.Decimal               `json:"amount"`
	CalculationMethod expenses.FeeCalculationMethod `json:"calculationMethod"`
	Participants      []FeeParticipantResponse      `json:"participants,omitempty"`
}

type NewOtherFeeRequest struct {
	UserProfileID     uuid.UUID                     `json:"-"`
	GroupExpenseID    uuid.UUID                     `json:"-"`
	Name              string                        `json:"name" binding:"required,min=3"`
	Amount            decimal.Decimal               `json:"amount" binding:"required"`
	CalculationMethod expenses.FeeCalculationMethod `json:"calculationMethod" binding:"required"`
}

type UpdateOtherFeeRequest struct {
	UserProfileID     uuid.UUID                     `json:"-"`
	ID                uuid.UUID                     `json:"-"`
	GroupExpenseID    uuid.UUID                     `json:"-"`
	Name              string                        `json:"name" binding:"required,min=3"`
	Amount            decimal.Decimal               `json:"amount" binding:"required"`
	CalculationMethod expenses.FeeCalculationMethod `json:"calculationMethod" binding:"required"`
}
