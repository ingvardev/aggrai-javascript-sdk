# @aiaggregator/sdk

JavaScript/TypeScript SDK for AI Aggregator async API.

## Features

- 🚀 **Async job-based API** — no timeout issues with long-running requests
- 📡 **SSE support** — real-time updates via Server-Sent Events
- 🔄 **Automatic retry** — exponential backoff for transient failures
- ⏹️ **Cancellation** — abort any request with `AbortController`
- 🔧 **Tool/function calling** — OpenAI and Claude compatible
- 📦 **Zero dependencies** — uses native `fetch`

## Installation

```bash
npm install @aiaggregator/sdk
# or
pnpm add @aiaggregator/sdk
# or
yarn add @aiaggregator/sdk
```

## Quick Start

```typescript
import { AIAggregator } from '@aiaggregator/sdk'

const client = new AIAggregator({
  baseUrl: 'https://api.example.com',
  apiKey: 'your-api-key',
})

// Simple chat - returns Promise that resolves when job completes
const result = await client.chat({
  messages: [{ role: 'user', content: 'Hello!' }],
  provider: 'openai',
  model: 'gpt-4o-mini',
})

console.log(result.content)
console.log(result.usage) // { tokensIn: 10, tokensOut: 20, cost: 0.001 }
```

## How It Works

The SDK uses an async job-based API:

1. **`client.chat()`** sends request to `/api/chat`
2. The server creates a job and returns `{ jobId, status }`
3. SDK uses SSE (or polls) `/api/chat/{jobId}/events` for updates
4. When job completes, the Promise resolves with the result

This approach allows for:
- Long-running AI requests without timeout issues
- Real-time status updates via SSE
- Job tracking and cancellation
- Better resource management on the server

## Usage

### Basic Chat

```typescript
const result = await client.chat({
  prompt: 'Write a haiku about programming',
  provider: 'claude',
  model: 'claude-3-haiku-20240307',
  maxTokens: 100,
})

console.log(result.content)
```

### Async Chat (Non-blocking)

```typescript
// Create job without waiting
const { jobId, status } = await client.chatAsync({
  messages: [{ role: 'user', content: 'Hello!' }],
})

console.log(`Job created: ${jobId}`) // Immediately available

// Do other work...

// Later: wait for result
const result = await client.waitForJob(jobId)
console.log(result.content)
```

### With Cancellation

```typescript
const controller = new AbortController()

// Cancel after 10 seconds
setTimeout(() => controller.abort(), 10000)

try {
  const result = await client.chat(
    { prompt: 'Write a long essay...' },
    controller.signal
  )
} catch (error) {
  if (error.code === 'aborted') {
    console.log('Request was cancelled')
  }
}
```

### Check Job Status

```typescript
const job = await client.getJobStatus(jobId)

console.log(job.status) // 'pending' | 'processing' | 'completed' | 'failed'
console.log(job.output) // Result when completed
```

### Cancel Job

```typescript
await client.cancelJob(jobId)
```

### With Tools/Functions

```typescript
const result = await client.chat({
  messages: [{ role: 'user', content: "What's the weather in Paris?" }],
  tools: [
    {
      type: 'function',
      function: {
        name: 'get_weather',
        description: 'Get weather for a location',
        parameters: {
          type: 'object',
          properties: {
            location: { type: 'string', description: 'City name' },
          },
          required: ['location'],
        },
      },
    },
  ],
  toolChoice: 'auto',
})

if (result.toolCalls) {
  for (const call of result.toolCalls) {
    console.log(call.function.name) // 'get_weather'
    console.log(call.function.arguments) // '{"location": "Paris"}'
  }
}
```

## Configuration

```typescript
import { AIAggregator } from '@aiaggregator/sdk'

const client = new AIAggregator({
  // Required
  baseUrl: 'https://api.example.com',
  apiKey: 'your-api-key',

  // Optional
  defaultProvider: 'openai',      // Default provider for requests
  defaultModel: 'gpt-4o-mini',    // Default model
  timeout: 300000,                // Request timeout (ms) - default 5 min
  pollingInterval: 1000,          // Job polling interval (ms)
  maxPollingAttempts: 300,        // Max polling attempts (5 min default)
  useSSE: 'auto',                 // SSE mode: 'auto' | true | false
})
```

