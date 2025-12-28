package graph

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/ingvar/aiaggregator/packages/providers"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// This file will not be regenerated automatically.
// It serves as dependency injection for your app.

type Resolver struct {
	jobService       *usecases.JobService
	authService      *usecases.AuthService
	providerRegistry *providers.ProviderRegistry
}

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

// NewServer creates a new GraphQL server.
func NewServer(resolver *Resolver) http.Handler {
	return handler.NewDefaultServer(NewExecutableSchema(Config{
		Resolvers: resolver,
	}))
}
