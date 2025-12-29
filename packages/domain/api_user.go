// Package domain contains core business entities and value objects.
package domain

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
)

// API User/Key errors.
var (
	ErrAPIUserNotFound  = errors.New("api user not found")
	ErrAPIKeyNotFound   = errors.New("api key not found")
	ErrAPIKeyExpired    = errors.New("api key expired")
	ErrAPIKeyInactive   = errors.New("api key is inactive")
	ErrAPIKeyRevoked    = errors.New("api key has been revoked")
	ErrInsufficientScope = errors.New("insufficient scope for this operation")
)

// APIKeyScope represents permission scopes for API keys.
type APIKeyScope string

const (
	ScopeRead   APIKeyScope = "read"   // Read-only access (list jobs, get usage)
	ScopeWrite  APIKeyScope = "write"  // Create jobs, use streaming
	ScopeAdmin  APIKeyScope = "admin"  // Manage users and keys
	ScopeAll    APIKeyScope = "*"      // Full access
)

// ValidScopes is the list of all valid scopes.
var ValidScopes = []APIKeyScope{ScopeRead, ScopeWrite, ScopeAdmin, ScopeAll}

// IsValidScope checks if a scope string is valid.
func IsValidScope(s string) bool {
	for _, scope := range ValidScopes {
		if string(scope) == s {
			return true
		}
	}
	return false
}

