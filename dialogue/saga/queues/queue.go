package queues

import (
	"context"
	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
)

type Queue interface {
	Publish(ctx context.Context, sg *entities.Saga) error
	Listen(ctx context.Context, consumeFunc func(message SagaInboundMessage))
}
