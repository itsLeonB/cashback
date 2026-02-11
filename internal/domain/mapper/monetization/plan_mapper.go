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
	}
}
