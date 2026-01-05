package llm

import (
	"context"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/ungerr"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

type LLMService interface {
	Prompt(ctx context.Context, systemMsg, userMsg string) (string, error)
}

type openAILLMService struct {
	client openai.Client
	model  string
}

func NewLLMService() LLMService {
	client := openai.NewClient(option.WithAPIKey(config.Global.LLM.ApiKey), option.WithBaseURL(config.Global.LLM.BaseUrl))
	return &openAILLMService{client, config.Global.LLM.Model}
}

func (llm *openAILLMService) Prompt(ctx context.Context, systemMsg, userMsg string) (string, error) {
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
