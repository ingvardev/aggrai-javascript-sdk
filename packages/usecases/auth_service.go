package usecases

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/rs/zerolog/log"
)

// ErrUnauthorized is returned when authentication fails.
var ErrUnauthorized = errors.New("unauthorized")

// ErrRateLimited is returned when too many auth attempts are made.
var ErrRateLimited = errors.New("rate limited")

// RateLimiter provides basic rate limiting for authentication.
type RateLimiter struct {
	mu       sync.RWMutex
	attempts map[string][]time.Time
	limit    int           // Max attempts
	window   time.Duration // Time window
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	// Start cleanup goroutine
	go rl.cleanup()
	return rl
}

// Allow checks if a request from the given key is allowed.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Filter old attempts
	var recent []time.Time
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= rl.limit {
		rl.attempts[key] = recent
		return false
	}

	rl.attempts[key] = append(recent, now)
	return true
}

// cleanup periodically removes old entries.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)
		for key, attempts := range rl.attempts {
			var recent []time.Time
			for _, t := range attempts {
				if t.After(cutoff) {
					recent = append(recent, t)
				}
			}
			if len(recent) == 0 {
				delete(rl.attempts, key)
			} else {
				rl.attempts[key] = recent
			}
		}
		rl.mu.Unlock()
	}
}

// AuthService handles API key authentication with audit logging.
type AuthService struct {
	apiKeyRepo  APIKeyRepository
	apiUserRepo APIUserRepository
	tenantRepo  TenantRepository
	auditRepo   AuditLogRepository
	rateLimiter *RateLimiter
}

// NewAuthService creates a new authentication service.
// If apiKeyRepo or apiUserRepo is nil, only legacy auth is supported.
func NewAuthService(
	tenantRepo TenantRepository,
	apiKeyRepo APIKeyRepository,
	apiUserRepo APIUserRepository,
	auditRepo AuditLogRepository,
) *AuthService {
	return &AuthService{
		tenantRepo:  tenantRepo,
		apiKeyRepo:  apiKeyRepo,
		apiUserRepo: apiUserRepo,
		auditRepo:   auditRepo,
		// Rate limit: 100 auth attempts per minute per IP
		rateLimiter: NewRateLimiter(100, time.Minute),
	}
}

// AuthenticateRequest contains the request context for authentication.
type AuthenticateRequest struct {
	RawKey    string
	ClientIP  net.IP
	UserAgent string
}

// AuthResult represents the result of authentication (for backward compatibility).
type AuthResult struct {
	Tenant     *domain.Tenant
	Authorized bool
	AuthCtx    *domain.AuthContext
}

// Authenticate resolves an API key to tenant and optional API user.
func (s *AuthService) Authenticate(ctx context.Context, apiKey string) (*AuthResult, error) {
	return s.AuthenticateWithContext(ctx, &AuthenticateRequest{
		RawKey: apiKey,
	})
}

// AuthenticateWithContext resolves an API key with full request context.
func (s *AuthService) AuthenticateWithContext(ctx context.Context, req *AuthenticateRequest) (*AuthResult, error) {
	if req.RawKey == "" {
		return &AuthResult{Authorized: false}, nil
	}

	// Rate limiting by IP
	ipKey := "unknown"
	if req.ClientIP != nil {
		ipKey = req.ClientIP.String()
	}
	if !s.rateLimiter.Allow(ipKey) {
		log.Warn().Str("ip", ipKey).Msg("Rate limited auth attempt")
		return nil, ErrRateLimited
	}

	// Try new API key system first (if repos are configured)
	if s.apiKeyRepo != nil && s.apiUserRepo != nil {
		authCtx, err := s.tryNewAuth(ctx, req)
		if err == nil && authCtx != nil {
			// Get tenant for backward compatibility
			tenant, _ := s.tenantRepo.GetByID(ctx, authCtx.TenantID)
			return &AuthResult{
				Authorized: true,
				Tenant:     tenant,
				AuthCtx:    authCtx,
			}, nil
		}
		// If key not found in new system, fallback to legacy
		if !errors.Is(err, domain.ErrAPIKeyNotFound) && err != nil {
			return &AuthResult{Authorized: false}, nil
		}
	}

	// Fallback: legacy tenant.api_key
	tenant, err := s.tenantRepo.GetByAPIKey(ctx, req.RawKey)
	if err != nil {
		// Log failed auth attempt
		s.logAuthEvent(ctx, nil, nil, nil, domain.AuditActionAuthFailed, req, map[string]interface{}{
			"reason":     "key_not_found",
			"key_prefix": safePrefix(req.RawKey),
		})
		return &AuthResult{Authorized: false}, nil
	}
	if !tenant.Active {
		s.logAuthEvent(ctx, &tenant.ID, nil, nil, domain.AuditActionAuthFailed, req, map[string]interface{}{
			"reason": "tenant_inactive",
		})
		return &AuthResult{Authorized: false}, nil
	}

	authCtx := &domain.AuthContext{
		TenantID:  tenant.ID,
		APIUserID: nil,
		KeyID:     nil,
		Scopes:    nil, // Legacy keys have full access (checked in AuthContext.HasScope)
		ClientIP:  req.ClientIP,
	}

	return &AuthResult{
		Authorized: true,
		Tenant:     tenant,
		AuthCtx:    authCtx,
	}, nil
}

