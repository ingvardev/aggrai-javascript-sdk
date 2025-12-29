'use client'

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Check,
  Copy,
  Key,
  Lock,
  Shield,
  Users,
  AlertTriangle,
  ChevronRight,
  Code2,
  Terminal,
} from 'lucide-react'
import { cn } from '@/lib/utils'

// Code block with copy functionality
function CodeBlock({
  code,
  language = 'bash',
  title,
}: {
  code: string
  language?: string
  title?: string
}) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="group relative rounded-lg border bg-zinc-950 dark:bg-zinc-900">
      {title && (
        <div className="flex items-center justify-between border-b bg-zinc-900/50 px-4 py-2">
          <span className="text-xs font-medium text-zinc-400">{title}</span>
          <Badge variant="outline" className="text-xs">
            {language}
          </Badge>
        </div>
      )}
      <div className="relative">
        <pre className="overflow-x-auto p-4 text-sm">
          <code className="text-zinc-100">{code}</code>
        </pre>
        <Button
          size="icon"
          variant="ghost"
          className="absolute right-2 top-2 h-8 w-8 opacity-0 transition-opacity group-hover:opacity-100"
          onClick={handleCopy}
        >
          {copied ? (
            <Check className="h-4 w-4 text-green-500" />
          ) : (
            <Copy className="h-4 w-4 text-zinc-400" />
          )}
        </Button>
      </div>
    </div>
  )
}

