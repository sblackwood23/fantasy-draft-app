package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sblackwood23/fantasy-draft-app/internal/database"
	"github.com/sblackwood23/fantasy-draft-app/internal/draft"
	"github.com/sblackwood23/fantasy-draft-app/internal/handlers"
)

// Dependencies contains all handlers and services needed for route registration
type Dependencies struct {
	Event       *handlers.EventHandler
	Player      *handlers.PlayerHandler
	User        *handlers.UserHandler
	EventPlayer *handlers.EventPlayerHandler
	DraftRoom   *handlers.DraftRoomHandler
	Draft       *draft.DraftService
}

func setupRoutes(r *chi.Mux, db *database.DB, deps *Dependencies) {
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
	r.Get("/ws/draft", deps.Draft.HandleWebSocket)

	// Events routes
	r.Get("/events/{id}", deps.Event.GetEvent)
	r.Get("/events", deps.Event.ListEvents)
	r.Post("/events", deps.Event.CreateEvent)
	r.Put("/events/{id}", deps.Event.UpdateEvent)
	r.Delete("/events/{id}", deps.Event.DeleteEvent)

	// Players routes
	r.Get("/players/{id}", deps.Player.GetPlayer)
	r.Get("/players", deps.Player.ListPlayers)
	r.Post("/players", deps.Player.CreatePlayer)
	r.Put("/players/{id}", deps.Player.UpdatePlayer)
	r.Delete("/players/{id}", deps.Player.DeletePlayer)

	// Users routes
	r.Get("/users/{id}", deps.User.GetUser)
	r.Get("/users", deps.User.ListUsers)
	r.Post("/users", deps.User.CreateUser)
	r.Put("/users/{id}", deps.User.UpdateUser)
	r.Delete("/users/{id}", deps.User.DeleteUser)

	// Event players routes
	r.Get("/events/{id}/players", deps.EventPlayer.GetEventPlayers)
	r.Post("/events/{id}/players", deps.EventPlayer.AddEventPlayers)
	r.Delete("/events/{id}/players/{playerID}", deps.EventPlayer.RemoveEventPlayer)

	// Draft room routes (HTTP)
	r.Post("/events/{id}/draft-room", deps.DraftRoom.CreateDraftRoom)
	r.Get("/events/{id}/draft-room", deps.DraftRoom.GetDraftRoom)
}
