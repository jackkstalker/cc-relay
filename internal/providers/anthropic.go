package providers

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// DefaultAnthropicBaseURL is the default Anthropic API base URL.
	DefaultAnthropicBaseURL = "https://api.anthropic.com"
)

// AnthropicProvider implements the Provider interface for Anthropic's API.
type AnthropicProvider struct {
	name    string
	baseURL string
	models  []string
}

// NewAnthropicProvider creates a new Anthropic provider instance.
// If baseURL is empty, DefaultAnthropicBaseURL is used.
func NewAnthropicProvider(name, baseURL string) *AnthropicProvider {
	return NewAnthropicProviderWithModels(name, baseURL, nil)
}

// NewAnthropicProviderWithModels creates a new Anthropic provider with configured models.
// If baseURL is empty, DefaultAnthropicBaseURL is used.
func NewAnthropicProviderWithModels(name, baseURL string, models []string) *AnthropicProvider {
	if baseURL == "" {
		baseURL = DefaultAnthropicBaseURL
	}

	return &AnthropicProvider{
		name:    name,
		baseURL: baseURL,
		models:  models,
	}
}

// Name returns the provider identifier.
func (p *AnthropicProvider) Name() string {
	return p.name
}

// BaseURL returns the backend API base URL.
func (p *AnthropicProvider) BaseURL() string {
	return p.baseURL
}

// Authenticate adds Anthropic-specific authentication to the request.
// Sets the x-api-key header with the provided API key.
func (p *AnthropicProvider) Authenticate(req *http.Request, key string) error {
	req.Header.Set("x-api-key", key)

	// Log authentication (key is redacted for security)
	log.Ctx(req.Context()).Debug().
		Str("provider", p.name).
		Msg("added authentication header")

	return nil
}

// ForwardHeaders returns headers to forward to the backend.
// Copies all anthropic-* headers from the original request and adds Content-Type.
func (p *AnthropicProvider) ForwardHeaders(originalHeaders http.Header) http.Header {
	headers := make(http.Header)

	// Copy all anthropic-* headers from the original request
	for key, values := range originalHeaders {
		// Check if key starts with "anthropic-" (case-insensitive)
		// http.Header stores keys in canonical form (Title-Case)
		canonicalKey := http.CanonicalHeaderKey(key)
		if len(canonicalKey) >= 10 && canonicalKey[:10] == "Anthropic-" {
			headers[canonicalKey] = append(headers[canonicalKey], values...)
		}
	}

	// Always set Content-Type for JSON requests
	headers.Set("Content-Type", "application/json")

	return headers
}

// SupportsStreaming indicates that Anthropic supports SSE streaming.
func (p *AnthropicProvider) SupportsStreaming() bool {
	return true
}

// Owner returns the owner identifier for Anthropic.
func (p *AnthropicProvider) Owner() string {
	return "anthropic"
}

// ListModels returns the list of available models for this provider.
func (p *AnthropicProvider) ListModels() []Model {
	if len(p.models) == 0 {
		return []Model{}
	}

	result := make([]Model, len(p.models))
	now := time.Now().Unix()

	for i, modelID := range p.models {
		result[i] = Model{
			ID:       modelID,
			Object:   "model",
			Created:  now,
			OwnedBy:  p.Owner(),
			Provider: p.name,
		}
	}

	return result
}
