package handlers

import (
	"github.com/go-chi/chi"
	"github.com/yelsukov/otus-ha/backend/models"
	"github.com/yelsukov/otus-ha/backend/server/v1/responses"
	"net/http"
	"strconv"
)

type friendAdder interface {
	Add(userId, friendId int64) error
}

func AddFriend(store friendAdder, userStore userCruder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("currentUserId").(int64)

		friendId, err := strconv.Atoi(chi.URLParam(r, "friend_id"))
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}
		friend, err := userStore.Get(int64(friendId))
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		if err := store.Add(userId, int64(friendId)); err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		responses.ResponseWithOk(w, NewUserResponse(&friend))
	}
}

type friendsFetcher interface {
	FetchFriends(userId int64, offset, limit uint32) ([]models.User, error)
}

func FetchFriends(userStore friendsFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("currentUserId").(int64)

		var offset, limit int
		if strOffset := r.URL.Query().Get("last_id"); strOffset != "" {
			offset, _ = strconv.Atoi(strOffset)
		}
		if strLimit := r.URL.Query().Get("limit"); strLimit != "" {
			limit, _ = strconv.Atoi(strLimit)
		}

		users, err := userStore.FetchFriends(userId, uint32(offset), uint32(limit))
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		responses.ResponseWithOk(w, NewUsersListResponse(users))
	}
}

type friendshipDeleter interface {
	Delete(userId, friendId int64) error
}

func DeleteFriend(store friendshipDeleter, userStore userCruder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("currentUserId").(int64)

		friendId, err := strconv.Atoi(chi.URLParam(r, "friend_id"))
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}
		friend, err := userStore.Get(int64(friendId))
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		if err := store.Delete(userId, int64(friendId)); err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		responses.ResponseWithOk(w, NewUserResponse(&friend))
	}
}
