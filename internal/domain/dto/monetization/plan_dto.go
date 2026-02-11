package monetization

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
)

type NewPlanRequest struct {
	Name string `json:"name" binding:"required,min=3"`
}

type PlanResponse struct {
	dto.BaseDTO
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
}

type UpdatePlanRequest struct {
	ID       uuid.UUID `json:"-"`
	Name     string    `json:"name" binding:"required,min=3"`
	IsActive bool      `json:"isActive"`
}
