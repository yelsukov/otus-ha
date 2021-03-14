package handlers

import (
	"encoding/json"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/news/domain/entities"
)

func ResponseWithError(w io.Writer, e error) {
	var code, message string
	switch e.(type) {
	case *entities.KernelError:
		e := e.(*entities.KernelError)
		code = e.Code
		message = e.Message
		break
	default:
		code = "5000"
		message = "Internal Server Error"
		log.Error("Internal Error: " + e.Error())
	}

	p, err := json.Marshal(map[string]string{
		"object":  "error",
		"code":    code,
		"message": message,
	})
	if err != nil {
		log.WithError(err).Warn("failed to response with Error")
	}
	if _, err = w.Write(p); err != nil {
		log.WithError(err).Warn("failed to write to connection")
	}
}

func ResponseWithOk(w io.Writer, payload interface{}) {
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.WithError(err).Warn("failed to response with OK")
	}
}

type ListResponse struct {
	Object string      `json:"object"`
	Data   interface{} `json:"data"`
}

func ResponseWithList(w io.Writer, payload interface{}) {
	ResponseWithOk(w, &ListResponse{"list", payload})
}
