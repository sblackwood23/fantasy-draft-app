package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sblackwood23/fantasy-draft-app/internal/database"
	"github.com/sblackwood23/fantasy-draft-app/internal/models"
	"github.com/sblackwood23/fantasy-draft-app/internal/repository"
)

func main() {
	ctx := context.Background()

	// Connect to database
	db, err := database.New(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Connected to database successfully")

	// Clear existing data
	if err := clearData(ctx, db); err != nil {
		log.Fatalf("Failed to clear data: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.Pool)
	playerRepo := repository.NewPlayerRepository(db.Pool)
	eventRepo := repository.NewEventRepository(db.Pool)

	// Seed users
	users := []models.User{
		{Username: "alice"},
		{Username: "bob"},
		{Username: "charlie"},
	}

	fmt.Println("\nSeeding users...")
	for i := range users {
		if err := userRepo.Create(ctx, &users[i]); err != nil {
			log.Fatalf("Failed to create user %s: %v", users[i].Username, err)
		}
		fmt.Printf("  ✓ Created user: %s\n", users[i].Username)
	}

	// Seed players
	players := []models.Player{
		{FirstName: "Tiger", LastName: "Woods", Status: "professional", CountryCode: "USA"},
		{FirstName: "Rory", LastName: "McIlroy", Status: "professional", CountryCode: "NIR"},
		{FirstName: "Jon", LastName: "Rahm", Status: "professional", CountryCode: "ESP"},
		{FirstName: "Scottie", LastName: "Scheffler", Status: "professional", CountryCode: "USA"},
		{FirstName: "Brooks", LastName: "Koepka", Status: "professional", CountryCode: "USA"},
		{FirstName: "Viktor", LastName: "Hovland", Status: "professional", CountryCode: "NOR"},
		{FirstName: "Xander", LastName: "Schauffele", Status: "professional", CountryCode: "USA"},
		{FirstName: "Patrick", LastName: "Cantlay", Status: "professional", CountryCode: "USA"},
		{FirstName: "Collin", LastName: "Morikawa", Status: "professional", CountryCode: "USA"},
		{FirstName: "Justin", LastName: "Thomas", Status: "professional", CountryCode: "USA"},
		{FirstName: "Jordan", LastName: "Spieth", Status: "professional", CountryCode: "USA"},
		{FirstName: "Hideki", LastName: "Matsuyama", Status: "professional", CountryCode: "JPN"},
		{FirstName: "Dustin", LastName: "Johnson", Status: "professional", CountryCode: "USA"},
		{FirstName: "Matt", LastName: "Fitzpatrick", Status: "professional", CountryCode: "ENG"},
		{FirstName: "Shane", LastName: "Lowry", Status: "professional", CountryCode: "IRL"},
	}

	fmt.Println("\nSeeding players...")
	for i := range players {
		if err := playerRepo.Create(ctx, &players[i]); err != nil {
			log.Fatalf("Failed to create player %s %s: %v", players[i].FirstName, players[i].LastName, err)
		}
		fmt.Printf("  ✓ Created player: %s %s\n", players[i].FirstName, players[i].LastName)
	}

	// Seed events
	events := []models.Event{
		{
			Name:              "2026 Masters Tournament Draft",
			MaxPicksPerTeam:   6,
			MaxTeamsPerPlayer: 2,
			Status:            "not_started",
			Stipulations:      models.Stipulations{"tournament": "Masters", "year": float64(2026)},
		},
	}

	fmt.Println("\nSeeding events...")
	for i := range events {
		if err := eventRepo.Create(ctx, &events[i]); err != nil {
			log.Fatalf("Failed to create event %s: %v", events[i].Name, err)
		}
		fmt.Printf("  ✓ Created event: %s\n", events[i].Name)
	}

	fmt.Printf("\nSeed completed successfully!\n")
	fmt.Printf("Summary: %d users, %d players, %d event\n", len(users), len(players), len(events))
}

func clearData(ctx context.Context, db *database.DB) error {
	fmt.Println("\nClearing existing data...")

	// Truncate tables in order (respecting foreign key constraints)
	// draft_results references both events and players, so clear it first
	tables := []string{"draft_results", "events", "players", "users"}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		if _, err := db.Pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
		fmt.Printf("  ✓ Cleared table: %s\n", table)
	}

	return nil
}
