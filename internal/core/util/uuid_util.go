package util

import "github.com/google/uuid"

func ToString(id uuid.UUID) string {
	return id.String()
}
