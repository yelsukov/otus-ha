package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"github.com/yelsukov/otus-ha/dialogue/vars"
)

type Server struct {
	ctx  context.Context
	mux  *chi.Mux
	port string
}

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

func InitMiddlewares(mux *chi.Mux) *chi.Mux {
	mux.Use(
		middleware.Recoverer,
		middleware.RequestID,
		middleware.Logger,
		middleware.Compress(3, "application/json"),
		middleware.SetHeader("Content-Type", "application/json"),
		middleware.SetHeader("Version", vars.VERSION),
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Request-Id", middleware.GetReqID(r.Context()))
				next.ServeHTTP(w, r)
			})
		},
		authMiddleware,
	)

	mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(http.StatusText(http.StatusNotFound))); err != nil {
			log.WithError(err).Warn("failed to response with Error")
		}
	})
	mux.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		if _, err := w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed))); err != nil {
			log.WithError(err).Warn("failed to response with Error")
		}
	})

	return mux
}

func NewServer(ctx context.Context, port string) *Server {
	return &Server{
		ctx,
		InitMiddlewares(chi.NewRouter()),
		port,
	}
}

func (s *Server) MountRoutes(pattern string, router *chi.Mux) {
	s.mux.Mount(pattern, router)
}

func (s *Server) Serve() {
	server := http.Server{
		Addr:         ":" + s.port,
		Handler:      s.mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	go func(s *http.Server) {
		err := s.ListenAndServe()
		log.WithError(err).Error("fail of listen and serve")
	}(&server)

	log.Info("http server started on port " + s.port)

	// Got gracefully shut down the http server
	<-s.ctx.Done()

	log.Printf("shut down the http server")
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() { cancel() }()
	if err := server.Shutdown(ctxShutDown); err != nil {
		log.WithError(err).Fatal("server shutdown failed")
	}
	log.Printf("server stopped properly")
}
