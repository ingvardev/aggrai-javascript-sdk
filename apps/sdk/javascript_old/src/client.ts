import type {
  SDKConfig,
  CreateChatRequest,
  ChatResponse,
  ChatResult,
  Job,
} from './types'
import { AIAggregatorError } from './types'

/**
 * AI Aggregator SDK Client
 *
 * Provides async job-based API for AI completions.
 * The SDK calls /api/chat which creates a job internally,
 * then automatically polls for the result.
 *
 * @example
 * ```typescript
 * const client = new AIAggregator({
 *   baseUrl: 'https://api.example.com',
 *   apiKey: 'your-api-key',
 * })
 *
 * // Simple completion - returns Promise that resolves when job completes
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
  private config: Required<SDKConfig>

  constructor(config: SDKConfig) {
    this.config = {
      baseUrl: config.baseUrl.replace(/\/$/, ''), // Remove trailing slash
      apiKey: config.apiKey,
      defaultProvider: config.defaultProvider ?? 'openai',
      defaultModel: config.defaultModel ?? '',
      timeout: config.timeout ?? 300000, // 5 minutes default for async jobs
      pollingInterval: config.pollingInterval ?? 1000,
      maxPollingAttempts: config.maxPollingAttempts ?? 300,
    }
  }

  /**
   * Send a chat request and wait for the result.
   *
   * This method:
   * 1. Calls /api/chat to create an async job
   * 2. Automatically subscribes to job updates
   * 3. Returns a Promise that resolves when the job completes
   *
   * @example
   * ```typescript
   * const result = await client.chat({
   *   messages: [{ role: 'user', content: 'Hello!' }],
   * })
   * console.log(result.content)
   * ```
   */
  async chat(request: CreateChatRequest): Promise<ChatResult> {
    // Step 1: Call /api/chat to create job
    const chatResponse = await this.createChat(request)

    // Step 2: Wait for job to complete
    return this.waitForJob(chatResponse.jobId)
  }

  /**
   * Send a chat request without waiting for the result.
   * Returns the job ID immediately for manual tracking.
   *
   * @example
   * ```typescript
   * const { jobId } = await client.chatAsync({
   *   messages: [{ role: 'user', content: 'Hello!' }],
   * })
   *
   * // Later: check status or wait
   * const result = await client.waitForJob(jobId)
   * ```
   */
  async chatAsync(request: CreateChatRequest): Promise<ChatResponse> {
    return this.createChat(request)
  }

  /**
   * Wait for a job to complete.
   * Returns a Promise that resolves with the result.
   */
  async waitForJob(jobId: string): Promise<ChatResult> {
    let attempts = 0

    while (attempts < this.config.maxPollingAttempts) {
      const job = await this.getJobStatus(jobId)

      if (job.status === 'completed') {
        return this.jobToResult(job)
      }

      if (job.status === 'failed') {
        throw new AIAggregatorError(
          job.error ?? 'Job failed',
          'job_failed',
          undefined,
          { jobId, job }
        )
      }

      await this.sleep(this.config.pollingInterval)
      attempts++
    }

    throw new AIAggregatorError(
      'Job polling timeout',
      'timeout',
      undefined,
      { jobId, attempts }
    )
  }

  /**
   * Get the current status of a job.
   */
  async getJobStatus(jobId: string): Promise<Job> {
    return this.request<Job>('GET', `/api/chat/${jobId}`)
  }

  /**
   * Cancel a pending job.
   */
  async cancelJob(jobId: string): Promise<void> {
    await this.request('DELETE', `/api/chat/${jobId}`)
  }

  // ============ Private Methods ============

  private async createChat(request: CreateChatRequest): Promise<ChatResponse> {
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

    return this.request<ChatResponse>('POST', '/api/chat', body)
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    const url = `${this.config.baseUrl}${path}`

    const controller = new AbortController()
    const timeoutId = setTimeout(
      () => controller.abort(),
      this.config.timeout
    )

    try {
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.config.apiKey,
        },
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      })

      clearTimeout(timeoutId)

      if (!response.ok) {
        let errorData: unknown
        try {
          errorData = await response.json()
        } catch {
          errorData = await response.text()
        }

        const message =
          typeof errorData === 'object' &&
          errorData !== null &&
          'message' in errorData
            ? String((errorData as { message: unknown }).message)
            : `Request failed with status ${response.status}`

        const code =
          typeof errorData === 'object' &&
          errorData !== null &&
          'error' in errorData
            ? String((errorData as { error: unknown }).error)
            : 'request_failed'

        throw new AIAggregatorError(message, code, response.status, errorData)
      }

      return response.json()
    } catch (error) {
      clearTimeout(timeoutId)

      if (error instanceof AIAggregatorError) {
        throw error
      }

      if (error instanceof Error && error.name === 'AbortError') {
        throw new AIAggregatorError('Request timeout', 'timeout')
      }

      throw new AIAggregatorError(
        error instanceof Error ? error.message : 'Network error',
        'network_error',
        undefined,
        error
      )
    }
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

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms))
  }
}
