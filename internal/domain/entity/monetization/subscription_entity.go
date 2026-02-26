package monetization

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/go-crud"
)

type SubscriptionStatus string

const (
	SubscriptionIncompletePayment SubscriptionStatus = "incomplete_payment"
	SubscriptionActive            SubscriptionStatus = "active"
	SubscriptionPastDuePayment    SubscriptionStatus = "past_due_payment"
	SubscriptionCanceled          SubscriptionStatus = "canceled"
)

type Subscription struct {
	crud.BaseEntity
	ProfileID          uuid.UUID
	PlanVersionID      uuid.UUID
	EndsAt             sql.NullTime
	CanceledAt         sql.NullTime
	AutoRenew          bool
	Status             SubscriptionStatus
	CurrentPeriodStart sql.NullTime
	CurrentPeriodEnd   sql.NullTime

	// Relationships
	Profile     users.UserProfile
	PlanVersion PlanVersion
	Payments    []Payment
}

func (s *Subscription) IsActive(t time.Time) bool {
	return s.PlanVersion.IsDefault || ((s.CurrentPeriodEnd.Valid && s.CurrentPeriodEnd.Time.After(t)) &&
		(s.CurrentPeriodStart.Valid && !s.CurrentPeriodStart.Time.After(t)) &&
		(s.Status == SubscriptionActive || s.Status == SubscriptionPastDuePayment))
}

func (s *Subscription) IsSubscribed(t time.Time) bool {
	return s.PlanVersion.IsDefault || ((!s.CanceledAt.Valid || s.CanceledAt.Time.After(t)) && s.Status != SubscriptionCanceled)
}

func (s *Subscription) ContinuedPeriods() (time.Time, time.Time) {
	startsAt := time.Now()
	if s.Status == SubscriptionActive && s.CurrentPeriodEnd.Valid && s.CurrentPeriodEnd.Time.After(startsAt) {
		startsAt = s.CurrentPeriodEnd.Time
	}
	endsAt := startsAt
	switch s.PlanVersion.BillingInterval {
	case MonthlyInterval:
		endsAt = endsAt.AddDate(0, 1, 0)
	case YearlyInterval:
		endsAt = endsAt.AddDate(1, 0, 0)
	}

	return startsAt, endsAt
}
