package util

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func GenerateSlug(name string, id uuid.UUID) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = nonAlphanumeric.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "profile"
	}
	suffix := strings.ReplaceAll(id.String(), "-", "")[:6]
	return slug + "-" + suffix
}
