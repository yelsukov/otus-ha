package endpoints

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"github.com/yelsukov/otus-ha/dialogue/domain/storages"
	"github.com/yelsukov/otus-ha/dialogue/server"
)

func GetMessagesRoutes(msgStorage storages.MessageStorage, chtStorage storages.ChatStorage) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", fetchMessages(msgStorage, chtStorage))
	r.Post("/", createMessage(msgStorage, chtStorage))
	return r
}

type messageResponse struct {
	Object string `json:"object"`
	*entities.Message
}

func prepareMessageList(messages []entities.Message) *[]messageResponse {
	qty := len(messages)
	list := make([]messageResponse, qty, qty)
	for i := 0; i < qty; i++ {
		list[i] = messageResponse{"message", &messages[i]}
	}
	return &list
}

func fetchMessages(ms storages.MessageStorage, cs storages.ChatStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := strconv.Atoi(r.URL.Query().Get("uid"))
		if err != nil {
			server.ResponseWithError(w, entities.NewError("4002", "invalid user id"))
			return
		}
		cid, err := primitive.ObjectIDFromHex(r.URL.Query().Get("cid"))
		if err != nil {
			server.ResponseWithError(w, entities.NewError("4001", "invalid chat id"))
			return
		}
		chat, err := cs.ReadOne(&cid)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				server.ResponseWithError(w, entities.NewError("4040", "Chat Not Found"))
			} else {
				server.ResponseWithError(w, err)
			}
			return
		}
		if !chat.HasUser(uid) {
			server.ResponseWithError(w, entities.NewError("4031", "user do not belongs to chat"))
			return
		}

		var lastId primitive.ObjectID
		if lid := r.URL.Query().Get("cursor"); lid != "" {
			lastId, err = primitive.ObjectIDFromHex(lid)
			if err != nil {
				server.ResponseWithError(w, entities.NewError("4005", "invalid cursor"))
				return
			}
		}
		var limit int
		if strLimit := r.URL.Query().Get("limit"); strLimit != "" {
			limit, _ = strconv.Atoi(strLimit)
		}

		messages, err := ms.ReadMany(&cid, &lastId, uint32(limit))
		if err != nil {
			server.ResponseWithError(w, err)
			return
		}

		server.ResponseWithList(w, prepareMessageList(messages))
	}
}

type postMessageBody struct {
	ChatId primitive.ObjectID `json:"cid"`
	UserId int                `json:"uid"`
	Text   string             `json:"txt"`
}

func createMessage(ms storages.MessageStorage, cs storages.ChatStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body postMessageBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			server.ResponseWithError(w, entities.NewError("4000", "invalid JSON payload"))
			return
		}

		chat, err := cs.ReadOne(&body.ChatId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				server.ResponseWithError(w, entities.NewError("4040", "Chat Not Found"))
			} else {
				server.ResponseWithError(w, err)
			}
			return
		}
		// if it is new user in chat
		if !chat.HasUser(body.UserId) {
			server.ResponseWithError(w, entities.NewError("4031", "user do not belongs to chat"))
			return
		}

		message := entities.Message{
			ChatId: body.ChatId,
			UserId: body.UserId,
			Text:   body.Text,
		}

		if err := ms.InsertOne(&message); err != nil {
			server.ResponseWithError(w, err)
			return
		}

		server.ResponseWithOk(w, &messageResponse{"message", &message})
	}
}
