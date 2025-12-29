package usecases

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/rs/zerolog/log"
)

const (
	// MaxFailedAttempts before account lockout
	MaxFailedAttempts = 5
	// LockoutDuration is how long an account is locked
	LockoutDuration = 15 * time.Minute
	// SessionDuration is how long a session is valid
	SessionDuration = 24 * time.Hour
)

// WebAuthService handles tenant owner authentication for the dashboard.
type WebAuthService struct {
	ownerRepo   TenantOwnerRepository
	sessionRepo OwnerSessionRepository
	tenantRepo  TenantRepository
	auditRepo   AuditLogRepository
}

// NewWebAuthService creates a new web auth service.
func NewWebAuthService(
	ownerRepo TenantOwnerRepository,
	sessionRepo OwnerSessionRepository,
	tenantRepo TenantRepository,
	auditRepo AuditLogRepository,
) *WebAuthService {
	return &WebAuthService{
		ownerRepo:   ownerRepo,
		sessionRepo: sessionRepo,
		tenantRepo:  tenantRepo,
		auditRepo:   auditRepo,
	}
}

// LoginRequest contains login credentials.
type LoginRequest struct {
	Email     string
	Password  string
	UserAgent string
	IP        net.IP
}

// LoginResult contains the result of a login attempt.
type LoginResult struct {
	Success      bool
	SessionToken string // Raw token, only returned once
	Owner        *domain.TenantOwner
	Tenant       *domain.Tenant
	Error        error
}

// Login authenticates a tenant owner and creates a session.
func (s *WebAuthService) Login(ctx context.Context, req *LoginRequest) *LoginResult {
	owner, err := s.ownerRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		log.Debug().Err(err).Str("email", req.Email).Msg("Login: owner not found")
		s.logAuditEvent(ctx, nil, nil, domain.AuditActionAuthFailed, req.IP, req.UserAgent, map[string]interface{}{
			"email":  req.Email,
			"reason": "not_found",
		})
		return &LoginResult{Success: false, Error: domain.ErrInvalidCredentials}
	}

	log.Debug().Str("email", owner.Email).Str("hash", owner.PasswordHash[:20]).Msg("Login: owner found, checking password")

	// Check if can login
	if err := owner.CanLogin(); err != nil {
		log.Debug().Err(err).Msg("Login: CanLogin failed")
		s.logAuditEvent(ctx, &owner.TenantID, &owner.ID, domain.AuditActionAuthFailed, req.IP, req.UserAgent, map[string]interface{}{
			"reason": err.Error(),
		})
		return &LoginResult{Success: false, Error: err}
	}

	// Verify password
	if !owner.CheckPassword(req.Password) {
		log.Debug().Str("password_len", fmt.Sprintf("%d", len(req.Password))).Msg("Login: password check failed")
		_ = s.ownerRepo.IncrementFailedAttempts(ctx, owner.ID)

		// Lock account after too many attempts
		if owner.FailedAttempts+1 >= MaxFailedAttempts {
			lockUntil := time.Now().Add(LockoutDuration)
			_ = s.ownerRepo.LockAccount(ctx, owner.ID, lockUntil)
		}

		s.logAuditEvent(ctx, &owner.TenantID, &owner.ID, domain.AuditActionAuthFailed, req.IP, req.UserAgent, map[string]interface{}{
			"reason": "wrong_password",
		})
		return &LoginResult{Success: false, Error: domain.ErrInvalidCredentials}
	}

	// Reset failed attempts on successful login
	_ = s.ownerRepo.ResetFailedAttempts(ctx, owner.ID)

	ipStr := ""
	if req.IP != nil {
		ipStr = req.IP.String()
	}
	_ = s.ownerRepo.UpdateLastLogin(ctx, owner.ID, ipStr)

	// Create session
	session, rawToken := domain.NewOwnerSession(owner.ID, req.UserAgent, req.IP, SessionDuration)

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return &LoginResult{Success: false, Error: err}
	}

	// Get tenant info
	tenant, _ := s.tenantRepo.GetByID(ctx, owner.TenantID)

	s.logAuditEvent(ctx, &owner.TenantID, &owner.ID, domain.AuditActionOwnerLoggedIn, req.IP, req.UserAgent, nil)

	return &LoginResult{
		Success:      true,
		SessionToken: rawToken,
		Owner:        owner,
		Tenant:       tenant,
	}
}

