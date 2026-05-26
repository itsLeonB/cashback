package util

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateSlug(t *testing.T) {
	id := uuid.MustParse("a3bf9c00-1234-5678-9abc-def012345678")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "John Doe", "john-doe-a3bf9c"},
		{"special chars", "José María!", "jos-mar-a-a3bf9c"},
		{"extra spaces", "  Hello   World  ", "hello-world-a3bf9c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlug(tt.input, id)
			if got != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
