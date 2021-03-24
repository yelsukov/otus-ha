package queues

import "context"

type Consumer interface {
	Listen(ctx context.Context, consumeFunc func(message []byte)) error
	Close() error
}

type Producer interface {
	Publish(ctx context.Context, message []byte) error
	Close() error
}
