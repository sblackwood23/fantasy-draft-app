package draft

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"sync"
	"time"
)

type DraftStatus string

const (
	StatusNotStarted DraftStatus = "not_started"
	StatusInProgress DraftStatus = "in_progress"
	StatusCompleted  DraftStatus = "completed"
)

type DraftState struct {
	mu               sync.Mutex    // Protects concurrent access to state
	eventID          int           // ID of the event for which the draft is occurring
	currentTurnID    int           // ID of the user whose turn it currently is
	pickTimer        *time.Timer   // Stores the timer for a pick
	roundNumber      int           // The number of what round it is
	draftStatus      DraftStatus   // Status of the draft
	outgoing         chan []byte   // Outgoing messages from the draft state
	pickOrder        []int         // Order of user IDs for drafting
	currentPickIndex int           // Current position in pickOrder
	timerDuration    time.Duration // How long each user has to pick
	turnDeadline     time.Time     // When the current turn expires (for client countdown)
	totalRounds      int           // Total rounds in the draft (picks per team)
	availablePlayers []int         // Player IDs available to draft
}

func NewDraftState(eventID int) *DraftState {
	return &DraftState{
		eventID:     eventID,
		draftStatus: StatusNotStarted,
		outgoing:    make(chan []byte, 256),
	}
}

// StartDraft initializes and starts the draft with the given pick order, total rounds, timer duration, and available players
func (d *DraftState) StartDraft(pickOrder []int, totalRounds int, timerDuration time.Duration, availablePlayers []int) error {
	if d.draftStatus != StatusNotStarted {
		return fmt.Errorf("draft already started")
	}
	if len(pickOrder) == 0 {
		return fmt.Errorf("pick order cannot be empty")
	}
	if len(availablePlayers) == 0 {
		return fmt.Errorf("available players cannot be empty")
	}

	d.pickOrder = pickOrder
	d.totalRounds = totalRounds
	d.timerDuration = timerDuration
	d.availablePlayers = availablePlayers
	d.currentPickIndex = 0
	d.currentTurnID = pickOrder[0]
	d.roundNumber = 1
	d.draftStatus = StatusInProgress

	// Start the pick timer (sets turnDeadline)
	d.startTimer()

	// Emit draft started message
	msg, _ := json.Marshal(map[string]interface{}{
		"type":         "draft_started",
		"eventID":      d.eventID,
		"currentTurn":  d.currentTurnID,
		"roundNumber":  d.roundNumber,
		"turnDeadline": d.turnDeadline.Unix(),
	})
	d.outgoing <- msg

	return nil
}

// startTimer starts the countdown for the current pick
func (d *DraftState) startTimer() {
	// Stop existing timer if any
	if d.pickTimer != nil {
		d.pickTimer.Stop()
	}

	d.turnDeadline = time.Now().Add(d.timerDuration)
	d.pickTimer = time.NewTimer(d.timerDuration)

	// Wait for timer in a goroutine
	go func() {
		<-d.pickTimer.C
		d.handleTimerExpired()
	}()
}

// handleTimerExpired is called when the pick timer runs out - triggers auto-draft
func (d *DraftState) handleTimerExpired() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.draftStatus != StatusInProgress {
		return
	}

	// Pick a random available player
	if len(d.availablePlayers) == 0 {
		return // No players left to draft
	}

	randomIndex := rand.Intn(len(d.availablePlayers))
	playerID := d.availablePlayers[randomIndex]
	userID := d.currentTurnID

	// Remove player from available list
	d.removePlayer(playerID)

	// Emit auto-draft pick message
	msg, _ := json.Marshal(map[string]interface{}{
		"type":      "pick_made",
		"userID":    userID,
		"playerID":  playerID,
		"round":     d.roundNumber,
		"autoDraft": true,
	})
	d.outgoing <- msg

	// Move to next turn
	d.advanceTurn()
}

