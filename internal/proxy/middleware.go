// Package proxy implements the HTTP proxy server for cc-relay.
package proxy

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"time"

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

			// Log request start
			zerolog.Ctx(r.Context()).Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Msg("request started")

			// Serve request
			next.ServeHTTP(wrapped, r)

			// Log request completion
			duration := time.Since(start)
			logger := zerolog.Ctx(r.Context()).With().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", wrapped.statusCode).
				Dur("duration_ms", duration).
				Logger()

			if wrapped.statusCode >= 500 {
				logger.Error().Msg("request failed")
			} else if wrapped.statusCode >= 400 {
				logger.Warn().Msg("request error")
			} else {
				logger.Info().Msg("request completed")
			}
		})
	}
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
