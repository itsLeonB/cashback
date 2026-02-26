package monetization

import (
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
)

func PlanToResponse(p entity.Plan) dto.PlanResponse {
	return dto.PlanResponse{
		BaseDTO:  mapper.BaseToDTO(p.BaseEntity),
		Name:     p.Name,
		IsActive: p.IsActive,
		Priority: p.Priority,
	}
}

func PlanVersionToResponse(pv entity.PlanVersion) dto.PlanVersionResponse {
	return dto.PlanVersionResponse{
		BaseDTO:            mapper.BaseToDTO(pv.BaseEntity),
		PlanID:             pv.PlanID,
		PlanName:           pv.Plan.Name,
		PriceAmount:        pv.PriceAmount,
		PriceCurrency:      pv.PriceCurrency,
		BillingInterval:    string(pv.BillingInterval),
		BillUploadsDaily:   pv.BillUploadsDaily,
		BillUploadsMonthly: pv.BillUploadsMonthly,
		EffectiveFrom:      pv.EffectiveFrom,
		EffectiveTo:        pv.EffectiveTo.Time,
		IsDefault:          pv.IsDefault,
	}
}
