package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

func TestAuthService_Authenticate(t *testing.T) {
	ctx := context.Background()

	t.Run("valid API key", func(t *testing.T) {
		repo := NewMockTenantRepository()

		tenant := domain.NewTenant("Test Tenant", "test-api-key-123")
		_ = repo.Create(ctx, tenant)

		svc := NewAuthService(repo, nil, nil, nil)
		result, err := svc.Authenticate(ctx, "test-api-key-123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Authorized {
			t.Error("expected authorized to be true")
		}

		if result.Tenant.ID != tenant.ID {
			t.Errorf("expected tenant ID %v, got %v", tenant.ID, result.Tenant.ID)
		}
	})

	t.Run("invalid API key", func(t *testing.T) {
		repo := NewMockTenantRepository()

		svc := NewAuthService(repo, nil, nil, nil)
		result, err := svc.Authenticate(ctx, "invalid-key")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Authorized {
			t.Error("expected authorized to be false")
		}
	})

	t.Run("empty API key", func(t *testing.T) {
		repo := NewMockTenantRepository()

		svc := NewAuthService(repo, nil, nil, nil)
		result, err := svc.Authenticate(ctx, "")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Authorized {
			t.Error("expected authorized to be false for empty key")
		}
	})

	t.Run("inactive tenant", func(t *testing.T) {
		repo := NewMockTenantRepository()

		tenant := domain.NewTenant("Test Tenant", "test-api-key")
		tenant.Deactivate()
		_ = repo.Create(ctx, tenant)

		svc := NewAuthService(repo, nil, nil, nil)
		result, err := svc.Authenticate(ctx, "test-api-key")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Authorized {
			t.Error("expected authorized to be false for inactive tenant")
		}
	})

	t.Run("repository error returns unauthorized (not error)", func(t *testing.T) {
		repo := NewMockTenantRepository()
		repo.GetByKeyErr = errors.New("db error")

		svc := NewAuthService(repo, nil, nil, nil)
		result, err := svc.Authenticate(ctx, "some-key")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Authorized {
			t.Error("expected authorized to be false on repository error")
		}
	})
}

func TestAuthService_AuthenticateByAPIKey(t *testing.T) {
	ctx := context.Background()

	t.Run("valid API key", func(t *testing.T) {
		repo := NewMockTenantRepository()

		tenant := domain.NewTenant("Test Tenant", "test-api-key")
		_ = repo.Create(ctx, tenant)

		svc := NewAuthService(repo, nil, nil, nil)
		result, err := svc.AuthenticateByAPIKey(ctx, "test-api-key")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.ID != tenant.ID {
			t.Errorf("expected tenant ID %v, got %v", tenant.ID, result.ID)
		}
	})

	t.Run("invalid API key returns error", func(t *testing.T) {
		repo := NewMockTenantRepository()

		svc := NewAuthService(repo, nil, nil, nil)
		_, err := svc.AuthenticateByAPIKey(ctx, "invalid-key")

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})
}

func TestAuthService_GetTenant(t *testing.T) {
	ctx := context.Background()

	t.Run("existing tenant", func(t *testing.T) {
		repo := NewMockTenantRepository()

		tenant := domain.NewTenant("Test Tenant", "test-api-key")
		_ = repo.Create(ctx, tenant)

		svc := NewAuthService(repo, nil, nil, nil)
		result, err := svc.GetTenant(ctx, tenant.ID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Name != tenant.Name {
			t.Errorf("expected tenant name %q, got %q", tenant.Name, result.Name)
		}
	})

	t.Run("non-existing tenant", func(t *testing.T) {
		repo := NewMockTenantRepository()

		svc := NewAuthService(repo, nil, nil, nil)
		_, err := svc.GetTenant(ctx, uuid.New())

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, domain.ErrTenantNotFound) {
			t.Errorf("expected ErrTenantNotFound, got %v", err)
		}
	})
}
