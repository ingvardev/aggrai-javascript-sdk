package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

type webAuthContextKey struct{}

// WebAuthMiddleware handles web session authentication
type WebAuthMiddleware struct {
	webAuthService *usecases.WebAuthService
}

// NewWebAuthMiddleware creates a new web auth middleware
func NewWebAuthMiddleware(webAuthService *usecases.WebAuthService) *WebAuthMiddleware {
	return &WebAuthMiddleware{
		webAuthService: webAuthService,
	}
}

// Handler returns HTTP middleware that extracts session token and validates it
func (m *WebAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Add client IP to context
		clientIP := getClientIPFromRequest(r)
		ctx = context.WithValue(ctx, "client_ip", clientIP)

		// Add user agent to context
		ctx = context.WithValue(ctx, "user_agent", r.UserAgent())

		// Try to get session token from Authorization header or cookie
		sessionToken := extractSessionToken(r)
		if sessionToken != "" {
			ctx = context.WithValue(ctx, "session_token", sessionToken)

			// Validate session
			if m.webAuthService != nil {
				webCtx, err := m.webAuthService.ValidateSession(ctx, sessionToken)
				if err == nil && webCtx != nil {
					ctx = context.WithValue(ctx, webAuthContextKey{}, webCtx)
				}
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WebAuthContextFromContext extracts WebAuthContext from context
func WebAuthContextFromContext(ctx context.Context) *domain.WebAuthContext {
	webCtx, ok := ctx.Value(webAuthContextKey{}).(*domain.WebAuthContext)
	if !ok {
		return nil
	}
	return webCtx
}

// extractSessionToken extracts session token from request
func extractSessionToken(r *http.Request) string {
	// Try Authorization header first (Bearer token)
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Try X-Session-Token header
	if token := r.Header.Get("X-Session-Token"); token != "" {
		return token
	}

	// Try cookie
	if cookie, err := r.Cookie("session_token"); err == nil {
		return cookie.Value
	}

	return ""
}

// getClientIPFromRequest extracts client IP from request
func getClientIPFromRequest(r *http.Request) string {
	// Try X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if parsed := net.ParseIP(ip); parsed != nil {
				return ip
			}
		}
	}

	// Try X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if parsed := net.ParseIP(xri); parsed != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
