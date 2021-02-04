package endpoints

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"github.com/yelsukov/otus-ha/dialogue/domain/storages"
	"github.com/yelsukov/otus-ha/dialogue/server"
)

func GetChatsRoutes(storage storages.ChatStorage) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", fetchChats(storage))
	r.Get("/{id:[0-9a-z]+}", getChat(storage))
	r.Post("/", createChat(storage))
	return r
}

type chatResponse struct {
	Object string `json:"object"`
	*entities.Chat
}

func prepareChatsList(chats []entities.Chat) *[]chatResponse {
	list := make([]chatResponse, len(chats), len(chats))
	for i, chat := range chats {
		list[i] = chatResponse{"chat", &chat}
	}
	return &list
}

func fetchChats(s storages.ChatStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.URL.Query().Get("user_id")
		if userId == "" {
			server.ResponseWithError(w, entities.NewError("4000", "user id can't be empty"))
			return
		}

		lastId := r.URL.Query().Get("last_id")
		var limit int
		if strLimit := r.URL.Query().Get("limit"); strLimit != "" {
			limit, _ = strconv.Atoi(strLimit)
		}

		chats, err := s.ReadMany(userId, lastId, uint32(limit))
		if err != nil {
			server.ResponseWithError(w, err)
			return
		}

		server.ResponseWithList(w, prepareChatsList(chats))
	}
}

func getChat(s storages.ChatStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			server.ResponseWithError(w, entities.NewError("4040", "Chat Not Found"))
			return
		}
		chat, err := s.ReadOne(id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				server.ResponseWithError(w, entities.NewError("4040", "Chat Not Found"))
			} else {
				server.ResponseWithError(w, err)
			}
			return
		}

		server.ResponseWithOk(w, &chatResponse{"chat", chat})
	}
}

type postChatBody struct {
	Users [2]string `json:"users"`
}

func createChat(s storages.ChatStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body postChatBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			server.ResponseWithError(w, entities.NewError("4000", "invalid JSON payload"))
			return
		}

		chat := entities.Chat{Users: body.Users}
		if err := s.InsertOne(&chat); err != nil {
			server.ResponseWithError(w, err)
			return
		}

		server.ResponseWithOk(w, &chatResponse{"chat", &chat})
	}
}
