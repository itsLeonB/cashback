package langfuse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/stretchr/testify/assert"
)

func TestLangfuseClient_GetPrompt(t *testing.T) {
	ctx := context.Background()
	name := "test-prompt"

	t.Run("success_chat_prompt", func(t *testing.T) {
		expectedPrompt := &Prompt{
			Type:    PromptTypeChat,
			Name:    name,
			Version: 1,
			RawPrompt: json.RawMessage(`[
				{"role": "system", "content": "You are a helpful assistant", "type": "chatmessage"},
				{"role": "user", "content": "Hello", "type": "chatmessage"}
			]`),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/public/v2/prompts/test-prompt", r.URL.Path)
			assert.Contains(t, r.Header.Get("Authorization"), "Basic")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedPrompt)
		}))
		defer server.Close()

		cfg := config.Langfuse{
			PublicKey: "pk-test",
			SecretKey: "sk-test",
			BaseUrl:   server.URL,
		}
		client := NewClient(cfg)

		prompt, err := client.GetPrompt(ctx, name)
		assert.NoError(t, err)
		assert.NotNil(t, prompt)
		assert.Equal(t, PromptTypeChat, prompt.Type)
		assert.Equal(t, name, prompt.Name)

		msgs, err := prompt.ChatMessages()
		assert.NoError(t, err)
		assert.Len(t, msgs, 2)
		assert.Equal(t, "system", msgs[0].Role)
		assert.Equal(t, "user", msgs[1].Role)
	})

	t.Run("success_text_prompt", func(t *testing.T) {
		expectedPrompt := &Prompt{
			Type:      PromptTypeText,
			Name:      name,
			Version:   1,
			RawPrompt: json.RawMessage(`"This is a text prompt"`),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedPrompt)
		}))
		defer server.Close()

		cfg := config.Langfuse{
			PublicKey: "pk-test",
			SecretKey: "sk-test",
			BaseUrl:   server.URL,
		}
		client := NewClient(cfg)

		prompt, err := client.GetPrompt(ctx, name)
		assert.NoError(t, err)
		assert.NotNil(t, prompt)
		assert.Equal(t, PromptTypeText, prompt.Type)

		text, err := prompt.Text()
		assert.NoError(t, err)
		assert.Equal(t, "This is a text prompt", text)
	})

	t.Run("error_not_found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		}))
		defer server.Close()

		cfg := config.Langfuse{
			PublicKey: "pk-test",
			SecretKey: "sk-test",
			BaseUrl:   server.URL,
		}
		client := NewClient(cfg)

		prompt, err := client.GetPrompt(ctx, name)
		assert.Error(t, err)
		assert.Nil(t, prompt)
		assert.Contains(t, err.Error(), "unexpected status code: 404")
	})
}
