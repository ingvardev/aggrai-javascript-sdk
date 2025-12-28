// Package graph contains GraphQL schema configuration.
package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

// Config holds the GraphQL schema configuration.
type Config struct {
	Resolvers ResolverRoot
}

// ResolverRoot is the root resolver interface.
type ResolverRoot interface {
	Query() QueryResolver
	Mutation() MutationResolver
}

// QueryResolver defines query resolvers.
type QueryResolver interface {
	Me(ctx context.Context) (*Tenant, error)
	Job(ctx context.Context, id uuid.UUID) (*Job, error)
	Jobs(ctx context.Context, filter *JobsFilter, pagination *PaginationInput) (*JobConnection, error)
	UsageSummary(ctx context.Context) ([]*UsageSummary, error)
	Providers(ctx context.Context) ([]*Provider, error)
}

// MutationResolver defines mutation resolvers.
type MutationResolver interface {
	CreateJob(ctx context.Context, input CreateJobInput) (*Job, error)
	CancelJob(ctx context.Context, id uuid.UUID) (*Job, error)
}

// NewExecutableSchema creates a new executable schema.
func NewExecutableSchema(cfg Config) graphql.ExecutableSchema {
	return &executableSchema{
		resolvers: cfg.Resolvers,
	}
}

// executableSchema wraps the resolvers.
type executableSchema struct {
	resolvers ResolverRoot
}

// Schema returns the schema string.
func (e *executableSchema) Schema() *graphql.Schema {
	return nil
}

// Complexity returns the complexity config.
func (e *executableSchema) Complexity(typeName, field string, childComplexity int, args map[string]interface{}) (int, bool) {
	return 0, false
}

// Exec executes a query.
func (e *executableSchema) Exec(ctx context.Context) graphql.ResponseHandler {
	return nil
}
