/**
 * AI Aggregator JavaScript SDK
 *
 * Async job-based API client for AI completions.
 *
 * @packageDocumentation
 */

export { AIAggregator } from './client'
export { WorkflowExecution } from './workflow'

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
  // Workflow types
  WorkflowExecutionConfig,
  WorkflowExecutionEvent,
  IntakeQuestionEvent,
  IntakeAnswerRecordedEvent,
  ExecutionCompletedEvent,
  ExecutionFailedEvent,
} from './types'

export { AIAggregatorError } from './types'

export { DEFAULT_CONFIG, ENDPOINTS, ERROR_CODES } from './constants'
