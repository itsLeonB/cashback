package authadapter

import (
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
	"github.com/itsLeonB/sekure"
)

type hashServiceAdapter struct {
	inner sekure.HashService
}

func NewHashService(inner sekure.HashService) auth.HashService {
	return &hashServiceAdapter{inner}
}

func (a *hashServiceAdapter) Hash(password string) (string, error) {
	return a.inner.Hash(password)
}

func (a *hashServiceAdapter) Verify(hash, password string) (bool, error) {
	return a.inner.CheckHash(hash, password)
}
