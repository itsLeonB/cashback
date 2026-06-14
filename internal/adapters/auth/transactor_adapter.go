package authadapter

import (
	"context"

	"github.com/itsLeonB/cashback/internal/domain/service/auth"
	"github.com/itsLeonB/go-crud"
)

type transactorAdapter struct {
	inner crud.Transactor
}

func NewTransactor(inner crud.Transactor) auth.Transactor {
	return &transactorAdapter{inner}
}

func (a *transactorAdapter) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return a.inner.WithinTransaction(ctx, fn)
}
