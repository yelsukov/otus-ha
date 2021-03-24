package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"github.com/yelsukov/otus-ha/dialogue/domain/storages"
	"github.com/yelsukov/otus-ha/dialogue/server"
)

func GetMessagesRoutes(msgStorage storages.MessageStorage, chtStorage storages.ChatStorage, sagaOrc entities.SagaOrchestrator) *chi.Mux {
	r := chi.NewRouter()
	r.Use(server.AuthMiddleware)

	r.Get("/", fetchMessages(msgStorage, chtStorage, sagaOrc))
	r.Post("/", createMessage(msgStorage, chtStorage, sagaOrc))
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

func fetchMessages(ms storages.MessageStorage, cs storages.ChatStorage, so entities.SagaOrchestrator) http.HandlerFunc {
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

		if !so.IsActive() {
			return
		}
		unread := make([]string, 10)
		for _, m := range messages {
			if m.UserId != uid && !m.Read {
				unread = append(unread, m.Id.String())
			}
		}
		if len(unread) == 0 {
			return
		}

		num, err := ms.SetReadFlag(unread, true)
		if err != nil {
			log.WithError(err).Error("failed to update read flag for messages")
			return
		}
		err = so.ExecuteSaga(context.Background(), &entities.Saga{
			Id:        "read-msg-" + strconv.Itoa(time.Now().Nanosecond()),
			DialogTrx: entities.DialogTrx{MessagesIds: unread},
			CounterTrx: entities.CounterTrx{
				Command: "dec",
				ChatId:  chat.Id.String(),
				UserId:  uid,
				Num:     uint(num),
			},
			// Compensate transaction for Read flag update
			Compensate: func(s *entities.Saga) error {
				_, err := ms.SetReadFlag(s.DialogTrx.MessagesIds, false)
				return err
			},
		})
		if err != nil {
			log.WithError(err).Error("failed to execute saga")
		}
	}
}

type postMessageBody struct {
	ChatId primitive.ObjectID `json:"cid"`
	UserId int                `json:"uid"`
	Text   string             `json:"txt"`
}

func createMessage(ms storages.MessageStorage, cs storages.ChatStorage, so entities.SagaOrchestrator) http.HandlerFunc {
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

		if err = ms.InsertOne(&message); err != nil {
			server.ResponseWithError(w, err)
			return
		}

		server.ResponseWithOk(w, &messageResponse{"message", &message})

		if !so.IsActive() {
			return
		}
		users := chat.UsersExceptOne(message.UserId)
		if users == nil {
			log.Warnf("message user %d not found in chat %s", message.UserId, chat.Id)
			return
		}
		err = so.ExecuteSaga(context.Background(), &entities.Saga{
			Id: "new-msg-" + strconv.Itoa(time.Now().Nanosecond()),
			CounterTrx: entities.CounterTrx{
				Command: "incr",
				ChatId:  message.ChatId.String(),
				UserId:  users[0],
				Num:     1,
			},
		})
		if err != nil {
			log.WithError(err).Error("failed to execute saga")
		}
	}
}
