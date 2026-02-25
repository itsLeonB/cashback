package monetization

import (
	"math"
	"time"

	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
)

func SimpleSubscriptionMapper() func(entity.Subscription) dto.SubscriptionResponse {
	return func(s entity.Subscription) dto.SubscriptionResponse {
		return SubscriptionToResponse(s, time.Now())
	}
}

func SubscriptionToResponse(s entity.Subscription, t time.Time) dto.SubscriptionResponse {
	dueDays := -1
	if s.CurrentPeriodEnd.Valid && !s.CurrentPeriodEnd.Time.Before(t) {
		dueHours := s.CurrentPeriodEnd.Time.Sub(t).Hours()
		dueDays = int(math.Ceil(dueHours / 24))
	}
	return dto.SubscriptionResponse{
		BaseDTO:            mapper.BaseToDTO(s.BaseEntity),
		ProfileID:          s.ProfileID,
		ProfileName:        s.Profile.Name,
		PlanVersionID:      s.PlanVersionID,
		PlanName:           s.PlanVersion.Plan.Name,
		EndsAt:             s.EndsAt.Time,
		CanceledAt:         s.CanceledAt.Time,
		AutoRenew:          s.AutoRenew,
		BillUploadsDaily:   int(s.PlanVersion.BillUploadsDaily),
		BillUploadsMonthly: int(s.PlanVersion.BillUploadsMonthly),
		Status:             string(s.Status),
		PaymentDueDays:     dueDays,
		CurrentPeriodStart: s.CurrentPeriodStart.Time,
		CurrentPeriodEnd:   s.CurrentPeriodEnd.Time,
	}
}
