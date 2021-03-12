package entities

import "github.com/yelsukov/otus-ha/news/domain/models"

type EventResponse struct {
	Object string `json:"object"`
	*models.Event
}
