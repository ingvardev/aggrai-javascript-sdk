package usecases
// Package usecases contains application business logic and use case implementations.
package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

// AuthService handles authentication business logic.

































}	return s.tenantRepo.GetByID(ctx, id)func (s *AuthService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {// GetTenant retrieves a tenant by ID.}	return tenant, nil	}		return nil, domain.ErrUnauthorized	if !tenant.Active {	}		return nil, domain.ErrUnauthorized	if err != nil {	tenant, err := s.tenantRepo.GetByAPIKey(ctx, apiKey)	}		return nil, domain.ErrUnauthorized	if apiKey == "" {func (s *AuthService) AuthenticateByAPIKey(ctx context.Context, apiKey string) (*domain.Tenant, error) {// AuthenticateByAPIKey authenticates a tenant by API key.}	}		tenantRepo: tenantRepo,	return &AuthService{func NewAuthService(tenantRepo TenantRepository) *AuthService {// NewAuthService creates a new auth service.}	tenantRepo TenantRepositorytype AuthService struct {
