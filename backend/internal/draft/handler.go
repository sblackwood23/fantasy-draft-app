package draft

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/coder/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	Conn *websocket.Conn
	Send chan []byte // Buffered channel for outgoing messages
}

// HandleWebSocket upgrades HTTP connection to WebSocket and handles messages
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
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

	// Start write pump in separate goroutine
	go client.writePump(r.Context())

	// Start read pump (blocks here until connection closes)
	client.readPump(r.Context())
}

// readPump handles incoming messages from the client
func (c *Client) readPump(ctx context.Context) {
	defer func() {
		close(c.Send) // Close send channel when done
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
		c.handleMessage(data)
	}
}

// writePump handles outgoing messages to the client
func (c *Client) writePump(ctx context.Context) {
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

// handleMessage processes incoming messages (currently just echoes back)
func (c *Client) handleMessage(data []byte) {
	// Parse message as JSON to validate format
	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		// Send error back to client
		errMsg := map[string]string{
			"type":  "error",
			"error": "invalid JSON format",
		}
		errData, _ := json.Marshal(errMsg)
		c.Send <- errData // Push to send channel (non-blocking)
		return
	}

	// Echo the message back (for now)
	response := map[string]interface{}{
		"type":    "echo",
		"message": "Message received",
		"data":    msg,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	// Push response to send channel
	c.Send <- responseData
}
