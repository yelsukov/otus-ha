package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

func Connect(ctx context.Context, uri string, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:       uri,
		Password:   password,
		DB:         0, // use default DB
		MaxRetries: 2,
	})
	var err error
	ctxPing, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	for i := 0; i < 10; i++ {
		if err = client.Ping(ctxPing).Err(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	return client, nil
}
