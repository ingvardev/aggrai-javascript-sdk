// Package graph contains GraphQL schema and executable schema.
package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/ast"
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

// DirectiveRoot is placeholder for directives.
type DirectiveRoot struct{}

// ComplexityRoot is placeholder for complexity calculations.
type ComplexityRoot struct{}

// NewExecutableSchema creates a new executable schema.
// This is a simplified version - in production, use gqlgen generate
func NewExecutableSchema(cfg Config) graphql.ExecutableSchema {
	return &simpleSchema{
		resolvers: cfg.Resolvers,
	}
}

type simpleSchema struct {
	resolvers ResolverRoot
}

func (s *simpleSchema) Schema() *ast.Schema {
	return nil
}

func (s *simpleSchema) Complexity(typeName, field string, childComplexity int, args map[string]interface{}) (int, bool) {
	return childComplexity + 1, true
}

func (s *simpleSchema) Exec(ctx context.Context) graphql.ResponseHandler {
	return graphql.OneShot(graphql.ErrorResponse(ctx, "not implemented - run gqlgen generate"))
}
