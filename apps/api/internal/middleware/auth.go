// Package middleware contains HTTP middleware for the API.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

type contextKey string

const (
	// TenantContextKey is the context key for the authenticated tenant.
	TenantContextKey contextKey = "tenant"
)

// TenantFromContext retrieves the tenant from the request context.
func TenantFromContext(ctx context.Context) *domain.Tenant {
	tenant, ok := ctx.Value(TenantContextKey).(*domain.Tenant)
	if !ok {
		return nil
	}
	return tenant
}

// AuthMiddleware creates authentication middleware.
func AuthMiddleware(authService *usecases.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for certain paths
			if r.URL.Path == "/health" || r.URL.Path == "/playground" || r.URL.Path == "/" {
				next.ServeHTTP(w, r)
				return
			}

			apiKey := extractAPIKey(r)
			if apiKey == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			result, err := authService.Authenticate(r.Context(), apiKey)
			if err != nil {
				http.Error(w, "Authentication error", http.StatusInternalServerError)
				return
			}

			if !result.Authorized {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), TenantContextKey, result.Tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
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
