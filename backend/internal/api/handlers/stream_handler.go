package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	TenantID  int                    `json:"tenant_id"`
	Timestamp time.Time              `json:"timestamp"`
}

// BroadcastRequest represents a request to broadcast an event
type BroadcastRequest struct {
	Type     string                 `json:"type" binding:"required"`
	Data     map[string]interface{} `json:"data" binding:"required"`
	TenantID int                    `json:"tenant_id" binding:"required"`
}

// Client represents a connected SSE client
type Client struct {
	ID       string
	TenantID int
	Channel  chan Event
	Context  context.Context
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
	// Get tenant ID from context (set by middleware)
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable buffering for nginx

	// Create client
	clientID := fmt.Sprintf("%d-%d", tenantID, time.Now().UnixNano())
	client := &Client{
		ID:       clientID,
		TenantID: tenantID,
		Channel:  make(chan Event, 10),
		Context:  c.Request.Context(),
	}

	// Register client
	h.mu.Lock()
	h.clients[clientID] = client
	h.mu.Unlock()

	log.Printf("[SSE] Client %s connected (tenant: %d)", clientID, tenantID)

	// Remove client on disconnect
	defer func() {
		h.mu.Lock()
		delete(h.clients, clientID)
		close(client.Channel)
		h.mu.Unlock()
		log.Printf("[SSE] Client %s disconnected", clientID)
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
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event.Type, h.formatEventData(event))
			c.Writer.Flush()
		}
	}
}

// BroadcastEvent broadcasts an event to all connected clients for a specific tenant
func (h *StreamHandler) BroadcastEvent(eventType string, data map[string]interface{}, tenantID int) {
	event := Event{
		Type:      eventType,
		Data:      data,
		TenantID:  tenantID,
		Timestamp: time.Now(),
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, client := range h.clients {
		// Only send to clients of the same tenant
		if client.TenantID == tenantID {
			select {
			case client.Channel <- event:
				count++
			default:
				// Channel full, skip this client
				log.Printf("[SSE] Warning: Client %s channel full, dropping event", client.ID)
			}
		}
	}

	if count > 0 {
		log.Printf("[SSE] Broadcasted %s event to %d clients (tenant: %d)", eventType, count, tenantID)
	}
}

// HandleBroadcast handles HTTP POST requests to broadcast events
// POST /internal/broadcast (internal API, no auth required - should be internal network only)
func (h *StreamHandler) HandleBroadcast(c *gin.Context) {
	var req BroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.BroadcastEvent(req.Type, req.Data, req.TenantID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Event broadcasted",
		"clients": h.GetConnectedClientsForTenant(req.TenantID),
	})
}

// GetConnectedClients returns the number of connected clients
func (h *StreamHandler) GetConnectedClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetConnectedClientsForTenant returns the number of connected clients for a specific tenant
func (h *StreamHandler) GetConnectedClientsForTenant(tenantID int) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, client := range h.clients {
		if client.TenantID == tenantID {
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
		log.Printf("[SSE] Error marshaling event data: %v", err)
		return "{}"
	}
	return string(jsonData)
}
