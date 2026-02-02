package draft

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

// DraftService manages WebSocket connections and draft state
type DraftService struct {
	manager *Manager
	state   *DraftState
	mu      sync.RWMutex // protects state
}

// NewDraftService creates a new DraftService and starts the manager
func NewDraftService() *DraftService {
	s := &DraftService{
		manager: NewManager(),
	}
	go s.manager.Run()
	return s
}

// CreateRoom creates a new draft room for the given event with available players
func (s *DraftService) CreateRoom(eventID int, playerIDs []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = NewDraftState(eventID)
	s.state.SetAvailablePlayers(playerIDs)
	return nil
}

// GetRoom returns the current draft room state
func (s *DraftService) GetRoom() *DraftState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Client represents a WebSocket client connection
type Client struct {
	Conn *websocket.Conn
	Send chan []byte // Buffered channel for outgoing messages
}

// SendError sends an error message to this client
func (c *Client) SendError(message string) {
	errMsg, _ := json.Marshal(map[string]string{
		"type":  "error",
		"error": message,
	})
	c.Send <- errMsg
}

// HandleWebSocket upgrades HTTP connection to WebSocket and handles messages
func (s *DraftService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Allow all origins for development (file:// and localhost)
		// In production, restrict this to your frontend domain
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}

	log.Println("WebSocket connection established")

	// Create client
	client := &Client{
		Conn: conn,
		Send: make(chan []byte, 256), // Buffered channel
	}
	// Register client with the draft manager
	s.manager.Register(client)

	// Start write pump in separate goroutine
	go s.writePump(r.Context(), client)

	// Start read pump (blocks here until connection closes)
	s.readPump(r.Context(), client)
}

// readPump handles incoming messages from the client
func (s *DraftService) readPump(ctx context.Context, c *Client) {
	defer func() {
		s.manager.Unregister(c) // Unregister client
		c.Conn.Close(websocket.StatusNormalClosure, "connection closed")
		log.Println("Client disconnected")
	}()

	// Set read limit to 32KB
	c.Conn.SetReadLimit(32768)

	// Read loop - wait for messages from client
	for {
		_, data, err := c.Conn.Read(ctx)
		if err != nil {
			// Check if connection was closed normally
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				log.Println("Client disconnected normally")
				return
			}
			log.Printf("Read error: %v", err)
			return
		}

		log.Printf("Received message: %s", string(data))

		// Handle the message
		s.handleMessage(c, data)
	}
}

// writePump handles outgoing messages to the client
func (s *DraftService) writePump(ctx context.Context, c *Client) {
	// Write loop - wait for messages from Send channel
	for msg := range c.Send {
		// Write message to client
		if err := c.Conn.Write(ctx, websocket.MessageText, msg); err != nil {
			log.Printf("Write error: %v", err)
			return
		}
		log.Printf("Sent message: %s", string(msg))
	}
}

// handleMessage routes incoming messages to appropriate handlers
func (s *DraftService) handleMessage(c *Client, data []byte) {
	// Parse message to extract type
	var msg struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		c.SendError("invalid JSON format")
		return
	}

	// Route to appropriate handler based on message type
	switch msg.Type {
	case MsgTypeStartDraft:
		s.handleStartDraft(c, data)
	case MsgTypeMakePick:
		s.handleMakePick(c, data)
	default:
		c.SendError("unknown message type: " + msg.Type)
	}
}
