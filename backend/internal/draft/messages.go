package draft

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/sblackwood23/fantasy-draft-app/internal/models"
)

// Incoming message types (from client)
const (
	MsgTypeStartDraft = "start_draft"
	MsgTypeMakePick   = "make_pick"
	MsgTypePauseDraft = "pause_draft"
	MsgTypeResumeDraft = "resume_draft"
)

// Outgoing message types (to client)
const (
	MsgTypeDraftStarted   = "draft_started"
	MsgTypeDraftPaused    = "draft_paused"
	MsgTypeDraftResumed   = "draft_resumed"
	MsgTypeDraftCompleted = "draft_completed"
	MsgTypePickMade       = "pick_made"
	MsgTypeTurnChanged    = "turn_changed"
	MsgTypeError          = "error"
)

// StartDraftMessage represents the payload for starting a draft
// Note: availablePlayers comes from CreateRoom (HTTP), not this message
type StartDraftMessage struct {
	Type          string `json:"type"`
	PickOrder     []int  `json:"pickOrder"`
	TotalRounds   int    `json:"totalRounds"`
	TimerDuration int    `json:"timerDuration"` // in seconds
}

// MakePickMessage represents the payload for making a pick
type MakePickMessage struct {
	Type     string `json:"type"`
	UserID   int    `json:"userID"`
	PlayerID int    `json:"playerID"`
}

// handleStartDraft initializes and starts the draft
// Requires CreateRoom to have been called first (via HTTP endpoint)
func (s *DraftService) handleStartDraft(c *Client, data []byte) {
	var msg StartDraftMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		c.SendError("invalid start_draft message format")
		return
	}

	s.mu.Lock()
	state := s.state
	if state == nil {
		s.mu.Unlock()
		c.SendError("no draft room created - call CreateRoom first")
		return
	}

	// Start the draft using existing state (which has available players from CreateRoom)
	timerDuration := time.Duration(msg.TimerDuration) * time.Second
	availablePlayers := state.GetAvailablePlayers()
	if err := state.StartDraft(msg.PickOrder, msg.TotalRounds, timerDuration, availablePlayers); err != nil {
		s.mu.Unlock()
		c.SendError(err.Error())
		return
	}
	s.mu.Unlock()

	// Update event status to in_progress
	eventID := state.GetEventID()
	if err := s.eventUpdater.UpdateStatus(context.Background(), eventID, models.EventStatusInProgress); err != nil {
		log.Printf("Failed to update event status to in_progress: %v", err)
	}

	// Start the bridge goroutine to broadcast outgoing messages
	go s.startOutgoingBridge(state)

	// Start the persistence goroutine to save picks to database
	go s.startPickPersistence(state)

	// Start the completion handler to update event status when draft ends
	go s.startCompletionHandler(state)

	log.Printf("Draft started for event %d", eventID)
}

// handleMakePick processes a pick from a user
func (s *DraftService) handleMakePick(c *Client, data []byte) {
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	if state == nil {
		c.SendError("no draft in progress")
		return
	}

	var msg MakePickMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		c.SendError("invalid make_pick message format")
		return
	}

	if _, err := state.MakePick(msg.UserID, msg.PlayerID); err != nil {
		c.SendError(err.Error())
		return
	}
}

// handlePauseDraft pauses an in-progress draft
func (s *DraftService) handlePauseDraft(c *Client) {
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	if state == nil {
		c.SendError("no draft in progress")
		return
	}

	if err := state.PauseDraft(); err != nil {
		c.SendError(err.Error())
		return
	}

	log.Printf("Draft paused for event %d", state.GetEventID())
}

// handleResumeDraft resumes a paused draft
func (s *DraftService) handleResumeDraft(c *Client) {
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	if state == nil {
		c.SendError("no draft in progress")
		return
	}

	if err := state.ResumeDraft(); err != nil {
		c.SendError(err.Error())
		return
	}

	log.Printf("Draft resumed for event %d", state.GetEventID())
}

// startOutgoingBridge reads from the draft state's outgoing channel and broadcasts to all clients
func (s *DraftService) startOutgoingBridge(state *DraftState) {
	for msg := range state.Outgoing() {
		s.manager.Broadcast(msg)
	}
}

// startPickPersistence reads from the draft state's pick results channel and saves to database
func (s *DraftService) startPickPersistence(state *DraftState) {
	for pick := range state.PickResults() {
		ctx := context.Background()
		if err := s.pickSaver.SavePick(ctx, pick.EventID, pick.UserID, pick.PlayerID, pick.PickNumber, pick.Round); err != nil {
			log.Printf("Failed to persist pick: %v", err)
		} else {
			log.Printf("Persisted pick: event=%d user=%d player=%d pick#=%d round=%d auto=%v",
				pick.EventID, pick.UserID, pick.PlayerID, pick.PickNumber, pick.Round, pick.AutoDraft)
		}
	}
}

// startCompletionHandler waits for the draft to complete and updates event status
func (s *DraftService) startCompletionHandler(state *DraftState) {
	<-state.Completed()
	eventID := state.GetEventID()
	if err := s.eventUpdater.UpdateStatus(context.Background(), eventID, models.EventStatusCompleted); err != nil {
		log.Printf("Failed to update event status to completed: %v", err)
	} else {
		log.Printf("Event %d marked as completed", eventID)
	}
}
