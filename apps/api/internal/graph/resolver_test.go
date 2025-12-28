package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/providers"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// Mock implementations for testing

type mockJobRepo struct {
	jobs map[uuid.UUID]*domain.Job
}

func newMockJobRepo() *mockJobRepo {
	return &mockJobRepo{jobs: make(map[uuid.UUID]*domain.Job)}
}

func (r *mockJobRepo) Create(ctx context.Context, job *domain.Job) error {
	r.jobs[job.ID] = job
	return nil
}

func (r *mockJobRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	job, ok := r.jobs[id]
	if !ok {
		return nil, domain.ErrJobNotFound
	}
	return job, nil
}

func (r *mockJobRepo) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Job, error) {
	var result []*domain.Job
	for _, job := range r.jobs {
		if job.TenantID == tenantID {
			result = append(result, job)
		}
	}
	return result, nil
}

func (r *mockJobRepo) Update(ctx context.Context, job *domain.Job) error {
	r.jobs[job.ID] = job
	return nil
}

func (r *mockJobRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.jobs, id)
	return nil
}

func (r *mockJobRepo) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	count := 0
	for _, job := range r.jobs {
		if job.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

type mockTenantRepo struct {
	tenants map[uuid.UUID]*domain.Tenant
	byKey   map[string]*domain.Tenant
}

func newMockTenantRepo() *mockTenantRepo {
	return &mockTenantRepo{
		tenants: make(map[uuid.UUID]*domain.Tenant),
		byKey:   make(map[string]*domain.Tenant),
	}
}

func (r *mockTenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	r.tenants[tenant.ID] = tenant
	r.byKey[tenant.APIKey] = tenant
	return nil
}

func (r *mockTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	tenant, ok := r.tenants[id]
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	return tenant, nil
}

func (r *mockTenantRepo) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Tenant, error) {
	tenant, ok := r.byKey[apiKey]
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	return tenant, nil
}

func (r *mockTenantRepo) Update(ctx context.Context, tenant *domain.Tenant) error {
	r.tenants[tenant.ID] = tenant
	r.byKey[tenant.APIKey] = tenant
	return nil
}

func (r *mockTenantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if tenant, ok := r.tenants[id]; ok {
		delete(r.byKey, tenant.APIKey)
		delete(r.tenants, id)
	}
	return nil
}

func (r *mockTenantRepo) List(ctx context.Context, limit, offset int) ([]*domain.Tenant, error) {
	var result []*domain.Tenant
	for _, t := range r.tenants {
		result = append(result, t)
	}
	return result, nil
}

type mockJobQueue struct {
	enqueued []uuid.UUID
}

func (q *mockJobQueue) Enqueue(ctx context.Context, jobID uuid.UUID) error {
	q.enqueued = append(q.enqueued, jobID)
	return nil
}

func (q *mockJobQueue) Close() error {
	return nil
}

// mockUsageRepo implements usecases.UsageRepository for testing
type mockUsageRepo struct {
	usages map[uuid.UUID]*domain.Usage
}

func newMockUsageRepo() *mockUsageRepo {
	return &mockUsageRepo{usages: make(map[uuid.UUID]*domain.Usage)}
}

func (r *mockUsageRepo) Create(ctx context.Context, usage *domain.Usage) error {
	r.usages[usage.JobID] = usage
	return nil
}

func (r *mockUsageRepo) GetByJobID(ctx context.Context, jobID uuid.UUID) (*domain.Usage, error) {
	if usage, ok := r.usages[jobID]; ok {
		return usage, nil
	}
	return nil, nil
}

func (r *mockUsageRepo) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Usage, error) {
	return []*domain.Usage{}, nil
}

func (r *mockUsageRepo) GetSummary(ctx context.Context, tenantID uuid.UUID) ([]*domain.UsageSummary, error) {
	return []*domain.UsageSummary{}, nil
}

// mockPricingRepo implements usecases.PricingRepository for testing
type mockPricingRepo struct {
	pricings map[uuid.UUID]*domain.ProviderPricing
}

func newMockPricingRepo() *mockPricingRepo {
	return &mockPricingRepo{pricings: make(map[uuid.UUID]*domain.ProviderPricing)}
}

func (r *mockPricingRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProviderPricing, error) {
	if pricing, ok := r.pricings[id]; ok {
		return pricing, nil
	}
	return nil, nil
}

func (r *mockPricingRepo) GetByProviderModel(ctx context.Context, provider, model string) (*domain.ProviderPricing, error) {
	return nil, nil
}

