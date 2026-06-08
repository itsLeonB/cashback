package provider

import (
	"errors"

	"github.com/go-playground/validator/v10"
	adapters "github.com/itsLeonB/cashback/internal/adapters/core/service/queue"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/service/langfuse"
	"github.com/itsLeonB/cashback/internal/core/service/llm"
	"github.com/itsLeonB/cashback/internal/core/service/mail"
	"github.com/itsLeonB/cashback/internal/core/service/ocr"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/core/service/store"
	"github.com/itsLeonB/cashback/internal/core/service/webpush"
	"github.com/itsLeonB/ungerr"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type CoreServices struct {
	LLM      llm.LLMService
	Mail     mail.MailService
	Image    storage.ImageService
	State    store.StateStore
	OCR      ocr.OCRService
	Storage  storage.StorageRepository
	Queue    queue.TaskQueue
	WebPush  webpush.Client
	Langfuse langfuse.Client

	NATSConn  *nats.Conn
	JetStream jetstream.JetStream
}

func (cs *CoreServices) Shutdown() error {
	var errs error
	if e := cs.State.Shutdown(); e != nil {
		errs = errors.Join(errs, e)
	}
	if e := cs.Queue.Shutdown(); e != nil {
		errs = errors.Join(errs, e)
	}
	if e := cs.Langfuse.Shutdown(); e != nil {
		errs = errors.Join(errs, e)
	}
	if e := cs.NATSConn.Drain(); e != nil {
		errs = errors.Join(errs, e)
	}
	return errs
}

func ProvideCoreServices() (*CoreServices, error) {
	storageRepo, err := storage.NewGCSStorageRepository()
	if err != nil {
		return nil, err
	}

	ocrClient, err := ocr.NewOCRClient()
	if err != nil {
		return nil, err
	}

	nc, err := nats.Connect(config.Global.Url)
	if err != nil {
		return nil, ungerr.Wrap(err, "error connecting to NATS")
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, ungerr.Wrap(err, "error creating JetStream context")
	}

	stateStore, err := store.NewStateStore(js)
	if err != nil {
		nc.Close()
		return nil, err
	}

	taskQueue := adapters.NewNATSTaskQueue(js)

	return &CoreServices{
		LLM:       llm.NewLLMService(config.Global.LLM),
		Mail:      mail.NewMailService(),
		Image:     storage.NewImageService(validator.New(), storageRepo),
		State:     stateStore,
		OCR:       ocrClient,
		Storage:   storageRepo,
		Queue:     taskQueue,
		WebPush:   webpush.NewWebPush(config.Global.Push),
		Langfuse:  langfuse.NewClient(config.Global.Langfuse),
		NATSConn:  nc,
		JetStream: js,
	}, nil
}
