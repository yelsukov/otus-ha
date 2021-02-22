package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/news/domain/storages"
	"github.com/yelsukov/otus-ha/news/vars"
)

type Server struct {
	ctx context.Context
	mux *chi.Mux
}

func InitRouter() *chi.Mux {
	mux := chi.NewRouter()
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

func NewServer(ctx context.Context, storage storages.EventStorage) *Server {
	mux := InitRouter()
	mux.Get("/events", fetchEvents(storage))
	return &Server{ctx, mux}
}

func (s *Server) Serve(port string) {
	server := http.Server{
		Addr:         ":" + port,
		Handler:      s.mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	go func(s *http.Server) {
		err := s.ListenAndServe()
		log.WithError(err).Error("fail of listen and serve")
	}(&server)

	log.Info("http server started on port " + port)

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