// APIUser represents an API user within a tenant.
type APIUser struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description string
	Active      bool
	Metadata    map[string]interface{}
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewAPIUser creates a new API user.
func NewAPIUser(tenantID uuid.UUID, name string) *APIUser {
	now := time.Now().UTC()
	return &APIUser{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      name,
		Active:    true,
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// APIKey represents an API key belonging to an API user.
type APIKey struct {
	ID         uuid.UUID
	APIUserID  uuid.UUID
	KeyHash    string   // HMAC-SHA256 hash of the key
	KeyPrefix  string   // First 12 chars for identification (e.g., "sk-abc12345...")
	Name       string
	Scopes     []string // Permission scopes
	Active     bool
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
	LastUsedIP net.IP
	UsageCount int64
	CreatedAt  time.Time
	RevokedAt  *time.Time
	RevokedBy  *uuid.UUID // API user who revoked this key
}

// APIKeyWithRaw is returned only on creation, containing the raw key.
// The raw key is shown ONLY ONCE and never stored.
type APIKeyWithRaw struct {
	APIKey
	RawKey string // Only available once, at creation time
}

// hmacSecret should be set from environment variable.
// In production, use a proper secret management system.
var hmacSecret []byte

// SetHMACSecret sets the secret used for HMAC hashing.
// Must be called during application initialization.
func SetHMACSecret(secret string) {
	hmacSecret = []byte(secret)
}

// GenerateAPIKey creates a new API key with a random value.
func GenerateAPIKey(apiUserID uuid.UUID, name string, scopes []string) *APIKeyWithRaw {
	// Generate 32 bytes of random data
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to UUID-based generation
		randomBytes = []byte(uuid.New().String() + uuid.New().String())[:32]
	}

	// Format: "sk-" + 48 hex chars = 51 chars total
	rawKey := "sk-" + hex.EncodeToString(randomBytes)

	// Default scopes if none provided
	if len(scopes) == 0 {
		scopes = []string{string(ScopeRead), string(ScopeWrite)}
	}

	return &APIKeyWithRaw{
		APIKey: APIKey{
			ID:        uuid.New(),
			APIUserID: apiUserID,
			KeyHash:   HashAPIKey(rawKey),
			KeyPrefix: rawKey[:12],
			Name:      name,
			Scopes:    scopes,
			Active:    true,
			CreatedAt: time.Now().UTC(),
		},
		RawKey: rawKey,
	}
}

// HashAPIKey computes HMAC-SHA256 hash of an API key.
// Uses server secret to prevent rainbow table attacks if hashes are leaked.
func HashAPIKey(rawKey string) string {
	if len(hmacSecret) == 0 {
		// Fallback to simple SHA256 if HMAC secret not set
		// This should only happen in tests
		hash := sha256.Sum256([]byte(rawKey))
		return hex.EncodeToString(hash[:])
	}

	h := hmac.New(sha256.New, hmacSecret)
	h.Write([]byte(rawKey))
	return hex.EncodeToString(h.Sum(nil))
}

// IsExpired checks if the API key has expired.
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// IsRevoked checks if the API key has been revoked.
func (k *APIKey) IsRevoked() bool {
	return k.RevokedAt != nil
}

// IsValid checks if the API key can be used.
func (k *APIKey) IsValid() bool {
	return k.Active && !k.IsExpired() && !k.IsRevoked()
}

// HasScope checks if the API key has the required scope.
func (k *APIKey) HasScope(required APIKeyScope) bool {
	for _, scope := range k.Scopes {
		if scope == string(ScopeAll) || scope == string(required) {
			return true
		}
	}
	return false
}

// HasAnyScope checks if the API key has any of the required scopes.
func (k *APIKey) HasAnyScope(required ...APIKeyScope) bool {
	for _, r := range required {
		if k.HasScope(r) {
			return true
		}
	}
	return false
}

// AuthContext holds resolved authentication information.
type AuthContext struct {
	TenantID  uuid.UUID
	APIUserID *uuid.UUID   // nil for legacy tenant-level keys
	KeyID     *uuid.UUID   // nil for legacy tenant-level keys
	Scopes    []string     // Scopes from the key, empty for legacy
	ClientIP  net.IP       // Client IP address
}

// IsLegacyAuth returns true if using legacy tenant.api_key.
func (a *AuthContext) IsLegacyAuth() bool {
	return a.APIUserID == nil
}

// HasScope checks if the auth context has the required scope.
// Legacy auth has all scopes for backward compatibility.
func (a *AuthContext) HasScope(required APIKeyScope) bool {
	if a.IsLegacyAuth() {
		return true // Legacy keys have full access
	}
	for _, scope := range a.Scopes {
		if scope == string(ScopeAll) || scope == string(required) {
			return true
		}
	}
	return false
}

// RequireScope returns an error if the auth context doesn't have the required scope.
func (a *AuthContext) RequireScope(required APIKeyScope) error {
	if !a.HasScope(required) {
		return ErrInsufficientScope
	}
	return nil
}

// AuditAction represents types of audit events.
type AuditAction string

const (
	AuditActionKeyCreated    AuditAction = "key_created"
	AuditActionKeyRevoked    AuditAction = "key_revoked"
	AuditActionKeyDeleted    AuditAction = "key_deleted"
	AuditActionUserCreated   AuditAction = "user_created"
	AuditActionUserUpdated   AuditAction = "user_updated"
	AuditActionUserDeleted   AuditAction = "user_deleted"
	AuditActionAuthSuccess   AuditAction = "auth_success"
	AuditActionAuthFailed    AuditAction = "auth_failed"
	AuditActionScopeViolation AuditAction = "scope_violation"
	// Request activity actions
	AuditActionCompletion    AuditAction = "completion"
	AuditActionStreaming     AuditAction = "streaming"
	AuditActionRequest       AuditAction = "request"
)

// AuditLogEntry represents an entry in the audit log.
type AuditLogEntry struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	APIUserID *uuid.UUID
	APIKeyID  *uuid.UUID
	Action    AuditAction
	Details   map[string]interface{}
	IPAddress net.IP
	UserAgent string
	CreatedAt time.Time
}

// NewAuditLogEntry creates a new audit log entry.
func NewAuditLogEntry(tenantID uuid.UUID, action AuditAction) *AuditLogEntry {
	return &AuditLogEntry{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Action:    action,
		Details:   make(map[string]interface{}),
		CreatedAt: time.Now().UTC(),
	}
}
