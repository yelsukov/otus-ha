package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/yelsukov/otus-ha/backend/errors"
	"github.com/yelsukov/otus-ha/backend/models"
	"github.com/yelsukov/otus-ha/backend/server/v1/responses"
	"net/http"
	"strconv"
)

type UserResponse struct {
	Object string `json:"object"`
	*models.User
}

type UsersListResponse struct {
	Object string         `json:"object"`
	Data   []UserResponse `json:"data"`
}

type userCruder interface {
	Create(user *models.User) (int64, error)
	Get(id int64) (models.User, error)
	Fetch(match [][2]string, lastId int64, limit uint32) ([]models.User, error)
	Update(user *models.User, clean *models.User) error
}

type friendshipChecker interface {
	IsFriend(userId int64, friendId int64) (bool, error)
}

func GetUser(store userCruder, friendship friendshipChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}
		user, err := store.Get(int64(userId))
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		// Check is it a friend
		curUserId := r.Context().Value("currentUserId").(int64)
		if curUserId != user.Id {
			user.IsFriend, err = friendship.IsFriend(curUserId, user.Id)
		}

		responses.ResponseWithOk(w, NewUserResponse(&user))
	}
}

func GetUsers(store userCruder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var lastId, limit int
		if strLastId := r.URL.Query().Get("last_id"); strLastId != "" {
			lastId, _ = strconv.Atoi(strLastId)
		}
		if strLimit := r.URL.Query().Get("limit"); strLimit != "" {
			limit, _ = strconv.Atoi(strLimit)
		}

		match := make([][2]string, 0, 2)
		firstName := r.URL.Query().Get("firstName")
		if firstName != "" {
			match = append(match, [2]string{"`first_name` LIKE (?)", firstName + "%"})
		}
		lastName := r.URL.Query().Get("lastName")
		if lastName != "" {
			match = append(match, [2]string{"`last_name` LIKE (?)", lastName + "%"})
		}

		users, err := store.Fetch(match, int64(lastId), uint32(limit))
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		responses.ResponseWithOk(w, NewUsersListResponse(users))
	}
}

func GetMe(store userCruder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("currentUserId").(int64)
		user, err := store.Get(int64(userId))
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		responses.ResponseWithOk(w, NewUserResponse(&user))
	}
}

func UpdateMe(store userCruder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("currentUserId").(int64)
		user, err := store.Get(userId)

		var body models.User
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			responses.ResponseWithError(w, errors.New("4000", "invalid JSON payload"))
			return
		}

		err = store.Update(&body, &user)
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		responses.ResponseWithOk(w, NewUserResponse(&user))
	}
}

func NewUserResponse(u *models.User) UserResponse {
	return UserResponse{"user", u}
}

func NewUsersListResponse(list []models.User) *UsersListResponse {
	qty := len(list)
	rl := make([]UserResponse, qty)
	for i := 0; i < qty; i++ {
		rl[i] = NewUserResponse(&list[i])
	}
	return &UsersListResponse{"list", rl}
}
