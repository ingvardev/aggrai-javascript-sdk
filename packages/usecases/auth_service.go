// Package usecases contains application business logic and use case implementations.
package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

// AuthResult represents the result of authentication.
type AuthResult struct {
	Tenant     *domain.Tenant
	Authorized bool
}

// AuthService handles authentication business logic.
type AuthService struct {
	tenantRepo TenantRepository
}

// NewAuthService creates a new auth service.
func NewAuthService(tenantRepo TenantRepository) *AuthService {
	return &AuthService{
		tenantRepo: tenantRepo,
	}
}

// Authenticate validates an API key and returns the associated tenant.
func (s *AuthService) Authenticate(ctx context.Context, apiKey string) (*AuthResult, error) {
	if apiKey == "" {
		return &AuthResult{Authorized: false}, nil
	}

	tenant, err := s.tenantRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		if err == domain.ErrTenantNotFound {
			return &AuthResult{Authorized: false}, nil
		}
		return nil, err
	}

	if !tenant.Active {
		return &AuthResult{Authorized: false}, nil
	}

	return &AuthResult{
		Tenant:     tenant,
		Authorized: true,
	}, nil
}

// AuthenticateByAPIKey is an alias for Authenticate for backward compatibility.
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