### SSE Configuration

| Value | Behavior |
|-------|----------|
| `'auto'` (default) | Use SSE if `fetch` is available (Node.js 18+ or browser) |
| `true` | Always use SSE |
| `false` | Always use polling |

## Error Handling

```typescript
import { AIAggregatorError, ERROR_CODES } from '@aiaggregator/sdk'

try {
  const result = await client.chat({ prompt: 'Hello' })
} catch (error) {
  if (error instanceof AIAggregatorError) {
    console.error('Code:', error.code)
    console.error('Message:', error.message)
    console.error('Status:', error.status)  // HTTP status if applicable
    console.error('Details:', error.details)

    // Handle specific errors
    switch (error.code) {
      case ERROR_CODES.TIMEOUT:
        console.log('Request timed out')
        break
      case ERROR_CODES.JOB_FAILED:
        console.log('Job failed:', error.details)
        break
      case ERROR_CODES.ABORTED:
        console.log('Request was cancelled')
        break
    }
  }
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `timeout` | Request or polling timeout exceeded |
| `job_failed` | Job failed on the server |
| `request_failed` | HTTP request failed |
| `network_error` | Network connection error |
| `sse_failed` | SSE connection failed |
| `validation_error` | Invalid input parameters |
| `aborted` | Request was cancelled via AbortSignal |

## Types

All types are exported for TypeScript users:

```typescript
import type {
  SDKConfig,
  ChatResult,
  ChatResponse,
  CreateChatRequest,
  Job,
  JobStatus,
  JobType,
  ChatMessage,
  MessageRole,
  Tool,
  ToolFunction,
  ToolCall,
} from '@aiaggregator/sdk'

// Constants
import { DEFAULT_CONFIG, ENDPOINTS, ERROR_CODES } from '@aiaggregator/sdk'
```

## API Reference

### `client.chat(request, signal?)`

Send a chat request and wait for the result.

**Parameters:**
- `request: CreateChatRequest`
  - `prompt?: string` - Simple text prompt
  - `messages?: ChatMessage[]` - Chat messages array
  - `provider?: string` - AI provider (openai, claude, ollama)
  - `model?: string` - Model name
  - `maxTokens?: number` - Maximum tokens to generate
  - `temperature?: number` - Temperature (0-2)
  - `tools?: Tool[]` - Tools/functions for AI to call
  - `toolChoice?: string` - Tool choice mode
  - `metadata?: Record<string, unknown>` - Custom metadata
- `signal?: AbortSignal` - Optional signal to cancel request

**Returns:** `Promise<ChatResult>`

```typescript
interface ChatResult {
  content: string
  toolCalls?: ToolCall[]
  finishReason: string
  usage: { tokensIn: number; tokensOut: number; cost: number }
  jobId: string
  provider?: string
  model?: string
}
```

### `client.chatAsync(request, signal?)`

Send a chat request without waiting for completion.

**Returns:** `Promise<ChatResponse>` with `{ jobId, status }`

### `client.waitForJob(jobId, signal?)`

Wait for a job to complete. Uses SSE when available, falls back to polling.

**Returns:** `Promise<ChatResult>`

### `client.getJobStatus(jobId, signal?)`

Get current job status.

**Returns:** `Promise<Job>`

### `client.cancelJob(jobId, signal?)`

Cancel a pending job.

**Returns:** `Promise<void>`

## Node.js Compatibility

| Node.js Version | Support |
|-----------------|---------|
| 18+ | ✅ Full support (native fetch) |
| 16-17 | ⚠️ Requires fetch polyfill |
| < 16 | ❌ Not supported |

## Development

```bash
# Install dependencies
pnpm install

# Build
pnpm build

# Watch mode
pnpm dev

# Run tests
pnpm test

# Type check
pnpm typecheck
```

## Architecture

```
src/
├── client.ts      # Main AIAggregator class
├── constants.ts   # Configuration defaults and error codes
├── http.ts        # HTTP client with retry logic
├── sse.ts         # SSE client for real-time updates
├── types.ts       # TypeScript types and interfaces
└── index.ts       # Public exports
```

## License

MIT
