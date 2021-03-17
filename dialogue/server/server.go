package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"github.com/yelsukov/otus-ha/dialogue/vars"
)

type Server struct {
	mux *chi.Mux
	ca  entities.ConsulAgent
	hs  *http.Server
}

func InitMiddlewares(mux *chi.Mux) *chi.Mux {
	mux.Use(
		middleware.Recoverer,
		middleware.RequestID, // Check Request ID in headers and create new if it is empty
		middleware.Logger,
		middleware.Compress(3, "application/json"),
		middleware.SetHeader("Content-Type", "application/json"),
		middleware.SetHeader("Version", vars.VERSION),
		// Put request id to response
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Request-Id", middleware.GetReqID(r.Context()))
				next.ServeHTTP(w, r)
			})
		},
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

	mux.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("pong")); err != nil {
			log.WithError(err).Warn("failed to pong")
		}
	})

	return mux
}

func NewServer(ca entities.ConsulAgent) *Server {
	return &Server{
		InitMiddlewares(chi.NewRouter()),
		ca,
		nil,
	}
}

func (s *Server) MountRoutes(pattern string, router *chi.Mux) {
	s.mux.Mount(pattern, router)
}

func (s *Server) Serve(port string) error {
	s.hs = &http.Server{
		Addr:         ":" + port,
		Handler:      s.mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	log.Info("registering service at consul")
	if err := s.ca.Register(); err != nil {
		return err
	}

	log.Info("http server listening on port " + port)
	err := s.hs.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}

	return err
}

func (s *Server) Shutdown() {
	log.Printf("shut down the http server")
	if err := s.ca.Unregister(); err != nil {
		log.WithError(err).Error("failed to unregister service from consul")
	}

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() { cancel() }()
	if err := s.hs.Shutdown(ctxShutDown); err != nil {
		log.WithError(err).Fatal("server shutdown failed")
	} else {
		log.Printf("server stopped properly")
	}
}
