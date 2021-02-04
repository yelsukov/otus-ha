package storages

import (
	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
)

type MessageStorage interface {
	InsertOne(message *entities.Message) error
	ReadOne(id string) (*entities.Message, error)
	ReadMany(chatId string, lastId string, limit uint32) ([]entities.Message, error)
}
