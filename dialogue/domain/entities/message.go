package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Message struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChatId    primitive.ObjectID `bson:"cid" json:"cid"`
	UserId    int                `bson:"aid" json:"uid"`
	CreatedAt int64              `bson:"ts" json:"createdAt"`
	Text      string             `bson:"txt" json:"text"`
}

func (m *Message) Validate() error {
	if m.UserId == 0 {
		return NewError("4002", "invalid user ID")
	}
	if m.Text == "" {
		return NewError("4003", "message cannot be empty")
	}
	return nil
}

func (m *Message) BeforeSave() {
	m.CreatedAt = time.Now().Unix()
}
