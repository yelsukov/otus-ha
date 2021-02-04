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

type ChatStorage struct {
	ctx context.Context
	col *mongo.Collection
}

func NewChatStorage(ctx context.Context, db *mongo.Database) *ChatStorage {
	return &ChatStorage{ctx, db.Collection("chats")}
}

func (c *ChatStorage) InsertOne(chat *entities.Chat) error {
	chat.CreatedAt = time.Now().Unix()
	result, err := c.col.InsertOne(c.ctx, chat)
	if err != nil {
		return err
	}
	chat.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (c *ChatStorage) ReadOne(id string) (*entities.Chat, error) {
	var chat entities.Chat
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = c.col.FindOne(c.ctx, bson.D{{"_id", _id}}).Decode(&chat)
	return &chat, err
}

func (c *ChatStorage) ReadMany(userId string, lastId string, limit uint32) ([]entities.Chat, error) {
	var chats []entities.Chat

	opts := options.Find()
	opts.SetSort(bson.D{{"$natural", -1}})
	if limit == 0 {
		limit = 25
	}
	opts.SetLimit(int64(limit))

	filter := bson.D{{"users", userId}}
	if lastId != "" {
		filter = append(filter, bson.E{"_id", bson.D{{"$gt", lastId}}})
	}

	cursor, err := c.col.Find(c.ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(c.ctx, &chats); err != nil {
		return nil, err
	}

	return chats, nil
}
