package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sblackwood23/fantasy-draft-app/internal/database"
	"github.com/sblackwood23/fantasy-draft-app/internal/draft"
	"github.com/sblackwood23/fantasy-draft-app/internal/handlers"
	"github.com/sblackwood23/fantasy-draft-app/internal/repository"
)

func main() {
	ctx := context.Background()

	// Initialize database connection
	db, err := database.New(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Database connected successfully")

	// Initialize repositories
	eventRepo := repository.NewEventRepository(db.Pool)
	playerRepo := repository.NewPlayerRepository(db.Pool)
	userRepo := repository.NewUserRepository(db.Pool)
	eventPlayerRepo := repository.NewEventPlayerRepository(db.Pool)
	draftResultRepo := repository.NewDraftResultRepository(db.Pool)

	// Initialize services
	draftService := draft.NewDraftService(draftResultRepo, eventRepo)

	// Initialize dependencies
	deps := &Dependencies{
		Event:       handlers.NewEventHandler(eventRepo),
		Player:      handlers.NewPlayerHandler(playerRepo),
		User:        handlers.NewUserHandler(userRepo),
		EventPlayer: handlers.NewEventPlayerHandler(eventPlayerRepo),
		DraftRoom:   handlers.NewDraftRoomHandler(eventPlayerRepo, eventRepo, userRepo, draftService),
		Draft:       draftService,
	}

	r := chi.NewRouter()

	// Setup all routes
	setupRoutes(r, db, deps)

	// Start server
	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)

	// Graceful shutdown handling
	server := &http.Server{
		Addr:    port,
		Handler: r,
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	fmt.Println("\nShutting down server...")

	// Graceful shutdown with 5 second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	fmt.Println("Server stopped gracefully")
}

// healthCheckHandler returns server and database health status
func healthCheckHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.Pool.Ping(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status": "error", "message": "database unavailable"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "database": "connected"}`))
	}
}
