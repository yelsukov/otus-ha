package responses

import (
	"encoding/json"
	"github.com/yelsukov/otus-ha/backend/errors"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/backend/models"
)

type User struct {
	Object string `json:"object"`
	models.User
}

type SignIn struct {
	Object   string        `json:"object"`
	Token    string        `json:"token"`
	UserId   uint32        `json:"user_id"`
	Username string        `json:"username"`
	ExpireAt time.Duration `json:"expire_at"`
}

func ResponseWithError(w http.ResponseWriter, e error) {
	var status int
	var code, message string
	switch e.(type) {
	case *errors.KernelError:
		e := e.(*errors.KernelError)
		code = e.Code
		message = e.Message
		status, _ = strconv.Atoi(e.Code[0:3])
		break
	default:
		status = 500
		code = "5000"
		message = "Internal Server Error"
		log.Error("Internal Error: " + e.Error())
	}

	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(map[string]string{
		"object":  "error",
		"code":    code,
		"message": message,
	})
	if err != nil {
		log.WithError(err).Warn("failed to response with Error")
	}
}

func ResponseWithOk(w http.ResponseWriter, payload interface{}) {
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.WithError(err).Warn("failed to response with OK")
	}
}
