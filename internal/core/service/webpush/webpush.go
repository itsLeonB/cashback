package webpush

import (
	"github.com/SherClockHolmes/webpush-go"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/ungerr"
)

type Client interface {
	Send(subscription Subscription) error
}

type webpushClient struct {
	opts *webpush.Options
}

func NewWebPush(cfg config.Push) *webpushClient {
	return &webpushClient{
		&webpush.Options{
			VAPIDPublicKey:  cfg.VapidPublicKey,
			VAPIDPrivateKey: cfg.VapidPrivateKey,
			Subscriber:      cfg.VapidSubject,
		},
	}
}

type Subscription struct {
	Endpoint string
	Keys     Keys
	Payload  []byte
}

type Keys struct {
	P256dh string `json:"p256dh"`
	Auth   string `json:"auth"`
}

func (wc *webpushClient) Send(subscription Subscription) error {
	// Create webpush subscription
	webpushSub := &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.Keys.P256dh,
			Auth:   subscription.Keys.Auth,
		},
	}

	resp, err := webpush.SendNotification(subscription.Payload, webpushSub, wc.opts)
	if err != nil {
		return err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			logger.Errorf("error closing response body: %v", e)
		}
	}()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ungerr.Unknownf("push service returned status %d", resp.StatusCode)
	}

	return nil
}

func GenerateKeys() (string, string, error) {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		return "", "", ungerr.Wrap(err, "error generating VAPID keys")
	}
	return privateKey, publicKey, nil
}
