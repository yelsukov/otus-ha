package handlers

import (
	"github.com/yelsukov/otus-ha/backend/providers/news"
	"net/http"

	"github.com/yelsukov/otus-ha/backend/server/v1/responses"
)

type newsReader interface {
	Read(uid int) (*news.ReadResponse, error)
}

func GetNews(service newsReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("currentUserId").(int64)
		res, err := service.Read(int(userId))
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}

		responses.ResponseWithOk(w, res)
	}
}
