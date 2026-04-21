package llm

import (
	"context"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/ungerr"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

type ChatMessage struct {
	Role    string
	Content string
}

type LLMService interface {
	Prompt(ctx context.Context, systemMsg, userMsg string) (string, error)
	Chat(ctx context.Context, msgs []ChatMessage) (string, error)
}

type openAILLMService struct {
	client openai.Client
	model  string
}

func NewLLMService(cfg config.LLM) LLMService {
	client := openai.NewClient(option.WithAPIKey(cfg.ApiKey), option.WithBaseURL(cfg.BaseUrl))
	return &openAILLMService{client, cfg.Model}
}

func (llm *openAILLMService) Prompt(ctx context.Context, systemMsg, userMsg string) (string, error) {
	ctx, span := otel.Tracer.Start(ctx, "openAILLMService.Prompt")
	defer span.End()

	if userMsg == "" {
		return "", ungerr.Unknown("empty user message")
	}

	msgs := make([]openai.ChatCompletionMessageParamUnion, 0, 2)

	if systemMsg != "" {
		msgs = append(msgs, openai.SystemMessage(systemMsg))
	}

	msgs = append(msgs, openai.UserMessage(userMsg))

	params := openai.ChatCompletionNewParams{
		Model:    llm.model,
		Messages: msgs,
	}

	response, err := llm.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", ungerr.Wrap(err, "error getting LLM response")
	}

	if len(response.Choices) < 1 {
		return "", ungerr.Unknown("no response from LLM")
	}

	return response.Choices[0].Message.Content, nil
}

func (llm *openAILLMService) Chat(ctx context.Context, msgs []ChatMessage) (string, error) {
	ctx, span := otel.Tracer.Start(ctx, "openAILLMService.Chat")
	defer span.End()

	if len(msgs) == 0 {
		return "", ungerr.Unknown("empty messages")
	}

	openaiMsgs := make([]openai.ChatCompletionMessageParamUnion, 0, len(msgs))
	for _, msg := range msgs {
		switch msg.Role {
		case "system":
			openaiMsgs = append(openaiMsgs, openai.SystemMessage(msg.Content))
		case "user":
			openaiMsgs = append(openaiMsgs, openai.UserMessage(msg.Content))
		case "assistant":
			openaiMsgs = append(openaiMsgs, openai.AssistantMessage(msg.Content))
		default:
			return "", ungerr.Unknownf("unhandled message role: %s", msg.Role)
		}
	}

	params := openai.ChatCompletionNewParams{
		Model:    llm.model,
		Messages: openaiMsgs,
	}

	response, err := llm.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", ungerr.Wrap(err, "error getting LLM response")
	}

	if len(response.Choices) < 1 {
		return "", ungerr.Unknown("no response from LLM")
	}

	return response.Choices[0].Message.Content, nil
}
