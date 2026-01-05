package provider

import (
	"github.com/go-playground/validator/v10"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/service/llm"
	"github.com/itsLeonB/cashback/internal/core/service/mail"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/core/service/store"
)

type CoreServices struct {
	LLM   llm.LLMService
	Mail  mail.MailService
	Image storage.ImageService
	State store.StateStore
}

func (cs *CoreServices) Shutdown() error {
	return cs.State.Shutdown()
}

func ProvideCoreServices() (*CoreServices, error) {
	storageRepo, err := storage.NewGCSStorageRepository()
	if err != nil {
		return nil, err
	}

	store, err := store.NewStateStore()
	if err != nil {
		return nil, err
	}

	return &CoreServices{
		LLM:   llm.NewLLMService(config.Global.LLM),
		Mail:  mail.NewMailService(),
		Image: storage.NewImageService(validator.New(), storageRepo),
		State: store,
	}, nil
}
