package storages

import (
	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatStorage interface {
	InsertOne(chat *entities.Chat) error
	ReadOne(id *primitive.ObjectID) (*entities.Chat, error)
	ReadMany(userId int, lastId *primitive.ObjectID, limit uint32) ([]entities.Chat, error)
	Update(chat *entities.Chat) error
}