// advanceTurn moves to the next player in the pick order
// Uses snake draft: 1→2→3→4→4→3→2→1→1→2→3→4...
func (d *DraftState) advanceTurn() {
	d.currentPickIndex++
	numPlayers := len(d.pickOrder)
	totalPicks := numPlayers * d.totalRounds

	// Check if draft is complete
	if d.currentPickIndex >= totalPicks {
		d.completeDraft()
		return
	}

	// Check if we've completed a round
	if d.currentPickIndex >= numPlayers*d.roundNumber {
		d.roundNumber++
	}

	// Snake draft logic: odd rounds go forward, even rounds go backward
	var positionInRound int
	if d.roundNumber%2 == 1 {
		// Odd round: forward (0, 1, 2, 3)
		positionInRound = d.currentPickIndex % numPlayers
	} else {
		// Even round: backward (3, 2, 1, 0)
		positionInRound = numPlayers - 1 - (d.currentPickIndex % numPlayers)
	}

	d.currentTurnID = d.pickOrder[positionInRound]

	// Restart timer for next pick
	d.startTimer()

	// Emit turn changed message
	msg, _ := json.Marshal(map[string]interface{}{
		"type":         "turn_changed",
		"currentTurn":  d.currentTurnID,
		"roundNumber":  d.roundNumber,
		"turnDeadline": d.turnDeadline.Unix(),
	})
	d.outgoing <- msg
}

// completeDraft finalizes the draft when all picks are made
func (d *DraftState) completeDraft() {
	d.draftStatus = StatusCompleted

	// Stop any running timer
	if d.pickTimer != nil {
		d.pickTimer.Stop()
	}

	// Emit draft completed message
	msg, _ := json.Marshal(map[string]interface{}{
		"type":        "draft_completed",
		"eventID":     d.eventID,
		"totalPicks":  d.currentPickIndex,
		"totalRounds": d.totalRounds,
	})
	d.outgoing <- msg
}

// MakePick processes a pick from a user
// Returns error if it's not their turn or draft isn't in progress
func (d *DraftState) MakePick(userID, playerID int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.draftStatus != StatusInProgress {
		return fmt.Errorf("draft is not in progress")
	}

	if userID != d.currentTurnID {
		return fmt.Errorf("not your turn")
	}

	if !d.isPlayerAvailable(playerID) {
		return fmt.Errorf("player not available")
	}

	// Stop the current timer (pick was made in time)
	if d.pickTimer != nil {
		d.pickTimer.Stop()
	}

	// Remove player from available list
	d.removePlayer(playerID)

	// Emit pick made message
	msg, _ := json.Marshal(map[string]interface{}{
		"type":      "pick_made",
		"userID":    userID,
		"playerID":  playerID,
		"round":     d.roundNumber,
		"autoDraft": false,
	})
	d.outgoing <- msg

	// Move to next turn
	d.advanceTurn()

	return nil
}

// isPlayerAvailable checks if a player is still available to draft
func (d *DraftState) isPlayerAvailable(playerID int) bool {
	return slices.Contains(d.availablePlayers, playerID)
}

// removePlayer removes a player from the available list
func (d *DraftState) removePlayer(playerID int) {
	d.availablePlayers = slices.DeleteFunc(d.availablePlayers, func(id int) bool {
		return id == playerID
	})
}

// Outgoing returns the channel for reading outgoing messages
func (d *DraftState) Outgoing() <-chan []byte {
	return d.outgoing
}

// GetCurrentTurn returns the user ID of the current turn
func (d *DraftState) GetCurrentTurn() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.currentTurnID
}

// GetStatus returns the current draft status
func (d *DraftState) GetStatus() DraftStatus {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.draftStatus
}

// GetRoundNumber returns the current round number
func (d *DraftState) GetRoundNumber() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.roundNumber
}

// GetEventID returns the event ID for this draft
func (d *DraftState) GetEventID() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.eventID
}

// SetAvailablePlayers sets the available players for the draft
func (d *DraftState) SetAvailablePlayers(playerIDs []int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.availablePlayers = playerIDs
}

// GetAvailablePlayers returns the available players for the draft
func (d *DraftState) GetAvailablePlayers() []int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.availablePlayers
}
