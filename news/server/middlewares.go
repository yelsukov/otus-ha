package server

import (
	"net/http"
	"strings"

	"github.com/yelsukov/otus-ha/news/domain/entities"
	"github.com/yelsukov/otus-ha/news/vars"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			ResponseWithError(w, entities.NewError("4010", "an authorization header is required"))
			return
		}

		tkn := strings.Split(authHeader, " ")
		if len(tkn) != 2 {
			ResponseWithError(w, entities.NewError("4011", "invalid authorization token"))
			return
		}

		if tkn[1] != vars.TOKEN {
			ResponseWithError(w, entities.NewError("4011", "invalid authorization token"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
