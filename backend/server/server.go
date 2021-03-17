package server

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/middleware"
	"github.com/tarantool/go-tarantool"
	"github.com/yelsukov/otus-ha/backend/bus"
	v1 "github.com/yelsukov/otus-ha/backend/server/v1"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/backend/conf"
)

type Server struct {
	ctx context.Context
	cfg *conf.Config
	mux *chi.Mux
	db  *sql.DB
	tt  *tarantool.Connection
	bus *bus.Producer
}

func NewServer(ctx context.Context, cfg *conf.Config, db *sql.DB, tt *tarantool.Connection, bus *bus.Producer) *Server {
	return &Server{
		ctx,
		cfg,
		chi.NewRouter(),
		db,
		tt,
		bus,
	}
}

func (s *Server) setupRoutes() {
	s.mux.Use(
		middleware.Recoverer,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				headers := w.Header()
				headers.Set("Access-Control-Allow-Origin", "*")
				headers.Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")
				headers.Set("Access-Control-Allow-Headers", "Origin, Accept, Content-Type, Authorization")
				if r.Method != http.MethodOptions {
					next.ServeHTTP(w, r)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			})
		},
	)

	// Attach handlers of api v1
	s.mux.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Request-Id", middleware.GetReqID(r.Context()))
				next.ServeHTTP(w, r)
			})
		})
		r.Mount("/v1", v1.InitApiMux(s.ctx, s.db, s.tt, s.bus, s.cfg))
	})

	s.mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(http.StatusText(http.StatusNotFound))); err != nil {
			log.WithError(err).Warn("failed to response with Error")
		}
	})
	s.mux.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		if _, err := w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed))); err != nil {
			log.WithError(err).Warn("failed to response with Error")
		}
	})
}

func (s *Server) Serve() {
	s.setupRoutes()

	server := http.Server{
		Addr:         ":" + s.cfg.ServerPort,
		Handler:      s.mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	go func(s *http.Server) {
		err := s.ListenAndServe()
		log.Error(err)
	}(&server)

	log.Info("http server started on port " + s.cfg.ServerPort)

	// Got gracefully shut down the http server
	<-s.ctx.Done()

	log.Printf("shut down the http server")
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() { cancel() }()
	if err := server.Shutdown(ctxShutDown); err != nil {
		log.WithError(err).Fatal("server shutdown failed")
	}
	log.Printf("server exited properly")
}
