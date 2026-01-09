package queue

import "context"

type TaskMessage interface {
	Type() string
}

type TaskQueue interface {
	Enqueue(ctx context.Context, message TaskMessage) error
	Shutdown() error
}
