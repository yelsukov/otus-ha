package v1

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/tarantool/go-tarantool"

	"github.com/yelsukov/otus-ha/backend/balancer"
	"github.com/yelsukov/otus-ha/backend/bus"
	"github.com/yelsukov/otus-ha/backend/conf"
	"github.com/yelsukov/otus-ha/backend/errors"
	"github.com/yelsukov/otus-ha/backend/jwt"
	"github.com/yelsukov/otus-ha/backend/providers/dialogue"
	"github.com/yelsukov/otus-ha/backend/server/v1/handlers"
	"github.com/yelsukov/otus-ha/backend/server/v1/responses"
	"github.com/yelsukov/otus-ha/backend/storages/mysql"
	"github.com/yelsukov/otus-ha/backend/storages/nosql"
)

const version = "1.0.0"

func InitApiMux(ctx context.Context, db *sql.DB, tdb *tarantool.Connection, bus *bus.Producer, cfg *conf.Config) *chi.Mux {
	authorizer := &jwt.JWT{Secret: []byte(cfg.JwtSecret), Ttl: cfg.JwtTtl}

	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Compress(3, "application/json"),
		middleware.SetHeader("Content-Type", "application/json"),
		middleware.SetHeader("X-Api-Version", version),
	)

	var prefixSearcher handlers.PrefixSearcher
	userStore := mysql.NewUsersStorage(db)
	if tdb == nil {
		prefixSearcher = userStore
	} else {
		prefixSearcher = nosql.NewUsersStorage(tdb)
	}
	friendsStore := mysql.NewFriendsStorage(db)
	router.Route("/users", func(r chi.Router) {
		r.Use(authenticationMiddleware(authorizer))
		r.Get("/", handlers.GetUsers(userStore))
		r.Get("/search", handlers.SearchUsers(prefixSearcher))
		r.Get("/{id:[0-9]+}", handlers.GetUser(userStore, friendsStore))
		r.Get("/me", handlers.GetMe(userStore))
		r.Put("/me", handlers.UpdateMe(userStore))
	})

	router.Route("/friends", func(r chi.Router) {
		r.Use(authenticationMiddleware(authorizer))
		r.Get("/", handlers.FetchFriends(userStore))
		r.Post("/{friend_id:[0-9]+}", handlers.AddFriend(friendsStore, userStore, bus))
		r.Delete("/{friend_id:[0-9]+}", handlers.DeleteFriend(friendsStore, userStore, bus))
	})

	router.Route("/news", func(r chi.Router) {
		r.Use(authParamMiddleware(authorizer))
		r.HandleFunc("/", handlers.GetNews(cfg.NewsServiceUrl, cfg.NewsServiceToken))
	})

	dialogueBalancer, err := balancer.New(cfg.ConsulDsn, cfg.DialogueServiceName, cfg.DialogueServiceHosts, 5*time.Second)
	if err == nil {
		go dialogueBalancer.RunHealthChecker(ctx)
	}
	dialogueService := &dialogue.ServiceProvider{
		Token:    cfg.DialogueServiceToken,
		Balancer: dialogueBalancer,
	}
	router.Route("/chats", func(r chi.Router) {
		r.Use(authenticationMiddleware(authorizer))
		// Fetch chats
		r.Get("/", handlers.FetchChats(dialogueService))
		// Read chat
		r.Get("/{cid:[0-9a-z]+}", handlers.GetChat(dialogueService))
		// Fetch messages
		r.Get("/{cid:[0-9a-z]+}/messages", handlers.FetchMessages(dialogueService))
		// Send message
		r.Post("/{cid:[0-9a-z]+}/messages", handlers.SendMessages(dialogueService))
		// Create chat
		r.Post("/", handlers.CreateChat(dialogueService))
		// Update chat
		r.Put("/{cid:[0-9a-z]+}", handlers.UpdateChat(dialogueService))
	})

	router.Post("/auth/sign-in", handlers.LoginHandler(userStore, authorizer, bus))
	router.With(authenticationMiddleware(authorizer)).Post("/auth/sign-out", handlers.LogoutHandler(userStore, bus))
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

// authParamMiddleware middleware to authenticate client by token from GET params
// used for ws connection coz rxjs websockets lib doesn't allow headers
func authParamMiddleware(jwt jwt.Tokenizer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.URL.Query().Get("token")
			if token == "" {
				responses.ResponseWithError(w, errors.New("4010", "an authorization header is required"))
				return
			}

			uid, err := jwt.ExtractUserId(token)
			if err != nil {
				responses.ResponseWithError(w, errors.New("4011", "invalid authorization token"))
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "currentUserId", uid)))
		})
	}
}
