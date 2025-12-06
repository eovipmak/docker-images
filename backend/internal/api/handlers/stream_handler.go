package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Event represents a Server-Sent Event
type Event struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	UserID    int                    `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
}

// BroadcastRequest represents a request to broadcast an event
type BroadcastRequest struct {
	Type   string                 `json:"type" binding:"required"`
	Data   map[string]interface{} `json:"data" binding:"required"`
	UserID int                    `json:"user_id" binding:"required"`
}

// Client represents a connected SSE client
type Client struct {
	ID      string
	UserID  int
	Channel chan Event
	Context context.Context
}

// StreamHandler handles Server-Sent Events
type StreamHandler struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler() *StreamHandler {
	return &StreamHandler{
		clients: make(map[string]*Client),
	}
}

// HandleSSE handles Server-Sent Events endpoint
// GET /api/stream/events
func (h *StreamHandler) HandleSSE(c *gin.Context) {
	// Get user ID from context (set by middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user context not found"})
		return
	}
	userID := userIDValue.(int)

	// Set CORS headers for SSE
	// When credentials are included, we must specify the exact origin, not wildcard
	origin := c.GetHeader("Origin")
	if origin == "" {
		origin = "http://localhost:3000" // Default for development
	}
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Access-Control-Allow-Credentials", "true")

	// Handle preflight OPTIONS request
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable buffering for nginx

	// Create client
	clientID := fmt.Sprintf("%d-%d", userID, time.Now().UnixNano())
	client := &Client{
		ID:      clientID,
		UserID:  userID,
		Channel: make(chan Event, 10),
		Context: c.Request.Context(),
	}

	// Register client
	h.mu.Lock()
	h.clients[clientID] = client
	h.mu.Unlock()

	log.Printf("Client %s connected (user: %d)", clientID, userID)

	// Remove client on disconnect
	defer func() {
		h.mu.Lock()
		delete(h.clients, clientID)
		close(client.Channel)
		h.mu.Unlock()
		log.Printf("Client %s disconnected", clientID)
	}()

	// Send initial connection event
	fmt.Fprintf(c.Writer, "event: connected\ndata: {\"message\": \"Connected to event stream\"}\n\n")
	c.Writer.Flush()

	// Stream events
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-client.Context.Done():
			// Client disconnected
			return
		case <-ticker.C:
			// Send heartbeat to keep connection alive
			fmt.Fprintf(c.Writer, ": heartbeat\n\n")
			c.Writer.Flush()
		case event := <-client.Channel:
			// Send event to client
			log.Printf("[SSE] Sending %s event to client %s", event.Type, clientID)
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event.Type, h.formatEventData(event))
			c.Writer.Flush()
		}
	}
}

// BroadcastEvent broadcasts an event to all connected clients for a specific user
func (h *StreamHandler) BroadcastEvent(eventType string, data map[string]interface{}, userID int) {
	event := Event{
		Type:      eventType,
		Data:      data,
		UserID:    userID,
		Timestamp: time.Now(),
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, client := range h.clients {
		// Only send to clients of the same user
		if client.UserID == userID {
			select {
			case client.Channel <- event:
				count++
				log.Printf("Sent %s event to client %s (user: %d)", eventType, client.ID, userID)
			default:
				// Channel full, skip this client
				log.Printf("Warning: Client %s channel full, dropping event %s (user: %d)", client.ID, eventType, userID)
			}
		}
	}

	log.Printf("Broadcasted %s event to %d/%d clients (user: %d)", eventType, count, len(h.clients), userID)
}

// HandleBroadcast handles HTTP POST requests to broadcast events
// POST /internal/broadcast (internal API, no auth required - should be internal network only)
func (h *StreamHandler) HandleBroadcast(c *gin.Context) {
	var req BroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Broadcast request failed: invalid JSON - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Received broadcast request: type=%s, user=%d", req.Type, req.UserID)

	h.BroadcastEvent(req.Type, req.Data, req.UserID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Event broadcasted",
		"clients": h.GetConnectedClientsForUser(req.UserID),
	})
}

// GetConnectedClients returns the number of connected clients
func (h *StreamHandler) GetConnectedClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetConnectedClientsForUser returns the number of connected clients for a specific user
func (h *StreamHandler) GetConnectedClientsForUser(userID int) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, client := range h.clients {
		if client.UserID == userID {
			count++
		}
	}
	return count
}

// formatEventData formats event data as JSON string
func (h *StreamHandler) formatEventData(event Event) string {
	jsonData, err := json.Marshal(map[string]interface{}{
		"type":      event.Type,
		"data":      event.Data,
		"timestamp": event.Timestamp.Format(time.RFC3339),
	})
	if err != nil {
		log.Printf("Error marshaling event data: %v", err)
		return "{}"
	}
	return string(jsonData)
}
