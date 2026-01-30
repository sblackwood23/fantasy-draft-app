package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sblackwood23/fantasy-draft-app/internal/database"
	"github.com/sblackwood23/fantasy-draft-app/internal/draft"
	"github.com/sblackwood23/fantasy-draft-app/internal/handlers"
)

func setupRoutes(r *chi.Mux, db *database.DB, eventHandler *handlers.EventHandler, playerHandler *handlers.PlayerHandler, userHandler *handlers.UserHandler) {
	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", healthCheckHandler(db))

	// WebSocket route for draft
	r.Get("/ws/draft", draft.HandleWebSocket)

	// Events routes
	r.Get("/events/{id}", eventHandler.GetEvent)
	r.Get("/events", eventHandler.ListEvents)
	r.Post("/events", eventHandler.CreateEvent)
	r.Put("/events/{id}", eventHandler.UpdateEvent)
	r.Delete("/events/{id}", eventHandler.DeleteEvent)

	// Players routes
	r.Get("/players/{id}", playerHandler.GetPlayer)
	r.Get("/players", playerHandler.ListPlayers)
	r.Post("/players", playerHandler.CreatePlayer)
	r.Put("/players/{id}", playerHandler.UpdatePlayer)
	r.Delete("/players/{id}", playerHandler.DeletePlayer)

	// Users routes
	r.Get("/users/{id}", userHandler.GetUser)
	r.Get("/users", userHandler.ListUsers)
	r.Post("/users", userHandler.CreateUser)
	r.Put("/users/{id}", userHandler.UpdateUser)
	r.Delete("/users/{id}", userHandler.DeleteUser)
}
