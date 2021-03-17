package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/yelsukov/otus-ha/backend/errors"
	"github.com/yelsukov/otus-ha/backend/providers/dialogue"
	"github.com/yelsukov/otus-ha/backend/server/v1/responses"
)

type dialogueProvider interface {
	FetchChats(ctx context.Context, uid, limit int) ([]byte, error)
	GetChat(ctx context.Context, uid int, cid string) ([]byte, error)
	CreateChat(ctx context.Context, users []int) ([]byte, error)
	AddUsers2Chat(ctx context.Context, uid int, cid string, users []int) ([]byte, error)
	FetchMessages(ctx context.Context, uid int, cid string, limit int) ([]byte, error)
	CreateMessage(ctx context.Context, message dialogue.PostMessageBody) ([]byte, error)
}

func FetchChats(provider dialogueProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := r.Context().Value("currentUserId").(int64)
		chats, err := provider.FetchChats(r.Context(), int(uid), 100)
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		_, _ = w.Write(chats)
	}
}

func GetChat(provider dialogueProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := r.Context().Value("currentUserId").(int64)

		cid := chi.URLParam(r, "cid")
		if cid == "" {
			responses.ResponseWithError(w, errors.New("4001", "invalid chat id"))
			return
		}

		chat, err := provider.GetChat(r.Context(), int(uid), cid)
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		_, _ = w.Write(chat)
	}
}

func CreateChat(provider dialogueProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := int(r.Context().Value("currentUserId").(int64))

		var body map[string][]int
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body["users"]) < 2 {
			responses.ResponseWithError(w, errors.New("4000", "invalid JSON payload"))
			return
		}

		var found bool
		for _, u := range body["users"] {
			if uid == u {
				found = true
				break
			}
		}
		if !found {
			responses.ResponseWithError(w, errors.New("4030", "Forbidden"))
			return
		}

		chat, err := provider.CreateChat(r.Context(), body["users"])
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		_, _ = w.Write(chat)
	}
}

func UpdateChat(provider dialogueProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := r.Context().Value("currentUserId").(int64)

		cid := chi.URLParam(r, "cid")
		if cid == "" {
			responses.ResponseWithError(w, errors.New("4001", "invalid chat id"))
			return
		}

		var body map[string][]int
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body["users"]) < 2 {
			responses.ResponseWithError(w, errors.New("4000", "invalid JSON payload"))
			return
		}

		chat, err := provider.AddUsers2Chat(r.Context(), int(uid), cid, body["users"])
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		_, _ = w.Write(chat)
	}
}

func FetchMessages(provider dialogueProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := r.Context().Value("currentUserId").(int64)

		cid := chi.URLParam(r, "cid")
		if cid == "" {
			responses.ResponseWithError(w, errors.New("4001", "invalid chat id"))
			return
		}

		messages, err := provider.FetchMessages(r.Context(), int(uid), cid, 100)
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		_, _ = w.Write(messages)
	}
}

func SendMessages(provider dialogueProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := r.Context().Value("currentUserId").(int64)

		var body dialogue.PostMessageBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			responses.ResponseWithError(w, errors.New("4000", "invalid JSON payload"))
			return
		}
		body.UserId = int(uid)
		message, err := provider.CreateMessage(r.Context(), body)
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		_, _ = w.Write(message)
	}
}
