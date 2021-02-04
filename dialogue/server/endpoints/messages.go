package endpoints

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"github.com/yelsukov/otus-ha/dialogue/domain/storages"
	"github.com/yelsukov/otus-ha/dialogue/server"
)

func GetMessagesRoutes(storage storages.MessageStorage) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", fetchMessages(storage))
	r.Get("/{id:[0-9]+}", getMessage(storage))
	r.Post("/", createMessage(storage))
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
		chatId := r.URL.Query().Get("chat_id")
		if chatId == "" {
			server.ResponseWithError(w, entities.NewError("4040", "Chat Not Found"))
			return
		}

		lastId := r.URL.Query().Get("last_id")
		var limit int
		if strLimit := r.URL.Query().Get("limit"); strLimit != "" {
			limit, _ = strconv.Atoi(strLimit)
		}

		messages, err := s.ReadMany(chatId, lastId, uint32(limit))
		if err != nil {
			server.ResponseWithError(w, err)
			return
		}

		server.ResponseWithList(w, prepareMessageList(messages))
	}
}

func getMessage(s storages.MessageStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			server.ResponseWithError(w, entities.NewError("4041", "Message Not Found"))
			return
		}
		message, err := s.ReadOne(id)
		if err != nil {
			server.ResponseWithError(w, err)
			return
		}

		server.ResponseWithOk(w, &messageResponse{"message", message})
	}
}

func createMessage(s storages.MessageStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cid, err := primitive.ObjectIDFromHex(chi.URLParam(r, "chat_id"))
		if err != nil {
			log.WithError(err).Error("failed to parse chat id to ObjectId")
			server.ResponseWithError(w, entities.NewError("4001", "invalid chat ID"))
			return
		}
		aid, err := strconv.Atoi(chi.URLParam(r, "author_id"))
		if err != nil || aid == 0 {
			server.ResponseWithError(w, entities.NewError("4002", "invalid author ID"))
			return
		}
		text := chi.URLParam(r, "message")
		if text == "" {
			server.ResponseWithError(w, entities.NewError("4003", "message cannot be empty"))
			return
		}

		message := entities.Message{
			ChatId:   cid,
			AuthorId: uint64(aid),
			Text:     text,
		}

		if err := s.InsertOne(&message); err != nil {
			server.ResponseWithError(w, err)
			return
		}

		server.ResponseWithOk(w, &messageResponse{"message", &message})
	}
}
