package subscriber

import (
	"context"
	"sync"

	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/ungerr"
	"github.com/nats-io/nats.go/jetstream"
)

type Subscriber struct {
	consumers []jetstream.ConsumeContext
	mu        sync.Mutex
}

func Setup(providers *provider.Providers) (*Subscriber, error) {
	queues := configureQueues(providers)

	js := providers.JetStream
	ctx := context.Background()

	// Ensure stream exists with all task subjects
	subjects := make([]string, 0, len(queues))
	for _, q := range queues {
		subjects = append(subjects, q.name)
	}

	_, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "TASKS",
		Subjects: subjects,
	})
	if err != nil {
		return nil, ungerr.Wrap(err, "error creating NATS stream")
	}

	s := &Subscriber{}

	for _, q := range queues {
		cons, err := js.CreateOrUpdateConsumer(ctx, "TASKS", jetstream.ConsumerConfig{
			Durable:       q.name,
			FilterSubject: q.name,
			AckPolicy:     jetstream.AckExplicitPolicy,
			MaxDeliver:    3,
		})
		if err != nil {
			s.Stop()
			return nil, ungerr.Wrap(err, "error creating consumer for "+q.name)
		}

		cc, err := cons.Consume(q.handler)
		if err != nil {
			s.Stop()
			return nil, ungerr.Wrap(err, "error starting consume for "+q.name)
		}

		s.mu.Lock()
		s.consumers = append(s.consumers, cc)
		s.mu.Unlock()
	}

	return s, nil
}

func (s *Subscriber) Start() error {
	logger.Info("subscriber started")
	return nil
}

func (s *Subscriber) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, cc := range s.consumers {
		cc.Stop()
	}
	s.consumers = nil
	logger.Info("subscriber stopped")
}
