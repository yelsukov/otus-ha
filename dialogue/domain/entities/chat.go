package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Chat struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Users     [2]string          `bson:"users" json:"users"`
	CreatedAt int64              `bson:"ts" json:"created_at"`
}

func (c *Chat) Validate() error {
	return nil
}

func (c *Chat) BeforeSave() error {
	return nil
}
