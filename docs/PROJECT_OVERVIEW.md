# AI Aggregator — Полное описание проекта

## Назначение

AI Aggregator — унифицированная API-платформа для работы с несколькими AI-провайдерами через единый интерфейс. Позволяет запускать запросы к OpenAI, Claude, Ollama через единый API с отслеживанием использования, стоимости и мультитенантной поддержкой.

## Архитектура

Монорепо с Go-бэкендом и Next.js-фронтендом:

```
AIAggregator/
├── apps/
│   ├── api/              # GraphQL API + SSE streaming (Go)
│   ├── worker/           # Async job processor (Go)
│   └── web/              # Dashboard (Next.js 14)
├── packages/
│   ├── domain/           # Доменные сущности
│   ├── usecases/         # Бизнес-логика, интерфейсы
│   ├── providers/        # AI-провайдеры
│   ├── adapters/         # PostgreSQL, in-memory репозитории
│   ├── queue/            # Asynq job queue
│   ├── pubsub/           # Redis pub/sub
│   └── shared/           # Config, logging
└── infrastructure/
    └── postgres/         # Миграции БД
```

## Технологический стек

### Backend (Go 1.24+)

| Компонент | Технология | Версия |
|-----------|------------|--------|
| GraphQL | gqlgen | v0.17.85 |
| HTTP Router | go-chi/chi | v5.0.12 |
| Database | PostgreSQL + pgx | v5.8.0 |
| Job Queue | Redis + asynq | v0.24.1 |
| Pub/Sub | go-redis | v9.5.1 |
| Logging | zerolog | v1.32.0 |
| WebSocket | gorilla/websocket | v1.5.0 |

### Frontend (Next.js 14)

| Компонент | Технология |
|-----------|------------|
| Framework | Next.js 14.1.3 |
| UI Components | Radix UI + shadcn/ui |
| Styling | Tailwind CSS |
| State Management | TanStack Query v5 |
| GraphQL Client | graphql-request |
| i18n | i18next + react-i18next |
| Theming | next-themes |
| Icons | Lucide React |

### Infrastructure

| Компонент | Технология |
|-----------|------------|
| Database | PostgreSQL 16 |
| Cache/Queue | Redis |
| Migrations | golang-migrate |
| Container | Docker Compose |

## Структура базы данных

### Таблица tenants (Арендаторы)

```sql
CREATE TABLE tenants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(255) NOT NULL,
    api_key       VARCHAR(255) UNIQUE NOT NULL,
    active        BOOLEAN DEFAULT true,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);
```

### Таблица jobs (Задания)

```sql
CREATE TABLE jobs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID REFERENCES tenants(id) ON DELETE CASCADE,
    type          VARCHAR(50) NOT NULL,        -- 'TEXT', 'IMAGE'
    input         TEXT NOT NULL,
    status        VARCHAR(50) DEFAULT 'pending', -- pending, processing, completed, failed
    result        TEXT,
    error         TEXT,
    provider      VARCHAR(100),
    tokens_in     INTEGER DEFAULT 0,
    tokens_out    INTEGER DEFAULT 0,
    cost          DECIMAL(10,6) DEFAULT 0,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    started_at    TIMESTAMPTZ,
    finished_at   TIMESTAMPTZ
);
```

### Таблица usage (Использование)

```sql
CREATE TABLE usage (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID REFERENCES tenants(id) ON DELETE CASCADE,
    job_id        UUID REFERENCES jobs(id) ON DELETE CASCADE,
    provider      VARCHAR(100) NOT NULL,
    model         VARCHAR(255),
    tokens_in     INTEGER DEFAULT 0,
    tokens_out    INTEGER DEFAULT 0,
    cost          DECIMAL(10,6) DEFAULT 0,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);
```

### Таблица provider_pricing (Ценообразование)

