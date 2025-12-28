package middleware
// Package middleware contains HTTP middleware for the API.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ingvar/aiaggregator/packages/domain"
)

type contextKey string
























































}	return ""	}		return apiKey	if apiKey := r.URL.Query().Get("api_key"); apiKey != "" {	// Check query parameter (for WebSocket connections)	}		return strings.TrimPrefix(authHeader, "Bearer ")	if strings.HasPrefix(authHeader, "Bearer ") {	authHeader := r.Header.Get("Authorization")	// Check Authorization header (Bearer token)	}		return apiKey	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {	// Check X-API-Key headerfunc extractAPIKey(r *http.Request) string {}	}		})			next.ServeHTTP(w, r.WithContext(ctx))			ctx := context.WithValue(r.Context(), TenantContextKey, tenant)			}				return				http.Error(w, "Unauthorized", http.StatusUnauthorized)			if err != nil {			tenant, err := authenticator(apiKey)			}				return				http.Error(w, "Unauthorized", http.StatusUnauthorized)			if apiKey == "" {			apiKey := extractAPIKey(r)		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {	return func(next http.Handler) http.Handler {func AuthMiddleware(authenticator func(apiKey string) (*domain.Tenant, error)) func(http.Handler) http.Handler {// AuthMiddleware creates authentication middleware.}	return tenant	}		return nil	if !ok {	tenant, ok := ctx.Value(TenantContextKey).(*domain.Tenant)func TenantFromContext(ctx context.Context) *domain.Tenant {// TenantFromContext retrieves the tenant from the request context.)	TenantContextKey contextKey = "tenant"	// TenantContextKey is the context key for the authenticated tenant.const (
