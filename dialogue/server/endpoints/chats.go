package endpoints

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	r.Use(authMiddleware)
	r.Get("/", fetchChats(storage))
	r.Get("/{cid:[0-9a-z]+}", getChat(storage))
	r.Post("/", createChat(storage))
	r.Put("/{cid:[0-9a-z]+}", addUsers(storage))
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
		userId, err := strconv.Atoi(r.URL.Query().Get("uid"))
		if err != nil {
			server.ResponseWithError(w, entities.NewError("4000", "user id is invalid"))
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

		chats, err := s.ReadMany(userId, &lastId, uint32(limit))
		if err != nil {
			server.ResponseWithError(w, err)
			return
		}

		server.ResponseWithList(w, prepareChatsList(chats))
	}
}

func getChat(s storages.ChatStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := strconv.Atoi(r.URL.Query().Get("uid"))
		if err != nil {
			server.ResponseWithError(w, entities.NewError("4002", "invalid user id"))
			return
		}
		cid, err := primitive.ObjectIDFromHex(chi.URLParam(r, "cid"))
		if err != nil {
			server.ResponseWithError(w, entities.NewError("4001", "invalid chat id"))
			return
		}
		chat, err := s.ReadOne(&cid)
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

		server.ResponseWithOk(w, &chatResponse{"chat", chat})
	}
}

type postChatBody struct {
	Users []int `json:"users"`
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

func addUsers(cs storages.ChatStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := strconv.Atoi(r.URL.Query().Get("uid"))
		if err != nil {
			server.ResponseWithError(w, entities.NewError("4002", "invalid user id"))
			return
		}

		cid, err := primitive.ObjectIDFromHex(chi.URLParam(r, "cid"))
		if err != nil {
			server.ResponseWithError(w, entities.NewError("4001", "invalid chat id"))
			return
		}

		var body postChatBody
		if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
			server.ResponseWithError(w, entities.NewError("4000", "invalid JSON payload"))
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

		if len(body.Users) > 0 {
			chat.Users = append(chat.Users, body.Users...)
			if err := cs.Update(chat); err != nil {
				server.ResponseWithError(w, err)
				return
			}
		}

		server.ResponseWithOk(w, &chatResponse{"chat", chat})
	}
}
