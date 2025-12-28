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

// AllJobType returns all possible job types.
func AllJobType() []JobType {
	return []JobType{JobTypeText, JobTypeImage}
}

// IsValid checks if the job type is valid.
func (e JobType) IsValid() bool {
	switch e {
	case JobTypeText, JobTypeImage:
		return true
	}
	return false
}

func (e JobType) String() string {
	return string(e)
}

// JobStatus represents the status of a job.
type JobStatus string

const (
	JobStatusPending    JobStatus = "PENDING"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
)

// AllJobStatus returns all possible job statuses.
func AllJobStatus() []JobStatus {
	return []JobStatus{JobStatusPending, JobStatusProcessing, JobStatusCompleted, JobStatusFailed}
}

// IsValid checks if the job status is valid.
func (e JobStatus) IsValid() bool {
	switch e {
	case JobStatusPending, JobStatusProcessing, JobStatusCompleted, JobStatusFailed:
		return true
	}
	return false
}

func (e JobStatus) String() string {
	return string(e)
}

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

// AllProviderType returns all possible provider types.
func AllProviderType() []ProviderType {
	return []ProviderType{ProviderTypeOpenai, ProviderTypeClaude, ProviderTypeLocal, ProviderTypeOllama}
}

// IsValid checks if the provider type is valid.
func (e ProviderType) IsValid() bool {
	switch e {
	case ProviderTypeOpenai, ProviderTypeClaude, ProviderTypeLocal, ProviderTypeOllama:
		return true
	}
	return false
}

func (e ProviderType) String() string {
	return string(e)
}

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
