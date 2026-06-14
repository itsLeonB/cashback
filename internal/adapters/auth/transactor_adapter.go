package authadapter

import (
	"context"

	"github.com/itsLeonB/go-authkit"
	"github.com/itsLeonB/go-crud"
)

type transactorAdapter struct {
	inner crud.Transactor
}

func NewTransactor(inner crud.Transactor) authkit.Transactor {
	return &transactorAdapter{inner}
}

func (a *transactorAdapter) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return a.inner.WithinTransaction(ctx, fn)
}
