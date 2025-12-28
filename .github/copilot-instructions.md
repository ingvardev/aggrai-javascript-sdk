# AI Aggregator - Copilot Instructions

## Project Overview

AI Aggregator is a unified API platform for running requests across multiple AI providers (OpenAI, Claude, Ollama). It provides:
- **SSE Streaming** — real-time text generation
- **Tools/Functions** — AI function calling support for OpenAI and Claude
- **Async job processing** — background task queue via Redis/asynq
- **Provider abstraction** — unified interface for all AI providers
- **Dynamic model selection** — fetch available models from provider APIs
- **Usage tracking** — tokens, cost by provider and tenant
- **Pricing configuration** — customizable per-model pricing
- **Multi-language UI** — English/Russian with i18next
- **Health checks** — Kubernetes-ready liveness/readiness probes
- **API key authentication** — multi-tenant support

## Architecture

This is a **Go monorepo** with a **Next.js frontend**:

```
apps/
├── api/          # GraphQL API + SSE streaming server (Go, gqlgen, Chi)
├── worker/       # Async job processor (Go, asynq, Redis)
└── web/          # Dashboard frontend (Next.js 14, React, Tailwind, i18next)

packages/
├── domain/       # Core entities (Job, Tenant, Usage, Provider, Pricing)
├── usecases/     # Business logic services and interfaces
├── providers/    # AI provider implementations (OpenAI, Claude, Ollama)
├── adapters/     # Repository implementations (PostgreSQL, in-memory)
├── queue/        # Job queue (asynq/Redis)
├── pubsub/       # Redis pub/sub for real-time updates
└── shared/       # Config, logging utilities

infrastructure/
└── postgres/     # Database migrations (golang-migrate)
```

## Tech Stack

### Backend (Go 1.24+)
- **GraphQL**: gqlgen v0.17 for code generation
- **Router**: go-chi/chi v5
- **Database**: PostgreSQL 16 with pgx/v5 driver
- **Queue**: Redis + hibiken/asynq
- **Migrations**: golang-migrate
- **Logging**: rs/zerolog
- **Streaming**: Server-Sent Events (SSE)

### Frontend (Next.js 14)
- **UI**: Radix UI + Tailwind CSS + shadcn/ui
- **State**: TanStack Query (React Query)
- **GraphQL Client**: graphql-request
- **i18n**: i18next + react-i18next + language detector
- **Theming**: next-themes (dark/light/system)

## Code Conventions

### Go
- Use Clean/Hexagonal Architecture: domain → usecases → adapters
- Interfaces defined in `usecases/` package
- Implementations in `adapters/` or `providers/`
- All repository methods take `context.Context` as first parameter
- Use `uuid.UUID` from google/uuid for IDs
- Error handling: return domain errors (ErrJobNotFound, ErrUnauthorized, etc.)
- Logging with `shared.NewLogger("component-name")`

### Domain Entities
```go
// Job states
JobStatusPending, JobStatusProcessing, JobStatusCompleted, JobStatusFailed

// Job types
JobTypeText, JobTypeImage

// Provider types
ProviderTypeOpenAI, ProviderTypeClaude, ProviderTypeOllama, ProviderTypeLocal
```

### GraphQL Schema
- Located at `apps/api/internal/graph/schema.graphql` (primary)
- Regenerate with: `cd apps/api && ~/go/bin/gqlgen generate`
- Input types use `Input` suffix (CreateJobInput)
- Connection pattern for pagination (JobConnection, JobEdge, PageInfo)

### Testing
- Test files: `*_test.go` in same package
- Mock implementations in `mocks_test.go`
- Use httptest for HTTP provider testing
- Table-driven tests with `t.Run()`

## Key Interfaces

```go
// AIProvider - implement for new AI providers
type AIProvider interface {
    Name() string
    Type() string
    Execute(ctx context.Context, job *domain.Job) (*ProviderResult, error)
    IsAvailable(ctx context.Context) bool
}

// StreamingProvider - for real-time text generation
type StreamingProvider interface {
    AIProvider
    CompleteStream(ctx context.Context, req *CompletionRequest, onChunk func(chunk string)) (*CompletionResponse, error)
}

// ModelListProvider - for dynamic model loading
type ModelListProvider interface {
    AIProvider
    ListModels(ctx context.Context) ([]ModelInfo, error)
}

// ToolsProvider - for function/tool calling (OpenAI, Claude)
type ToolsProvider interface {
    StreamingProvider
    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
}

// JobRepository - job persistence
type JobRepository interface {
    Create(ctx context.Context, job *domain.Job) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error)
    Update(ctx context.Context, job *domain.Job) error
    // ...
}
```

## Tools/Functions API

OpenAI and Claude providers support function calling:

