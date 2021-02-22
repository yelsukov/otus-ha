package entities

type Event struct {
	Name      string `json:"name"`
	Message   string `json:"msg"`
	ObjectId  int    `json:"oid"`
	SubjectId int    `json:"sid"`
	Timestamp int    `json:"ts"`
}