// ValidateSession validates a session token and returns auth context.
func (s *WebAuthService) ValidateSession(ctx context.Context, rawToken string) (*domain.WebAuthContext, error) {
	if rawToken == "" {
		return nil, domain.ErrSessionNotFound
	}

	tokenHash := domain.HashSessionToken(rawToken)

	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	if session.IsExpired() {
		_ = s.sessionRepo.Delete(ctx, session.ID)
		return nil, domain.ErrSessionExpired
	}

	owner, err := s.ownerRepo.GetByID(ctx, session.OwnerID)
	if err != nil {
		return nil, err
	}

	if !owner.Active {
		return nil, domain.ErrAccountInactive
	}

	return &domain.WebAuthContext{
		OwnerID:   owner.ID,
		TenantID:  owner.TenantID,
		Email:     owner.Email,
		Name:      owner.Name,
		Role:      owner.Role,
		SessionID: session.ID,
	}, nil
}

// Logout invalidates a session.
func (s *WebAuthService) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}

	tokenHash := domain.HashSessionToken(rawToken)
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil // Already logged out
	}

	// Get owner for audit log
	owner, _ := s.ownerRepo.GetByID(ctx, session.OwnerID)
	if owner != nil {
		s.logAuditEvent(ctx, &owner.TenantID, &owner.ID, domain.AuditActionOwnerLoggedOut, nil, "", nil)
	}

	return s.sessionRepo.Delete(ctx, session.ID)
}

// LogoutAll invalidates all sessions for an owner.
func (s *WebAuthService) LogoutAll(ctx context.Context, ownerID uuid.UUID) error {
	return s.sessionRepo.DeleteByOwnerID(ctx, ownerID)
}

// GetOwnerByID retrieves an owner by ID.
func (s *WebAuthService) GetOwnerByID(ctx context.Context, id uuid.UUID) (*domain.TenantOwner, error) {
	return s.ownerRepo.GetByID(ctx, id)
}

// GetOwnersByTenantID retrieves all owners for a tenant.
func (s *WebAuthService) GetOwnersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.TenantOwner, error) {
	return s.ownerRepo.GetByTenantID(ctx, tenantID)
}

// CreateOwner creates a new tenant owner.
func (s *WebAuthService) CreateOwner(ctx context.Context, tenantID uuid.UUID, email, password, name string, role domain.OwnerRole) (*domain.TenantOwner, error) {
	// Check if email already exists
	existing, err := s.ownerRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	owner, err := domain.NewTenantOwner(tenantID, email, password, name, role)
	if err != nil {
		return nil, err
	}

	if err := s.ownerRepo.Create(ctx, owner); err != nil {
		return nil, err
	}

	s.logAuditEvent(ctx, &tenantID, &owner.ID, domain.AuditActionOwnerCreated, nil, "", map[string]interface{}{
		"email": email,
		"name":  name,
		"role":  string(role),
	})

	return owner, nil
}

// UpdateOwner updates an existing owner.
func (s *WebAuthService) UpdateOwner(ctx context.Context, owner *domain.TenantOwner) error {
	return s.ownerRepo.Update(ctx, owner)
}

// ChangePassword changes an owner's password.
func (s *WebAuthService) ChangePassword(ctx context.Context, ownerID uuid.UUID, newPassword string) error {
	owner, err := s.ownerRepo.GetByID(ctx, ownerID)
	if err != nil {
		return err
	}

	if err := owner.SetPassword(newPassword); err != nil {
		return err
	}

	if err := s.ownerRepo.Update(ctx, owner); err != nil {
		return err
	}

	// Invalidate all sessions on password change
	_ = s.sessionRepo.DeleteByOwnerID(ctx, ownerID)

	s.logAuditEvent(ctx, &owner.TenantID, &ownerID, domain.AuditActionPasswordChanged, nil, "", nil)

	return nil
}

// CleanupExpiredSessions removes all expired sessions.
func (s *WebAuthService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	return s.sessionRepo.DeleteExpired(ctx)
}

// logAuditEvent logs an audit event asynchronously.
func (s *WebAuthService) logAuditEvent(ctx context.Context, tenantID, ownerID *uuid.UUID, action domain.AuditAction, ip net.IP, userAgent string, details map[string]interface{}) {
	if s.auditRepo == nil {
		return
	}

	go func() {
		tid := uuid.Nil
		if tenantID != nil {
			tid = *tenantID
		}

		entry := domain.NewAuditLogEntry(tid, action)
		entry.IPAddress = ip
		entry.UserAgent = userAgent
		entry.Details = details

		if err := s.auditRepo.Create(context.Background(), entry); err != nil {
			log.Warn().Err(err).Str("action", string(action)).Msg("Failed to create audit log entry")
		}
	}()
}
