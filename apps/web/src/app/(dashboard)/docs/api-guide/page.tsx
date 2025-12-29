'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Check,
  Copy,
  Key,
  Zap,
  MessageSquare,
  Code2,
  Terminal,
  ChevronRight,
  Sparkles,
  Clock,
  Globe,
  Shield,
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

// Table of contents
function TableOfContents() {
  const items = [
    { id: 'quickstart', label: 'Быстрый старт', icon: Zap },
    { id: 'authentication', label: 'Аутентификация', icon: Key },
    { id: 'streaming', label: 'SSE Streaming', icon: Sparkles },
    { id: 'graphql', label: 'GraphQL API', icon: Code2 },
    { id: 'providers', label: 'Провайдеры', icon: Globe },
    { id: 'errors', label: 'Обработка ошибок', icon: Shield },
    { id: 'examples', label: 'Примеры', icon: Terminal },
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

export default function APIGuidePage() {
  return (
    <div className="container mx-auto py-8">
      <div className="grid gap-8 lg:grid-cols-[1fr_220px]">
        <div className="space-y-12">
          {/* Header */}
          <div className="space-y-4">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span>API Reference</span>
              <ChevronRight className="h-4 w-4" />
              <span>Getting Started</span>
            </div>
            <h1 className="text-4xl font-bold tracking-tight">API Integration Guide</h1>
            <p className="text-xl text-muted-foreground">
              Полное руководство по использованию AI Aggregator API с вашим API ключом.
            </p>
          </div>

          {/* Quick Start */}
          <section id="quickstart" className="scroll-mt-16 space-y-6">
            <h2 className="text-2xl font-bold">🚀 Быстрый старт</h2>
            <p className="text-muted-foreground">
              Получите ответ от AI за 30 секунд. Замените <code>YOUR_API_KEY</code> на ваш ключ.
            </p>

            <CodeBlock
              language="bash"
              title="Первый запрос"
              code={`curl -N "http://localhost:8080/stream?provider=openai&model=gpt-4o-mini" \\
  -H "X-API-Key: YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"prompt": "Привет! Расскажи о себе."}'`}
            />

            <Card className="border-green-500/30 bg-green-500/5">
              <CardContent className="flex items-start gap-3 pt-4">
                <Check className="h-5 w-5 text-green-500" />
                <div>
                  <p className="font-medium">Готово!</p>
                  <p className="text-sm text-muted-foreground">
                    Вы увидите потоковый ответ от GPT-4o-mini через Server-Sent Events.
                  </p>
                </div>
              </CardContent>
            </Card>
          </section>

          {/* Authentication */}
          <section id="authentication" className="scroll-mt-16 space-y-6">
            <h2 className="text-2xl font-bold">🔑 Аутентификация</h2>
            <p className="text-muted-foreground">
              Все запросы к API требуют API ключ. Передавайте его в заголовке <code>X-API-Key</code>.
            </p>

            <Card>
              <CardHeader>
                <CardTitle className="text-base">Формат ключа</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <CodeBlock code="agg_abc123xxxxxxxxxxxxxxxxxxxxxxxx" language="text" />
                <p className="text-sm text-muted-foreground">
                  Ключи начинаются с префикса <code>agg_</code> и содержат 32+ символа.
                </p>
              </CardContent>
            </Card>

            <div className="space-y-4">
              <h3 className="font-medium">Способы передачи ключа</h3>

              <Tabs defaultValue="header" className="w-full">
                <TabsList className="grid w-full grid-cols-3">
                  <TabsTrigger value="header">X-API-Key Header</TabsTrigger>
                  <TabsTrigger value="bearer">Bearer Token</TabsTrigger>
                  <TabsTrigger value="query">Query Parameter</TabsTrigger>
                </TabsList>
                <TabsContent value="header" className="mt-4">
                  <CodeBlock
                    language="http"
                    code={`GET /stream HTTP/1.1
Host: localhost:8080
X-API-Key: agg_abc123xxxxxxxxxxxxxxxxxxxxxxxx`}
                  />
                  <p className="mt-2 text-sm text-muted-foreground">
                    ✅ Рекомендуемый способ
                  </p>
                </TabsContent>
                <TabsContent value="bearer" className="mt-4">
                  <CodeBlock
                    language="http"
                    code={`GET /stream HTTP/1.1
Host: localhost:8080
Authorization: Bearer agg_abc123xxxxxxxxxxxxxxxxxxxxxxxx`}
                  />
                </TabsContent>
                <TabsContent value="query" className="mt-4">
                  <CodeBlock
                    language="http"
                    code={`GET /stream?api_key=agg_abc123xxxxxxxxxxxxxxxxxxxxxxxx HTTP/1.1
Host: localhost:8080`}
                  />
                  <p className="mt-2 text-sm text-muted-foreground">
                    ⚠️ Только для WebSocket соединений
                  </p>
                </TabsContent>
              </Tabs>
            </div>
          </section>

          {/* SSE Streaming */}
          <section id="streaming" className="scroll-mt-16 space-y-6">
            <h2 className="text-2xl font-bold">⚡ SSE Streaming</h2>
            <p className="text-muted-foreground">
              Получайте ответы AI в реальном времени через Server-Sent Events.
            </p>

            <Card>
              <CardHeader className="bg-muted/30">
                <div className="flex items-center gap-3">
                  <MethodBadge method="POST" />
                  <code className="text-sm font-medium">/stream</code>
                </div>
                <CardDescription>Streaming completion endpoint</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6 pt-6">
                <div className="space-y-4">
                  <h4 className="font-medium">Query Parameters</h4>
                  <div className="overflow-hidden rounded-lg border">
                    <table className="w-full text-sm">
                      <thead className="bg-muted/50">
                        <tr>
                          <th className="px-4 py-2 text-left font-medium">Параметр</th>
                          <th className="px-4 py-2 text-left font-medium">Тип</th>
                          <th className="px-4 py-2 text-left font-medium">Описание</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr className="border-t">
                          <td className="px-4 py-2">
                            <code className="rounded bg-muted px-1.5 py-0.5 text-xs">provider</code>
                          </td>
                          <td className="px-4 py-2 text-muted-foreground">string</td>
                          <td className="px-4 py-2 text-muted-foreground">
                            openai, claude, ollama
                          </td>
                        </tr>
                        <tr className="border-t">
                          <td className="px-4 py-2">
                            <code className="rounded bg-muted px-1.5 py-0.5 text-xs">model</code>
                          </td>
                          <td className="px-4 py-2 text-muted-foreground">string</td>
                          <td className="px-4 py-2 text-muted-foreground">
                            gpt-4o-mini, claude-3-haiku, llama3.2 и др.
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>

                <div className="space-y-4">
                  <h4 className="font-medium">Request Body</h4>
                  <CodeBlock
                    language="json"
                    code={`{
  "prompt": "Напиши короткое стихотворение о программировании",
  "system_prompt": "Ты - творческий AI ассистент",
  "max_tokens": 500,
  "temperature": 0.7
}`}
                  />
                </div>

                <div className="space-y-4">
                  <h4 className="font-medium">Response (SSE)</h4>
                  <CodeBlock
                    language="text"
                    code={`event: message
data: {"content": "В "}

event: message
data: {"content": "мире "}

event: message
data: {"content": "кода..."}

event: done
data: {"usage": {"prompt_tokens": 15, "completion_tokens": 42}}`}
                  />
                </div>
              </CardContent>
            </Card>

            <div className="space-y-4">
              <h3 className="font-medium">Полный пример с curl</h3>
              <CodeBlock
                language="bash"
                code={`curl -N "http://localhost:8080/stream?provider=openai&model=gpt-4o-mini" \\
  -H "X-API-Key: YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "prompt": "Объясни квантовые компьютеры простыми словами",
    "system_prompt": "Отвечай кратко и понятно",
    "max_tokens": 300,
    "temperature": 0.5
  }'`}
              />
            </div>
          </section>

          {/* GraphQL API */}
          <section id="graphql" className="scroll-mt-16 space-y-6">
            <h2 className="text-2xl font-bold">📊 GraphQL API</h2>
            <p className="text-muted-foreground">
              Используйте GraphQL для управления jobs, получения статистики и работы с историей.
            </p>

            <Card>
              <CardHeader className="bg-muted/30">
                <div className="flex items-center gap-3">
                  <MethodBadge method="POST" />
                  <code className="text-sm font-medium">/graphql</code>
                </div>
                <CardDescription>GraphQL endpoint</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6 pt-6">
                <div className="space-y-4">
                  <h4 className="font-medium">Создать Job</h4>
                  <CodeBlock
                    language="graphql"
                    code={`mutation {
  createJob(input: {
    type: TEXT
    input: "Напиши функцию сортировки на Python"
    model: "gpt-4o-mini"
    provider: "openai"
  }) {
    id
    status
    createdAt
  }
}`}
                  />
                </div>

                <div className="space-y-4">
                  <h4 className="font-medium">Получить список Jobs</h4>
                  <CodeBlock
                    language="graphql"
                    code={`query {
  jobs(first: 10) {
    edges {
      node {
        id
        status
        input
        output
        provider
        model
        createdAt
        completedAt
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}`}
                  />
                </div>

                <div className="space-y-4">
                  <h4 className="font-medium">Получить статистику использования</h4>
                  <CodeBlock
                    language="graphql"
                    code={`query {
  usageStats(period: THIS_MONTH) {
    totalTokens
    totalCost
    requestCount
    byProvider {
      provider
      tokens
      cost
      requests
    }
  }
}`}
                  />
                </div>

                <div className="space-y-4">
                  <h4 className="font-medium">Пример curl запроса</h4>
                  <CodeBlock
                    language="bash"
                    code={`curl -X POST http://localhost:8080/graphql \\
  -H "X-API-Key: YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "query": "query { jobs(first: 5) { edges { node { id status } } } }"
  }'`}
                  />
                </div>
              </CardContent>
            </Card>

            <Card className="border-blue-500/30 bg-blue-500/5">
              <CardContent className="flex items-start gap-3 pt-4">
                <Sparkles className="h-5 w-5 text-blue-500" />
                <div>
                  <p className="font-medium">GraphQL Playground</p>
                  <p className="text-sm text-muted-foreground">
                    Откройте <code>http://localhost:8080/playground</code> для интерактивного тестирования.
                  </p>
                </div>
              </CardContent>
            </Card>
          </section>

          {/* Providers */}
          <section id="providers" className="scroll-mt-16 space-y-6">
            <h2 className="text-2xl font-bold">🌐 Провайдеры</h2>
            <p className="text-muted-foreground">
              AI Aggregator поддерживает несколько AI провайдеров с единым API.
            </p>

            <div className="grid gap-4 md:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">OpenAI</CardTitle>
                  <CardDescription>GPT-4, GPT-4o, GPT-3.5</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Provider</span>
                      <code>openai</code>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Models</span>
                      <span>gpt-4o, gpt-4o-mini</span>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Claude</CardTitle>
                  <CardDescription>Anthropic Claude 3</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Provider</span>
                      <code>claude</code>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Models</span>
                      <span>opus, sonnet, haiku</span>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Ollama</CardTitle>
                  <CardDescription>Локальные модели</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Provider</span>
                      <code>ollama</code>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Models</span>
                      <span>llama3.2, mistral</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>

            <div className="space-y-4">
              <h4 className="font-medium">Получить список доступных моделей</h4>
              <CodeBlock
                language="graphql"
                code={`query {
  providerModels(provider: "openai") {
    id
    name
    description
  }
}`}
              />
            </div>
          </section>

          {/* Errors */}
          <section id="errors" className="scroll-mt-16 space-y-6">
            <h2 className="text-2xl font-bold">⚠️ Обработка ошибок</h2>
            <p className="text-muted-foreground">
              API возвращает стандартные HTTP коды статуса и JSON с описанием ошибки.
            </p>

            <div className="overflow-hidden rounded-lg border">
              <table className="w-full text-sm">
                <thead className="bg-muted/50">
                  <tr>
                    <th className="px-4 py-2 text-left font-medium">Код</th>
                    <th className="px-4 py-2 text-left font-medium">Описание</th>
                    <th className="px-4 py-2 text-left font-medium">Решение</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-t">
                    <td className="px-4 py-2">
                      <Badge variant="outline" className="text-yellow-500">401</Badge>
                    </td>
                    <td className="px-4 py-2">Unauthorized</td>
                    <td className="px-4 py-2 text-muted-foreground">Проверьте API ключ</td>
                  </tr>
                  <tr className="border-t">
                    <td className="px-4 py-2">
                      <Badge variant="outline" className="text-red-500">403</Badge>
                    </td>
                    <td className="px-4 py-2">Forbidden</td>
                    <td className="px-4 py-2 text-muted-foreground">Недостаточно прав (scope)</td>
                  </tr>
                  <tr className="border-t">
                    <td className="px-4 py-2">
                      <Badge variant="outline" className="text-orange-500">429</Badge>
                    </td>
                    <td className="px-4 py-2">Too Many Requests</td>
                    <td className="px-4 py-2 text-muted-foreground">Подождите и повторите</td>
                  </tr>
                  <tr className="border-t">
                    <td className="px-4 py-2">
                      <Badge variant="outline" className="text-red-500">500</Badge>
                    </td>
                    <td className="px-4 py-2">Internal Error</td>
                    <td className="px-4 py-2 text-muted-foreground">Свяжитесь с поддержкой</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <CodeBlock
              language="json"
              title="Пример ошибки"
              code={`{
  "error": "unauthorized",
  "message": "Invalid API key"
}`}
            />
          </section>

          {/* Examples */}
          <section id="examples" className="scroll-mt-16 space-y-6">
            <h2 className="text-2xl font-bold">💻 Примеры кода</h2>

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

              <TabsContent value="javascript" className="mt-4 space-y-4">
                <h4 className="font-medium">SSE Streaming с fetch</h4>
                <CodeBlock
                  language="javascript"
                  code={`const API_KEY = 'agg_your_api_key_here';
const API_BASE = 'http://localhost:8080';

async function streamCompletion(prompt) {
  const response = await fetch(
    \`\${API_BASE}/stream?provider=openai&model=gpt-4o-mini\`,
    {
      method: 'POST',
      headers: {
        'X-API-Key': API_KEY,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ prompt }),
    }
  );

  const reader = response.body.getReader();
  const decoder = new TextDecoder();

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    const chunk = decoder.decode(value);
    const lines = chunk.split('\\n');

    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const data = JSON.parse(line.slice(6));
        if (data.content) {
          process.stdout.write(data.content);
        }
      }
    }
  }
}

streamCompletion('Объясни, что такое API');`}
                />

                <h4 className="font-medium">GraphQL запрос</h4>
                <CodeBlock
                  language="javascript"
                  code={`async function getJobs() {
  const response = await fetch(\`\${API_BASE}/graphql\`, {
    method: 'POST',
    headers: {
      'X-API-Key': API_KEY,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      query: \`
        query {
          jobs(first: 10) {
            edges {
              node { id status input output }
            }
          }
        }
      \`,
    }),
  });

  const { data } = await response.json();
  return data.jobs.edges.map(e => e.node);
}`}
                />
              </TabsContent>

              <TabsContent value="python" className="mt-4 space-y-4">
                <h4 className="font-medium">SSE Streaming</h4>
                <CodeBlock
                  language="python"
                  code={`import requests
import json

API_KEY = 'agg_your_api_key_here'
API_BASE = 'http://localhost:8080'

def stream_completion(prompt: str):
    response = requests.post(
        f'{API_BASE}/stream?provider=openai&model=gpt-4o-mini',
        headers={
            'X-API-Key': API_KEY,
            'Content-Type': 'application/json',
        },
        json={'prompt': prompt},
        stream=True
    )

    for line in response.iter_lines():
        if line:
            line = line.decode('utf-8')
            if line.startswith('data: '):
                data = json.loads(line[6:])
                if 'content' in data:
                    print(data['content'], end='', flush=True)

stream_completion('Напиши haiku о Python')`}
                />

                <h4 className="font-medium">GraphQL с requests</h4>
                <CodeBlock
                  language="python"
                  code={`def get_jobs():
    response = requests.post(
        f'{API_BASE}/graphql',
        headers={
            'X-API-Key': API_KEY,
            'Content-Type': 'application/json',
        },
        json={
            'query': '''
                query {
                    jobs(first: 10) {
                        edges {
                            node { id status input output }
                        }
                    }
                }
            '''
        }
    )
    data = response.json()['data']
    return [edge['node'] for edge in data['jobs']['edges']]`}
                />
              </TabsContent>

              <TabsContent value="go" className="mt-4 space-y-4">
                <h4 className="font-medium">SSE Streaming</h4>
                <CodeBlock
                  language="go"
                  code={`package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

const (
    apiKey  = "agg_your_api_key_here"
    apiBase = "http://localhost:8080"
)

func streamCompletion(prompt string) error {
    body, _ := json.Marshal(map[string]string{"prompt": prompt})

    req, _ := http.NewRequest("POST",
        apiBase+"/stream?provider=openai&model=gpt-4o-mini",
        bytes.NewBuffer(body))
    req.Header.Set("X-API-Key", apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "data: ") {
            var data map[string]interface{}
            json.Unmarshal([]byte(line[6:]), &data)
            if content, ok := data["content"].(string); ok {
                fmt.Print(content)
            }
        }
    }
    return nil
}

func main() {
    streamCompletion("Напиши функцию сортировки")
}`}
                />
              </TabsContent>
            </Tabs>
          </section>

          {/* Rate Limits */}
          <section className="scroll-mt-16 space-y-6">
            <h2 className="text-2xl font-bold">⏱️ Rate Limits</h2>

            <div className="grid gap-4 md:grid-cols-3">
              <Card>
                <CardContent className="flex items-center gap-3 pt-6">
                  <div className="rounded-lg bg-blue-500/10 p-2">
                    <Clock className="h-5 w-5 text-blue-500" />
                  </div>
                  <div>
                    <p className="font-medium">100 req/min</p>
                    <p className="text-sm text-muted-foreground">Аутентификация</p>
                  </div>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="flex items-center gap-3 pt-6">
                  <div className="rounded-lg bg-green-500/10 p-2">
                    <MessageSquare className="h-5 w-5 text-green-500" />
                  </div>
                  <div>
                    <p className="font-medium">По плану</p>
                    <p className="text-sm text-muted-foreground">API запросы</p>
                  </div>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="flex items-center gap-3 pt-6">
                  <div className="rounded-lg bg-purple-500/10 p-2">
                    <Zap className="h-5 w-5 text-purple-500" />
                  </div>
                  <div>
                    <p className="font-medium">Retry-After</p>
                    <p className="text-sm text-muted-foreground">Header в 429</p>
                  </div>
                </CardContent>
              </Card>
            </div>
          </section>
        </div>

        {/* Table of Contents */}
        <TableOfContents />
      </div>
    </div>
  )
}
