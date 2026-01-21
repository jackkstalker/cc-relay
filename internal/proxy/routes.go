// Package proxy implements the HTTP proxy server for cc-relay.
package proxy

import (
	"fmt"
	"net/http"

	"github.com/omarluq/cc-relay/internal/config"
	"github.com/omarluq/cc-relay/internal/providers"
)

// SetupRoutes creates the HTTP handler with all routes configured.
// Routes:
//   - POST /v1/messages - Proxy to backend provider (with auth if configured)
//   - GET /health - Health check endpoint (no auth required)
func SetupRoutes(cfg *config.Config, provider providers.Provider, providerKey string) (http.Handler, error) {
	mux := http.NewServeMux()

	// Create proxy handler
	handler, err := NewHandler(provider, providerKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create handler: %w", err)
	}

	// Apply middleware in order:
	// 1. RequestIDMiddleware (first - generates ID)
	// 2. LoggingMiddleware (second - logs with ID)
	// 3. AuthMiddleware (third - auth logs include ID)
	// 4. Handler
	var messagesHandler http.Handler = handler
	if cfg.Server.APIKey != "" {
		messagesHandler = AuthMiddleware(cfg.Server.APIKey)(messagesHandler)
	}

	messagesHandler = LoggingMiddleware()(messagesHandler)
	messagesHandler = RequestIDMiddleware()(messagesHandler)

	// Register routes
	mux.Handle("POST /v1/messages", messagesHandler)

	// Health check endpoint (no auth required)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck // Health check write error is non-critical
		w.Write([]byte(`{"status":"ok"}`))
	})

	return mux, nil
}
