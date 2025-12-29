/**
 * AI Aggregator SDK Types
 */

/** Job status */
export type JobStatus = 'pending' | 'processing' | 'completed' | 'failed'

/** Job type */
export type JobType = 'text' | 'image'

/** Chat message role */
export type MessageRole = 'system' | 'user' | 'assistant' | 'tool'

/** Chat message */
export interface ChatMessage {
  role: MessageRole
  content: string
  toolCalls?: ToolCall[]
  toolCallId?: string
}

/** Tool/function definition */
export interface Tool {
  type: 'function'
  function: ToolFunction
}

/** Tool function definition */
export interface ToolFunction {
  name: string
  description?: string
  parameters?: Record<string, unknown>
}

/** Tool call from AI */
export interface ToolCall {
  id: string
  type: 'function'
  function: {
    name: string
    arguments: string
  }
}

/** Create chat request */
export interface CreateChatRequest {
  /** Simple prompt (alternative to messages) */
  prompt?: string
  /** Chat messages */
  messages?: ChatMessage[]
  /** Provider name (openai, claude, ollama) */
  provider?: string
  /** Model name */
  model?: string
  /** Maximum tokens to generate */
  maxTokens?: number
  /** Temperature (0-2) */
  temperature?: number
  /** Tools/functions for AI to call */
  tools?: Tool[]
  /** Tool choice: "auto", "none", "required", or function name */
  toolChoice?: string | { type: 'function'; function: { name: string } }
  /** Custom metadata */
  metadata?: Record<string, unknown>
}

/** Response from /api/chat (job creation) */
export interface ChatResponse {
  /** Job ID for tracking */
  jobId: string
  /** Initial job status */
  status: JobStatus
}

/** Job object */
export interface Job {
  id: string
  type: JobType
  status: JobStatus
  input: string
  output?: string
  provider?: string
  model?: string
  error?: string
  tokensIn?: number
  tokensOut?: number
  cost?: number
  toolCalls?: ToolCall[]
  finishReason?: string
  metadata?: Record<string, unknown>
  createdAt: string
  updatedAt: string
  completedAt?: string
}

/** SDK configuration */
export interface SDKConfig {
  /** API base URL */
  baseUrl: string
  /** API key */
  apiKey: string
  /** Default provider */
  defaultProvider?: string
  /** Default model */
  defaultModel?: string
  /** Request timeout in ms (default: 30000) */
  timeout?: number
  /** Polling interval in ms (default: 1000) */
  pollingInterval?: number
  /** Maximum polling attempts (default: 300 = 5 minutes) */
  maxPollingAttempts?: number
}

/** Chat completion result */
export interface ChatResult {
  /** Generated content */
  content: string
  /** Tool calls (if any) */
  toolCalls?: ToolCall[]
  /** Finish reason: stop, tool_calls, length, etc. */
  finishReason: string
  /** Token usage and cost */
  usage: {
    tokensIn: number
    tokensOut: number
    cost: number
  }
  /** Job ID */
  jobId: string
  /** Provider used */
  provider?: string
  /** Model used */
  model?: string
}

/** SDK error */
export class AIAggregatorError extends Error {
  constructor(
    message: string,
    public code: string,
    public status?: number,
    public details?: unknown
  ) {
    super(message)
    this.name = 'AIAggregatorError'
  }
}