// HTTP method badge
function MethodBadge({ method }: { method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' }) {
  const colors = {
    GET: 'bg-green-500/10 text-green-500 border-green-500/20',
    POST: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
    PUT: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
    DELETE: 'bg-red-500/10 text-red-500 border-red-500/20',
    PATCH: 'bg-purple-500/10 text-purple-500 border-purple-500/20',
  }

  return (
    <Badge variant="outline" className={cn('font-mono text-xs font-bold', colors[method])}>
      {method}
    </Badge>
  )
}

// Endpoint card
function EndpointCard({
  method,
  path,
  title,
  description,
  children,
}: {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  path: string
  title: string
  description: string
  children: React.ReactNode
}) {
  return (
    <Card className="overflow-hidden">
      <CardHeader className="bg-muted/30 pb-4">
        <div className="flex items-center gap-3">
          <MethodBadge method={method} />
          <code className="text-sm font-medium">{path}</code>
        </div>
        <CardTitle className="mt-3 text-lg">{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6 pt-6">{children}</CardContent>
    </Card>
  )
}

// Parameter table
function ParamTable({
  params,
}: {
  params: { name: string; type: string; required: boolean; description: string }[]
}) {
  return (
    <div className="overflow-hidden rounded-lg border">
      <table className="w-full text-sm">
        <thead className="bg-muted/50">
          <tr>
            <th className="px-4 py-2 text-left font-medium">Параметр</th>
            <th className="px-4 py-2 text-left font-medium">Тип</th>
            <th className="px-4 py-2 text-left font-medium">Обязательный</th>
            <th className="px-4 py-2 text-left font-medium">Описание</th>
          </tr>
        </thead>
        <tbody>
          {params.map((param) => (
            <tr key={param.name} className="border-t">
              <td className="px-4 py-2">
                <code className="rounded bg-muted px-1.5 py-0.5 text-xs">{param.name}</code>
              </td>
              <td className="px-4 py-2 text-muted-foreground">{param.type}</td>
              <td className="px-4 py-2">
                {param.required ? (
                  <Badge variant="default" className="bg-green-500">
                    Да
                  </Badge>
                ) : (
                  <Badge variant="secondary">Нет</Badge>
                )}
              </td>
              <td className="px-4 py-2 text-muted-foreground">{param.description}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// Response example
function ResponseExample({ status, body }: { status: number; body: string }) {
  const statusColors: Record<number, string> = {
    200: 'text-green-500',
    201: 'text-green-500',
    204: 'text-green-500',
    400: 'text-yellow-500',
    401: 'text-red-500',
    403: 'text-red-500',
    404: 'text-red-500',
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <span className={cn('font-mono text-sm font-bold', statusColors[status] || 'text-muted-foreground')}>
          {status}
        </span>
        <span className="text-sm text-muted-foreground">
          {status === 200 && 'OK'}
          {status === 201 && 'Created'}
          {status === 204 && 'No Content'}
          {status === 400 && 'Bad Request'}
          {status === 401 && 'Unauthorized'}
          {status === 403 && 'Forbidden'}
          {status === 404 && 'Not Found'}
        </span>
      </div>
      {body && <CodeBlock code={body} language="json" />}
    </div>
  )
}

// Table of contents
function TableOfContents() {
  const items = [
    { id: 'overview', label: 'Обзор', icon: Users },
    { id: 'auth', label: 'Аутентификация', icon: Lock },
    { id: 'create-user', label: 'Создать API User', icon: Users },
    { id: 'list-users', label: 'Список API Users', icon: Users },
    { id: 'create-key', label: 'Создать API Key', icon: Key },
    { id: 'list-keys', label: 'Список API Keys', icon: Key },
    { id: 'revoke-key', label: 'Отозвать API Key', icon: Key },
    { id: 'security', label: 'Безопасность', icon: Shield },
  ]

  return (
    <nav className="sticky top-4 hidden lg:block">
      <div className="space-y-1">
        <p className="mb-3 text-sm font-medium">Содержание</p>
        {items.map((item) => (
          <a
            key={item.id}
            href={`#${item.id}`}
            className="flex items-center gap-2 rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            <item.icon className="h-4 w-4" />
            {item.label}
          </a>
        ))}
      </div>
    </nav>
  )
}

export default function APIUsersDocsPage() {
  const { t } = useTranslation()

  return (
    <div className="container mx-auto py-8">
      <div className="grid gap-8 lg:grid-cols-[1fr_220px]">
        <div className="space-y-12">
          {/* Header */}
          <div className="space-y-4">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span>API Reference</span>
              <ChevronRight className="h-4 w-4" />
              <span>API Users</span>
            </div>
            <h1 className="text-4xl font-bold tracking-tight">API Users Management</h1>
            <p className="text-xl text-muted-foreground">
              Управление API пользователями и ключами доступа для программного взаимодействия с AI
              Aggregator.
            </p>
          </div>

          {/* Overview */}
          <section id="overview" className="scroll-mt-16 space-y-6">
            <h2 className="text-2xl font-bold">Обзор</h2>
            <p className="text-muted-foreground">
              API Users — это сервисные аккаунты для программного доступа к AI Aggregator. Каждый
              API User может иметь несколько API ключей с различными правами доступа (scopes).
            </p>

            <Card className="bg-muted/30">
              <CardContent className="pt-6">
                <div className="flex items-start gap-4">
                  <div className="rounded-lg bg-primary/10 p-3">
                    <Users className="h-6 w-6 text-primary" />
                  </div>
                  <div className="space-y-2">
                    <h3 className="font-semibold">Модель доступа</h3>
                    <div className="font-mono text-sm text-muted-foreground">
                      <div>Tenant (организация)</div>
                      <div className="ml-4">└── API Users (сервисные аккаунты)</div>
                      <div className="ml-8">└── API Keys (ключи доступа с scopes)</div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            <div className="grid gap-4 md:grid-cols-3">
              <Card>
                <CardContent className="flex items-center gap-3 pt-6">
                  <div className="rounded-lg bg-green-500/10 p-2">
                    <Check className="h-5 w-5 text-green-500" />
                  </div>
                  <div>
                    <p className="font-medium">Множество ключей</p>
                    <p className="text-sm text-muted-foreground">Один user — много keys</p>
                  </div>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="flex items-center gap-3 pt-6">
                  <div className="rounded-lg bg-blue-500/10 p-2">
                    <Shield className="h-5 w-5 text-blue-500" />
                  </div>
                  <div>
                    <p className="font-medium">Granular scopes</p>
                    <p className="text-sm text-muted-foreground">Точный контроль прав</p>
                  </div>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="flex items-center gap-3 pt-6">
                  <div className="rounded-lg bg-purple-500/10 p-2">
                    <Key className="h-5 w-5 text-purple-500" />
                  </div>
                  <div>
                    <p className="font-medium">Безопасность</p>
                    <p className="text-sm text-muted-foreground">HMAC-SHA256 хеширование</p>
                  </div>
                </CardContent>
              </Card>
            </div>
          </section>

          {/* Authentication */}
          <section id="auth" className="scroll-mt-16 space-y-6">
            <h2 className="text-2xl font-bold">Аутентификация</h2>
            <p className="text-muted-foreground">
              Все запросы к Admin API требуют авторизации одним из способов:
            </p>

            <Tabs defaultValue="session" className="w-full">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="session">Session Token (Dashboard)</TabsTrigger>
                <TabsTrigger value="apikey">API Key (Programmatic)</TabsTrigger>
              </TabsList>
              <TabsContent value="session" className="mt-4">
                <Card>
                  <CardContent className="pt-6">
                    <p className="mb-4 text-sm text-muted-foreground">
                      Получается после логина через GraphQL mutation <code>login</code>.
                    </p>
                    <CodeBlock
                      code="Authorization: Bearer <session_token>"
                      language="http"
                      title="Header"
                    />
                  </CardContent>
                </Card>
              </TabsContent>
              <TabsContent value="apikey" className="mt-4">
                <Card>
                  <CardContent className="pt-6">
                    <p className="mb-4 text-sm text-muted-foreground">
                      API ключ с scope <code>admin</code>.
                    </p>
                    <CodeBlock
                      code="X-API-Key: agg_xxxxxxxxxxxx"
                      language="http"
                      title="Header"
                    />
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          </section>

          {/* Create User */}
          <section id="create-user" className="scroll-mt-16">
            <EndpointCard
              method="POST"
              path="/api/admin/users"
              title="Создать API User"
              description="Создаёт нового API пользователя в вашем tenant."
            >
              <div className="space-y-4">
                <h4 className="font-medium">Request Body</h4>
                <ParamTable
                  params={[
                    {
                      name: 'name',
                      type: 'string',
                      required: true,
                      description: 'Имя пользователя (уникальное в рамках tenant)',
                    },
                    {
                      name: 'description',
                      type: 'string',
                      required: false,
                      description: 'Описание назначения',
                    },
                  ]}
                />
              </div>

              <div className="space-y-4">
                <h4 className="font-medium">Пример запроса</h4>
                <CodeBlock
                  language="bash"
                  code={`curl -X POST http://localhost:8080/api/admin/users \\
  -H "Authorization: Bearer <session_token>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "Production Backend",
    "description": "Backend service for production"
  }'`}
                />
              </div>

              <div className="space-y-4">
                <h4 className="font-medium">Response</h4>
                <ResponseExample
                  status={201}
                  body={`{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  "name": "Production Backend",
  "description": "Backend service for production",
  "active": true,
  "created_at": "2025-12-29T10:30:00Z",
  "updated_at": "2025-12-29T10:30:00Z"
}`}
                />
              </div>
            </EndpointCard>
          </section>

          {/* List Users */}
          <section id="list-users" className="scroll-mt-16">
            <EndpointCard
              method="GET"
              path="/api/admin/users"
              title="Получить список API Users"
              description="Возвращает всех API пользователей tenant."
            >
              <div className="space-y-4">
                <h4 className="font-medium">Пример запроса</h4>
                <CodeBlock
                  language="bash"
                  code={`curl http://localhost:8080/api/admin/users \\
  -H "Authorization: Bearer <session_token>"`}
                />
              </div>

              <div className="space-y-4">
                <h4 className="font-medium">Response</h4>
                <ResponseExample
                  status={200}
                  body={`[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "name": "Production Backend",
    "description": "Backend service for production",
    "active": true,
    "created_at": "2025-12-29T10:30:00Z",
    "updated_at": "2025-12-29T10:30:00Z"
  }
]`}
                />
              </div>
            </EndpointCard>
          </section>

          {/* Create Key */}
          <section id="create-key" className="scroll-mt-16">
            <EndpointCard
              method="POST"
              path="/api/admin/api-keys"
              title="Создать API Key"
              description="Создаёт новый ключ для указанного API User."
            >
              <Card className="border-yellow-500/50 bg-yellow-500/10">
                <CardContent className="flex items-start gap-3 pt-4">
                  <AlertTriangle className="h-5 w-5 text-yellow-500" />
                  <p className="text-sm">
                    <strong>Важно:</strong> Ключ показывается только один раз! Сохраните его сразу.
                  </p>
                </CardContent>
              </Card>

              <div className="space-y-4">
                <h4 className="font-medium">Request Body</h4>
                <ParamTable
                  params={[
                    {
                      name: 'user_id',
                      type: 'string (UUID)',
                      required: true,
                      description: 'ID API User',
                    },
                    {
                      name: 'name',
                      type: 'string',
                      required: false,
                      description: 'Название ключа (по умолчанию "Default")',
                    },
                    {
                      name: 'scopes',
                      type: 'string[]',
                      required: false,
                      description: 'Права доступа (по умолчанию ["read", "write"])',
                    },
                  ]}
                />
              </div>

              <div className="space-y-4">
                <h4 className="font-medium">Доступные Scopes</h4>
                <div className="grid gap-2 sm:grid-cols-2">
                  <div className="flex items-center gap-2 rounded-lg border p-3">
                    <Badge variant="outline">read</Badge>
                    <span className="text-sm text-muted-foreground">Чтение данных</span>
                  </div>
                  <div className="flex items-center gap-2 rounded-lg border p-3">
                    <Badge variant="outline">write</Badge>
                    <span className="text-sm text-muted-foreground">Создание и изменение</span>
                  </div>
                  <div className="flex items-center gap-2 rounded-lg border p-3">
                    <Badge variant="outline">admin</Badge>
                    <span className="text-sm text-muted-foreground">Полный доступ</span>
                  </div>
                  <div className="flex items-center gap-2 rounded-lg border p-3">
                    <Badge variant="outline">*</Badge>
                    <span className="text-sm text-muted-foreground">Все права</span>
                  </div>
                </div>
              </div>

              <div className="space-y-4">
                <h4 className="font-medium">Пример запроса</h4>
                <CodeBlock
                  language="bash"
                  code={`curl -X POST http://localhost:8080/api/admin/api-keys \\
  -H "Authorization: Bearer <session_token>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Production Key",
    "scopes": ["read", "write"]
  }'`}
                />
              </div>

              <div className="space-y-4">
                <h4 className="font-medium">Response</h4>
                <ResponseExample
                  status={201}
                  body={`{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "key_prefix": "agg_abc123",
  "key": "agg_abc123xxxxxxxxxxxxxxxxxxxxxxxx",
  "name": "Production Key",
  "scopes": ["read", "write"],
  "active": true,
  "created_at": "2025-12-29T10:35:00Z"
}`}
                />
                <p className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Lock className="h-4 w-4" />
                  Поле <code>key</code> содержит полный ключ — сохраните его сейчас!
                </p>
              </div>
            </EndpointCard>
          </section>

          {/* List Keys */}
          <section id="list-keys" className="scroll-mt-16">
            <EndpointCard
              method="GET"
              path="/api/admin/users/{user_id}/api-keys"
              title="Получить список API Keys"
              description="Возвращает все ключи для указанного API User."
            >
              <div className="space-y-4">
                <h4 className="font-medium">Path Parameters</h4>
                <ParamTable
                  params={[
                    { name: 'user_id', type: 'UUID', required: true, description: 'ID API User' },
                  ]}
                />
              </div>

              <div className="space-y-4">
                <h4 className="font-medium">Пример запроса</h4>
                <CodeBlock
                  language="bash"
                  code={`curl http://localhost:8080/api/admin/users/550e8400-e29b-41d4-a716-446655440000/api-keys \\
  -H "Authorization: Bearer <session_token>"`}
                />
              </div>

              <div className="space-y-4">
                <h4 className="font-medium">Response</h4>
                <ResponseExample
                  status={200}
                  body={`[
  {
    "id": "770e8400-e29b-41d4-a716-446655440002",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "key_prefix": "agg_abc123",
    "name": "Production Key",
    "scopes": ["read", "write"],
    "active": true,
    "last_used_at": "2025-12-29T11:00:00Z",
    "usage_count": 42,
    "created_at": "2025-12-29T10:35:00Z"
  }
]`}
                />
                <p className="text-sm text-muted-foreground">
                  💡 Полный ключ не возвращается — только <code>key_prefix</code> для
                  идентификации.
                </p>
              </div>
            </EndpointCard>
          </section>

          {/* Revoke Key */}
          <section id="revoke-key" className="scroll-mt-16">
            <EndpointCard
              method="DELETE"
              path="/api/admin/api-keys/{id}"
              title="Отозвать API Key"
              description="Деактивирует ключ. Отозванный ключ больше не может использоваться."
            >
              <div className="space-y-4">
                <h4 className="font-medium">Path Parameters</h4>
                <ParamTable
                  params={[{ name: 'id', type: 'UUID', required: true, description: 'ID API Key' }]}
                />
              </div>

              <div className="space-y-4">
                <h4 className="font-medium">Пример запроса</h4>
                <CodeBlock
                  language="bash"
                  code={`curl -X DELETE http://localhost:8080/api/admin/api-keys/770e8400-e29b-41d4-a716-446655440002 \\
  -H "Authorization: Bearer <session_token>"`}
                />
              </div>

              <div className="space-y-4">
                <h4 className="font-medium">Response</h4>
                <ResponseExample status={204} body="" />
                <p className="text-sm text-muted-foreground">
                  Пустой ответ означает успешное удаление.
                </p>
              </div>
            </EndpointCard>
          </section>

          {/* Security */}
          <section id="security" className="scroll-mt-16 space-y-6">
            <h2 className="text-2xl font-bold">Безопасность</h2>

            <div className="grid gap-4 md:grid-cols-2">
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <Lock className="h-5 w-5 text-primary" />
                    <CardTitle className="text-base">Хранение ключей</CardTitle>
                  </div>
                </CardHeader>
                <CardContent className="text-sm text-muted-foreground">
                  <ul className="list-inside list-disc space-y-1">
                    <li>HMAC-SHA256 хеширование</li>
                    <li>Полный ключ показывается только при создании</li>
                    <li>Для идентификации используется key_prefix</li>
                  </ul>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <Shield className="h-5 w-5 text-primary" />
                    <CardTitle className="text-base">Rate Limiting</CardTitle>
                  </div>
                </CardHeader>
                <CardContent className="text-sm text-muted-foreground">
                  <ul className="list-inside list-disc space-y-1">
                    <li>100 попыток аутентификации в минуту на IP</li>
                    <li>429 Too Many Requests при превышении</li>
                  </ul>
                </CardContent>
              </Card>
            </div>

            <Card className="border-primary/20 bg-primary/5">
              <CardHeader>
                <CardTitle className="text-base">Рекомендации</CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="grid gap-2 text-sm md:grid-cols-2">
                  <li className="flex items-start gap-2">
                    <Check className="mt-0.5 h-4 w-4 text-green-500" />
                    <span>Давайте ключам только необходимые scopes</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <Check className="mt-0.5 h-4 w-4 text-green-500" />
                    <span>Используйте разные ключи для разных сред</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <Check className="mt-0.5 h-4 w-4 text-green-500" />
                    <span>Регулярно ротируйте ключи</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <Check className="mt-0.5 h-4 w-4 text-green-500" />
                    <span>При компрометации сразу отзывайте ключ</span>
                  </li>
                </ul>
              </CardContent>
            </Card>
          </section>

          {/* Code Examples */}
          <section className="space-y-6">
            <h2 className="text-2xl font-bold">Примеры кода</h2>

            <Tabs defaultValue="javascript" className="w-full">
              <TabsList>
                <TabsTrigger value="javascript" className="gap-2">
                  <Code2 className="h-4 w-4" />
                  JavaScript
                </TabsTrigger>
                <TabsTrigger value="python" className="gap-2">
                  <Terminal className="h-4 w-4" />
                  Python
                </TabsTrigger>
                <TabsTrigger value="go" className="gap-2">
                  <Terminal className="h-4 w-4" />
                  Go
                </TabsTrigger>
              </TabsList>

              <TabsContent value="javascript" className="mt-4">
                <CodeBlock
                  language="javascript"
                  code={`const API_BASE = 'http://localhost:8080';

// Создание API User
async function createAPIUser(sessionToken, name) {
  const response = await fetch(\`\${API_BASE}/api/admin/users\`, {
    method: 'POST',
    headers: {
      'Authorization': \`Bearer \${sessionToken}\`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ name }),
  });
  return response.json();
}

// Создание API Key
async function createAPIKey(sessionToken, userId, name, scopes) {
  const response = await fetch(\`\${API_BASE}/api/admin/api-keys\`, {
    method: 'POST',
    headers: {
      'Authorization': \`Bearer \${sessionToken}\`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ user_id: userId, name, scopes }),
  });
  return response.json();
}`}
                />
              </TabsContent>

              <TabsContent value="python" className="mt-4">
                <CodeBlock
                  language="python"
                  code={`import requests

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

def create_api_key(session_token: str, user_id: str,
                   name: str, scopes: list) -> dict:
    response = requests.post(
        f'{API_BASE}/api/admin/api-keys',
        headers={
            'Authorization': f'Bearer {session_token}',
            'Content-Type': 'application/json',
        },
        json={'user_id': user_id, 'name': name, 'scopes': scopes}
    )
    response.raise_for_status()
    return response.json()`}
                />
              </TabsContent>

              <TabsContent value="go" className="mt-4">
                <CodeBlock
                  language="go"
                  code={`package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

const apiBase = "http://localhost:8080"

func createAPIUser(sessionToken, name string) (map[string]interface{}, error) {
    body, _ := json.Marshal(map[string]string{"name": name})

    req, _ := http.NewRequest("POST",
        apiBase+"/api/admin/users", bytes.NewBuffer(body))
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
}`}
                />
              </TabsContent>
            </Tabs>
          </section>
        </div>

        {/* Table of Contents */}
        <TableOfContents />
      </div>
    </div>
  )
}
