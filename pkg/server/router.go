package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Load handlers
	h := NewHandler()

	// Health check
	r.Get("/health", h.HealthCheckHandler)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/validate", func(r chi.Router) {
			r.Post("/tool", h.ValidateToolHandler)
			r.Post("/tools", h.ValidateToolsHandler)
		})
		r.Route("/tools", func(r chi.Router) {
			r.Route("/register", func(r chi.Router) {
				r.Post("/", h.ToolRegistrationHandler)
			})
			r.Route("/list", func(r chi.Router) {
				r.Get("/", h.ListToolsHandler)
			})
		})
	})

	return r
}
