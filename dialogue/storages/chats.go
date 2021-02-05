package storages

import (
	"context"
	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatStorage struct {
	ctx context.Context
	col *mongo.Collection
}

func NewChatStorage(ctx context.Context, db *mongo.Database) *ChatStorage {
	return &ChatStorage{ctx, db.Collection("chats")}
}

func (c *ChatStorage) InsertOne(chat *entities.Chat) error {
	if err := chat.Validate(); err != nil {
		return err
	}
	chat.BeforeSave()

	result, err := c.col.InsertOne(c.ctx, chat)
	if err != nil {
		return err
	}
	chat.Id = result.InsertedID.(primitive.ObjectID)

	return nil
}

func (c *ChatStorage) ReadOne(id *primitive.ObjectID) (*entities.Chat, error) {
	var chat entities.Chat
	err := c.col.FindOne(c.ctx, bson.D{{"_id", id}}).Decode(&chat)
	return &chat, err
}

func (c *ChatStorage) ReadMany(userId int, lastId *primitive.ObjectID, limit uint32) ([]entities.Chat, error) {
	var chats []entities.Chat

	opts := options.Find()
	opts.SetSort(bson.D{{"$natural", -1}})
	if limit == 0 {
		limit = 25
	}
	opts.SetLimit(int64(limit))

	filter := bson.D{{"users", userId}}
	if lastId != nil {
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

func (c *ChatStorage) Update(chat *entities.Chat) error {
	if err := chat.Validate(); err != nil {
		return err
	}

	_, err := c.col.UpdateOne(
		c.ctx,
		bson.M{"_id": chat.Id},
		bson.D{
			{"$set", bson.D{{"users", chat.Users}}},
		},
	)

	return err
}
