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
	r.Get("/", fetchMessages(msgStorage))
	r.Get("/{id:[0-9]+}", getMessage(msgStorage))
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

func fetchMessages(s storages.MessageStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chatId, err := primitive.ObjectIDFromHex(r.URL.Query().Get("chat_id"))
		if err != nil {
			server.ResponseWithError(w, entities.NewError("4001", "invalid chat id"))
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

		messages, err := s.ReadMany(&chatId, &lastId, uint32(limit))
		if err != nil {
			server.ResponseWithError(w, err)
			return
		}

		server.ResponseWithList(w, prepareMessageList(messages))
	}
}

func getMessage(s storages.MessageStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
		if err != nil {
			server.ResponseWithError(w, entities.NewError("4006", "invalid message id"))
			return
		}
		message, err := s.ReadOne(&id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				server.ResponseWithError(w, entities.NewError("4041", "Message Not Found"))
			} else {
				server.ResponseWithError(w, err)
			}
			return
		}

		server.ResponseWithOk(w, &messageResponse{"message", message})
	}
}

type postMessageBody struct {
	CID primitive.ObjectID `json:"chat_id"`
	AID int                `json:"author_id"`
	Txt string             `json:"text"`
}

func createMessage(ms storages.MessageStorage, cs storages.ChatStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body postMessageBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			server.ResponseWithError(w, entities.NewError("4000", "invalid JSON payload"))
			return
		}

		chat, err := cs.ReadOne(&body.CID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				server.ResponseWithError(w, entities.NewError("4040", "Chat Not Found"))
			} else {
				server.ResponseWithError(w, err)
			}
			return
		}
		var found = false
		for _, uid := range chat.Users { // TODO method with split search
			if uid == body.AID {
				found = true
				break
			}
		}
		// if it is new user in chat
		if !found {
			chat.Users = append(chat.Users, body.AID)
			if err := cs.Update(chat); err != nil {
				server.ResponseWithError(w, err)
				return
			}
		}

		message := entities.Message{
			ChatId:   body.CID,
			AuthorId: body.AID,
			Text:     body.Txt,
		}

		if err := ms.InsertOne(&message); err != nil {
			server.ResponseWithError(w, err)
			return
		}

		server.ResponseWithOk(w, &messageResponse{"message", &message})
	}
}
