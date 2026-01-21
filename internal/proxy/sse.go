// Package proxy implements the HTTP proxy server for cc-relay.
package proxy

import (
	"encoding/json"
	"net/http"
)

// IsStreamingRequest checks if request body contains "stream": true.
// Returns false if the body is invalid JSON or stream field is missing/false.
func IsStreamingRequest(body []byte) bool {
	// Parse as map to check stream field
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return false
	}

	stream, ok := req["stream"].(bool)

	return ok && stream
}

// - Connection: keep-alive - maintain streaming connection.
func SetSSEHeaders(h http.Header) {
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache, no-transform")
	h.Set("X-Accel-Buffering", "no")
	h.Set("Connection", "keep-alive")
}
