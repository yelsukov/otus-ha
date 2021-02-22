package models

type News struct {
	Object    string `json:"object"`
	Name      string `json:"name"`
	Message   string `json:"message"`
	Timestamp int64  `json:"ts"`
}