// tryNewAuth tries to authenticate using the new API key system.
func (s *AuthService) tryNewAuth(ctx context.Context, req *AuthenticateRequest) (*domain.AuthContext, error) {
	keyHash := domain.HashAPIKey(req.RawKey)
	apiKey, err := s.apiKeyRepo.GetByHash(ctx, keyHash)
	if err != nil {
		return nil, err
	}

	// Validate key
	if !apiKey.IsValid() {
		// Log failed auth
		s.logAuthEvent(ctx, nil, &apiKey.APIUserID, &apiKey.ID, domain.AuditActionAuthFailed, req, map[string]interface{}{
			"reason": "key_invalid",
		})
		return nil, domain.ErrAPIKeyInactive
	}

	// Get API user
	apiUser, err := s.apiUserRepo.GetByID(ctx, apiKey.APIUserID)
	if err != nil {
		return nil, ErrUnauthorized
	}
	if !apiUser.Active {
		s.logAuthEvent(ctx, &apiUser.TenantID, &apiUser.ID, &apiKey.ID, domain.AuditActionAuthFailed, req, map[string]interface{}{
			"reason": "user_inactive",
		})
		return nil, ErrUnauthorized
	}

	// Update last used asynchronously (don't block the request)
	go func() {
		clientIPStr := ""
		if req.ClientIP != nil {
			clientIPStr = req.ClientIP.String()
		}
		if err := s.apiKeyRepo.UpdateLastUsed(context.Background(), apiKey.ID, clientIPStr); err != nil {
			log.Warn().Err(err).Str("key_id", apiKey.ID.String()).Msg("Failed to update last used")
		}
	}()

	return &domain.AuthContext{
		TenantID:  apiUser.TenantID,
		APIUserID: &apiUser.ID,
		KeyID:     &apiKey.ID,
		Scopes:    apiKey.Scopes,
		ClientIP:  req.ClientIP,
	}, nil
}

