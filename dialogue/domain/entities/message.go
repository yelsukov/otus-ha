package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Message struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChatId    primitive.ObjectID `bson:"cid" json:"chat_id"`
	AuthorId  int                `bson:"aid" json:"author_id"`
	Date      string             `bson:"dt" json:"-"`
	CreatedAt int64              `bson:"ts" json:"created_at"`
	Text      string             `bson:"txt" json:"text"`
}

func (m *Message) Validate() error {
	if m.AuthorId == 0 {
		return NewError("4002", "invalid author ID")
	}
	if m.Text == "" {
		return NewError("4003", "message cannot be empty")
	}
	return nil
}

func (m *Message) BeforeSave() {
	now := time.Now()
	m.Date = now.Format("YYYY-MM-DD")
	m.CreatedAt = now.Unix()
}
