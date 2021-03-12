package server

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/yelsukov/otus-ha/news/domain/entities"
	"github.com/yelsukov/otus-ha/news/domain/models"
	"github.com/yelsukov/otus-ha/news/domain/storages"
)

type WsHandlerFunc func(w io.Writer, r io.Reader, cid string)

func prepareEventsList(messages []*models.Event) *[]entities.EventResponse {
	qty := len(messages)
	list := make([]entities.EventResponse, qty, qty)
	for i := 0; i < qty; i++ {
		list[i] = entities.EventResponse{Object: "event", Event: messages[i]}
	}
	return &list
}

func FetchEvents(es storages.EventStorage) WsHandlerFunc {
	return func(w io.Writer, r io.Reader, cid string) {
		buf, _ := ioutil.ReadAll(r)
		act := string(bytes.TrimRight(buf, "\n"))
		if act != "list" {
			ResponseWithError(w, entities.NewError("4000", "unknown action"))
			return
		}

		uid, err := strconv.Atoi(cid)
		if err != nil {
			ResponseWithError(w, entities.NewError("5000", "invalid connection id"))
			return
		}

		events, err := es.ReadMany(uid)
		if err != nil {
			ResponseWithError(w, err)
			return
		}

		ResponseWithList(w, prepareEventsList(events))
	}
}
