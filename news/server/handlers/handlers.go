package handlers

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/yelsukov/otus-ha/news/domain/entities"
	"github.com/yelsukov/otus-ha/news/domain/models"
	"github.com/yelsukov/otus-ha/news/domain/storages"
)

func prepareEventsList(messages []*models.Event) *[]entities.EventResponse {
	qty := len(messages)
	list := make([]entities.EventResponse, qty, qty)
	for i := 0; i < qty; i++ {
		list[i] = entities.EventResponse{Object: "event", Event: messages[i]}
	}
	return &list
}

// OnNewConnect handle new connection event
func OnNewConnect(es storages.EventStorage) func(w io.Writer, cid string) {
	return func(w io.Writer, cid string) {
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

// OnRequest handle inbound requests
func OnRequest(es storages.EventStorage) func(w io.Writer, r io.Reader, cid string) {
	return func(w io.Writer, r io.Reader, cid string) {
		buf, _ := ioutil.ReadAll(r)
		act := string(bytes.Trim(buf, "\n\""))
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
