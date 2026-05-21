package payment

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
	"github.com/stripe/stripe-go/v85"
	billingportalsession "github.com/stripe/stripe-go/v85/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v85/checkout/session"
	"github.com/stripe/stripe-go/v85/customer"
	"github.com/stripe/stripe-go/v85/webhook"
)

type stripeGateway struct {
	webhookSecret string
}

func NewGateway(cfg config.Payment) Gateway {
	stripe.Key = cfg.ServerKey
	return &stripeGateway{webhookSecret: cfg.WebhookSecret}
}

func (sg *stripeGateway) CreateCheckoutSession(params CheckoutParams) (CheckoutResult, error) {
	var customerID string
	if params.CustomerID != "" {
		customerID = params.CustomerID
	} else {
		cp := &stripe.CustomerParams{
			Email: stripe.String(params.ProfileEmail),
			Metadata: map[string]string{
				"profile_id": params.ProfileID.String(),
			},
		}
		c, err := customer.New(cp)
		if err != nil {
			return CheckoutResult{}, ungerr.Wrap(err, "error creating stripe customer")
		}
		customerID = c.ID
	}

	metadata := map[string]string{
		"subscription_id": params.Payment.SubscriptionID.String(),
		"profile_id":      params.ProfileID.String(),
	}

	sp := &stripe.CheckoutSessionParams{
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		Customer: stripe.String(customerID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(params.PlanVersion.StripePriceID.String),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(params.SuccessURL),
		CancelURL:  stripe.String(params.CancelURL),
		Metadata:   metadata,
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: metadata,
		},
	}

	sess, err := checkoutsession.New(sp)
	if err != nil {
		return CheckoutResult{}, ungerr.Wrap(err, "error creating stripe checkout session")
	}

	return CheckoutResult{
		CheckoutURL:       sess.URL,
		GatewayCustomerID: customerID,
		GatewaySessionID:  sess.ID,
	}, nil
}

func (sg *stripeGateway) HandleWebhook(payload []byte, signature string) (*WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, sg.webhookSecret)
	if err != nil {
		return nil, ungerr.Wrap(err, "invalid stripe webhook signature")
	}

	switch event.Type {
	case "checkout.session.completed":
		return sg.handleCheckoutSessionCompleted(event)
	case "invoice.paid":
		return sg.handleInvoicePaid(event)
	case "invoice.payment_failed":
		return sg.handleInvoicePaymentFailed(event)
	case "customer.subscription.deleted":
		return sg.handleCustomerSubscriptionCanceled(event)
	default:
		return nil, ungerr.Unknownf("Stripe webhook event: %s is unhandled", event.Type)
	}
}

func (sg *stripeGateway) CreatePortalSession(customerID string, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	sess, err := billingportalsession.New(params)
	if err != nil {
		return "", ungerr.Wrap(err, "error creating portal session")
	}

	return sess.URL, nil
}

func (sg *stripeGateway) handleCustomerSubscriptionCanceled(event stripe.Event) (*WebhookEvent, error) {
	sub, err := ezutil.Unmarshal[stripe.Subscription](event.Data.Raw)
	if err != nil {
		return nil, ungerr.Wrap(err, "error parsing subscription")
	}

	subID, err := ezutil.Parse[uuid.UUID](sub.Metadata["subscription_id"])
	if err != nil {
		return nil, err
	}

	return &WebhookEvent{
		Type:           "subscription_canceled",
		GatewayEventID: event.ID,
		SubscriptionID: subID,
		GatewaySubID:   sub.ID,
	}, nil
}

func (sg *stripeGateway) handleInvoicePaymentFailed(event stripe.Event) (*WebhookEvent, error) {
	inv, err := ezutil.Unmarshal[stripe.Invoice](event.Data.Raw)
	if err != nil {
		return nil, ungerr.Wrap(err, "error parsing invoice")
	}

	subID, gatewaySubID, err := sg.parseInvoiceIDs(inv)
	if err != nil {
		return nil, err
	}

	return &WebhookEvent{
		Type:           "payment_failed",
		GatewayEventID: event.ID,
		SubscriptionID: subID,
		GatewaySubID:   gatewaySubID,
	}, nil
}

func (sg *stripeGateway) handleInvoicePaid(event stripe.Event) (*WebhookEvent, error) {
	inv, err := ezutil.Unmarshal[stripe.Invoice](event.Data.Raw)
	if err != nil {
		return nil, ungerr.Wrap(err, "error parsing invoice")
	}

	subID, gatewaySubID, err := sg.parseInvoiceIDs(inv)
	if err != nil {
		return nil, err
	}

	periodStart, periodEnd := sg.parseInvoicePeriod(inv)

	return &WebhookEvent{
		Type:           "payment_success",
		GatewayEventID: event.ID,
		SubscriptionID: subID,
		GatewaySubID:   gatewaySubID,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
	}, nil
}

func (sg *stripeGateway) parseInvoicePeriod(inv stripe.Invoice) (time.Time, time.Time) {
	if inv.Lines != nil && len(inv.Lines.Data) > 0 {
		line := inv.Lines.Data[0]
		if line.Period != nil && line.Period.Start > 0 && line.Period.End > 0 {
			return time.Unix(line.Period.Start, 0), time.Unix(line.Period.End, 0)
		}
	}
	// Fallback to invoice-level period
	return time.Unix(inv.PeriodStart, 0), time.Unix(inv.PeriodEnd, 0)
}

func (sg *stripeGateway) handleCheckoutSessionCompleted(event stripe.Event) (*WebhookEvent, error) {
	sess, err := ezutil.Unmarshal[stripe.CheckoutSession](event.Data.Raw)
	if err != nil {
		return nil, ungerr.Wrap(err, "error parsing checkout session")
	}

	subID, err := ezutil.Parse[uuid.UUID](sess.Metadata["subscription_id"])
	if err != nil {
		return nil, err
	}

	if sess.Subscription == nil {
		return nil, ungerr.Unknown("checkout session missing subscription data")
	}

	return &WebhookEvent{
		Type:           "payment_success",
		GatewayEventID: event.ID,
		SubscriptionID: subID,
		GatewaySubID:   sess.Subscription.ID,
	}, nil
}

func (sg *stripeGateway) parseInvoiceIDs(inv stripe.Invoice) (uuid.UUID, string, error) {
	var subID uuid.UUID
	var gatewaySubID string
	var err error

	if inv.Parent != nil && inv.Parent.SubscriptionDetails != nil {
		sd := inv.Parent.SubscriptionDetails

		// Try expanded subscription metadata first (only if metadata is populated)
		if sd.Subscription != nil && sd.Subscription.ID != "" {
			gatewaySubID = sd.Subscription.ID
			if len(sd.Subscription.Metadata) > 0 {
				if raw := sd.Subscription.Metadata["subscription_id"]; raw != "" {
					subID, err = ezutil.Parse[uuid.UUID](raw)
					if err != nil {
						return uuid.Nil, "", err
					}
				}
			}
		}

		// Fallback to SubscriptionDetails.Metadata if subID still not resolved
		if subID == uuid.Nil && len(sd.Metadata) > 0 {
			if raw := sd.Metadata["subscription_id"]; raw != "" {
				subID, err = ezutil.Parse[uuid.UUID](raw)
				if err != nil {
					return uuid.Nil, "", err
				}
			}
		}
	}

	if subID == uuid.Nil {
		return uuid.Nil, "", ungerr.Unknown("unable to extract subscription_id from invoice")
	}

	return subID, gatewaySubID, nil
}
