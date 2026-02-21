package mapper

import (
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/monetization"
)

func SubscriptionToResponse(s monetization.Subscription) dto.SubscriptionResponse {
	return dto.SubscriptionResponse{
		BaseDTO:    BaseToDTO(s.BaseEntity),
		ProfileID:  s.ProfileID,
		EndsAt:     s.EndsAt.Time,
		CanceledAt: s.CanceledAt.Time,
	}
}
