package graph
// Package graph contains the GraphQL server implementation.
package graph



























}	return srv	srv.Use(extension.Introspection{})	// Add extensions	srv.AddTransport(transport.POST{})	srv.AddTransport(transport.GET{})	srv.AddTransport(transport.Options{})	// Add transports	}))		Resolvers: resolver,	srv := handler.New(NewExecutableSchema(Config{		resolver := NewResolver()func NewServer() http.Handler {// NewServer creates a new GraphQL server handler.)	"github.com/99designs/gqlgen/graphql/handler/transport"	"github.com/99designs/gqlgen/graphql/handler/extension"	"github.com/99designs/gqlgen/graphql/handler"	"net/http"import (
