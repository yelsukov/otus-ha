package storages

import (
	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageStorage interface {
	InsertOne(message *entities.Message) error
	ReadOne(id *primitive.ObjectID) (*entities.Message, error)
	ReadMany(chatId *primitive.ObjectID, lastId *primitive.ObjectID, limit uint32) ([]entities.Message, error)
	SetReadFlag(ids []string, flag bool) (int64, error)
}
