package providers

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// DefaultZAIBaseURL is the default Z.AI API base URL.
	// Z.AI provides an Anthropic-compatible API endpoint.
	DefaultZAIBaseURL = "https://api.z.ai/api/anthropic"
)

// ZAIProvider implements the Provider interface for Z.AI's Anthropic-compatible API.
// Z.AI (Zhipu AI) offers GLM models through an API that is compatible with Anthropic's
// Messages API format, making it a drop-in replacement for cost optimization.
type ZAIProvider struct {
	name    string
	baseURL string
	models  []string
}

// NewZAIProvider creates a new Z.AI provider instance.
// If baseURL is empty, DefaultZAIBaseURL is used.
func NewZAIProvider(name, baseURL string) *ZAIProvider {
	return NewZAIProviderWithModels(name, baseURL, nil)
}

// NewZAIProviderWithModels creates a new Z.AI provider with configured models.
// If baseURL is empty, DefaultZAIBaseURL is used.
func NewZAIProviderWithModels(name, baseURL string, models []string) *ZAIProvider {
	if baseURL == "" {
		baseURL = DefaultZAIBaseURL
	}

	return &ZAIProvider{
		name:    name,
		baseURL: baseURL,
		models:  models,
	}
}

// Name returns the provider identifier.
func (p *ZAIProvider) Name() string {
	return p.name
}

// BaseURL returns the backend API base URL.
func (p *ZAIProvider) BaseURL() string {
	return p.baseURL
}

// Authenticate adds Z.AI authentication to the request.
// Z.AI uses the same x-api-key header as Anthropic for authentication.
func (p *ZAIProvider) Authenticate(req *http.Request, key string) error {
	req.Header.Set("x-api-key", key)

	// Log authentication (key is redacted for security)
	log.Ctx(req.Context()).Debug().
		Str("provider", p.name).
		Msg("added authentication header")

	return nil
}

// ForwardHeaders returns headers to forward to the backend.
// Copies all anthropic-* headers from the original request and adds Content-Type.
// Z.AI is Anthropic-compatible, so it accepts the same headers.
func (p *ZAIProvider) ForwardHeaders(originalHeaders http.Header) http.Header {
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

// SupportsStreaming indicates that Z.AI supports SSE streaming.
// Z.AI's Anthropic-compatible API supports the same streaming format.
func (p *ZAIProvider) SupportsStreaming() bool {
	return true
}

// Owner returns the owner identifier for Z.AI.
func (p *ZAIProvider) Owner() string {
	return "zhipu"
}

// ListModels returns the list of available models for this provider.
func (p *ZAIProvider) ListModels() []Model {
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
