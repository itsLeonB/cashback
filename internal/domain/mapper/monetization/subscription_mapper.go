package monetization

import (
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
)

func SubscriptionToResponse(s entity.Subscription) dto.SubscriptionResponse {
	return dto.SubscriptionResponse{
		BaseDTO:       mapper.BaseToDTO(s.BaseEntity),
		ProfileID:     s.ProfileID,
		ProfileName:   s.Profile.Name,
		PlanVersionID: s.PlanVersionID,
		PlanName:      s.PlanVersion.Plan.Name,
		EndsAt:        s.EndsAt.Time,
		CanceledAt:    s.CanceledAt.Time,
		AutoRenew:     s.AutoRenew,
	}
}
