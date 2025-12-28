# AI Aggregator — Схема базы данных

PostgreSQL 16+

## ER-диаграмма

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│  ┌──────────────┐         ┌──────────────┐         ┌──────────────┐        │
│  │   tenants    │────────<│     jobs     │────────<│    usage     │        │
│  └──────────────┘  1:N    └──────────────┘  1:1    └──────────────┘        │
│                                                                             │
│  ┌──────────────┐         ┌──────────────────────┐                         │
│  │  providers   │         │   provider_pricing   │                         │
│  └──────────────┘         └──────────────────────┘                         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Таблицы

### tenants

Арендаторы/пользователи системы. Каждый tenant имеет уникальный API-ключ.

```sql
CREATE TABLE tenants (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             VARCHAR(255) NOT NULL,
    api_key          VARCHAR(255) UNIQUE NOT NULL,
    active           BOOLEAN DEFAULT true,
    default_provider VARCHAR(255),
    settings         JSONB DEFAULT '{
        "darkMode": true,
        "notifications": {
            "jobCompleted": true,
            "jobFailed": true,
            "providerOffline": true,
            "usageThreshold": false,
            "weeklySummary": false,
            "marketingEmails": false
        }
    }'::jsonb,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);
```

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | Первичный ключ |
| name | VARCHAR(255) | Имя арендатора |
| api_key | VARCHAR(255) | Уникальный API-ключ для аутентификации |
| active | BOOLEAN | Активен ли tenant |
| default_provider | VARCHAR(255) | Провайдер по умолчанию |
| settings | JSONB | Настройки пользователя (тема, уведомления) |
| created_at | TIMESTAMPTZ | Дата создания |
| updated_at | TIMESTAMPTZ | Дата обновления |

**Индексы:**
- `idx_tenants_api_key` — для быстрого поиска по API-ключу
- `idx_tenants_active` — для фильтрации активных

---

### jobs

Задания на обработку AI-провайдерами.

```sql
CREATE TABLE jobs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type        VARCHAR(50) NOT NULL,
    input       TEXT NOT NULL,
    status      VARCHAR(50) NOT NULL DEFAULT 'pending',
    result      TEXT,
    error       TEXT,
    provider    VARCHAR(100),
    tokens_in   INTEGER DEFAULT 0,
    tokens_out  INTEGER DEFAULT 0,
    cost        DECIMAL(10, 6) DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);
```

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | Первичный ключ |
| tenant_id | UUID | FK → tenants.id |
| type | VARCHAR(50) | Тип задания: `TEXT`, `IMAGE` |
| input | TEXT | Входные данные (prompt) |
| status | VARCHAR(50) | Статус: `pending`, `processing`, `completed`, `failed` |
| result | TEXT | Результат выполнения |
| error | TEXT | Текст ошибки (если failed) |
| provider | VARCHAR(100) | Использованный провайдер |
| tokens_in | INTEGER | Входные токены |
| tokens_out | INTEGER | Выходные токены |
| cost | DECIMAL(10,6) | Стоимость в USD |
| created_at | TIMESTAMPTZ | Дата создания |
| updated_at | TIMESTAMPTZ | Дата обновления |
| started_at | TIMESTAMPTZ | Начало обработки |
| finished_at | TIMESTAMPTZ | Завершение обработки |

**Индексы:**
- `idx_jobs_tenant_id` — фильтрация по tenant
- `idx_jobs_status` — фильтрация по статусу
- `idx_jobs_created_at` — сортировка по дате
- `idx_jobs_tenant_status` — составной для частых запросов

---

### providers

Конфигурация AI-провайдеров (хранятся в БД, но сейчас не используются — провайдеры создаются в коде).

