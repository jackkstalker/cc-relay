// Package providers defines the interface for LLM backend providers.
package providers

import "net/http"

// Provider defines the interface for LLM backend providers.
// All provider implementations must implement this interface to be compatible with cc-relay.
type Provider interface {
	// Name returns the provider identifier.
	Name() string

	// BaseURL returns the backend API base URL.
	BaseURL() string

	// Authenticate adds provider-specific authentication to the request.
	// The key parameter is the API key to use for authentication.
	Authenticate(req *http.Request, key string) error

	// ForwardHeaders returns headers to add when forwarding the request.
	// This includes provider-specific headers and any anthropic-* headers from the original request.
	ForwardHeaders(originalHeaders http.Header) http.Header

	// SupportsStreaming indicates if the provider supports SSE streaming.
	SupportsStreaming() bool
}
