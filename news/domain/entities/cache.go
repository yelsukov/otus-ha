package entities

import (
	"github.com/yelsukov/otus-ha/news/domain/models"
	"time"
)

type Cache interface {
	Set(key string, value interface{}, ttl time.Duration) error
	Get(key string) (string, error)

	AddFollowers(uid int, fids ...int) error
	ReadFollowers(uid int) ([]int, error)

	AddEvents(uid int, events ...*models.Event)
	AddEventToFollowers(event *models.Event, fids ...int)
	ReadEvents(uid int) ([]string, error)
	DeleteEvents(uid int) error

	Connect(uri string, password string) error
	Disconnect() error
}
