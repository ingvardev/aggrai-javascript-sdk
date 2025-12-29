/**
 * AI Aggregator SDK Constants
 */

/** Default configuration values */
export const DEFAULT_CONFIG = {
  /** Default timeout for requests (5 minutes) */
  TIMEOUT: 300_000,
  /** Default polling interval (1 second) */
  POLLING_INTERVAL: 1_000,
  /** Default max polling attempts (5 minutes at 1s interval) */
  MAX_POLLING_ATTEMPTS: 300,
  /** Default provider */
  DEFAULT_PROVIDER: 'openai',
  /** Default retry attempts for failed requests */
  MAX_RETRIES: 3,
  /** Base delay for exponential backoff (ms) */
  RETRY_BASE_DELAY: 1_000,
} as const

/** API endpoints */
export const ENDPOINTS = {
  CHAT: '/api/chat',
  JOB: (id: string) => `/api/chat/${id}`,
  JOB_EVENTS: (id: string) => `/api/chat/${id}/events`,
} as const

/** Error codes */
export const ERROR_CODES = {
  TIMEOUT: 'timeout',
  NETWORK_ERROR: 'network_error',
  REQUEST_FAILED: 'request_failed',
  JOB_FAILED: 'job_failed',
  SSE_FAILED: 'sse_failed',
  VALIDATION_ERROR: 'validation_error',
  ABORTED: 'aborted',
} as const
