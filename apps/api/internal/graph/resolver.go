package graph

import (
	"context"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
	"github.com/ingvar/aiaggregator/apps/api/internal/middleware"
	"github.com/ingvar/aiaggregator/packages/providers"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// This file will not be regenerated automatically.
// It serves as dependency injection for your app.

type Resolver struct {
	jobService       *usecases.JobService
	authService      *usecases.AuthService
	tenantRepo       usecases.TenantRepository
	providerRegistry *providers.ProviderRegistry
}

func NewResolver(
	jobService *usecases.JobService,
	authService *usecases.AuthService,
	tenantRepo usecases.TenantRepository,
	providerRegistry *providers.ProviderRegistry,
) *Resolver {
	return &Resolver{
		jobService:       jobService,
		authService:      authService,
		tenantRepo:       tenantRepo,
		providerRegistry: providerRegistry,
	}
}

// NewServer creates a new GraphQL server with WebSocket and SSE support for subscriptions.
func NewServer(resolver *Resolver) http.Handler {
	srv := handler.New(NewExecutableSchema(Config{
		Resolvers: resolver,
	}))

	// Add SSE transport (for subscriptions via Server-Sent Events)
	srv.AddTransport(transport.SSE{})

	// Add WebSocket transport (for subscriptions via WebSocket)
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, validate the origin properly
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		// Handle connectionParams from WebSocket clients for authentication
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
			// Extract API key from connectionParams
			apiKey, _ := initPayload["X-API-Key"].(string)
			if apiKey == "" {
				apiKey, _ = initPayload["x-api-key"].(string)
			}
			if apiKey == "" {
				apiKey, _ = initPayload["apiKey"].(string)
			}

			if apiKey != "" && resolver.authService != nil {
				result, err := resolver.authService.Authenticate(ctx, apiKey)
				if err == nil && result.Authorized {
					// Add tenant to context for WebSocket subscriptions
					ctx = context.WithValue(ctx, middleware.TenantContextKey, result.Tenant)
				}
			}

			return ctx, &initPayload, nil
		},
	})

	// Add standard HTTP transports
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	// Add introspection extension
	srv.Use(extension.Introspection{})

	return srv
}
