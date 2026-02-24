package subscription

import (
	"database/sql"
	"time"

	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/ungerr"
)

func TransitionStatus(sub entity.Subscription, newStatus entity.SubscriptionStatus) (entity.Subscription, error) {
	transitioner, err := getTransitioner(sub.Status)
	if err != nil {
		return entity.Subscription{}, err
	}

	if err = transitioner.Transition(sub.Payments, newStatus); err != nil {
		return entity.Subscription{}, err
	}

	if newStatus == entity.SubscriptionActive {
		sub.EndsAt = sql.NullTime{
			Time:  time.Now().AddDate(0, 1, 0),
			Valid: true,
		}
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
	Transition(payments []entity.Payment, target entity.SubscriptionStatus) error
}

type fromIncomplete struct{}

func (fromIncomplete) Transition(payments []entity.Payment, target entity.SubscriptionStatus) error {
	switch target {
	case entity.SubscriptionActive, entity.SubscriptionCanceled:
		return anyValidPayments(payments)
	default:
		return ungerr.Unknownf("illegal state transition from incomplete to %s", target)
	}
}

type fromActive struct{}

func (fromActive) Transition(payments []entity.Payment, target entity.SubscriptionStatus) error {
	switch target {
	case entity.SubscriptionPastDuePayment:
		return noValidPayments(payments)
	case entity.SubscriptionCanceled:
		return anyValidPayments(payments)
	default:
		return ungerr.Unknownf("illegal state transition from active to %s", target)
	}
}

type fromPastDue struct{}

func (fromPastDue) Transition(payments []entity.Payment, target entity.SubscriptionStatus) error {
	switch target {
	case entity.SubscriptionActive:
		return anyValidPayments(payments)
	case entity.SubscriptionCanceled:
		return noValidPayments(payments)
	default:
		return ungerr.Unknownf("illegal state transition from past due to %s", target)
	}
}

func anyValidPayments(payments []entity.Payment) error {
	for _, payment := range payments {
		if payment.Status == entity.PaidPayment {
			return nil
		}
	}
	return ungerr.ForbiddenError("no valid payments found")
}

func noValidPayments(payments []entity.Payment) error {
	for _, payment := range payments {
		if payment.Status == entity.PaidPayment {
			return ungerr.ForbiddenError("payment already paid")
		}
	}
	return nil
}
