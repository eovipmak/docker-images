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
	TenantID int                    `json:"tenant_id"`
}

// broadcastEvent sends an event to the backend SSE handler
func broadcastEvent(eventType string, data map[string]interface{}, tenantID int) {
	// Get backend URL from environment
	backendURL := os.Getenv("BACKEND_API_URL")
	if backendURL == "" {
		backendURL = "http://backend:8080"
	}

	// Prepare broadcast request
	req := BroadcastRequest{
		Type:     eventType,
		Data:     data,
		TenantID: tenantID,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("[Event] Failed to marshal broadcast request: %v", err)
		return
	}

	// Send POST request to backend internal endpoint
	url := fmt.Sprintf("%s/internal/broadcast", backendURL)
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[Event] Failed to broadcast event %s: %v", eventType, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Event] Broadcast event %s failed with status: %d", eventType, resp.StatusCode)
		return
	}

	log.Printf("[Event] Successfully broadcasted %s event for tenant %d", eventType, tenantID)
}
