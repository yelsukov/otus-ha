package entities

import "github.com/go-chi/chi"

type Server interface {
	Serve()
	MountRoutes(pattern string, router *chi.Mux)
}
