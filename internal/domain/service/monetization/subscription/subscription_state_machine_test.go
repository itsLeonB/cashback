package subscription_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization/subscription"
	"github.com/itsLeonB/go-crud"
	"github.com/stretchr/testify/assert"
)

func TestTransitionStatus(t *testing.T) {
	t.Run("from PastDue to Active", func(t *testing.T) {
		t.Run("should fail if historical payment is Paid but latest is Pending (Reproduction)", func(t *testing.T) {
			sub := entity.Subscription{
				Status: entity.SubscriptionPastDuePayment,
				Payments: []entity.Payment{
					{
						BaseEntity: crud.BaseEntity{
							ID:        uuid.New(),
							CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
						},
						Status: entity.PaidPayment,
					},
					{
						BaseEntity: crud.BaseEntity{
							ID:        uuid.New(),
							CreatedAt: time.Now(),
						},
						Status: entity.PendingPayment,
					},
				},
				PlanVersion: entity.PlanVersion{
					BillingInterval: entity.MonthlyInterval,
				},
			}

			_, err := subscription.TransitionStatus(sub, entity.SubscriptionActive)
			// This is expected to FAIL currently because anyValidPayments sees the first Paid payment
			assert.Error(t, err, "Expected error because latest payment is not Paid")
		})

		t.Run("should succeed if latest payment is Paid", func(t *testing.T) {
			sub := entity.Subscription{
				Status: entity.SubscriptionPastDuePayment,
				Payments: []entity.Payment{
					{
						BaseEntity: crud.BaseEntity{
							ID:        uuid.New(),
							CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
						},
						Status: entity.PaidPayment,
					},
					{
						BaseEntity: crud.BaseEntity{
							ID:        uuid.New(),
							CreatedAt: time.Now(),
						},
						Status: entity.PaidPayment,
					},
				},
				PlanVersion: entity.PlanVersion{
					BillingInterval: entity.MonthlyInterval,
				},
			}

			updatedSub, err := subscription.TransitionStatus(sub, entity.SubscriptionActive)
			assert.NoError(t, err)
			assert.Equal(t, entity.SubscriptionActive, updatedSub.Status)
		})
	})

	t.Run("from Incomplete to Active", func(t *testing.T) {
		t.Run("should succeed if latest payment is Paid", func(t *testing.T) {
			sub := entity.Subscription{
				Status: entity.SubscriptionIncompletePayment,
				Payments: []entity.Payment{
					{
						BaseEntity: crud.BaseEntity{
							ID:        uuid.New(),
							CreatedAt: time.Now(),
						},
						Status: entity.PaidPayment,
					},
				},
				PlanVersion: entity.PlanVersion{
					BillingInterval: entity.MonthlyInterval,
				},
			}

			updatedSub, err := subscription.TransitionStatus(sub, entity.SubscriptionActive)
			assert.NoError(t, err)
			assert.Equal(t, entity.SubscriptionActive, updatedSub.Status)
		})

		t.Run("should fail if latest payment is Pending", func(t *testing.T) {
			sub := entity.Subscription{
				Status: entity.SubscriptionIncompletePayment,
				Payments: []entity.Payment{
					{
						BaseEntity: crud.BaseEntity{
							ID:        uuid.New(),
							CreatedAt: time.Now(),
						},
						Status: entity.PendingPayment,
					},
				},
				PlanVersion: entity.PlanVersion{
					BillingInterval: entity.MonthlyInterval,
				},
			}

			_, err := subscription.TransitionStatus(sub, entity.SubscriptionActive)
			assert.Error(t, err)
		})
	})

	t.Run("from Active to PastDue", func(t *testing.T) {
		t.Run("should succeed if latest payment is Error", func(t *testing.T) {
			sub := entity.Subscription{
				Status: entity.SubscriptionActive,
				Payments: []entity.Payment{
					{
						BaseEntity: crud.BaseEntity{
							ID:        uuid.New(),
							CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
						},
						Status: entity.PaidPayment,
					},
					{
						BaseEntity: crud.BaseEntity{
							ID:        uuid.New(),
							CreatedAt: time.Now(),
						},
						Status: entity.ErrorPayment,
					},
				},
				PlanVersion: entity.PlanVersion{
					BillingInterval: entity.MonthlyInterval,
				},
			}

			updatedSub, err := subscription.TransitionStatus(sub, entity.SubscriptionPastDuePayment)
			assert.NoError(t, err)
			assert.Equal(t, entity.SubscriptionPastDuePayment, updatedSub.Status)
		})

		t.Run("should fail if latest payment is Paid", func(t *testing.T) {
			sub := entity.Subscription{
				Status: entity.SubscriptionActive,
				Payments: []entity.Payment{
					{
						BaseEntity: crud.BaseEntity{
							ID:        uuid.New(),
							CreatedAt: time.Now(),
						},
						Status: entity.PaidPayment,
					},
				},
				PlanVersion: entity.PlanVersion{
					BillingInterval: entity.MonthlyInterval,
				},
			}

			_, err := subscription.TransitionStatus(sub, entity.SubscriptionPastDuePayment)
			assert.Error(t, err)
		})
	})

	t.Run("from Active to Active", func(t *testing.T) {
		t.Run("should succeed if latest payment is Paid (Renewal Extension)", func(t *testing.T) {
			sub := entity.Subscription{
				Status: entity.SubscriptionActive,
				Payments: []entity.Payment{
					{
						BaseEntity: crud.BaseEntity{
							ID:        uuid.New(),
							CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
						},
						Status: entity.PaidPayment,
					},
					{
						BaseEntity: crud.BaseEntity{
							ID:        uuid.New(),
							CreatedAt: time.Now(),
						},
						Status: entity.PaidPayment,
					},
				},
				PlanVersion: entity.PlanVersion{
					BillingInterval: entity.MonthlyInterval,
				},
			}

			updatedSub, err := subscription.TransitionStatus(sub, entity.SubscriptionActive)
			assert.NoError(t, err)
			assert.Equal(t, entity.SubscriptionActive, updatedSub.Status)
		})
	})
}
