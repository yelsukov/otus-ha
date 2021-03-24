package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/yelsukov/otus-ha/counter/storages"
)

type Client struct {
	client *redis.Client
}

func New() storages.Storage {
	return &Client{}
}

func (c *Client) Connect(ctx context.Context, uri string, password string) error {
	c.client = redis.NewClient(&redis.Options{
		Addr:       uri,
		Password:   password,
		DB:         0, // use default DB
		MaxRetries: 2,
	})
	var err error
	ctxPing, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	for i := 0; i < 10; i++ {
		if err = c.client.Ping(ctxPing).Err(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}) (string, error) {
	return c.client.Set(ctx, key, value, -1).Result()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Incr(ctx context.Context, key string, num int64) (int64, error) {
	return c.client.IncrBy(ctx, key, num).Result()
}

func (c *Client) Decr(ctx context.Context, key string, num int64) (int64, error) {
	return c.client.DecrBy(ctx, key, num).Result()
}

func (c *Client) Close() error {
	return c.client.Close()
}
