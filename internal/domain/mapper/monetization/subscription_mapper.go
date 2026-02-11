package monetization

import (
	"time"

	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
)

func SubscriptionToResponse(s entity.Subscription) dto.SubscriptionResponse {
	var endsAt *time.Time
	if s.EndsAt.Valid {
		endsAt = &s.EndsAt.Time
	}

	var canceledAt *time.Time
	if s.CanceledAt.Valid {
		canceledAt = &s.CanceledAt.Time
	}

	return dto.SubscriptionResponse{
		BaseDTO:       mapper.BaseToDTO(s.BaseEntity),
		ProfileID:     s.ProfileID,
		ProfileName:   s.Profile.Name,
		PlanVersionID: s.PlanVersionID,
		PlanName:      s.PlanVersion.Plan.Name,
		EndsAt:        endsAt,
		CanceledAt:    canceledAt,
		AutoRenew:     s.AutoRenew,
	}
}
