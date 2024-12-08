package server

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"outreach-generator/internal/handlers"
)

type Server struct {
	db       *sql.DB
	handlers *handlers.Handlers
}

func New(db *sql.DB) *Server {
	h := handlers.New(db)
	return &Server{
		db:       db,
		handlers: h,
	}
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Static files
	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Pages
	r.Get("/", s.handlers.HandleHome())
	r.Get("/config", s.handlers.HandleConfig())

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/companies", s.handlers.HandleGetCompanies())
		r.Post("/generate-outreach", s.handlers.HandleGenerateOutreach())
		r.Post("/generate-all", s.handlers.HandleGenerateAll())
		r.Get("/config", s.handlers.HandleGetConfig())
		r.Post("/config", s.handlers.HandleSaveConfig())
	})

	return r
}
