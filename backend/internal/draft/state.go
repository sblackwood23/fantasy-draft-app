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
	StatusPaused     DraftStatus = "paused"
	StatusCompleted  DraftStatus = "completed"
)

// PickResult contains the details of a completed pick for persistence
type PickResult struct {
	EventID    int
	UserID     int
	PlayerID   int
	PickNumber int
	Round      int
	AutoDraft  bool
}

type DraftState struct {
	mu               sync.Mutex      // Protects concurrent access to state
	eventID          int             // ID of the event for which the draft is occurring
	currentTurnID    int             // ID of the user whose turn it currently is
	pickTimer        *time.Timer     // Stores the timer for a pick
	roundNumber      int             // The number of what round it is
	draftStatus      DraftStatus     // Status of the draft
	outgoing         chan []byte     // Outgoing messages from the draft state
	pickResults      chan PickResult // Channel for completed picks (for persistence)
	completed        chan struct{}   // Closed when draft completes (signals DraftService)
	pickOrder        []int           // Order of user IDs for drafting
	currentPickIndex int             // Current position in pickOrder
	timerDuration    time.Duration   // How long each user has to pick
	turnDeadline     time.Time       // When the current turn expires (for client countdown)
	remainingTime    time.Duration   // Time remaining when paused (for resume)
	totalRounds      int             // Total rounds in the draft (picks per team)
	availablePlayers []int           // Player IDs available to draft
}

func NewDraftState(eventID int) *DraftState {
	return &DraftState{
		eventID:     eventID,
		draftStatus: StatusNotStarted,
		outgoing:    make(chan []byte, 256),
		pickResults: make(chan PickResult, 256),
		completed:   make(chan struct{}),
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
	d.startTimer(d.timerDuration)

	// Emit draft started message
	msg, _ := json.Marshal(map[string]interface{}{
		"type":         MsgTypeDraftStarted,
		"eventID":      d.eventID,
		"currentTurn":  d.currentTurnID,
		"roundNumber":  d.roundNumber,
		"turnDeadline": d.turnDeadline.Unix(),
	})
	d.outgoing <- msg

	return nil
}

// startTimer starts the countdown for the current pick with the given duration
func (d *DraftState) startTimer(duration time.Duration) {
	// Stop existing timer if any
	if d.pickTimer != nil {
		d.pickTimer.Stop()
	}

	d.turnDeadline = time.Now().Add(duration)
	d.pickTimer = time.NewTimer(duration)

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

	// Create pick result (pick_number is 1-indexed)
	pickResult := PickResult{
		EventID:    d.eventID,
		UserID:     userID,
		PlayerID:   playerID,
		PickNumber: d.currentPickIndex + 1,
		Round:      d.roundNumber,
		AutoDraft:  true,
	}

	// Emit auto-draft pick message
	msg, _ := json.Marshal(map[string]interface{}{
		"type":       MsgTypePickMade,
		"userID":     userID,
		"playerID":   playerID,
		"pickNumber": pickResult.PickNumber,
		"round":      d.roundNumber,
		"autoDraft":  true,
	})
	d.outgoing <- msg

	// Emit pick result for persistence
	d.pickResults <- pickResult

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
	d.startTimer(d.timerDuration)

	// Emit turn changed message
	msg, _ := json.Marshal(map[string]interface{}{
		"type":         MsgTypeTurnChanged,
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
		"type":        MsgTypeDraftCompleted,
		"eventID":     d.eventID,
		"totalPicks":  d.currentPickIndex,
		"totalRounds": d.totalRounds,
	})
	d.outgoing <- msg

	// Signal completion to DraftService
	close(d.completed)
}

// MakePick processes a pick from a user
// Returns pick details for persistence, or error if invalid
func (d *DraftState) MakePick(userID, playerID int) (*PickResult, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.draftStatus != StatusInProgress {
		return nil, fmt.Errorf("draft is not in progress")
	}

	if userID != d.currentTurnID {
		return nil, fmt.Errorf("not your turn")
	}

	if !d.isPlayerAvailable(playerID) {
		return nil, fmt.Errorf("player not available")
	}

	// Stop the current timer (pick was made in time)
	if d.pickTimer != nil {
		d.pickTimer.Stop()
	}

	// Remove player from available list
	d.removePlayer(playerID)

	// Create pick result (pick_number is 1-indexed)
	pickResult := &PickResult{
		EventID:    d.eventID,
		UserID:     userID,
		PlayerID:   playerID,
		PickNumber: d.currentPickIndex + 1,
		Round:      d.roundNumber,
		AutoDraft:  false,
	}

	// Emit pick made message
	msg, _ := json.Marshal(map[string]interface{}{
		"type":       MsgTypePickMade,
		"userID":     userID,
		"playerID":   playerID,
		"pickNumber": pickResult.PickNumber,
		"round":      d.roundNumber,
		"autoDraft":  false,
	})
	d.outgoing <- msg

	// Emit pick result for persistence
	d.pickResults <- *pickResult

	// Move to next turn
	d.advanceTurn()

	return pickResult, nil
}

// PauseDraft pauses the draft, stopping the timer and saving remaining time
func (d *DraftState) PauseDraft() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.draftStatus != StatusInProgress {
		return fmt.Errorf("can only pause an in-progress draft")
	}

	// Calculate remaining time before stopping timer
	d.remainingTime = max(time.Until(d.turnDeadline), 0)

	// Stop the timer
	if d.pickTimer != nil {
		d.pickTimer.Stop()
	}

	d.draftStatus = StatusPaused

	// Emit draft paused message
	msg, _ := json.Marshal(map[string]interface{}{
		"type":          MsgTypeDraftPaused,
		"eventID":       d.eventID,
		"remainingTime": d.remainingTime.Seconds(),
	})
	d.outgoing <- msg

	return nil
}

// ResumeDraft resumes a paused draft, restarting the timer with remaining time
func (d *DraftState) ResumeDraft() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.draftStatus != StatusPaused {
		return fmt.Errorf("can only resume a paused draft")
	}

	d.draftStatus = StatusInProgress

	// Restart timer with remaining time
	d.startTimer(d.remainingTime)

	// Emit draft resumed message
	msg, _ := json.Marshal(map[string]interface{}{
		"type":         MsgTypeDraftResumed,
		"eventID":      d.eventID,
		"currentTurn":  d.currentTurnID,
		"turnDeadline": d.turnDeadline.Unix(),
	})
	d.outgoing <- msg

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

// PickResults returns the channel for reading completed picks (for persistence)
func (d *DraftState) PickResults() <-chan PickResult {
	return d.pickResults
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

// Completed returns a channel that is closed when the draft completes
func (d *DraftState) Completed() <-chan struct{} {
	return d.completed
}
