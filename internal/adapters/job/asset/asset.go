package asset

import (
	"context"

	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/provider"
)

type Asset struct {
	transferMethodService service.TransferMethodService
}

func Setup(providers *provider.Providers) *Asset {
	return &Asset{providers.Services.TransferMethod}
}

func (a *Asset) Run() error {
	return a.transferMethodService.SyncMethods(context.Background())
}
