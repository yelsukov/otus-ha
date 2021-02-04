package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Message struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChatId    primitive.ObjectID `bson:"cid" json:"chat_id"`
	AuthorId  uint64             `bson:"aid" json:"author_id"`
	Date      string             `bson:"dt" json:"-"`
	CreatedAt int64              `bson:"ts" json:"created_at"`
	Text      string             `bson:"txt" json:"text"`
}

func (m *Message) Validate() error {
	return nil
}

func (m *Message) BeforeSave() error {
	return nil
}