// AuthenticateByAPIKey is a convenience method for backward compatibility.
func (s *AuthService) AuthenticateByAPIKey(ctx context.Context, apiKey string) (*domain.Tenant, error) {
	result, err := s.Authenticate(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	if !result.Authorized {
		return nil, domain.ErrUnauthorized
	}
	return result.Tenant, nil
}

// GetTenant retrieves a tenant by ID.
func (s *AuthService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	return s.tenantRepo.GetByID(ctx, id)
}

// CreateAPIUser creates a new API user within a tenant.
func (s *AuthService) CreateAPIUser(ctx context.Context, authCtx *domain.AuthContext, name, description string) (*domain.APIUser, error) {
	// Check scope
	if err := authCtx.RequireScope(domain.ScopeAdmin); err != nil {
		s.logAuthEvent(ctx, &authCtx.TenantID, authCtx.APIUserID, authCtx.KeyID, domain.AuditActionScopeViolation, nil, map[string]interface{}{
			"required_scope": "admin",
			"action":         "create_user",
		})
		return nil, err
	}

	user := domain.NewAPIUser(authCtx.TenantID, name)
	user.Description = description

	if err := s.apiUserRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Log user creation
	s.logAuthEvent(ctx, &authCtx.TenantID, &user.ID, authCtx.KeyID, domain.AuditActionUserCreated, nil, map[string]interface{}{
		"user_name": name,
	})

	return user, nil
}

// CreateAPIKey generates a new API key for a user.
func (s *AuthService) CreateAPIKey(ctx context.Context, authCtx *domain.AuthContext, apiUserID uuid.UUID, name string, scopes []string) (*domain.APIKeyWithRaw, error) {
	// Check scope
	if err := authCtx.RequireScope(domain.ScopeAdmin); err != nil {
		s.logAuthEvent(ctx, &authCtx.TenantID, authCtx.APIUserID, authCtx.KeyID, domain.AuditActionScopeViolation, nil, map[string]interface{}{
			"required_scope": "admin",
			"action":         "create_key",
		})
		return nil, err
	}

	// Verify user exists and belongs to this tenant
	apiUser, err := s.apiUserRepo.GetByID(ctx, apiUserID)
	if err != nil {
		return nil, err
	}
	if apiUser.TenantID != authCtx.TenantID {
		return nil, domain.ErrAPIUserNotFound // Don't leak info about other tenants
	}

	// Validate scopes
	if len(scopes) == 0 {
		scopes = []string{string(domain.ScopeRead), string(domain.ScopeWrite)}
	}
	for _, scope := range scopes {
		if !domain.IsValidScope(scope) {
			return nil, errors.New("invalid scope: " + scope)
		}
	}

	keyWithRaw := domain.GenerateAPIKey(apiUserID, name, scopes)
	if err := s.apiKeyRepo.Create(ctx, &keyWithRaw.APIKey); err != nil {
		return nil, err
	}

	// Log key creation
	s.logAuthEvent(ctx, &authCtx.TenantID, &apiUserID, &keyWithRaw.ID, domain.AuditActionKeyCreated, nil, map[string]interface{}{
		"key_name":   name,
		"key_prefix": keyWithRaw.KeyPrefix,
		"scopes":     scopes,
	})

	return keyWithRaw, nil
}

// RevokeAPIKey revokes an API key.
func (s *AuthService) RevokeAPIKey(ctx context.Context, authCtx *domain.AuthContext, keyID uuid.UUID) error {
	// Check scope
	if err := authCtx.RequireScope(domain.ScopeAdmin); err != nil {
		s.logAuthEvent(ctx, &authCtx.TenantID, authCtx.APIUserID, authCtx.KeyID, domain.AuditActionScopeViolation, nil, map[string]interface{}{
			"required_scope": "admin",
			"action":         "revoke_key",
		})
		return err
	}

	// Revoke with tenant check to prevent cross-tenant revocation
	revokedBy := uuid.Nil
	if authCtx.APIUserID != nil {
		revokedBy = *authCtx.APIUserID
	}

	if err := s.apiKeyRepo.RevokeWithTenantCheck(ctx, keyID, authCtx.TenantID, revokedBy); err != nil {
		return err
	}

	// Log key revocation
	s.logAuthEvent(ctx, &authCtx.TenantID, authCtx.APIUserID, &keyID, domain.AuditActionKeyRevoked, nil, map[string]interface{}{
		"revoked_key_id": keyID.String(),
	})

	return nil
}

// ListAPIUsers lists all API users for the authenticated tenant.
func (s *AuthService) ListAPIUsers(ctx context.Context, authCtx *domain.AuthContext) ([]*domain.APIUser, error) {
	if err := authCtx.RequireScope(domain.ScopeAdmin); err != nil {
		return nil, err
	}
	return s.apiUserRepo.GetByTenantID(ctx, authCtx.TenantID)
}

// ListAPIKeys lists all API keys for a user.
func (s *AuthService) ListAPIKeys(ctx context.Context, authCtx *domain.AuthContext, userID uuid.UUID) ([]*domain.APIKey, error) {
	if err := authCtx.RequireScope(domain.ScopeAdmin); err != nil {
		return nil, err
	}

	// Verify user belongs to this tenant
	user, err := s.apiUserRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.TenantID != authCtx.TenantID {
		return nil, domain.ErrAPIUserNotFound
	}

	return s.apiKeyRepo.GetByUserID(ctx, userID)
}

// logAuthEvent logs an audit event asynchronously.
func (s *AuthService) logAuthEvent(ctx context.Context, tenantID *uuid.UUID, userID, keyID *uuid.UUID, action domain.AuditAction, req *AuthenticateRequest, details map[string]interface{}) {
	if s.auditRepo == nil {
		return
	}

	go func() {
		// Use nil tenant ID if not provided
		tid := uuid.Nil
		if tenantID != nil {
			tid = *tenantID
		}

		entry := domain.NewAuditLogEntry(tid, action)
		entry.APIUserID = userID
		entry.APIKeyID = keyID
		entry.Details = details

		if req != nil {
			entry.IPAddress = req.ClientIP
			entry.UserAgent = req.UserAgent
		}

		if err := s.auditRepo.Create(context.Background(), entry); err != nil {
			log.Warn().Err(err).Str("action", string(action)).Msg("Failed to create audit log entry")
		}
	}()
}

// safePrefix returns first 8 chars of a key for logging (doesn't expose full key).
func safePrefix(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:8] + "..."
}
