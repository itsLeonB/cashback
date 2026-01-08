package util

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func GenerateObjectKey(filename string) string {
	ext := filepath.Ext(filename)
	timestamp := time.Now().Format("2006/01/02")
	return fmt.Sprintf("%s/%s%s", timestamp, uuid.NewString(), ext)
}