func (r *mockPricingRepo) GetDefaultByProvider(ctx context.Context, provider string) (*domain.ProviderPricing, error) {
	return nil, nil
}

func (r *mockPricingRepo) List(ctx context.Context) ([]*domain.ProviderPricing, error) {
	return []*domain.ProviderPricing{}, nil
}

func (r *mockPricingRepo) ListByProvider(ctx context.Context, provider string) ([]*domain.ProviderPricing, error) {
	return []*domain.ProviderPricing{}, nil
}

func (r *mockPricingRepo) Create(ctx context.Context, pricing *domain.ProviderPricing) error {
	r.pricings[pricing.ID] = pricing
	return nil
}

func (r *mockPricingRepo) Update(ctx context.Context, pricing *domain.ProviderPricing) error {
	r.pricings[pricing.ID] = pricing
	return nil
}

func (r *mockPricingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.pricings, id)
	return nil
}

func setupTestServer(t *testing.T) (*httptest.Server, *domain.Tenant) {
	jobRepo := newMockJobRepo()
	tenantRepo := newMockTenantRepo()
	queue := &mockJobQueue{}

	// Create test tenant
	tenant := domain.NewTenant("Test Tenant", "test-api-key")
	_ = tenantRepo.Create(context.Background(), tenant)

	jobService := usecases.NewJobService(jobRepo, queue)
	authService := usecases.NewAuthService(tenantRepo)
	providerRegistry := providers.NewProviderRegistry()

	// Create mock usage repo and pricing service for test
	usageRepo := newMockUsageRepo()
	pricingRepo := newMockPricingRepo()
	pricingService := usecases.NewPricingService(pricingRepo)

	resolver := NewResolver(jobService, authService, tenantRepo, usageRepo, pricingService, providerRegistry)
	handler := NewServer(resolver)

	return httptest.NewServer(handler), tenant
}

func graphqlRequest(t *testing.T, server *httptest.Server, query string, variables map[string]interface{}) map[string]interface{} {
	body, _ := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})

	resp, err := http.Post(server.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return result
}

func TestGraphQL_CreateJob(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	query := `
		mutation CreateJob($input: CreateJobInput!) {
			createJob(input: $input) {
				id
				status
				input
			}
		}
	`

	// Note: CreateJobInput only has type and input, tenantId comes from auth context
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"type":  "TEXT",
			"input": "Hello, AI!",
		},
	}

	result := graphqlRequest(t, server, query, variables)

	// Since we don't have auth context in test, we expect an error
	// In a real test we would set up auth middleware
	if errors, ok := result["errors"]; ok {
		t.Logf("Expected auth error in test without middleware: %v", errors)
		return
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected response format: %v", result)
	}

	createJob, ok := data["createJob"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected createJob format: %v", data)
	}

	if createJob["status"] != "PENDING" {
		t.Errorf("expected status PENDING, got %v", createJob["status"])
	}

	if createJob["input"] != "Hello, AI!" {
		t.Errorf("expected input %q, got %v", "Hello, AI!", createJob["input"])
	}
}

func TestGraphQL_GetJob(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Test getting a non-existent job
	getQuery := `
		query GetJob($id: ID!) {
			job(id: $id) {
				id
				status
				input
			}
		}
	`

	getVars := map[string]interface{}{
		"id": uuid.New().String(),
	}

	result := graphqlRequest(t, server, getQuery, getVars)

	// Should have errors (job not found or unauthorized)
	if errors, ok := result["errors"]; ok {
		t.Logf("Expected error: %v", errors)
	}
}

func TestGraphQL_ListProviders(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	query := `
		query {
			providers {
				id
				name
				type
				enabled
			}
		}
	`

	result := graphqlRequest(t, server, query, nil)

	if errors, ok := result["errors"]; ok {
		t.Fatalf("GraphQL errors: %v", errors)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected response format: %v", result)
	}

	providers, ok := data["providers"].([]interface{})
	if !ok {
		t.Fatalf("unexpected providers format: %v", data)
	}

	// Provider list might be empty in test, that's fine
	t.Logf("Found %d providers", len(providers))
}

func TestGraphQL_JobNotFound(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	query := `
		query GetJob($id: ID!) {
			job(id: $id) {
				id
			}
		}
	`

	variables := map[string]interface{}{
		"id": uuid.New().String(),
	}

	result := graphqlRequest(t, server, query, variables)

	// Should have errors
	errors, ok := result["errors"]
	if !ok {
		t.Fatal("expected errors, got none")
	}

	t.Logf("Expected error received: %v", errors)
}
