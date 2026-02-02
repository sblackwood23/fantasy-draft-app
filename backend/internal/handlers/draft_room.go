package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sblackwood23/fantasy-draft-app/internal/draft"
	"github.com/sblackwood23/fantasy-draft-app/internal/repository"
)

// DraftRoomHandler handles HTTP endpoints for draft room management
type DraftRoomHandler struct {
	eventPlayerRepo *repository.EventPlayerRepository
	draftService    *draft.DraftService
}

// NewDraftRoomHandler creates a new DraftRoomHandler
func NewDraftRoomHandler(eventPlayerRepo *repository.EventPlayerRepository, draftService *draft.DraftService) *DraftRoomHandler {
	return &DraftRoomHandler{
		eventPlayerRepo: eventPlayerRepo,
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