```sql
CREATE TABLE provider_pricing (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider                 VARCHAR(50) NOT NULL,
    model                    VARCHAR(100) NOT NULL,
    input_price_per_million  DECIMAL(10,6) NOT NULL DEFAULT 0,
    output_price_per_million DECIMAL(10,6) NOT NULL DEFAULT 0,
    image_price              DECIMAL(10,4) DEFAULT NULL,
    is_default               BOOLEAN NOT NULL DEFAULT false,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, model)
);
```

Предустановленные цены:
- OpenAI: gpt-4o-mini, gpt-4o, gpt-4-turbo, gpt-3.5-turbo, dall-e-3
- Claude: claude-3-haiku, claude-3.5-sonnet, claude-3-opus

## Реализованные фичи

### AI Провайдеры

| Провайдер | Streaming | Models API | Tools/Functions |
|-----------|-----------|------------|-----------------|
| OpenAI | да | да (/v1/models) | да |
| Claude | да | статический список | да |
| Ollama | да | да (/api/tags) | нет |
| Stub | да (dev) | нет | нет |

### API Features

- GraphQL API — полноценный API с queries, mutations, subscriptions
- SSE Streaming — real-time генерация текста с /stream endpoint
- WebSocket Subscriptions — подписка на обновления jobs и usage
- Multi-tenant — изоляция данных по API-ключам
- Health Checks — /healthz, /readyz, /health для Kubernetes

### Tools/Functions API

Поддержка вызова функций через AI (OpenAI и Claude):

```go
resp, _ := provider.Complete(ctx, &CompletionRequest{
    Messages: []ChatMessage{{Role: "user", Content: "..."}},
    Tools: []Tool{{
        Type: "function",
        Function: ToolFunction{
            Name: "get_weather",
            Parameters: map[string]interface{}{...},
        },
    }},
    ToolChoice: "auto",
})

if resp.FinishReason == "tool_calls" {
    for _, tc := range resp.ToolCalls {
        // tc.Function.Name = "get_weather"
        // tc.Function.Arguments = `{"location": "Paris"}`
    }
}
```

### Usage и Pricing

- Отслеживание токенов (input/output) по каждому запросу
- Расчёт стоимости по настраиваемым ценам
- Агрегация по провайдерам и моделям
- GraphQL Subscriptions для real-time обновлений usage

### Frontend Dashboard

| Страница | Описание |
|----------|----------|
| Chat | Streaming чат с выбором провайдера/модели |
| Jobs | Список заданий с фильтрацией и пагинацией |
| Dashboard | Обзор использования и статистика |
| Settings | Настройка цен провайдеров |

### Локализация

- Языки: English, Русский
- Автодетект: i18next-browser-languagedetector
- Переключатель: компонент в header

### UI/UX

- Dark/Light/System темы
- Responsive дизайн
- Toast уведомления (sonner)
- Command palette (cmdk)
- Markdown рендеринг в чате

### DevOps Ready

- Health Checks: liveness, readiness, full health
- Graceful Shutdown: корректное завершение сервера
- Docker Compose: postgres, redis, api, worker, web
- Environment Config: через .env файлы

## Ключевые файлы

| Файл | Описание |
|------|----------|
| apps/api/cmd/server/main.go | Entry point API сервера |
| apps/api/internal/graph/schema.graphql | GraphQL схема |
| packages/usecases/provider.go | Интерфейсы провайдеров |
| packages/providers/openai_provider.go | OpenAI с tools |
| packages/providers/claude_provider.go | Claude с tools |
| apps/web/src/components/streaming-chat.tsx | Streaming чат |
| apps/web/src/lib/i18n/ | Локализация |

## Запуск

```bash
# Инфраструктура
docker compose up -d postgres redis

# API сервер
go run ./apps/api/cmd/server

# Worker
go run ./apps/worker/cmd/worker

# Frontend
cd apps/web && pnpm dev
```

URLs:
- GraphQL Playground: http://localhost:8080/playground
- Frontend: http://localhost:3000
- Health: http://localhost:8080/health

## Переменные окружения

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
ENABLE_STUB_PROVIDER=false
```
