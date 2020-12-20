package v1

import (
	"context"
	"database/sql"
	"github.com/yelsukov/otus-ha/backend/errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/yelsukov/otus-ha/backend/conf"
	"github.com/yelsukov/otus-ha/backend/jwt"
	"github.com/yelsukov/otus-ha/backend/server/v1/handlers"
	"github.com/yelsukov/otus-ha/backend/server/v1/responses"
	"github.com/yelsukov/otus-ha/backend/storages"
)

const version = "1.0.0"

func InitApiMux(db *sql.DB, cfg *conf.Config) *chi.Mux {
	authorizer := &jwt.JWT{Secret: []byte(cfg.JwtSecret), Ttl: cfg.JwtTtl}

	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Compress(3, "application/json"),
		middleware.SetHeader("Content-Type", "application/json"),
		middleware.SetHeader("Version", version),
	)

	userStore := storages.NewUsersStorage(db)
	friendsStore := storages.NewFriendsStorage(db)
	router.Route("/users", func(r chi.Router) {
		r.Use(authenticationMiddleware(authorizer))
		r.Get("/", handlers.GetUsers(userStore))
		r.Get("/{id:[0-9]+}", handlers.GetUser(userStore, friendsStore))
		r.Get("/me", handlers.GetMe(userStore))
		r.Put("/me", handlers.UpdateMe(userStore))
	})

	router.Route("/friends", func(r chi.Router) {
		r.Use(authenticationMiddleware(authorizer))
		r.Get("/", handlers.FetchFriends(userStore))
		r.Post("/{friend_id:[0-9]+}", handlers.AddFriend(friendsStore, userStore))
		r.Delete("/{friend_id:[0-9]+}", handlers.DeleteFriend(friendsStore, userStore))
	})

	router.Post("/auth/sign-in", handlers.LoginHandler(userStore, authorizer))
	router.Post("/auth/sign-up", handlers.SignupHandler(userStore, authorizer))

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		responses.ResponseWithError(w, errors.New("4040", "resource not found"))
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		responses.ResponseWithError(w, errors.New("4050", "method not allowed"))
	})

	return router
}

func authenticationMiddleware(jwt jwt.Tokenizer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				responses.ResponseWithError(w, errors.New("4010", "an authorization header is required"))
				return
			}

			tkn := strings.Split(authHeader, " ")
			if len(tkn) != 2 {
				responses.ResponseWithError(w, errors.New("4011", "invalid authorization token"))
				return
			}

			uid, err := jwt.ExtractUserId(tkn[1])
			if err != nil {
				responses.ResponseWithError(w, errors.New("4011", "invalid authorization token"))
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "currentUserId", uid)))
		})
	}
}
