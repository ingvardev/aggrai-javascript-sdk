import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  BookOpen,
  Code,
  Zap,
  Key,
  Cpu,
  BarChart3,
  Settings,
  ExternalLink
} from 'lucide-react'

export default function DocsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Documentation</h1>
        <p className="text-muted-foreground">
          Learn how to use the AI Aggregator API
        </p>
      </div>

      <Tabs defaultValue="quickstart" className="space-y-6">
        <TabsList>
          <TabsTrigger value="quickstart">Quick Start</TabsTrigger>
          <TabsTrigger value="api">API Reference</TabsTrigger>
          <TabsTrigger value="providers">Providers</TabsTrigger>
          <TabsTrigger value="examples">Examples</TabsTrigger>
        </TabsList>

        {/* Quick Start */}
        <TabsContent value="quickstart" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Zap className="h-5 w-5" />
                Getting Started
              </CardTitle>
              <CardDescription>
                Get up and running with AI Aggregator in minutes
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <div className="flex items-start gap-4">
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground text-sm font-bold">
                    1
                  </div>
                  <div>
                    <h3 className="font-medium">Get your API Key</h3>
                    <p className="text-sm text-muted-foreground mt-1">
                      Navigate to Settings → API Keys to generate a new API key for authentication.
                    </p>
                  </div>
                </div>

                <div className="flex items-start gap-4">
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground text-sm font-bold">
                    2
                  </div>
                  <div>
                    <h3 className="font-medium">Make your first request</h3>
                    <p className="text-sm text-muted-foreground mt-1">
                      Use the GraphQL API to create a job:
                    </p>
                    <pre className="mt-2 rounded-lg bg-muted p-4 text-sm overflow-x-auto">
{`curl -X POST http://localhost:8080/graphql \\
  -H "Content-Type: application/json" \\
  -H "X-API-Key: your-api-key" \\
  -d '{
    "query": "mutation { createJob(input: {type: TEXT, input: \\"Hello AI\\"}) { id status } }"
  }'`}
                    </pre>
                  </div>
                </div>

                <div className="flex items-start gap-4">
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground text-sm font-bold">
                    3
                  </div>
                  <div>
                    <h3 className="font-medium">Check job status</h3>
                    <p className="text-sm text-muted-foreground mt-1">
                      Poll for job completion or use GraphQL subscriptions for real-time updates:
                    </p>
                    <pre className="mt-2 rounded-lg bg-muted p-4 text-sm overflow-x-auto">
{`subscription {
  jobUpdated {
    id
    status
    result
  }
}`}
                    </pre>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Key className="h-5 w-5" />
                Authentication
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                All API requests require authentication using an API key. Include the key in the <code className="rounded bg-muted px-1.5 py-0.5">X-API-Key</code> header:
              </p>
              <pre className="rounded-lg bg-muted p-4 text-sm overflow-x-auto">
{`curl -H "X-API-Key: your-api-key" \\
  http://localhost:8080/graphql`}
              </pre>
            </CardContent>
          </Card>
        </TabsContent>

        {/* API Reference */}
        <TabsContent value="api" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Code className="h-5 w-5" />
                GraphQL API
              </CardTitle>
              <CardDescription>
                The AI Aggregator uses GraphQL for all API operations
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div>
                <h3 className="font-medium mb-2">Endpoint</h3>
                <code className="rounded bg-muted px-3 py-2 text-sm block">
                  POST http://localhost:8080/graphql
                </code>
              </div>

              <div>
                <h3 className="font-medium mb-2">Playground</h3>
                <p className="text-sm text-muted-foreground">
                  Explore the API interactively at{' '}
                  <a href="http://localhost:8080/playground" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline inline-flex items-center gap-1">
                    http://localhost:8080/playground
                    <ExternalLink className="h-3 w-3" />
                  </a>
                </p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Queries</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="border rounded-lg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <Badge>Query</Badge>
                  <code className="font-medium">me</code>
                </div>
                <p className="text-sm text-muted-foreground">Get current tenant information</p>
              </div>

              <div className="border rounded-lg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <Badge>Query</Badge>
                  <code className="font-medium">jobs</code>
                </div>
                <p className="text-sm text-muted-foreground">List jobs with filtering and pagination</p>
              </div>

              <div className="border rounded-lg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <Badge>Query</Badge>
                  <code className="font-medium">job(id: ID!)</code>
                </div>
                <p className="text-sm text-muted-foreground">Get a specific job by ID</p>
              </div>

              <div className="border rounded-lg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <Badge>Query</Badge>
                  <code className="font-medium">providers</code>
                </div>
                <p className="text-sm text-muted-foreground">List available AI providers</p>
              </div>

              <div className="border rounded-lg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <Badge>Query</Badge>
                  <code className="font-medium">usageSummary</code>
                </div>
                <p className="text-sm text-muted-foreground">Get usage statistics by provider</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Mutations</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="border rounded-lg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <Badge variant="secondary">Mutation</Badge>
                  <code className="font-medium">createJob(input: CreateJobInput!)</code>
                </div>
                <p className="text-sm text-muted-foreground">Create a new AI job</p>
                <pre className="mt-2 rounded bg-muted p-3 text-xs overflow-x-auto">
{`input CreateJobInput {
  type: JobType!    # TEXT or IMAGE
  input: String!    # The prompt or request
}`}
                </pre>
              </div>

              <div className="border rounded-lg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <Badge variant="secondary">Mutation</Badge>
                  <code className="font-medium">cancelJob(id: ID!)</code>
                </div>
                <p className="text-sm text-muted-foreground">Cancel a pending job</p>
              </div>

              <div className="border rounded-lg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <Badge variant="secondary">Mutation</Badge>
                  <code className="font-medium">updateTenant(input: UpdateTenantInput!)</code>
                </div>
                <p className="text-sm text-muted-foreground">Update tenant settings</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Subscriptions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="border rounded-lg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <Badge variant="outline">Subscription</Badge>
                  <code className="font-medium">jobUpdated</code>
                </div>
                <p className="text-sm text-muted-foreground">Real-time updates for all tenant jobs</p>
              </div>

              <div className="border rounded-lg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <Badge variant="outline">Subscription</Badge>
                  <code className="font-medium">jobStatusChanged(jobId: ID!)</code>
                </div>
                <p className="text-sm text-muted-foreground">Real-time updates for a specific job</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Providers */}
        <TabsContent value="providers" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Cpu className="h-5 w-5" />
                Supported Providers
              </CardTitle>
              <CardDescription>
                AI Aggregator supports multiple AI providers
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <Badge className="bg-green-500">OpenAI</Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    GPT-4, GPT-3.5-turbo, and other OpenAI models
                  </p>
                  <p className="text-xs text-muted-foreground mt-2">
                    Requires: <code>OPENAI_API_KEY</code>
                  </p>
                </div>

                <div className="border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <Badge className="bg-orange-500">Claude</Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Anthropic Claude 3, Claude 3.5 models
                  </p>
                  <p className="text-xs text-muted-foreground mt-2">
                    Requires: <code>ANTHROPIC_API_KEY</code>
                  </p>
                </div>

                <div className="border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <Badge className="bg-blue-500">Ollama</Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Local open-source models (Llama, Mistral, etc.)
                  </p>
                  <p className="text-xs text-muted-foreground mt-2">
                    Requires: <code>OLLAMA_URL</code> (default: localhost:11434)
                  </p>
                </div>

                <div className="border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <Badge variant="outline">Stub</Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Test provider for development (always available)
                  </p>
                  <p className="text-xs text-muted-foreground mt-2">
                    No configuration needed
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Provider Selection</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                The system automatically selects the best available provider based on:
              </p>
              <ul className="space-y-2 text-sm text-muted-foreground">
                <li className="flex items-start gap-2">
                  <span className="text-primary">1.</span>
                  Your default provider setting (if configured)
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-primary">2.</span>
                  Provider priority (OpenAI {'>'} Claude {'>'} Ollama {'>'} Stub)
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-primary">3.</span>
                  Provider availability (health check)
                </li>
              </ul>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Examples */}
        <TabsContent value="examples" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <BookOpen className="h-5 w-5" />
                Code Examples
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <div>
                <h3 className="font-medium mb-2">JavaScript/TypeScript</h3>
                <pre className="rounded-lg bg-muted p-4 text-sm overflow-x-auto">
{`import { GraphQLClient } from 'graphql-request'

const client = new GraphQLClient('http://localhost:8080/graphql', {
  headers: { 'X-API-Key': 'your-api-key' },
})

const CREATE_JOB = \`
  mutation CreateJob($input: CreateJobInput!) {
    createJob(input: $input) {
      id
      status
    }
  }
\`

const { createJob } = await client.request(CREATE_JOB, {
  input: { type: 'TEXT', input: 'Explain quantum computing' }
})

console.log('Job created:', createJob.id)`}
                </pre>
              </div>

              <div>
                <h3 className="font-medium mb-2">Python</h3>
                <pre className="rounded-lg bg-muted p-4 text-sm overflow-x-auto">
{`import requests

url = "http://localhost:8080/graphql"
headers = {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key"
}

query = """
mutation CreateJob($input: CreateJobInput!) {
    createJob(input: $input) {
        id
        status
    }
}
"""

response = requests.post(url, json={
    "query": query,
    "variables": {"input": {"type": "TEXT", "input": "Hello AI"}}
}, headers=headers)

print(response.json())`}
                </pre>
              </div>

              <div>
                <h3 className="font-medium mb-2">Go</h3>
                <pre className="rounded-lg bg-muted p-4 text-sm overflow-x-auto">
{`package main

import (
    "context"
    "github.com/machinebox/graphql"
)

func main() {
    client := graphql.NewClient("http://localhost:8080/graphql")

    req := graphql.NewRequest(\`
        mutation CreateJob($input: CreateJobInput!) {
            createJob(input: $input) {
                id
                status
            }
        }
    \`)

    req.Var("input", map[string]string{
        "type":  "TEXT",
        "input": "Hello AI",
    })
    req.Header.Set("X-API-Key", "your-api-key")

    var resp struct {
        CreateJob struct {
            ID     string
            Status string
        }
    }

    if err := client.Run(context.Background(), req, &resp); err != nil {
        panic(err)
    }
}`}
                </pre>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