```go
// Define tools
tools := []usecases.Tool{
    {
        Type: "function",
        Function: usecases.ToolFunction{
            Name:        "get_weather",
            Description: "Get weather for a location",
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "location": map[string]interface{}{"type": "string"},
                },
                "required": []string{"location"},
            },
        },
    },
}

// Send request with tools
resp, _ := provider.Complete(ctx, &usecases.CompletionRequest{
    Messages: []usecases.ChatMessage{
        {Role: "user", Content: "What's the weather in Paris?"},
    },
    Tools:      tools,
    ToolChoice: "auto", // "auto", "none", "required", or function name
})

// Handle tool calls
if resp.FinishReason == "tool_calls" {
    for _, tc := range resp.ToolCalls {
        // tc.Function.Name = "get_weather"
        // tc.Function.Arguments = `{"location": "Paris"}`
    }
}
```

## Health Check Endpoints

Production-ready health checks:

| Endpoint | Purpose | Kubernetes Probe |
|----------|---------|------------------|
| `/healthz` | Liveness — process is running | `livenessProbe` |
| `/readyz` | Readiness — dependencies available | `readinessProbe` |
| `/health` | Full status with latency metrics | Monitoring |

Response example:
```json
{
  "status": "healthy",
  "service": "ai-aggregator-api",
  "version": "0.1.0",
  "checks": {
    "postgres": {"status": "healthy", "latency": "1.2ms"},
    "redis": {"status": "healthy", "latency": "0.5ms"}
  }
}
```

## Environment Variables

```env
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/aiaggregator?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# AI Providers
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
OLLAMA_URL=http://localhost:11434

# Server
API_PORT=8080
ENABLE_PLAYGROUND=true
ENABLE_STUB_PROVIDER=false  # Enable stub provider for testing
```

## Common Tasks

### Add new AI provider
1. Create `packages/providers/newprovider_provider.go`
2. Implement `usecases.AIProvider` interface
3. Optionally implement `StreamingProvider` for SSE support
4. Optionally implement `ModelListProvider` for dynamic models
5. Optionally implement `ToolsProvider` for function calling
6. Register in `apps/api/cmd/server/main.go` and `apps/worker/cmd/worker/main.go`
7. Add tests in `newprovider_provider_test.go`

### Add new GraphQL mutation/query
1. Update `apps/api/internal/graph/schema.graphql`
2. Run `cd apps/api && ~/go/bin/gqlgen generate`
3. Implement resolver in `apps/api/internal/graph/schema.resolvers.go`

### Add new domain entity
1. Create in `packages/domain/`
2. Add repository interface in `packages/usecases/repositories.go`
3. Implement in `packages/adapters/` (postgres.go, memory.go)
4. Add migration in `infrastructure/postgres/migrations/`

### Add translations
1. Add keys to `apps/web/src/lib/i18n/locales/en.json`
2. Add Russian translations to `apps/web/src/lib/i18n/locales/ru.json`
3. Use `const { t } = useTranslation()` in components

### Run migrations
```bash
./scripts/migrate.sh
```

### Start development
```bash
# Infrastructure
docker compose up -d postgres redis

# API server
lsof -ti:8080 | xargs kill -9 2>/dev/null; sleep 1 && go run ./apps/api/cmd/server

# Worker (separate terminal)
go run ./apps/worker/cmd/worker

# Frontend (separate terminal)
cd apps/web && pnpm dev
```

### Check for dead code
```bash
go install golang.org/x/tools/cmd/deadcode@latest
~/go/bin/deadcode -test ./...
```

## API Examples

### GraphQL Playground
http://localhost:8080/playground

### Create Job
```graphql
mutation {
  createJob(input: {type: TEXT, input: "Hello AI"}) {
    id
    status
  }
}
```

### List Provider Models
```graphql
query {
  providerModels(provider: "openai") {
    id
    name
    description
  }
}
```

### Headers
```
X-API-Key: dev-api-key-12345
```

### SSE Streaming
```bash
curl -N "http://localhost:8080/stream?provider=openai&model=gpt-4o-mini" \
  -H "X-API-Key: dev-api-key-12345" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello!"}'
```

## Testing

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Run specific package
go test ./packages/usecases/... -v
```

## Docker

```bash
# Start infrastructure
docker compose up -d postgres redis

# Build and run
docker compose up --build
```

## File Naming

- Go: `snake_case.go` (job_service.go, openai_provider.go)
- Tests: `*_test.go` (job_service_test.go)
- React: `PascalCase.tsx` (JobList.tsx, CreateJobDialog.tsx)
- Styles: `kebab-case.css`

## Provider Feature Matrix

| Provider | Streaming | Models API | Tools/Functions |
|----------|-----------|------------|-----------------|
| OpenAI   | ✅        | ✅         | ✅              |
| Claude   | ✅        | Static     | ✅              |
| Ollama   | ✅        | ✅         | ❌              |
| Stub     | ✅        | ❌         | ❌              |
