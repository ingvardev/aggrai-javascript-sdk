# AI Aggregator

Унифицированная платформа для работы с различными AI-провайдерами (OpenAI, Claude, Ollama) через единый GraphQL API.

## ✨ Возможности

- 🚀 **Асинхронная обработка** — создание AI-запросов с фоновой обработкой через Redis/asynq
- 🔄 **Мульти-провайдеры** — OpenAI, Claude, Ollama, Stub (для тестов)
- 📊 **Отслеживание использования** — токены, стоимость по провайдерам и тенантам
- 🔐 **API-ключи** — мультитенантная аутентификация
- 🎮 **GraphQL Playground** — интерактивное тестирование API

## 🏗️ Архитектура

```
┌─────────────────┐     ┌──────────────────┐
│   GraphQL API   │     │     Worker       │
│    (:8080)      │     │   (asynq)        │
└────────┬────────┘     └────────┬─────────┘
         │                       │
         └───────────┬───────────┘
                     │
              ┌──────▼──────┐
              │  PostgreSQL │
              │   (jobs)    │
              └──────┬──────┘
                     │
              ┌──────▼──────┐
              │    Redis    │
              │   (queue)   │
              └─────────────┘
```

## 🚀 Быстрый старт

### 1. Запуск инфраструктуры

```bash
docker compose up -d postgres redis
```

### 2. Применение миграций

```bash
./scripts/migrate.sh
```

### 3. Запуск API сервера

```bash
go run ./apps/api/cmd/server
```

### 4. Запуск Worker (в отдельном терминале)

```bash
go run ./apps/worker/cmd/worker
```

### 5. Тестирование

```bash
# GraphQL Playground
open http://localhost:8080/playground

# Или через curl
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-api-key-12345" \
  -d '{"query":"mutation { createJob(input: { type: TEXT, input: \"Hello AI!\" }) { id status } }"}'
```

## 📁 Структура проекта

```
AIAggregator/
├── apps/
│   ├── api/              # GraphQL API сервер
│   │   ├── cmd/server/   # Точка входа
│   │   └── internal/
│   │       ├── graph/    # GraphQL resolvers (gqlgen)
│   │       ├── handlers/ # HTTP handlers
│   │       └── middleware/
│   ├── worker/           # Asynq worker
│   └── web/              # Next.js фронтенд (WIP)
├── packages/
│   ├── domain/           # Доменные сущности (Job, Tenant, Usage)
│   ├── usecases/         # Бизнес-логика (JobService, AuthService)
│   ├── providers/        # AI провайдеры (OpenAI, Claude, Stub)
│   ├── adapters/         # Репозитории (PostgreSQL, InMemory)
│   ├── queue/            # Очередь задач (asynq)
│   └── shared/           # Конфигурация, логгер
├── infrastructure/
│   ├── postgres/
│   │   ├── migrations/   # SQL миграции
│   │   ├── queries/      # sqlc queries
│   │   └── db/           # Сгенерированный sqlc код
│   └── docker/           # Dockerfiles
├── docker-compose.yml
├── gqlgen.yml
└── go.mod
```

## ⚙️ Конфигурация

Создайте `.env` файл (см. `.env.example`):

```env
# Server
API_HOST=0.0.0.0
API_PORT=8080

# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/aiaggregator?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# AI Providers (опционально)
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
OLLAMA_URL=http://localhost:11434

# Features
ENABLE_PLAYGROUND=true
LOG_LEVEL=debug
```

## 🔌 GraphQL API

### Queries

```graphql
# Текущий тенант
query { me { id name active } }

# Получить job по ID
query { job(id: "...") { id status result provider tokensIn tokensOut cost } }

# Список jobs
query { jobs { edges { node { id status type } } pageInfo { totalCount } } }

# Список провайдеров
query { providers { id name type enabled } }
```

### Mutations

```graphql
# Создать job
mutation {
  createJob(input: { type: TEXT, input: "Расскажи про Go" }) {
    id status
  }
}

# Отменить job
mutation { cancelJob(id: "...") { id status } }
```

## 🔧 Разработка

### Генерация GraphQL

```bash
cd apps/api && ~/go/bin/gqlgen generate
```

### Генерация sqlc

```bash
cd infrastructure/postgres && ~/go/bin/sqlc generate
```

### Сборка

```bash
go build ./...
```

### Тесты

```bash
go test ./...
```

## 📦 Технологии

| Компонент | Технология |
|-----------|------------|
| Backend | Go 1.24+ |
| GraphQL | gqlgen |
| Database | PostgreSQL 16 |
| ORM | sqlc |
| Queue | Redis + asynq |
| Frontend | Next.js 14 (WIP) |

## 📝 License

MIT

---

## License

MIT
