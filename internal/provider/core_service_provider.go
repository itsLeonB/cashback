package provider

import (
	"crypto/tls"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	adapters "github.com/itsLeonB/cashback/internal/adapters/core/service/queue"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/service/llm"
	"github.com/itsLeonB/cashback/internal/core/service/mail"
	"github.com/itsLeonB/cashback/internal/core/service/ocr"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/core/service/store"
	"github.com/itsLeonB/cashback/internal/core/service/webpush"
)

type CoreServices struct {
	LLM     llm.LLMService
	Mail    mail.MailService
	Image   storage.ImageService
	State   store.StateStore
	OCR     ocr.OCRService
	Storage storage.StorageRepository
	Queue   queue.TaskQueue
	WebPush webpush.Client
}

func (cs *CoreServices) Shutdown() error {
	var errs error
	if e := cs.State.Shutdown(); e != nil {
		errs = errors.Join(errs, e)
	}
	if e := cs.Queue.Shutdown(); e != nil {
		errs = errors.Join(errs, e)
	}
	return errs
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

	ocr, err := ocr.NewOCRClient()
	if err != nil {
		return nil, err
	}

	taskQueue, err := adapters.NewTaskQueue(RedisClientOpts(config.Global.Valkey))
	if err != nil {
		return nil, err
	}

	return &CoreServices{
		llm.NewLLMService(config.Global.LLM),
		mail.NewMailService(),
		storage.NewImageService(validator.New(), storageRepo),
		store,
		ocr,
		storageRepo,
		taskQueue,
		webpush.NewWebPush(config.Global.Push),
	}, nil
}

func RedisClientOpts(cfg config.Valkey) asynq.RedisClientOpt {
	opt := asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.Db,
	}

	if cfg.EnableTls {
		opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	return opt
}
