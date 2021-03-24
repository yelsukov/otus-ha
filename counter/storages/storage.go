package storages

import (
	"context"
)

type Storage interface {
	Connect(ctx context.Context, dsn string, pass string) error
	Set(ctx context.Context, key string, value interface{}) (string, error)
	Get(ctx context.Context, key string) (string, error)

	Incr(ctx context.Context, key string, num int64) (int64, error)
	Decr(ctx context.Context, key string, num int64) (int64, error)

	Close() error
}
