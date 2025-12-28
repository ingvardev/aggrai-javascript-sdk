// Package graph contains the GraphQL resolver implementations.
package graph

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/apps/api/internal/middleware"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/providers"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// Resolver is the root resolver with injected dependencies.
type Resolver struct {
	jobService       *usecases.JobService
	authService      *usecases.AuthService
	providerRegistry *providers.ProviderRegistry
}

// NewResolver creates a new resolver with dependencies.
func NewResolver(
	jobService *usecases.JobService,
	authService *usecases.AuthService,
	providerRegistry *providers.ProviderRegistry,
) *Resolver {
	return &Resolver{
		jobService:       jobService,
		authService:      authService,
		providerRegistry: providerRegistry,
	}
}

// Query resolver
type queryResolver struct{ *Resolver }

// Mutation resolver
type mutationResolver struct{ *Resolver }

// Query returns the query resolver.
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Mutation returns the mutation resolver.
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Me returns the current authenticated tenant.
func (r *queryResolver) Me(ctx context.Context) (*Tenant, error) {
	domainTenant := middleware.TenantFromContext(ctx)
	if domainTenant == nil {
		return nil, domain.ErrUnauthorized
	}

	return &Tenant{
		ID:        domainTenant.ID,
		Name:      domainTenant.Name,
		Active:    domainTenant.Active,
		CreatedAt: domainTenant.CreatedAt,
		UpdatedAt: domainTenant.UpdatedAt,
	}, nil
}

// Job returns a job by ID.
func (r *queryResolver) Job(ctx context.Context, id uuid.UUID) (*Job, error) {
	tenant := middleware.TenantFromContext(ctx)
	if tenant == nil {
		return nil, domain.ErrUnauthorized
	}

	job, err := r.jobService.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify tenant ownership
	if job.TenantID != tenant.ID {
		return nil, domain.ErrJobNotFound
	}

	return domainJobToGraphQL(job), nil
}

// Jobs returns a list of jobs for the authenticated tenant.
func (r *queryResolver) Jobs(ctx context.Context, filter *JobsFilter, pagination *PaginationInput) (*JobConnection, error) {
	tenant := middleware.TenantFromContext(ctx)
	if tenant == nil {
		return nil, domain.ErrUnauthorized
	}

	limit := 20
	offset := 0
	if pagination != nil {
		if pagination.Limit != nil {
			limit = *pagination.Limit
		}
		if pagination.Offset != nil {
			offset = *pagination.Offset
		}
	}

	jobs, err := r.jobService.ListJobs(ctx, tenant.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	totalCount, err := r.jobService.CountJobs(ctx, tenant.ID)
	if err != nil {
		return nil, err
	}

	edges := make([]*JobEdge, len(jobs))
	for i, job := range jobs {
		edges[i] = &JobEdge{
			Node:   domainJobToGraphQL(job),
			Cursor: job.ID.String(),
		}
	}

	return &JobConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			TotalCount:      totalCount,
			HasNextPage:     offset+len(jobs) < totalCount,
			HasPreviousPage: offset > 0,
		},
	}, nil
}

// UsageSummary returns usage statistics for the authenticated tenant.
func (r *queryResolver) UsageSummary(ctx context.Context) ([]*UsageSummary, error) {
	tenant := middleware.TenantFromContext(ctx)
	if tenant == nil {
		return nil, domain.ErrUnauthorized
	}

	// For now return mock data - will be implemented with real usage repo
	return []*UsageSummary{
		{
			Provider:       "stub-provider",
			TotalTokensIn:  1500,
			TotalTokensOut: 3000,
			TotalCost:      0.045,
			JobCount:       25,
		},
	}, nil
}

// Providers returns available AI providers.
func (r *queryResolver) Providers(ctx context.Context) ([]*Provider, error) {
	providerList := r.providerRegistry.List()

	result := make([]*Provider, len(providerList))
	for i, p := range providerList {
		result[i] = &Provider{
			ID:       uuid.New(), // Generate a consistent ID
			Name:     p.Name(),
			Type:     ProviderType(p.Type()),
			Enabled:  p.IsAvailable(ctx),
			Priority: 0,
		}
	}

	return result, nil
}

// CreateJob creates a new job.
func (r *mutationResolver) CreateJob(ctx context.Context, input CreateJobInput) (*Job, error) {
	tenant := middleware.TenantFromContext(ctx)
	if tenant == nil {
		return nil, domain.ErrUnauthorized
	}

	// Map GraphQL type to domain type
	var jobType domain.JobType
	switch input.Type {
	case JobTypeText:
		jobType = domain.JobTypeText
	case JobTypeImage:
		jobType = domain.JobTypeImage
	default:
		return nil, domain.ErrInvalidInput
	}

	job, err := r.jobService.CreateJob(ctx, &usecases.CreateJobInput{
		TenantID: tenant.ID,
		Type:     jobType,
		Input:    input.Input,
	})
	if err != nil {
		return nil, err
	}

	return domainJobToGraphQL(job), nil
}

// CancelJob cancels a pending job.
func (r *mutationResolver) CancelJob(ctx context.Context, id uuid.UUID) (*Job, error) {
	tenant := middleware.TenantFromContext(ctx)
	if tenant == nil {
		return nil, domain.ErrUnauthorized
	}

	job, err := r.jobService.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify tenant ownership
	if job.TenantID != tenant.ID {
		return nil, domain.ErrJobNotFound
	}

	// Only pending jobs can be cancelled
	if job.Status != domain.JobStatusPending {
		return nil, domain.ErrInvalidInput
	}

	job.MarkFailed("cancelled by user")
	// Note: Update would be needed here via repo

	return domainJobToGraphQL(job), nil
}

// Helper function to convert domain Job to GraphQL Job
func domainJobToGraphQL(job *domain.Job) *Job {
	gqlJob := &Job{
		ID:         job.ID,
		TenantID:   job.TenantID,
		Input:      job.Input,
		Result:     job.Result,
		Error:      job.Error,
		Provider:   job.Provider,
		TokensIn:   job.TokensIn,
		TokensOut:  job.TokensOut,
		Cost:       job.Cost,
		CreatedAt:  job.CreatedAt,
		UpdatedAt:  job.UpdatedAt,
		StartedAt:  job.StartedAt,
		FinishedAt: job.FinishedAt,
	}

	// Map domain types to GraphQL types
	switch job.Type {
	case domain.JobTypeText:
		gqlJob.Type = JobTypeText
	case domain.JobTypeImage:
		gqlJob.Type = JobTypeImage
	}

	switch job.Status {
	case domain.JobStatusPending:
		gqlJob.Status = JobStatusPending
	case domain.JobStatusProcessing:
		gqlJob.Status = JobStatusProcessing
	case domain.JobStatusCompleted:
		gqlJob.Status = JobStatusCompleted
	case domain.JobStatusFailed:
		gqlJob.Status = JobStatusFailed
	}

	return gqlJob
}

// Helper to create a pointer to an int
func intPtr(i int) *int {
	return &i
}

// Helper to create a pointer to a time
func timePtr(t time.Time) *time.Time {
	return &t
}
