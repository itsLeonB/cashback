package billparse

import (
	"regexp"
	"strings"
)

var (
	// Number-looking token regex with leading word boundary
	// Matches tokens that contain digits and common Indonesian separators, including trailing dash
	numericTokenRegex = regexp.MustCompile(`(?i)\b(?:Rp\.?\s?)?[\dO]+(?:[\.,\s][\dO]+)*(?:,-)?`)

	// Currency and trailing dash regex
	currencyRegex     = regexp.MustCompile(`(?i)Rp\.?\s?`)
	trailingDashRegex = regexp.MustCompile(`,-$`)
)

// Normalize processes OCR text to handle Indonesian number formats and OCR errors.
func Normalize(text string) string {
	return numericTokenRegex.ReplaceAllStringFunc(text, func(match string) string {
		// 1. Initial cleanup
		cleaned := currencyRegex.ReplaceAllString(match, "")
		cleaned = trailingDashRegex.ReplaceAllString(cleaned, "")
		cleaned = strings.ReplaceAll(cleaned, "O", "0") // OCR error: O -> 0
		cleaned = strings.ReplaceAll(cleaned, "o", "0")
		cleaned = strings.ReplaceAll(cleaned, " ", "") // Inner spaces

		return applyHeuristics(cleaned)
	})
}

func applyHeuristics(cleaned string) string {
	dots := strings.Count(cleaned, ".")
	commas := strings.Count(cleaned, ",")

	// If no separators, just return as is
	if dots == 0 && commas == 0 {
		return cleaned
	}

	// 2. Heuristic Logic

	// Multiple separators: determine which is thousands vs decimal
	if dots > 1 && commas == 1 {
		// e.g. 1.500.000,50 — dots are thousands, comma is decimal
		cleaned = strings.ReplaceAll(cleaned, ".", "")
		cleaned = strings.ReplaceAll(cleaned, ",", ".")
		return cleaned
	}
	if commas > 1 && dots == 1 {
		// e.g. 1,500,000.50 — commas are thousands, dot is decimal
		cleaned = strings.ReplaceAll(cleaned, ",", "")
		return cleaned
	}
	if dots > 1 || commas > 1 {
		// Ambiguous: strip all separators
		cleaned = strings.ReplaceAll(cleaned, ".", "")
		cleaned = strings.ReplaceAll(cleaned, ",", "")
		return cleaned
	}

	// Both dot and comma present (e.g., 15.500,50)
	if dots == 1 && commas == 1 {
		// Typical ID: . is thousands, , is decimal
		cleaned = strings.ReplaceAll(cleaned, ".", "")
		cleaned = strings.ReplaceAll(cleaned, ",", ".")
		return cleaned
	}

	return handleSingleSeparator(cleaned, commas)
}

func handleSingleSeparator(cleaned string, commas int) string {
	// Only dots or only commas (exactly one)
	separator := "."
	if commas == 1 {
		separator = ","
	}

	parts := strings.Split(cleaned, separator)
	if len(parts) == 2 {
		if len(parts[1]) == 3 {
			// Followed by 3 digits: likely thousands separator (e.g., 500.000)
			return strings.ReplaceAll(cleaned, separator, "")
		}
		if len(parts[1]) == 1 || len(parts[1]) == 2 {
			// Followed by 1 or 2 digits: likely decimal separator (e.g., 500,50 or 500.5)
			return strings.ReplaceAll(cleaned, separator, ".")
		}
	}

	return cleaned
}
