package storages

import (
	"context"
	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageStorage struct {
	ctx context.Context
	col *mongo.Collection
}

func NewMessageStorage(ctx context.Context, db *mongo.Database) *MessageStorage {
	return &MessageStorage{ctx, db.Collection("messages")}
}

func (m *MessageStorage) InsertOne(message *entities.Message) error {
	if err := message.Validate(); err != nil {
		return err
	}
	message.BeforeSave()

	result, err := m.col.InsertOne(m.ctx, message)
	if err != nil {
		return err
	}
	message.Id = result.InsertedID.(primitive.ObjectID)

	return nil
}

func (m *MessageStorage) ReadOne(id *primitive.ObjectID) (*entities.Message, error) {
	var message entities.Message
	err := m.col.FindOne(m.ctx, bson.M{"_id": id}).Decode(&message)
	return &message, err
}

func (m *MessageStorage) ReadMany(chatId, lastId *primitive.ObjectID, limit uint32) ([]entities.Message, error) {
	var messages []entities.Message

	opts := options.Find()
	opts.SetSort(bson.D{{"ts", -1}})
	if limit == 0 {
		limit = 25
	}
	opts.SetLimit(int64(limit))

	filter := bson.D{{"cid", chatId}}
	if lastId != nil {
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

func (m *MessageStorage) SetReadFlag(ids []primitive.ObjectID, flag bool) (int64, error) {
	res, err := m.col.UpdateMany(m.ctx,
		bson.M{"_id": bson.M{"$in": ids}},
		bson.M{"$set": bson.M{"rdn": flag}},
	)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}
