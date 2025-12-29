# API Users Management

> Управление API пользователями и ключами доступа для программного взаимодействия с AI Aggregator.

## Обзор

API Users — это сервисные аккаунты для программного доступа к AI Aggregator. Каждый API User может иметь несколько API ключей с различными правами доступа (scopes).

### Модель доступа

```
Tenant (организация)
  └── API Users (сервисные аккаунты)
        └── API Keys (ключи доступа с scopes)
```

---

## Аутентификация

Все запросы к Admin API требуют авторизации одним из способов:

### Session Token (Dashboard)

```http
Authorization: Bearer <session_token>
```

Получается после логина через GraphQL mutation `login`.

### API Key (Programmatic)

```http
X-API-Key: agg_xxxxxxxxxxxx
```

API ключ с scope `admin`.

---

## API Users

### Создать API User

Создаёт нового API пользователя в вашем tenant.

```http
POST /api/admin/users
```

#### Request Body

| Параметр | Тип | Обязательный | Описание |
|----------|-----|--------------|----------|
| `name` | string | ✅ | Имя пользователя (уникальное в рамках tenant) |
| `description` | string | ❌ | Описание назначения |

#### Пример запроса

```bash
curl -X POST http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer <session_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Backend",
    "description": "Backend service for production environment"
  }'
```

#### Успешный ответ `201 Created`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  "name": "Production Backend",
  "description": "Backend service for production environment",
  "active": true,
  "created_at": "2025-12-29T10:30:00Z",
  "updated_at": "2025-12-29T10:30:00Z"
}
```

#### Ошибки

| Код | Описание |
|-----|----------|
| `400` | Невалидные данные (отсутствует `name`) |
| `401` | Не авторизован |
| `403` | Недостаточно прав (требуется admin scope) |

---

### Получить список API Users

Возвращает всех API пользователей tenant.

```http
GET /api/admin/users
```

#### Пример запроса

```bash
curl http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer <session_token>"
```

#### Успешный ответ `200 OK`

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "name": "Production Backend",
    "description": "Backend service for production",
    "active": true,
    "created_at": "2025-12-29T10:30:00Z",
    "updated_at": "2025-12-29T10:30:00Z"
  },
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "name": "Staging Backend",
    "description": "Backend service for staging",
    "active": true,
    "created_at": "2025-12-28T15:00:00Z",
    "updated_at": "2025-12-28T15:00:00Z"
  }
]
```

---

## API Keys

### Создать API Key

Создаёт новый ключ для указанного API User.

> ⚠️ **Важно**: Ключ показывается только один раз! Сохраните его сразу.

```http
POST /api/admin/api-keys
```

#### Request Body

| Параметр | Тип | Обязательный | Описание |
|----------|-----|--------------|----------|
| `user_id` | string (UUID) | ✅ | ID API User |
| `name` | string | ❌ | Название ключа (по умолчанию "Default") |
| `scopes` | string[] | ❌ | Права доступа (по умолчанию `["read", "write"]`) |

#### Доступные Scopes

| Scope | Описание |
|-------|----------|
| `read` | Чтение данных (list jobs, view usage) |
| `write` | Создание и изменение ресурсов |
| `admin` | Полный доступ, включая управление пользователями |
| `*` | Все права |

#### Пример запроса

```bash
curl -X POST http://localhost:8080/api/admin/api-keys \
  -H "Authorization: Bearer <session_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Production Key",
    "scopes": ["read", "write"]
  }'
```

#### Успешный ответ `201 Created`

```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "key_prefix": "agg_abc123",
  "key": "agg_abc123xxxxxxxxxxxxxxxxxxxxxxxx",
  "name": "Production Key",
  "scopes": ["read", "write"],
  "active": true,
  "expires_at": null,
  "last_used_at": null,
  "usage_count": 0,
  "created_at": "2025-12-29T10:35:00Z",
  "revoked_at": null
}
```

> 🔒 Поле `key` содержит полный ключ — сохраните его сейчас!

#### Ошибки

| Код | Описание |
|-----|----------|
| `400` | Невалидный `user_id` |
| `401` | Не авторизован |
| `403` | Недостаточно прав |
| `404` | API User не найден |

---

### Получить список API Keys

Возвращает все ключи для указанного API User.

```http
GET /api/admin/users/{user_id}/api-keys
```

#### Path Parameters

| Параметр | Тип | Описание |
|----------|-----|----------|
| `user_id` | UUID | ID API User |

#### Пример запроса

```bash
curl http://localhost:8080/api/admin/users/550e8400-e29b-41d4-a716-446655440000/api-keys \
  -H "Authorization: Bearer <session_token>"
```

#### Успешный ответ `200 OK`

```json
[
  {
    "id": "770e8400-e29b-41d4-a716-446655440002",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "key_prefix": "agg_abc123",
    "name": "Production Key",
    "scopes": ["read", "write"],
    "active": true,
    "expires_at": null,
    "last_used_at": "2025-12-29T11:00:00Z",
    "usage_count": 42,
    "created_at": "2025-12-29T10:35:00Z",
    "revoked_at": null
  }
]
```

