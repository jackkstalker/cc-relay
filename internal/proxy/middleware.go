// Package proxy implements the HTTP proxy server for cc-relay.
package proxy

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/omarluq/cc-relay/internal/auth"
	"github.com/omarluq/cc-relay/internal/config"
	"github.com/rs/zerolog"
)

// AuthMiddleware creates middleware that validates x-api-key header.
// Uses constant-time comparison to prevent timing attacks.
func AuthMiddleware(expectedAPIKey string) func(http.Handler) http.Handler {
	// Pre-hash expected key at creation time (not per-request)
	expectedHash := sha256.Sum256([]byte(expectedAPIKey))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			providedKey := r.Header.Get("x-api-key")

			if providedKey == "" {
				zerolog.Ctx(r.Context()).Warn().Msg("authentication failed: missing x-api-key header")
				WriteError(w, http.StatusUnauthorized, "authentication_error", "missing x-api-key header")

				return
			}

			providedHash := sha256.Sum256([]byte(providedKey))

			// CRITICAL: Constant-time comparison prevents timing attacks
			if subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) != 1 {
				zerolog.Ctx(r.Context()).Warn().Msg("authentication failed: invalid x-api-key")
				WriteError(w, http.StatusUnauthorized, "authentication_error", "invalid x-api-key")

				return
			}

			zerolog.Ctx(r.Context()).Debug().Msg("authentication succeeded")
			next.ServeHTTP(w, r)
		})
	}
}

// MultiAuthMiddleware creates middleware supporting multiple authentication methods.
// Supports both x-api-key and Authorization: Bearer token authentication.
// If authConfig has no methods enabled, all requests pass through.
func MultiAuthMiddleware(authConfig *config.AuthConfig) func(http.Handler) http.Handler {
	// Build the authenticator chain based on config
	var authenticators []auth.Authenticator

	// Bearer token auth (checked first as it's more specific)
	if authConfig.AllowBearer {
		authenticators = append(authenticators, auth.NewBearerAuthenticator(authConfig.BearerSecret))
	}

	// API key auth
	if authConfig.APIKey != "" {
		authenticators = append(authenticators, auth.NewAPIKeyAuthenticator(authConfig.APIKey))
	}

	// If no auth configured, return pass-through middleware
	if len(authenticators) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	chainAuth := auth.NewChainAuthenticator(authenticators...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			result := chainAuth.Validate(r)

			if !result.Valid {
				zerolog.Ctx(r.Context()).Warn().
					Str("auth_type", string(result.Type)).
					Str("error", result.Error).
					Msg("authentication failed")
				WriteError(w, http.StatusUnauthorized, "authentication_error", result.Error)

				return
			}

			zerolog.Ctx(r.Context()).Debug().
				Str("auth_type", string(result.Type)).
				Msg("authentication succeeded")
			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDMiddleware adds X-Request-ID header and logger with request ID to context.
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract or generate request ID
			requestID := r.Header.Get("X-Request-ID")
			ctx := AddRequestID(r.Context(), requestID)

			// Write request ID to response header
			if requestID == "" {
				requestID = GetRequestID(ctx)
			}

			w.Header().Set("X-Request-ID", requestID)

			// Attach logger to request
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs each request with method, path, and duration.
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap ResponseWriter to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Get request ID for logging
			requestID := GetRequestID(r.Context())
			shortID := requestID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}

			// Log request start with arrow and request ID
			zerolog.Ctx(r.Context()).Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("req_id", shortID).
				Msgf("%s %s", r.Method, r.URL.Path)

			// Serve request
			next.ServeHTTP(wrapped, r)

			// Log request completion with detailed info
			duration := time.Since(start)
			durationStr := formatDuration(duration)

			logger := zerolog.Ctx(r.Context()).With().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", wrapped.statusCode).
				Str("duration", durationStr).
				Str("req_id", shortID).
				Logger()

			// Format completion message based on status
			var statusMsg string
			if wrapped.statusCode >= 500 {
				statusMsg = "✗"
			} else if wrapped.statusCode >= 400 {
				statusMsg = "⚠"
			} else {
				statusMsg = "✓"
			}

			completionMsg := formatCompletionMessage(wrapped.statusCode, statusMsg, durationStr)

			if wrapped.statusCode >= 500 {
				logger.Error().Msg(completionMsg)
			} else if wrapped.statusCode >= 400 {
				logger.Warn().Msg(completionMsg)
			} else {
				logger.Info().Msg(completionMsg)
			}
		})
	}
}

// formatDuration formats duration in human-readable format.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return d.String()
	}
	return d.Round(time.Millisecond).String()
}

// formatCompletionMessage formats the completion message with status.
func formatCompletionMessage(status int, symbol, duration string) string {
	return symbol + " " + http.StatusText(status) + " (" + duration + ")"
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
