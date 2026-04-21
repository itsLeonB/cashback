package langfuse

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/ungerr"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

type PromptType string

const (
	PromptTypeChat PromptType = "chat"
	PromptTypeText PromptType = "text"
)

type Prompt struct {
	Type            PromptType      `json:"type"`
	Name            string          `json:"name"`
	Version         int             `json:"version"`
	Config          any             `json:"config"`
	Labels          []string        `json:"labels"`
	Tags            []string        `json:"tags"`
	CommitMessage   *string         `json:"commitMessage"`
	ResolutionGraph any             `json:"resolutionGraph"`
	RawPrompt       json.RawMessage `json:"prompt"`
}

func (p *Prompt) ChatMessages() ([]ChatMessage, error) {
	if p.Type != PromptTypeChat {
		return nil, ungerr.Unknown("prompt is not a chat prompt")
	}
	var msgs []ChatMessage
	if err := json.Unmarshal(p.RawPrompt, &msgs); err != nil {
		return nil, ungerr.Wrap(err, "failed to unmarshal chat messages")
	}
	return msgs, nil
}

func (p *Prompt) Text() (string, error) {
	if p.Type != PromptTypeText {
		return "", ungerr.Unknown("prompt is not a text prompt")
	}
	var text string
	if err := json.Unmarshal(p.RawPrompt, &text); err != nil {
		return "", ungerr.Wrap(err, "failed to unmarshal text prompt")
	}
	return text, nil
}

func (p *Prompt) Compile(vars map[string]any) ([]ChatMessage, error) {
	if p.Type == PromptTypeChat {
		msgs, err := p.ChatMessages()
		if err != nil {
			return nil, err
		}
		for i := range msgs {
			msgs[i].Content = injectVariables(msgs[i].Content, vars)
		}
		return msgs, nil
	}

	text, err := p.Text()
	if err != nil {
		return nil, err
	}
	return []ChatMessage{
		{
			Role:    "user",
			Content: injectVariables(text, vars),
			Type:    "chatmessage",
		},
	}, nil
}

func injectVariables(template string, vars map[string]any) string {
	result := template
	for k, v := range vars {
		placeholder := fmt.Sprintf("{{%s}}", k)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", v))
	}
	return result
}

type GetPromptOptions struct {
	Version *int
	Label   *string
	Resolve *bool
}

type Client interface {
	GetPrompt(ctx context.Context, name string, opts ...GetPromptOptions) (*Prompt, error)
	Shutdown() error
}

type langfuseClient struct {
	publicKey  string
	secretKey  string
	baseUrl    string
	httpClient *http.Client
}

func NewClient(cfg config.Langfuse) Client {
	return &langfuseClient{
		publicKey: cfg.PublicKey,
		secretKey: cfg.SecretKey,
		baseUrl:   cfg.BaseUrl,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *langfuseClient) GetPrompt(ctx context.Context, name string, opts ...GetPromptOptions) (*Prompt, error) {
	ctx, span := otel.Tracer.Start(ctx, "langfuseClient.GetPrompt")
	defer span.End()

	parsedUrl, err := url.Parse(fmt.Sprintf("%s/api/public/v2/prompts/%s", c.baseUrl, url.PathEscape(name)))
	if err != nil {
		return nil, ungerr.Wrap(err, "failed to parse URL")
	}

	if len(opts) > 0 {
		q := parsedUrl.Query()
		if opts[0].Version != nil {
			q.Set("version", strconv.Itoa(*opts[0].Version))
		}
		if opts[0].Label != nil {
			q.Set("label", *opts[0].Label)
		}
		if opts[0].Resolve != nil {
			q.Set("resolve", strconv.FormatBool(*opts[0].Resolve))
		}
		parsedUrl.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedUrl.String(), nil)
	if err != nil {
		return nil, ungerr.Wrap(err, "failed to create request")
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.publicKey, c.secretKey)))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, ungerr.Wrap(err, "failed to execute request")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			span.RecordError(err)
			logger.Errorf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, ungerr.Unknown(fmt.Sprintf("unexpected status code: %d, body: %s", resp.StatusCode, string(body)))
	}

	var prompt Prompt
	if err := json.NewDecoder(resp.Body).Decode(&prompt); err != nil {
		return nil, ungerr.Wrap(err, "failed to decode response")
	}

	return &prompt, nil
}

func (c *langfuseClient) Shutdown() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
