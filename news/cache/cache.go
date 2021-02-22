package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/news/domain/models"
)

const (
	MaxListSize     = 1000
	FollowersSuffix = "followers"
	EventsSuffix    = "events"
)

type Cache struct {
	ctx    context.Context
	client *redis.Client
}

func NewCache(ctx context.Context) *Cache {
	return &Cache{ctx: ctx}
}

func (c *Cache) Connect(uri string, password string) error {
	c.client = redis.NewClient(&redis.Options{
		Addr:       uri,
		Password:   password,
		DB:         0, // use default DB
		MaxRetries: 2,
	})
	var err error
	for i := 0; i < 10; i++ {
		if err = c.client.Ping(c.ctx).Err(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return err
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(c.ctx, key, value, ttl).Err()
}

func (c *Cache) Get(key string) (string, error) {
	r, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return r, nil
}

func (c *Cache) AddFollowers(uid int, fids ...int) error {
	if fids == nil {
		return nil
	}
	followers := make([]interface{}, len(fids), len(fids))
	for i, fid := range fids {
		followers[i] = fid
	}
	_, err := c.client.LPush(c.ctx, CreateCacheKey(uid, FollowersSuffix), followers...).Result()
	return err
}

func (c *Cache) AddEvents(uid int, events ...*models.Event) {
	if events == nil {
		return
	}

	key := CreateCacheKey(uid, EventsSuffix)
	for _, event := range events {
		_, err := c.client.TxPipelined(c.ctx, func(pipe redis.Pipeliner) error {
			_, err := pipe.LPush(c.ctx, key, event.String()).Result()
			if err != nil {
				return err
			}
			_, err = pipe.LTrim(c.ctx, key, 0, MaxListSize).Result()
			return err
		})
		if err != nil {
			log.WithError(err).Errorf("fail on adding event %+v to %s", event, key)
		}
	}
}

func (c *Cache) AddEventToFollowers(event *models.Event, fids ...int) {
	if fids == nil {
		return
	}

	for _, fid := range fids {
		c.AddEvents(fid, event)
	}
}

func (c *Cache) ReadList(key string, start, stop int64) ([]string, error) {
	res := c.client.LRange(c.ctx, key, start, stop)
	list, err := res.Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	return list, nil
}

func (c *Cache) ReadFollowers(uid int) ([]int, error) {
	list, err := c.ReadList(CreateCacheKey(uid, FollowersSuffix), 0, -1)
	if err != nil || len(list) == 0 {
		return nil, err
	}

	result := make([]int, len(list), len(list))
	for i, f := range list {
		fid, err := strconv.Atoi(f)
		if err != nil {
			return nil, err
		}
		result[i] = fid
	}
	return result, nil
}

func (c *Cache) ReadEvents(uid int) ([]string, error) {
	return c.ReadList(CreateCacheKey(uid, EventsSuffix), 0, MaxListSize-1)
}

func (c *Cache) DeleteEvents(uid int) error {
	_, err := c.client.Del(c.ctx, CreateCacheKey(uid, EventsSuffix)).Result()
	if err == redis.Nil {
		err = nil
	}
	return err
}

func (c *Cache) Disconnect() error {
	return c.client.Close()
}

func CreateCacheKey(uid int, what string) string {
	return "user:" + strconv.Itoa(uid) + ":" + what
}