> 💡 Полный ключ не возвращается — только `key_prefix` для идентификации.

---

### Отозвать API Key

Деактивирует ключ. Отозванный ключ больше не может использоваться для аутентификации.

```http
DELETE /api/admin/api-keys/{id}
```

#### Path Parameters

| Параметр | Тип | Описание |
|----------|-----|----------|
| `id` | UUID | ID API Key |

#### Пример запроса

```bash
curl -X DELETE http://localhost:8080/api/admin/api-keys/770e8400-e29b-41d4-a716-446655440002 \
  -H "Authorization: Bearer <session_token>"
```

#### Успешный ответ `204 No Content`

Пустой ответ означает успешное удаление.

#### Ошибки

| Код | Описание |
|-----|----------|
| `400` | Невалидный ID ключа |
| `401` | Не авторизован |
| `403` | Недостаточно прав |
| `404` | Ключ не найден |

---

## Использование API Key

После создания ключа, используйте его для доступа к API:

### GraphQL запросы

```bash
curl -X POST http://localhost:8080/graphql \
  -H "X-API-Key: agg_abc123xxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ jobs(first: 10) { edges { node { id status } } } }"
  }'
```

### SSE Streaming

```bash
curl -N "http://localhost:8080/stream?provider=openai&model=gpt-4o-mini" \
  -H "X-API-Key: agg_abc123xxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello, world!"}'
```

---

## Модели данных

### API User

```typescript
interface APIUser {
  id: string;           // UUID
  tenant_id: string;    // UUID владельца tenant
  name: string;         // Уникальное имя
  description: string;  // Описание
  active: boolean;      // Активен ли пользователь
  created_at: string;   // ISO 8601 timestamp
  updated_at: string;   // ISO 8601 timestamp
}
```

### API Key

```typescript
interface APIKey {
  id: string;              // UUID
  user_id: string;         // UUID владельца API User
  key_prefix: string;      // Первые символы ключа для идентификации
  name: string;            // Название ключа
  scopes: string[];        // Права доступа
  active: boolean;         // Активен ли ключ
  expires_at?: string;     // ISO 8601, null = бессрочный
  last_used_at?: string;   // Последнее использование
  usage_count: number;     // Количество использований
  created_at: string;      // ISO 8601 timestamp
  revoked_at?: string;     // Дата отзыва, null = активен
}
```

---

## Безопасность

### Хранение ключей

- Ключи хешируются с использованием **HMAC-SHA256** перед сохранением в БД
- Полный ключ показывается только при создании
- Для идентификации используется `key_prefix`

### Rate Limiting

- **100 попыток аутентификации** в минуту на IP адрес
- При превышении лимита возвращается `429 Too Many Requests`

### Аудит

Все операции с ключами логируются:
- Создание ключа
- Использование ключа
- Отзыв ключа
- Неудачные попытки аутентификации

### Рекомендации

1. **Минимальные права** — давайте ключам только необходимые scopes
2. **Отдельные ключи** — используйте разные ключи для разных сред (dev/staging/prod)
3. **Ротация** — регулярно обновляйте ключи
4. **Мониторинг** — следите за `usage_count` и `last_used_at`
5. **Немедленный отзыв** — при компрометации сразу отзывайте ключ

---

## Примеры

### JavaScript/TypeScript

```typescript
const API_BASE = 'http://localhost:8080';

// Создание API User
async function createAPIUser(sessionToken: string, name: string) {
  const response = await fetch(`${API_BASE}/api/admin/users`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${sessionToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ name }),
  });
  return response.json();
}

// Создание API Key
async function createAPIKey(sessionToken: string, userId: string, name: string, scopes: string[]) {
  const response = await fetch(`${API_BASE}/api/admin/api-keys`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${sessionToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ user_id: userId, name, scopes }),
  });
  return response.json();
}
```

### Python

```python
import requests

API_BASE = 'http://localhost:8080'

def create_api_user(session_token: str, name: str) -> dict:
    response = requests.post(
        f'{API_BASE}/api/admin/users',
        headers={
            'Authorization': f'Bearer {session_token}',
            'Content-Type': 'application/json',
        },
        json={'name': name}
    )
    response.raise_for_status()
    return response.json()

def create_api_key(session_token: str, user_id: str, name: str, scopes: list) -> dict:
    response = requests.post(
        f'{API_BASE}/api/admin/api-keys',
        headers={
            'Authorization': f'Bearer {session_token}',
            'Content-Type': 'application/json',
        },
        json={'user_id': user_id, 'name': name, 'scopes': scopes}
    )
    response.raise_for_status()
    return response.json()
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

const apiBase = "http://localhost:8080"

func createAPIUser(sessionToken, name string) (map[string]interface{}, error) {
    body, _ := json.Marshal(map[string]string{"name": name})
    req, _ := http.NewRequest("POST", apiBase+"/api/admin/users", bytes.NewBuffer(body))
    req.Header.Set("Authorization", "Bearer "+sessionToken)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}
```

---

## Changelog

| Версия | Дата | Изменения |
|--------|------|-----------|
| 1.0.0 | 2025-12-29 | Первый релиз API Users |
