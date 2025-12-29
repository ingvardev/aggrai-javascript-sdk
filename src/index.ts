/**
 * AI Aggregator JavaScript SDK
 *
 * Async job-based API client for AI completions.
 *
 * @packageDocumentation
 */

export { AIAggregator } from './client'

export type {
  SDKConfig,
  CreateChatRequest,
  ChatResponse,
  ChatResult,
  Job,
  JobStatus,
  JobType,
  ChatMessage,
  MessageRole,
  Tool,
  ToolFunction,
  ToolCall,
} from './types'

export { AIAggregatorError } from './types'
