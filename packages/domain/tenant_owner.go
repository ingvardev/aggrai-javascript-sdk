package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Tenant Owner errors
var (
	ErrOwnerNotFound      = errors.New("owner not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account is locked")
	ErrAccountInactive    = errors.New("account is inactive")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrSessionExpired     = errors.New("session expired")
	ErrSessionNotFound    = errors.New("session not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// OwnerRole represents the role of a tenant owner
type OwnerRole string

const (
	OwnerRoleOwner  OwnerRole = "owner"  // Full access, can delete tenant
	OwnerRoleAdmin  OwnerRole = "admin"  // Manage users and keys
	OwnerRoleMember OwnerRole = "member" // Read-only access
)

// TenantOwner represents a human user who can log in to the dashboard
type TenantOwner struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Email          string
	PasswordHash   string
	Name           string
	Role           OwnerRole
	Active         bool
	EmailVerified  bool
	LastLoginAt    *time.Time
	LastLoginIP    net.IP
	FailedAttempts int
	LockedUntil    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewTenantOwner creates a new tenant owner with hashed password
func NewTenantOwner(tenantID uuid.UUID, email, password, name string, role OwnerRole) (*TenantOwner, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &TenantOwner{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         role,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// CheckPassword verifies the password against the stored hash
func (o *TenantOwner) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(o.PasswordHash), []byte(password))
	return err == nil
}

// SetPassword updates the password hash
func (o *TenantOwner) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	o.PasswordHash = string(hash)
	o.UpdatedAt = time.Now().UTC()
	return nil
}

// IsLocked checks if the account is currently locked
func (o *TenantOwner) IsLocked() bool {
	if o.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*o.LockedUntil)
}

// CanLogin checks if the owner can log in
func (o *TenantOwner) CanLogin() error {
	if !o.Active {
		return ErrAccountInactive
	}
	if o.IsLocked() {
		return ErrAccountLocked
	}
	return nil
}

// OwnerSession represents an active login session
type OwnerSession struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	TokenHash string
	UserAgent string
	IPAddress net.IP
	ExpiresAt time.Time
	CreatedAt time.Time
}

// NewOwnerSession creates a new session with generated token
func NewOwnerSession(ownerID uuid.UUID, userAgent string, ip net.IP, duration time.Duration) (*OwnerSession, string) {
	rawToken := generateSessionToken()
	tokenHash := HashSessionToken(rawToken)

	return &OwnerSession{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		TokenHash: tokenHash,
		UserAgent: userAgent,
		IPAddress: ip,
		ExpiresAt: time.Now().Add(duration),
		CreatedAt: time.Now().UTC(),
	}, rawToken
}

// IsExpired checks if the session has expired
func (s *OwnerSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// WebAuthContext holds resolved web authentication information
type WebAuthContext struct {
	OwnerID   uuid.UUID
	TenantID  uuid.UUID
	Email     string
	Name      string
	Role      OwnerRole
	SessionID uuid.UUID
}

// HasPermission checks if the owner has required permission
func (w *WebAuthContext) HasPermission(required OwnerRole) bool {
	switch w.Role {
	case OwnerRoleOwner:
		return true // Owner can do everything
	case OwnerRoleAdmin:
		return required != OwnerRoleOwner
	case OwnerRoleMember:
		return required == OwnerRoleMember
	}
	return false
}

// generateSessionToken creates a cryptographically secure random token
func generateSessionToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback to UUID if crypto/rand fails
		return uuid.New().String() + uuid.New().String()
	}
	return hex.EncodeToString(b)
}

// HashSessionToken computes SHA256 hash of a session token
func HashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// Audit actions for owner events
const (
	AuditActionOwnerLoggedIn  AuditAction = "owner_logged_in"
	AuditActionOwnerLoggedOut AuditAction = "owner_logged_out"
	AuditActionOwnerCreated   AuditAction = "owner_created"
	AuditActionPasswordChanged AuditAction = "password_changed"
)
