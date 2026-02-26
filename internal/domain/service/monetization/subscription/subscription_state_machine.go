package subscription

import (
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/ungerr"
)

func TransitionStatus(sub entity.Subscription, newStatus entity.SubscriptionStatus) (entity.Subscription, error) {
	transitioner, err := getTransitioner(sub.Status)
	if err != nil {
		return entity.Subscription{}, err
	}

	payment := latestPayment(sub.Payments)

	if err = transitioner.Transition(payment, newStatus); err != nil {
		return entity.Subscription{}, err
	}

	if newStatus == entity.SubscriptionActive {
		sub.CurrentPeriodStart = payment.StartsAt
		sub.CurrentPeriodEnd = payment.EndsAt
	}

	sub.Status = newStatus

	return sub, nil
}

func getTransitioner(oldStatus entity.SubscriptionStatus) (transitioner, error) {
	switch oldStatus {
	case entity.SubscriptionIncompletePayment:
		return fromIncomplete{}, nil
	case entity.SubscriptionActive:
		return fromActive{}, nil
	case entity.SubscriptionPastDuePayment:
		return fromPastDue{}, nil
	default:
		return nil, ungerr.Unknownf("unhandled transitioner from status: %s", oldStatus)
	}
}

type transitioner interface {
	Transition(payment entity.Payment, target entity.SubscriptionStatus) error
}

type fromIncomplete struct{}

func (fromIncomplete) Transition(payment entity.Payment, target entity.SubscriptionStatus) error {
	switch target {
	case entity.SubscriptionActive, entity.SubscriptionCanceled:
		return isValidPayment(payment)
	default:
		return ungerr.Unknownf("illegal state transition from incomplete to %s", target)
	}
}

type fromActive struct{}

func (fromActive) Transition(payment entity.Payment, target entity.SubscriptionStatus) error {
	switch target {
	case entity.SubscriptionActive:
		return isValidPayment(payment)
	case entity.SubscriptionPastDuePayment:
		return isInvalidPayment(payment)
	case entity.SubscriptionCanceled:
		return isValidPayment(payment)
	default:
		return ungerr.Unknownf("illegal state transition from active to %s", target)
	}
}

type fromPastDue struct{}

func (fromPastDue) Transition(payment entity.Payment, target entity.SubscriptionStatus) error {
	switch target {
	case entity.SubscriptionActive:
		return isValidPayment(payment)
	case entity.SubscriptionCanceled:
		return isInvalidPayment(payment)
	default:
		return ungerr.Unknownf("illegal state transition from past due to %s", target)
	}
}

func isValidPayment(payment entity.Payment) error {
	if payment.Status == entity.PaidPayment {
		return nil
	}

	return ungerr.ForbiddenError("no valid payments found")
}

func isInvalidPayment(payment entity.Payment) error {
	if payment.Status == entity.PaidPayment {
		return ungerr.ForbiddenError("payment already paid")
	}
	return nil
}

func latestPayment(payments []entity.Payment) entity.Payment {
	if len(payments) == 0 {
		return entity.Payment{}
	}

	latest := payments[0]
	for i := 1; i < len(payments); i++ {
		if payments[i].CreatedAt.After(latest.CreatedAt) {
			latest = payments[i]
		}
	}

	return latest
}
