import type { SDKConfig, CreateChatRequest, ChatResponse, ChatResult, Job } from './types'
import { AIAggregatorError } from './types'
import { DEFAULT_CONFIG, ENDPOINTS, ERROR_CODES } from './constants'
import { HttpClient } from './http'
import { SSEClient } from './sse'

/** Internal configuration with all fields required */
interface ResolvedConfig {
  baseUrl: string
  apiKey: string
  defaultProvider: string
  defaultModel: string
  timeout: number
  pollingInterval: number
  maxPollingAttempts: number
  useSSE: boolean | 'auto'
}

/**
 * AI Aggregator SDK Client
 *
 * Provides async job-based API for AI completions.
 *
 * @example
 * ```typescript
 * const client = new AIAggregator({
 *   baseUrl: 'https://api.example.com',
 *   apiKey: 'your-api-key',
 * })
 *
 * const result = await client.chat({
 *   messages: [{ role: 'user', content: 'Hello!' }],
 *   provider: 'openai',
 *   model: 'gpt-4o-mini',
 * })
 *
 * console.log(result.content)
 * ```
 */
export class AIAggregator {
  private readonly config: ResolvedConfig
  private readonly http: HttpClient
  private readonly sse: SSEClient

  constructor(config: SDKConfig) {
    this.validateConfig(config)

    this.config = {
      baseUrl: config.baseUrl.replace(/\/$/, ''),
      apiKey: config.apiKey,
      defaultProvider: config.defaultProvider ?? DEFAULT_CONFIG.DEFAULT_PROVIDER,
      defaultModel: config.defaultModel ?? '',
      timeout: config.timeout ?? DEFAULT_CONFIG.TIMEOUT,
      pollingInterval: config.pollingInterval ?? DEFAULT_CONFIG.POLLING_INTERVAL,
      maxPollingAttempts: config.maxPollingAttempts ?? DEFAULT_CONFIG.MAX_POLLING_ATTEMPTS,
      useSSE: config.useSSE ?? 'auto',
    }

    this.http = new HttpClient({
      baseUrl: this.config.baseUrl,
      apiKey: this.config.apiKey,
      timeout: this.config.timeout,
    })

    this.sse = new SSEClient({
      baseUrl: this.config.baseUrl,
      apiKey: this.config.apiKey,
      timeout: this.config.timeout,
    })
  }

  /**
   * Send a chat request and wait for the result.
   *
   * @param request - Chat request parameters
   * @param signal - Optional AbortSignal to cancel the request
   * @returns Promise resolving to chat completion result
   *
   * @example
   * ```typescript
   * const result = await client.chat({
   *   messages: [{ role: 'user', content: 'Hello!' }],
   * })
   * console.log(result.content)
   * ```
   *
   * @example With cancellation
   * ```typescript
   * const controller = new AbortController()
   * setTimeout(() => controller.abort(), 10000)
   *
   * try {
   *   const result = await client.chat({ prompt: 'Hello' }, controller.signal)
   * } catch (e) {
   *   if (e.code === 'aborted') console.log('Cancelled!')
   * }
   * ```
   */
  async chat(request: CreateChatRequest, signal?: AbortSignal): Promise<ChatResult> {
    this.validateRequest(request)

    const chatResponse = await this.createChat(request, signal)
    return this.waitForJob(chatResponse.jobId, signal)
  }

  /**
   * Send a chat request without waiting for the result.
   * Returns the job ID immediately for manual tracking.
   *
   * @param request - Chat request parameters
   * @param signal - Optional AbortSignal to cancel the request
   * @returns Promise resolving to job creation response
   */
  async chatAsync(request: CreateChatRequest, signal?: AbortSignal): Promise<ChatResponse> {
    this.validateRequest(request)
    return this.createChat(request, signal)
  }

  /**
   * Wait for a job to complete.
   * Uses SSE when available, falls back to polling.
   *
   * @param jobId - Job ID to wait for
   * @param signal - Optional AbortSignal to cancel waiting
   * @returns Promise resolving to chat completion result
   */
  async waitForJob(jobId: string, signal?: AbortSignal): Promise<ChatResult> {
    if (!jobId) {
      throw new AIAggregatorError('Job ID is required', ERROR_CODES.VALIDATION_ERROR)
    }

    if (this.shouldUseSSE()) {
      try {
        return await this.sse.waitForJob(jobId, signal)
      } catch (error) {
        // SSE failed, fallback to polling (unless aborted)
        if (error instanceof AIAggregatorError && error.code === ERROR_CODES.ABORTED) {
          throw error
        }
        return this.waitForJobPolling(jobId, signal)
      }
    }

    return this.waitForJobPolling(jobId, signal)
  }

