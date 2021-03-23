package storages

import (
	"context"
	"errors"
	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
)

var ErrSagaNotFound = errors.New("saga not found")

type Storage interface {
	Get(ctx context.Context, id string) (entities.Saga, error)
	Del(ctx context.Context, id string) error
	Save(ctx context.Context, sg *entities.Saga) error
	Connected() bool
}
