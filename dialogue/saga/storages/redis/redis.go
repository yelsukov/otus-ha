package redis

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"github.com/yelsukov/otus-ha/dialogue/saga/storages"
)

type SagaStorageRedis struct {
	client entities.RedisClient
}

func NewSagaStorageRedis(client entities.RedisClient) *SagaStorageRedis {
	return &SagaStorageRedis{client}
}

func (r *SagaStorageRedis) Connected() bool {
	return r.client != nil
}

func (r *SagaStorageRedis) Get(ctx context.Context, id string) (entities.Saga, error) {
	var saga entities.Saga
	found, err := r.client.Get(ctx, id).Result()
	if err != nil {
		if err == redis.Nil {
			return saga, storages.ErrSagaNotFound
		}
		return saga, err
	}
	if err = json.Unmarshal([]byte(found), &saga); err != nil {
		return saga, err
	}

	saga.Id = id

	return saga, nil
}

func (r *SagaStorageRedis) Del(ctx context.Context, id string) error {
	return r.client.Del(ctx, id).Err()
}

func (r *SagaStorageRedis) Save(ctx context.Context, saga *entities.Saga) error {
	payload, err := json.Marshal(saga)
	if err != nil {
		return err
	}

	_, err = r.client.Set(ctx, saga.Id, payload, -1).Result()

	return err
}
