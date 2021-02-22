package models

import "encoding/json"

type Event struct {
	Name       string `bson:"name" json:"name"`
	Message    string `bson:"msg" json:"message"`
	ProducerId int    `bson:"uid" json:"uid"`
	Followers  []int  `bson:"flr" json:"-"`
	Timestamp  int    `bson:"ts" json:"ts"`
}

func (e *Event) String() string {
	res, _ := json.Marshal(e)
	return string(res)
}
