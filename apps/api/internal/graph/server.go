// Package graph contains the GraphQL server implementation.
package graph

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
)

// NewServer creates a new GraphQL server handler.
func NewServer(resolver *Resolver) http.Handler {
	srv := handler.NewDefaultServer(NewExecutableSchema(Config{
		Resolvers: resolver,
	}))

	// Add transports
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	// Add extensions
	srv.Use(extension.Introspection{})
	srv.SetQueryCache(lru.New(1000))

	return srv
}
