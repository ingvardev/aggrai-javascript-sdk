// Package middleware contains HTTP middleware for the API.
package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

type contextKey string

const (
	// TenantContextKey is the context key for the authenticated tenant.
	TenantContextKey contextKey = "tenant"
	// AuthContextKey is the context key for the full auth context.
	AuthContextKey contextKey = "auth_context"
	// TenantIDContextKey is the context key for the tenant ID.
	TenantIDContextKey contextKey = "tenant_id"
	// APIUserIDContextKey is the context key for the API user ID.
	APIUserIDContextKey contextKey = "api_user_id"
	// SessionTokenContextKey is the context key for the session token (owner auth).
	SessionTokenContextKey contextKey = "session_token"
)

// TenantFromContext retrieves the tenant from the request context.
func TenantFromContext(ctx context.Context) *domain.Tenant {
	tenant, ok := ctx.Value(TenantContextKey).(*domain.Tenant)
	if !ok {
		return nil
	}
	return tenant
}

// AuthContextFromContext retrieves the auth context from the request context.
func AuthContextFromContext(ctx context.Context) *domain.AuthContext {
	authCtx, ok := ctx.Value(AuthContextKey).(*domain.AuthContext)
	if !ok {
		return nil
	}
	return authCtx
}

// TenantIDFromContext retrieves the tenant ID from the request context.
func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(TenantIDContextKey).(uuid.UUID)
	return id, ok
}

// APIUserIDFromContext retrieves the API user ID from the request context (may be nil).
func APIUserIDFromContext(ctx context.Context) *uuid.UUID {
	id, ok := ctx.Value(APIUserIDContextKey).(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}

// AuthMiddleware creates authentication middleware.
func AuthMiddleware(authService *usecases.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for certain paths
			if r.URL.Path == "/health" || r.URL.Path == "/healthz" || r.URL.Path == "/readyz" ||
				r.URL.Path == "/playground" || r.URL.Path == "/" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip authentication for WebSocket upgrade requests
			// Auth will be handled via connectionParams in the WebSocket InitFunc
			if r.Header.Get("Upgrade") == "websocket" {
				next.ServeHTTP(w, r)
				return
			}

			// Check for session token first (owner auth for dashboard)
			sessionToken := extractSessionToken(r)
			if sessionToken != "" {
				// Session-based auth - pass through, resolver will validate
				ctx := context.WithValue(r.Context(), SessionTokenContextKey, sessionToken)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			apiKey := extractAPIKey(r)
			if apiKey == "" {
				// Allow GraphQL requests without auth for public operations (login, register)
				if r.URL.Path == "/graphql" || r.URL.Path == "/query" {
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, "Unauthorized: missing API key", http.StatusUnauthorized)
				return
			}

			// Build auth request with client context
			authReq := &usecases.AuthenticateRequest{
				RawKey:    apiKey,
				ClientIP:  extractClientIP(r),
				UserAgent: r.UserAgent(),
			}

			result, err := authService.AuthenticateWithContext(r.Context(), authReq)
			if err != nil {
				if err == usecases.ErrRateLimited {
					http.Error(w, "Too many requests", http.StatusTooManyRequests)
					return
				}
				http.Error(w, "Authentication error", http.StatusInternalServerError)
				return
			}

			if !result.Authorized {
				http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
				return
			}

			// Add tenant to context (backward compatibility)
			ctx := context.WithValue(r.Context(), TenantContextKey, result.Tenant)

			// Add auth context with tenant ID, API user ID, and scopes
			if result.AuthCtx != nil {
				ctx = context.WithValue(ctx, AuthContextKey, result.AuthCtx)
				ctx = context.WithValue(ctx, TenantIDContextKey, result.AuthCtx.TenantID)
				if result.AuthCtx.APIUserID != nil {
					ctx = context.WithValue(ctx, APIUserIDContextKey, *result.AuthCtx.APIUserID)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireScope creates middleware that checks for required scope.
func RequireScope(scope domain.APIKeyScope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx := AuthContextFromContext(r.Context())
			if authCtx == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !authCtx.HasScope(scope) {
				http.Error(w, "Forbidden: insufficient scope", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractAPIKey extracts the API key from the request.
func extractAPIKey(r *http.Request) string {
	// Check X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Check Authorization header (Bearer token)
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Check query parameter (for WebSocket connections)
	if apiKey := r.URL.Query().Get("api_key"); apiKey != "" {
		return apiKey
	}

	return ""
}

// extractClientIP extracts the client IP from the request.
func extractClientIP(r *http.Request) net.IP {
	// Check X-Forwarded-For header (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if parsed := net.ParseIP(ip); parsed != nil {
				return parsed
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if parsed := net.ParseIP(xri); parsed != nil {
			return parsed
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil
	}
	return net.ParseIP(host)
}
