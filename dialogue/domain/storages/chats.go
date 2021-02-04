package storages

import (
	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
)

type ChatStorage interface {
	InsertOne(chat *entities.Chat) error
	ReadOne(id string) (*entities.Chat, error)
	ReadMany(userId string, lastId string, limit uint32) ([]entities.Chat, error)
}
