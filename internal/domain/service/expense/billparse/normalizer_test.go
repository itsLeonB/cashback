package billparse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	const expectedDecimal = "15500.50"
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "thousands with dots",
			input:    "500.000",
			expected: "500000",
		},
		{
			name:     "thousands with commas",
			input:    "500,000",
			expected: "500000",
		},
		{
			name:     "multiple dots",
			input:    "1.500.000",
			expected: "1500000",
		},
		{
			name:     "mixed separators ID format",
			input:    "15.500,50",
			expected: expectedDecimal,
		},
		{
			name:     "with Rp prefix and dots",
			input:    "Rp 15.500,50",
			expected: expectedDecimal,
		},
		{
			name:     "OCR error O instead of 0",
			input:    "5OO.OOO",
			expected: "500000",
		},
		{
			name:     "OCR error with decimal",
			input:    "15.5OO,5O",
			expected: expectedDecimal,
		},
		{
			name:     "short number no separator",
			input:    "500",
			expected: "500",
		},
		{
			name:     "decimal with comma",
			input:    "50,50",
			expected: "50.50",
		},
		{
			name:     "trailing dash",
			input:    "500.000,-",
			expected: "500000",
		},
		{
			name:     "quantity x price",
			input:    "2 x 15.000",
			expected: "2 x 15000",
		},
		{
			name:     "text with numbers",
			input:    "Total: Rp 1.500.000, Subtotal: 1.450.000",
			expected: "Total: 1500000, Subtotal: 1450000",
		},
		{
			name:     "multi-dot thousands with single decimal comma",
			input:    "1.500.000,50",
			expected: "1500000.50",
		},
		{
			name:     "multi-comma thousands with single decimal dot",
			input:    "1,500,000.50",
			expected: "1500000.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
