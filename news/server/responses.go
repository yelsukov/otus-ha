package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/news/domain/entities"
)

func ResponseWithError(w http.ResponseWriter, e error) {
	var status int
	var code, message string
	switch e.(type) {
	case *entities.KernelError:
		e := e.(*entities.KernelError)
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

type ListResponse struct {
	Object string      `json:"object"`
	Data   interface{} `json:"data"`
}

func ResponseWithList(w http.ResponseWriter, payload interface{}) {
	ResponseWithOk(w, &ListResponse{"list", payload})
}