  /**
   * Get the current status of a job.
   *
   * @param jobId - Job ID to check
   * @param signal - Optional AbortSignal
   * @returns Promise resolving to job object
   */
  async getJobStatus(jobId: string, signal?: AbortSignal): Promise<Job> {
    if (!jobId) {
      throw new AIAggregatorError('Job ID is required', ERROR_CODES.VALIDATION_ERROR)
    }

    return this.http.request<Job>({
      method: 'GET',
      path: ENDPOINTS.JOB(jobId),
      signal,
      retries: 1, // Don't retry status checks
    })
  }

  /**
   * Cancel a pending job.
   *
   * @param jobId - Job ID to cancel
   * @param signal - Optional AbortSignal
   */
  async cancelJob(jobId: string, signal?: AbortSignal): Promise<void> {
    if (!jobId) {
      throw new AIAggregatorError('Job ID is required', ERROR_CODES.VALIDATION_ERROR)
    }

    await this.http.request<void>({
      method: 'DELETE',
      path: ENDPOINTS.JOB(jobId),
      signal,
      retries: 1,
    })
  }

  // ============ Private Methods ============

  private validateConfig(config: SDKConfig): void {
    if (!config.baseUrl) {
      throw new AIAggregatorError('baseUrl is required', ERROR_CODES.VALIDATION_ERROR)
    }
    if (!config.apiKey) {
      throw new AIAggregatorError('apiKey is required', ERROR_CODES.VALIDATION_ERROR)
    }
  }

  private validateRequest(request: CreateChatRequest): void {
    if (!request.prompt && (!request.messages || request.messages.length === 0)) {
      throw new AIAggregatorError(
        "Either 'prompt' or 'messages' is required",
        ERROR_CODES.VALIDATION_ERROR
      )
    }
  }

  private async createChat(request: CreateChatRequest, signal?: AbortSignal): Promise<ChatResponse> {
    const body = {
      prompt: request.prompt,
      messages: request.messages,
      provider: request.provider ?? this.config.defaultProvider,
      model: request.model ?? this.config.defaultModel,
      maxTokens: request.maxTokens,
      temperature: request.temperature,
      tools: request.tools,
      toolChoice: request.toolChoice,
      metadata: request.metadata,
    }

    return this.http.request<ChatResponse>({
      method: 'POST',
      path: ENDPOINTS.CHAT,
      body,
      signal,
    })
  }

  private async waitForJobPolling(jobId: string, signal?: AbortSignal): Promise<ChatResult> {
    let attempts = 0

    while (attempts < this.config.maxPollingAttempts) {
      if (signal?.aborted) {
        throw new AIAggregatorError('Request aborted', ERROR_CODES.ABORTED, undefined, { jobId })
      }

      const job = await this.getJobStatus(jobId, signal)

      if (job.status === 'completed') {
        return this.jobToResult(job)
      }

      if (job.status === 'failed') {
        throw new AIAggregatorError(job.error ?? 'Job failed', ERROR_CODES.JOB_FAILED, undefined, {
          jobId,
          job,
        })
      }

      await this.sleep(this.config.pollingInterval, signal)
      attempts++
    }

    throw new AIAggregatorError('Job polling timeout', ERROR_CODES.TIMEOUT, undefined, {
      jobId,
      attempts,
    })
  }

  private shouldUseSSE(): boolean {
    if (this.config.useSSE === true) return true
    if (this.config.useSSE === false) return false
    return typeof fetch !== 'undefined'
  }

  private jobToResult(job: Job): ChatResult {
    return {
      content: job.output ?? '',
      toolCalls: job.toolCalls,
      finishReason: job.finishReason ?? 'stop',
      usage: {
        tokensIn: job.tokensIn ?? 0,
        tokensOut: job.tokensOut ?? 0,
        cost: job.cost ?? 0,
      },
      jobId: job.id,
      provider: job.provider,
      model: job.model,
    }
  }

  private sleep(ms: number, signal?: AbortSignal): Promise<void> {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(resolve, ms)

      if (signal) {
        const abortHandler = () => {
          clearTimeout(timeout)
          reject(new AIAggregatorError('Request aborted', ERROR_CODES.ABORTED))
        }
        signal.addEventListener('abort', abortHandler, { once: true })
      }
    })
  }
}
