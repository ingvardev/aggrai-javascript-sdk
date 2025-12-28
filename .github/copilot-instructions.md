# AI Aggregator - Copilot Instructions

## Project Overview

AI Aggregator is a unified API platform for running requests across multiple AI providers (OpenAI, Claude, Ollama). It provides:
- **SSE Streaming** — real-time text generation
- **Async job processing** — background task queue via Redis/asynq
- **Provider abstraction** — unified interface for all AI providers
- **Dynamic model selection** — fetch available models from provider APIs
- **Usage tracking** — tokens, cost by provider and tenant
- **Pricing configuration** — customizable per-model pricing
- **Multi-language UI** — English/Russian with i18next
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
├── usecases/     # Business logic services
├── providers/    # AI provider implementations (OpenAI, Claude, Ollama)
├── adapters/     # Repository implementations (PostgreSQL, in-memory)
├── queue/        # Job queue (asynq/Redis)
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

// JobRepository - job persistence
type JobRepository interface {
    Create(ctx context.Context, job *domain.Job) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error)
    Update(ctx context.Context, job *domain.Job) error
    // ...
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
```

## Common Tasks

### Add new AI provider
1. Create `packages/providers/newprovider_provider.go`
2. Implement `usecases.AIProvider` interface
3. Optionally implement `StreamingProvider` for SSE support
4. Optionally implement `ModelListProvider` for dynamic models
5. Register in `apps/api/cmd/server/main.go` and `apps/worker/cmd/worker/main.go`
6. Add tests in `newprovider_provider_test.go`

### Add new GraphQL mutation/query
1. Update `apps/api/internal/graph/schema.graphql`
2. Run `cd apps/api && ~/go/bin/gqlgen generate`
3. Implement resolver in `apps/api/internal/graph/schema.resolvers.go`

### Add new domain entity
1. Create in `packages/domain/`
2. Add repository interface in `packages/usecases/repositories.go`
3. Implement in `packages/adapters/` (postgres.go, inmemory.go)
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

### Headers
```
X-API-Key: dev-api-key-12345
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
