package jobs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// BroadcastRequest represents a request to broadcast an event
type BroadcastRequest struct {
	Type     string                 `json:"type"`
	Data     map[string]interface{} `json:"data"`
	UserID   int                    `json:"user_id"`
}

// broadcastEvent sends an event to the backend SSE handler
func broadcastEvent(eventType string, data map[string]interface{}, userID int) {
	log.Printf("[Event] Broadcasting %s event for user %d", eventType, userID)

	// Get backend URL from environment
	backendURL := os.Getenv("BACKEND_API_URL")
	if backendURL == "" {
		backendURL = "http://backend:8080"
	}

	log.Printf("[Event] Using backend URL: %s", backendURL)

	// Prepare broadcast request
	req := BroadcastRequest{
		Type:     eventType,
		Data:     data,
		UserID:   userID,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("[Event] Failed to marshal broadcast request: %v", err)
		return
	}

	log.Printf("[Event] Broadcast payload: %s", string(jsonData))

	// Send POST request to backend internal endpoint
	url := fmt.Sprintf("%s/internal/broadcast", backendURL)
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	log.Printf("[Event] Sending POST to %s", url)

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[Event] Failed to broadcast event %s: %v", eventType, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("[Event] Broadcast response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Event] Broadcast event %s failed with status: %d", eventType, resp.StatusCode)
		return
	}

	log.Printf("[Event] Successfully broadcasted %s event for user %d", eventType, userID)
}
