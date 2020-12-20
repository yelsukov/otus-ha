package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/yelsukov/otus-ha/backend/errors"
	"github.com/yelsukov/otus-ha/backend/jwt"
	"github.com/yelsukov/otus-ha/backend/models"
	"github.com/yelsukov/otus-ha/backend/server/v1/responses"
)

type SignUpRequest struct {
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
	models.User
}

type userCreator interface {
	Create(*models.User) (int64, error)
}

func SignupHandler(store userCreator, jwt jwt.Tokenizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			responses.ResponseWithError(w, errors.New("4000", "invalid JSON payload"))
			return
		}

		pass := models.PassChange{Password: body.Password, PasswordConfirm: body.PasswordConfirm}
		if err := pass.Validate(); err != nil {
			responses.ResponseWithError(w, err)
			return
		}
		if err := body.SetPassword(pass.Password); err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		userId, err := store.Create(&body.User)
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		token, err := jwt.Tokenize(userId, body.Username)
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		responses.ResponseWithOk(w, token)
	}
}

type SignInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type authenticator interface {
	Login(username, password string) (models.User, error)
}

func LoginHandler(store authenticator, jwt jwt.Tokenizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body SignInRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			responses.ResponseWithError(w, errors.New("4000", "invalid JSON payload"))
			return
		}
		if body.Username == "" {
			responses.ResponseWithError(w, errors.New("4001", "username is required"))
			return
		}
		if body.Password == "" {
			responses.ResponseWithError(w, errors.New("4002", "password is required"))
			return
		}

		user, err := store.Login(body.Username, body.Password)
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		token, err := jwt.Tokenize(user.Id, user.Username)
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		responses.ResponseWithOk(w, token)
	}
}
