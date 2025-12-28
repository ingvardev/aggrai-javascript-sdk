# AI Aggregator

[English](#english) | [Русский](#русский)

---

## Русский

### Описание

AI Aggregator — платформа для унифицированного доступа к различным AI-провайдерам (OpenAI, Claude, локальные модели) через единый API и веб-дашборд.

### Возможности

- 🚀 Создание AI-запросов (текст/изображения) с асинхронной обработкой
- 🔄 Абстракция провайдеров с автоматической маршрутизацией
- 📊 Отслеживание использования (токены/стоимость) по провайдерам и тенантам
- 🔐 Аутентификация через API-ключи
- 🖥️ Веб-дашборд: список задач, детали, статус провайдеров, метрики

### Технологии

**Backend:**
- Go 1.22+
- GraphQL (gqlgen)
- PostgreSQL + sqlc
- Redis + asynq (очереди)

**Frontend:**
- Next.js 14 (App Router)
- TypeScript
- shadcn/ui + Radix UI
- Tailwind CSS

### Быстрый старт

```bash
# Клонирование
git clone <repo-url>
cd AIAggregator

# Запуск инфраструктуры
docker-compose up -d

# Backend API
cd apps/api
go run ./cmd/server

# Worker
cd apps/worker
go run ./cmd/worker

# Frontend
cd apps/web
npm install
npm run dev
```

### Структура проекта

```
AIAggregator/
├── apps/
│   ├── api/          # GraphQL API сервер
│   ├── worker/       # Asynq воркер
│   └── web/          # Next.js фронтенд
├── packages/
│   ├── domain/       # Доменные сущности
│   ├── usecases/     # Бизнес-логика
│   ├── providers/    # AI провайдеры
│   └── shared/       # Общие утилиты
└── infrastructure/
    └── postgres/     # Миграции БД
```

### Переменные окружения

```env
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/aiaggregator?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# Server
API_PORT=8080
```

---

## English

### Description

AI Aggregator is a platform for unified access to various AI providers (OpenAI, Claude, local models) through a single API and web dashboard.

### Features

- 🚀 Create AI requests (text/images) with async processing
- 🔄 Provider abstraction with automatic routing
- 📊 Usage tracking (tokens/cost) per provider per tenant
- 🔐 API key authentication
- 🖥️ Web dashboard: job list, details, provider status, metrics

### Tech Stack

**Backend:**
- Go 1.22+
- GraphQL (gqlgen)
- PostgreSQL + sqlc
- Redis + asynq (queues)

**Frontend:**
- Next.js 14 (App Router)
- TypeScript
- shadcn/ui + Radix UI
- Tailwind CSS

### Quick Start

```bash
# Clone
git clone <repo-url>
cd AIAggregator

# Start infrastructure
docker-compose up -d

# Backend API
cd apps/api
go run ./cmd/server

# Worker
cd apps/worker
go run ./cmd/worker

# Frontend
cd apps/web
npm install
npm run dev
```

### Project Structure

```
AIAggregator/
├── apps/
│   ├── api/          # GraphQL API server
│   ├── worker/       # Asynq worker
│   └── web/          # Next.js frontend
├── packages/
│   ├── domain/       # Domain entities
│   ├── usecases/     # Business logic
│   ├── providers/    # AI providers
│   └── shared/       # Shared utilities
└── infrastructure/
    └── postgres/     # DB migrations
```

### Environment Variables

```env
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/aiaggregator?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# Server
API_PORT=8080
```

---

## License

MIT
