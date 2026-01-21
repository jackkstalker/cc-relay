// Package proxy implements the HTTP proxy server for cc-relay.
package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/omarluq/cc-relay/internal/config"
	"github.com/omarluq/cc-relay/internal/providers"
)

func TestSetupRoutes_CreatesHandler(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Server: config.ServerConfig{
			Listen: "127.0.0.1:0",
			APIKey: "test-key",
		},
	}
	provider := providers.NewAnthropicProvider("test", "https://api.anthropic.com")

	handler, err := SetupRoutes(cfg, provider, "backend-key")
	if err != nil {
		t.Fatalf("SetupRoutes failed: %v", err)
	}

	if handler == nil {
		t.Fatal("handler is nil")
	}
}

func TestSetupRoutes_AuthMiddlewareApplied(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Server: config.ServerConfig{
			APIKey: "test-key",
		},
	}
	provider := providers.NewAnthropicProvider("test", "https://api.anthropic.com")

	handler, err := SetupRoutes(cfg, provider, "backend-key")
	if err != nil {
		t.Fatalf("SetupRoutes failed: %v", err)
	}

	// Request without API key should return 401
	req := httptest.NewRequest("POST", "/v1/messages", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestSetupRoutes_AuthMiddlewareWithValidKey(t *testing.T) {
	t.Parallel()

	// Create mock backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer backend.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{
			APIKey: "test-key",
		},
	}
	provider := providers.NewAnthropicProvider("test", backend.URL)

	handler, err := SetupRoutes(cfg, provider, "backend-key")
	if err != nil {
		t.Fatalf("SetupRoutes failed: %v", err)
	}

	// Request with valid API key should pass auth and reach backend
	req := httptest.NewRequest("POST", "/v1/messages", http.NoBody)
	req.Header.Set("x-api-key", "test-key")
	req.Header.Set("anthropic-version", "2023-06-01")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Errorf("expected auth to pass, got 401: %s", rec.Body.String())
	}
}

func TestSetupRoutes_NoAuthWhenAPIKeyEmpty(t *testing.T) {
	t.Parallel()

	// Create mock backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer backend.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{
			APIKey: "", // No auth configured
		},
	}
	provider := providers.NewAnthropicProvider("test", backend.URL)

	handler, err := SetupRoutes(cfg, provider, "backend-key")
	if err != nil {
		t.Fatalf("SetupRoutes failed: %v", err)
	}

	// Request without API key should NOT return 401 when auth is disabled
	req := httptest.NewRequest("POST", "/v1/messages", http.NoBody)
	req.Header.Set("anthropic-version", "2023-06-01")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Errorf("expected no auth when APIKey is empty, got 401: %s", rec.Body.String())
	}
}

func TestSetupRoutes_HealthEndpoint(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Server: config.ServerConfig{
			APIKey: "test-key", // Auth enabled
		},
	}
	provider := providers.NewAnthropicProvider("test", "https://api.anthropic.com")

	handler, err := SetupRoutes(cfg, provider, "backend-key")
	if err != nil {
		t.Fatalf("SetupRoutes failed: %v", err)
	}

	// Health endpoint should work without auth
	req := httptest.NewRequest("GET", "/health", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	expectedBody := `{"status":"ok"}`
	if rec.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rec.Body.String())
	}
}

func TestSetupRoutes_HealthEndpointWithAuth(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Server: config.ServerConfig{
			APIKey: "test-key",
		},
	}
	provider := providers.NewAnthropicProvider("test", "https://api.anthropic.com")

	handler, err := SetupRoutes(cfg, provider, "backend-key")
	if err != nil {
		t.Fatalf("SetupRoutes failed: %v", err)
	}

	// Health endpoint should work even when server has auth enabled
	// (health check should never require auth)
	req := httptest.NewRequest("GET", "/health", http.NoBody)
	// Intentionally NOT setting x-api-key header
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("health endpoint should not require auth, got status %d", rec.Code)
	}
}

func TestSetupRoutes_OnlyPOSTToMessages(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Server: config.ServerConfig{
			APIKey: "", // No auth for simpler test
		},
	}
	provider := providers.NewAnthropicProvider("test", "https://api.anthropic.com")

	handler, err := SetupRoutes(cfg, provider, "backend-key")
	if err != nil {
		t.Fatalf("SetupRoutes failed: %v", err)
	}

	// GET to /v1/messages should not be handled
	req := httptest.NewRequest("GET", "/v1/messages", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Should return 405 Method Not Allowed (Go 1.22+ router behavior)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET, got %d", rec.Code)
	}
}

func TestSetupRoutes_OnlyGETToHealth(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Server: config.ServerConfig{
			APIKey: "",
		},
	}
	provider := providers.NewAnthropicProvider("test", "https://api.anthropic.com")

	handler, err := SetupRoutes(cfg, provider, "backend-key")
	if err != nil {
		t.Fatalf("SetupRoutes failed: %v", err)
	}

	// POST to /health should not be handled
	req := httptest.NewRequest("POST", "/health", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Should return 405 Method Not Allowed
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST to /health, got %d", rec.Code)
	}
}

func TestSetupRoutes_InvalidProviderBaseURL(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Server: config.ServerConfig{
			APIKey: "test-key",
		},
	}

	// Create provider with invalid base URL
	provider := providers.NewAnthropicProvider("test", "://invalid-url")

	handler, err := SetupRoutes(cfg, provider, "backend-key")
	if err == nil {
		t.Fatal("expected error for invalid provider base URL, got nil")
	}

	if handler != nil {
		t.Errorf("expected nil handler on error, got %v", handler)
	}
}

func TestSetupRoutes_404ForUnknownPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Server: config.ServerConfig{
			APIKey: "",
		},
	}
	provider := providers.NewAnthropicProvider("test", "https://api.anthropic.com")

	handler, err := SetupRoutes(cfg, provider, "backend-key")
	if err != nil {
		t.Fatalf("SetupRoutes failed: %v", err)
	}

	// Unknown path should return 404
	req := httptest.NewRequest("GET", "/unknown", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown path, got %d", rec.Code)
	}
}

func TestSetupRoutes_MessagesPathMustBeExact(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Server: config.ServerConfig{
			APIKey: "",
		},
	}
	provider := providers.NewAnthropicProvider("test", "https://api.anthropic.com")

	handler, err := SetupRoutes(cfg, provider, "backend-key")
	if err != nil {
		t.Fatalf("SetupRoutes failed: %v", err)
	}

	// /v1/messages/extra should not match the route
	req := httptest.NewRequest("POST", "/v1/messages/extra", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-exact path, got %d", rec.Code)
	}
}
