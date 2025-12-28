// Package graph contains generated GraphQL types and interfaces.
package graph

import (
	"time"

	"github.com/google/uuid"
)

// Job represents an AI processing request.
type Job struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenantId"`
	Type       JobType    `json:"type"`
	Input      string     `json:"input"`
	Status     JobStatus  `json:"status"`
	Result     *string    `json:"result"`
	Error      *string    `json:"error"`
	Provider   *string    `json:"provider"`
	TokensIn   int        `json:"tokensIn"`
	TokensOut  int        `json:"tokensOut"`
	Cost       float64    `json:"cost"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	StartedAt  *time.Time `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
}

// JobType represents the type of AI request.
type JobType string

const (
	JobTypeText  JobType = "TEXT"
	JobTypeImage JobType = "IMAGE"
)

// JobStatus represents the status of a job.
type JobStatus string

const (
	JobStatusPending    JobStatus = "PENDING"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
)

// Tenant represents an organization with API access.
type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Provider represents an AI provider.
type Provider struct {
	ID       uuid.UUID    `json:"id"`
	Name     string       `json:"name"`
	Type     ProviderType `json:"type"`
	Enabled  bool         `json:"enabled"`
	Priority int          `json:"priority"`
}

// ProviderType represents the type of AI provider.
type ProviderType string

const (
	ProviderTypeOpenai ProviderType = "OPENAI"
	ProviderTypeClaude ProviderType = "CLAUDE"
	ProviderTypeLocal  ProviderType = "LOCAL"
	ProviderTypeOllama ProviderType = "OLLAMA"
)

// Usage represents resource consumption.
type Usage struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	JobID     uuid.UUID `json:"jobId"`
	Provider  string    `json:"provider"`
	Model     *string   `json:"model"`
	TokensIn  int       `json:"tokensIn"`
	TokensOut int       `json:"tokensOut"`
	Cost      float64   `json:"cost"`
	CreatedAt time.Time `json:"createdAt"`
}

// UsageSummary represents aggregated usage statistics.
type UsageSummary struct {
	Provider       string  `json:"provider"`
	TotalTokensIn  int     `json:"totalTokensIn"`
	TotalTokensOut int     `json:"totalTokensOut"`
	TotalCost      float64 `json:"totalCost"`
	JobCount       int     `json:"jobCount"`
}

// PageInfo contains pagination information.
type PageInfo struct {
	TotalCount      int  `json:"totalCount"`
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
}

// JobConnection represents a paginated list of jobs.
type JobConnection struct {
	Edges    []*JobEdge `json:"edges"`
	PageInfo *PageInfo  `json:"pageInfo"`
}

// JobEdge represents an edge in the job connection.
type JobEdge struct {
	Node   *Job   `json:"node"`
	Cursor string `json:"cursor"`
}

// CreateJobInput represents input for creating a job.
type CreateJobInput struct {
	Type  JobType `json:"type"`
	Input string  `json:"input"`
}

// JobsFilter represents filter options for jobs.
type JobsFilter struct {
	Status *JobStatus `json:"status"`
	Type   *JobType   `json:"type"`
}

// PaginationInput represents pagination options.
type PaginationInput struct {
	Limit  *int `json:"limit"`
	Offset *int `json:"offset"`
}
