package storages

import (
	"context"
	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MessageStorage struct {
	ctx context.Context
	col *mongo.Collection
}

func NewMessageStorage(ctx context.Context, db *mongo.Database) *MessageStorage {
	return &MessageStorage{ctx, db.Collection("messages")}
}

func (m *MessageStorage) InsertOne(message *entities.Message) error {
	now := time.Now()
	message.Date = now.Format("YYYY-MM-DD")
	message.CreatedAt = now.Unix()
	result, err := m.col.InsertOne(m.ctx, message)
	if err != nil {
		return err
	}
	message.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (m *MessageStorage) ReadOne(id string) (*entities.Message, error) {
	var message entities.Message
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = m.col.FindOne(m.ctx, bson.M{"_id": _id}).Decode(&message)
	return &message, err
}

func (m *MessageStorage) ReadMany(chatId string, lastId string, limit uint32) ([]entities.Message, error) {
	var messages []entities.Message

	opts := options.Find()
	opts.SetSort(bson.D{{"$natural", -1}})
	if limit == 0 {
		limit = 25
	}
	opts.SetLimit(int64(limit))

	cid, err := primitive.ObjectIDFromHex(chatId)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"cid", cid}}
	if lastId != "" {
		filter = append(filter, bson.E{"_id", bson.D{{"$gt", lastId}}})
	}

	cursor, err := m.col.Find(m.ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(m.ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}
