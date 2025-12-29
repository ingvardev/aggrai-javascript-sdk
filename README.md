# @aiaggregator/sdk

JavaScript/TypeScript SDK for AI Aggregator async API.

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
3. SDK automatically polls `/api/chat/{jobId}` for status
4. When job completes, the Promise resolves with the result

This approach allows for:
- Long-running AI requests without timeout issues
- Job tracking and cancellation
- Better resource management on the server

## Usage

### Synchronous Chat (Wait for Result)

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
})
```

## Error Handling

```typescript
import { AIAggregatorError } from '@aiaggregator/sdk'

try {
  const result = await client.chat({ prompt: 'Hello' })
} catch (error) {
  if (error instanceof AIAggregatorError) {
    console.error('Code:', error.code)      // 'job_failed', 'timeout', etc.
    console.error('Message:', error.message)
    console.error('Status:', error.status)  // HTTP status if applicable
    console.error('Details:', error.details)
  }
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `job_failed` | Job failed on the server |
| `timeout` | Polling timeout exceeded |
| `request_failed` | HTTP request failed |
| `network_error` | Network connection error |
| `unauthorized` | Invalid API key |

## Types

All types are exported for TypeScript users:

```typescript
import type {
  ChatResult,
  ChatResponse,
  CreateChatRequest,
  Job,
  JobStatus,
  ChatMessage,
  Tool,
  ToolCall,
} from '@aiaggregator/sdk'
```

## API Reference

### `client.chat(request)`

Send a chat request and wait for the result.

**Parameters:**
- `prompt?: string` - Simple text prompt
- `messages?: ChatMessage[]` - Chat messages array
- `provider?: string` - AI provider (openai, claude, ollama)
- `model?: string` - Model name
- `maxTokens?: number` - Maximum tokens to generate
- `temperature?: number` - Temperature (0-2)
- `tools?: Tool[]` - Tools/functions for AI to call
- `toolChoice?: string` - Tool choice mode

**Returns:** `Promise<ChatResult>`

### `client.chatAsync(request)`

Send a chat request without waiting.

**Returns:** `Promise<ChatResponse>` with `{ jobId, status }`

### `client.waitForJob(jobId)`

Wait for a job to complete.

**Returns:** `Promise<ChatResult>`

### `client.getJobStatus(jobId)`

Get current job status.

**Returns:** `Promise<Job>`

### `client.cancelJob(jobId)`

Cancel a pending job.

**Returns:** `Promise<void>`

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

## License

MIT
