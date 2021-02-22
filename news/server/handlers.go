package server

import (
	"net/http"
	"strconv"

	"github.com/yelsukov/otus-ha/news/domain/entities"
	"github.com/yelsukov/otus-ha/news/domain/models"
	"github.com/yelsukov/otus-ha/news/domain/storages"
)

type eventResponse struct {
	Object string `json:"object"`
	*models.Event
}

func prepareEventsList(messages []*models.Event) *[]eventResponse {
	qty := len(messages)
	list := make([]eventResponse, qty, qty)
	for i := 0; i < qty; i++ {
		list[i] = eventResponse{"event", messages[i]}
	}
	return &list
}

func fetchEvents(es storages.EventStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fid, err := strconv.Atoi(r.URL.Query().Get("fid"))
		if err != nil {
			ResponseWithError(w, entities.NewError("4002", "invalid fid (follower id)"))
			return
		}

		events, err := es.ReadMany(fid)
		if err != nil {
			ResponseWithError(w, err)
			return
		}

		ResponseWithList(w, prepareEventsList(events))
	}
}
