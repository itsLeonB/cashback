package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/ungerr"
)

type ipaymuGateway struct {
	apiKey       string
	va           string
	baseURL      string
	returnURL    string
	notifyURL    string
	cancelURL    string
	notifySecret string
	client       *http.Client
}

func newIPaymuGateway(cfg config.Payment) (*ipaymuGateway, error) {
	notifyURL := cfg.NotifyUrl
	if cfg.NotifySecret != "" {
		sep := "?"
		if strings.Contains(notifyURL, "?") {
			sep = "&"
		}
		notifyURL += sep + "secret=" + cfg.NotifySecret
	}

	return &ipaymuGateway{
		apiKey:       cfg.ServerKey,
		va:           cfg.Va,
		baseURL:      cfg.BaseUrl,
		returnURL:    cfg.ReturnUrl,
		notifyURL:    notifyURL,
		cancelURL:    cfg.CancelUrl,
		notifySecret: cfg.NotifySecret,
		client:       &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (g *ipaymuGateway) Provider() string {
	return "ipaymu"
}

func (g *ipaymuGateway) CreateTransaction(ctx context.Context, payment entity.Payment) (entity.Payment, error) {
	ctx, span := otel.Tracer.Start(ctx, "ipaymuGateway.CreateTransaction")
	defer span.End()

	body := map[string]any{
		"product":     []string{"Cashus Subscription"},
		"qty":         []int{1},
		"price":       []int64{payment.Amount.IntPart()},
		"amount":      payment.Amount.IntPart(),
		"returnUrl":   g.returnURL,
		"notifyUrl":   g.notifyURL,
		"cancelUrl":   g.cancelURL,
		"referenceId": payment.ID.String(),
		"expired":     24,
		"comments":    "Cashus subscription",
	}

	respBody, err := g.doRequest(ctx, "/api/v2/payment", body)
	if err != nil {
		return entity.Payment{}, ungerr.Wrap(err, "error creating ipaymu transaction")
	}

	var result struct {
		Status  int    `json:"Status"`
		Message string `json:"Message"`
		Data    struct {
			SessionID     string `json:"SessionId"`
			TransactionID int    `json:"TransactionId"`
			URL           string `json:"Url"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return entity.Payment{}, ungerr.Wrap(err, "error parsing ipaymu response")
	}
	if result.Status != 200 {
		return entity.Payment{}, ungerr.Unknownf("ipaymu error: %s", result.Message)
	}

	// ponytail: GatewayTransactionID stores the redirect URL for iPaymu (frontend reads this to redirect user).
	// GatewayEventID stores the numeric transaction ID for status checks.
	payment.GatewayTransactionID = sql.NullString{String: result.Data.URL, Valid: true}
	payment.GatewayEventID = sql.NullString{String: strconv.Itoa(result.Data.TransactionID), Valid: true}

	return payment, nil
}

func (g *ipaymuGateway) ValidateAndCheckStatus(ctx context.Context, payload NotificationPayload) (entity.PaymentStatus, error) {
	ctx, span := otel.Tracer.Start(ctx, "ipaymuGateway.ValidateAndCheckStatus")
	defer span.End()

	// Validate shared secret from notify URL query param
	if g.notifySecret != "" {
		if subtle.ConstantTimeCompare([]byte(payload.Signature), []byte(g.notifySecret)) != 1 {
			return entity.ErrorPayment, ungerr.Unknown("invalid notification secret")
		}
	}

	// Server-side confirmation: call iPaymu transaction API to verify actual status
	trxID := payload.Extra["trx_id"]
	if trxID == "" {
		return entity.ErrorPayment, ungerr.Unknown("missing trx_id in ipaymu notification")
	}

	body := map[string]any{"transactionId": trxID}
	respBody, err := g.doRequest(ctx, "/api/v2/transaction", body)
	if err != nil {
		return entity.ErrorPayment, ungerr.Wrap(err, "error confirming ipaymu transaction status")
	}

	var result struct {
		Status int `json:"Status"`
		Data   struct {
			StatusCode int    `json:"StatusCode"`
			Status     string `json:"Status"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return entity.ErrorPayment, ungerr.Wrap(err, "error parsing ipaymu status response")
	}

	return g.mapStatus(result.Data.Status)
}

func (g *ipaymuGateway) mapStatus(status string) (entity.PaymentStatus, error) {
	switch strings.ToLower(status) {
	case "berhasil", "successful":
		return entity.PaidPayment, nil
	case "pending":
		return entity.PendingPayment, nil
	case "expired":
		return entity.ExpiredPayment, nil
	case "gagal", "failed":
		return entity.ErrorPayment, fmt.Errorf("payment failed")
	default:
		return entity.ErrorPayment, ungerr.Unknownf("unhandled ipaymu status: %s", status)
	}
}

func (g *ipaymuGateway) doRequest(ctx context.Context, path string, body map[string]any) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Signature: HMAC-SHA256
	bodyHash := sha256.Sum256(jsonBody)
	bodyHashHex := hex.EncodeToString(bodyHash[:])
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	stringToSign := "POST:" + g.va + ":" + bodyHashHex + ":" + g.apiKey
	mac := hmac.New(sha256.New, []byte(g.apiKey))
	mac.Write([]byte(stringToSign))
	signature := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("va", g.va)
	req.Header.Set("signature", signature)
	req.Header.Set("timestamp", ts)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Errorf("error closing response body: %v", err)
		}
	}()

	return io.ReadAll(resp.Body)
}
