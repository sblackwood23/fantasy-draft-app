package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sblackwood23/fantasy-draft-app/internal/draft"
	"github.com/sblackwood23/fantasy-draft-app/internal/models"
	"github.com/sblackwood23/fantasy-draft-app/internal/repository"
)

// DraftRoomHandler handles HTTP endpoints for draft room management
type DraftRoomHandler struct {
	eventPlayerRepo *repository.EventPlayerRepository
	eventRepo       *repository.EventRepository
	userRepo        *repository.UserRepository
	draftService    *draft.DraftService
}

// NewDraftRoomHandler creates a new DraftRoomHandler
func NewDraftRoomHandler(
	eventPlayerRepo *repository.EventPlayerRepository,
	eventRepo *repository.EventRepository,
	userRepo *repository.UserRepository,
	draftService *draft.DraftService,
) *DraftRoomHandler {
	return &DraftRoomHandler{
		eventPlayerRepo: eventPlayerRepo,
		eventRepo:       eventRepo,
		userRepo:        userRepo,
		draftService:    draftService,
	}
}

// CreateDraftRoom handles POST /events/{id}/draft-room
// Fetches available players from the database and creates a draft room
func (h *DraftRoomHandler) CreateDraftRoom(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid event ID"}`, http.StatusBadRequest)
		return
	}

	// Get available players for this event from the database
	playerIDs, err := h.eventPlayerRepo.GetPlayerIDsByEvent(r.Context(), eventID)
	if err != nil {
		http.Error(w, `{"error": "failed to get players"}`, http.StatusInternalServerError)
		return
	}

	if len(playerIDs) == 0 {
		http.Error(w, `{"error": "no players assigned to this event"}`, http.StatusBadRequest)
		return
	}

	// Delegate to draft handler to create the room
	if err := h.draftService.CreateRoom(eventID, playerIDs); err != nil {
		http.Error(w, `{"error": "failed to create draft room"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"status":           "draft room created",
		"eventID":          eventID,
		"availablePlayers": len(playerIDs),
	})
}

// GetDraftRoom handles GET /events/{id}/draft-room
// Returns the current draft room state
func (h *DraftRoomHandler) GetDraftRoom(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid event ID"}`, http.StatusBadRequest)
		return
	}

	room := h.draftService.GetRoom()
	if room == nil || room.GetEventID() != eventID {
		http.Error(w, `{"error": "no draft room for this event"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"eventID":     room.GetEventID(),
		"status":      room.GetStatus(),
		"roundNumber": room.GetRoundNumber(),
		"currentTurn": room.GetCurrentTurn(),
	})
}

// JoinEvent handles POST /events/join
// Validates passkey and registers/authenticates user for the draft
func (h *DraftRoomHandler) JoinEvent(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		TeamName string `json:"team_name"`
		Passkey  string `json:"passkey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if req.TeamName == "" {
		http.Error(w, `{"error": "team_name is required"}`, http.StatusBadRequest)
		return
	}

	if req.Passkey == "" {
		http.Error(w, `{"error": "passkey is required"}`, http.StatusBadRequest)
		return
	}

	// Look up event by passkey
	event, err := h.eventRepo.GetByPasskey(r.Context(), req.Passkey)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error": "invalid passkey"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Check if user already exists for this event
	existingUser, err := h.userRepo.GetByEventAndUsername(r.Context(), event.ID, req.TeamName)
	if err == nil {
		// User exists - return success (reconnection case)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(existingUser)
		return
	}

	if err != pgx.ErrNoRows {
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	// User doesn't exist - check if room has capacity
	const maxTeams = 12
	count, err := h.userRepo.CountByEvent(r.Context(), event.ID)
	if err != nil {
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	if count >= maxTeams {
		http.Error(w, `{"error": "draft room is full"}`, http.StatusConflict)
		return
	}

	// Create new user for this event
	newUser := &models.User{
		EventID:  event.ID,
		Username: req.TeamName,
	}
	if err := h.userRepo.Create(r.Context(), newUser); err != nil {
		http.Error(w, `{"error": "failed to register team"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}
