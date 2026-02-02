package draft

import (
	"encoding/json"
	"log"
	"time"
)

// Incoming message types (from client)
const (
	MsgTypeStartDraft = "start_draft"
	MsgTypeMakePick   = "make_pick"
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

	// Start the bridge goroutine to broadcast outgoing messages
	go s.startOutgoingBridge(state)

	log.Printf("Draft started for event %d", state.GetEventID())
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

	if err := state.MakePick(msg.UserID, msg.PlayerID); err != nil {
		c.SendError(err.Error())
		return
	}
}

// startOutgoingBridge reads from the draft state's outgoing channel and broadcasts to all clients
func (s *DraftService) startOutgoingBridge(state *DraftState) {
	for msg := range state.Outgoing() {
		s.manager.Broadcast(msg)
	}
}