```sql
CREATE TABLE providers (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) UNIQUE NOT NULL,
    type       VARCHAR(50) NOT NULL,
    endpoint   VARCHAR(500),
    api_key    VARCHAR(500),
    model      VARCHAR(255),
    enabled    BOOLEAN DEFAULT true,
    priority   INTEGER DEFAULT 0,
    config     JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | Первичный ключ |
| name | VARCHAR(255) | Уникальное имя провайдера |
| type | VARCHAR(50) | Тип: `openai`, `claude`, `ollama`, `local` |
| endpoint | VARCHAR(500) | URL API провайдера |
| api_key | VARCHAR(500) | API-ключ провайдера |
| model | VARCHAR(255) | Модель по умолчанию |
| enabled | BOOLEAN | Активен ли провайдер |
| priority | INTEGER | Приоритет (выше = предпочтительнее) |
| config | JSONB | Дополнительные настройки |
| created_at | TIMESTAMPTZ | Дата создания |
| updated_at | TIMESTAMPTZ | Дата обновления |

**Индексы:**
- `idx_providers_type` — фильтрация по типу
- `idx_providers_enabled` — фильтрация активных
- `idx_providers_priority` — сортировка по приоритету

---

### usage

Детальная статистика использования для биллинга и аналитики.

```sql
CREATE TABLE usage (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    job_id     UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    provider   VARCHAR(100) NOT NULL,
    model      VARCHAR(255),
    tokens_in  INTEGER DEFAULT 0,
    tokens_out INTEGER DEFAULT 0,
    cost       DECIMAL(10, 6) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | Первичный ключ |
| tenant_id | UUID | FK → tenants.id |
| job_id | UUID | FK → jobs.id |
| provider | VARCHAR(100) | Имя провайдера |
| model | VARCHAR(255) | Использованная модель |
| tokens_in | INTEGER | Входные токены |
| tokens_out | INTEGER | Выходные токены |
| cost | DECIMAL(10,6) | Стоимость в USD |
| created_at | TIMESTAMPTZ | Дата записи |

**Индексы:**
- `idx_usage_tenant_id` — агрегация по tenant
- `idx_usage_job_id` — связь с job
- `idx_usage_provider` — группировка по провайдеру
- `idx_usage_created_at` — временные выборки
- `idx_usage_tenant_provider` — составной для отчётов

---

### provider_pricing

Цены на токены и изображения по провайдерам и моделям.

```sql
CREATE TABLE provider_pricing (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider                 VARCHAR(50) NOT NULL,
    model                    VARCHAR(100) NOT NULL,
    input_price_per_million  DECIMAL(10, 6) NOT NULL DEFAULT 0,
    output_price_per_million DECIMAL(10, 6) NOT NULL DEFAULT 0,
    image_price              DECIMAL(10, 4) DEFAULT NULL,
    is_default               BOOLEAN NOT NULL DEFAULT false,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, model)
);
```

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | Первичный ключ |
| provider | VARCHAR(50) | Имя провайдера |
| model | VARCHAR(100) | Имя модели |
| input_price_per_million | DECIMAL(10,6) | Цена за 1M входных токенов (USD) |
| output_price_per_million | DECIMAL(10,6) | Цена за 1M выходных токенов (USD) |
| image_price | DECIMAL(10,4) | Цена за изображение (USD) |
| is_default | BOOLEAN | Модель по умолчанию для провайдера |
| created_at | TIMESTAMPTZ | Дата создания |
| updated_at | TIMESTAMPTZ | Дата обновления |

**Индексы:**
- `idx_provider_pricing_provider` — фильтрация по провайдеру
- `idx_provider_pricing_model` — составной для поиска цены

**Trigger:** `trigger_provider_pricing_updated_at` — автообновление `updated_at`

---

## Предустановленные данные

### Тестовый tenant

```sql
INSERT INTO tenants (id, name, api_key, active) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Default Tenant', 'dev-api-key-12345', true);
```

### Цены провайдеров

| Provider | Model | Input ($/1M) | Output ($/1M) | Image ($) |
|----------|-------|--------------|---------------|-----------|
| openai | gpt-4o-mini | 0.15 | 0.60 | — |
| openai | gpt-4o | 2.50 | 10.00 | — |
| openai | gpt-4-turbo | 10.00 | 30.00 | — |
| openai | gpt-3.5-turbo | 0.50 | 1.50 | — |
| openai | dall-e-3 | — | — | 0.04 |
| openai | dall-e-2 | — | — | 0.02 |
| claude | claude-3-haiku-20240307 | 0.25 | 1.25 | — |
| claude | claude-3-5-sonnet-20241022 | 3.00 | 15.00 | — |
| claude | claude-3-opus-20240229 | 15.00 | 75.00 | — |
| ollama | llama2, mistral, codellama | 0 | 0 | — |
| stub | stub-model | 0 | 0 | — |

---

## Связи

```
tenants (1) ─────< (N) jobs (1) ─────< (1) usage
    │
    └── api_key используется для аутентификации

provider_pricing ── не связана с другими таблицами (справочник)
providers ── не связана (legacy, конфигурация в коде)
```

---

## Миграции

| # | Файл | Описание |
|---|------|----------|
| 1 | 000001_create_tenants | Таблица tenants |
| 2 | 000002_create_jobs | Таблица jobs |
| 3 | 000003_create_providers | Таблица providers |
| 4 | 000004_create_usage | Таблица usage |
| 5 | 000005_seed_data | Тестовые данные |
| 6 | 000006_add_tenant_settings | Колонки default_provider, settings |
| 7 | 000007_create_provider_pricing | Таблица provider_pricing с ценами |

**Запуск миграций:**
```bash
./scripts/migrate.sh
# или
migrate -path infrastructure/postgres/migrations -database "$DATABASE_URL" up
```
