package storages

import (
	"github.com/yelsukov/otus-ha/news/domain/models"
)

type EventStorage interface {
	InsertOne(event *models.Event) error
	ReadMany(uid int) ([]*models.Event, error)
}
