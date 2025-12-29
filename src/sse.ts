/**
 * SSE (Server-Sent Events) client for job status updates
 */

import type { ChatResult } from './types'
import { AIAggregatorError } from './types'
import { ERROR_CODES, ENDPOINTS } from './constants'

export interface SSEClientConfig {
  baseUrl: string
  apiKey: string
  timeout: number
}

export interface SSEEvent {
  status: string
  output?: string
  error?: string
  provider?: string
  model?: string
  tokensIn?: number
  tokensOut?: number
  cost?: number
  finishReason?: string
  toolCalls?: ChatResult['toolCalls']
}

/**
 * SSE client for receiving real-time job updates
 */
export class SSEClient {
  constructor(private config: SSEClientConfig) {}

  /**
   * Subscribe to job events and wait for completion
   */
  async waitForJob(jobId: string, signal?: AbortSignal): Promise<ChatResult> {
    return new Promise((resolve, reject) => {
      const url = `${this.config.baseUrl}${ENDPOINTS.JOB_EVENTS(jobId)}`

      const controller = new AbortController()
      const timeoutId = setTimeout(() => {
        controller.abort()
        reject(new AIAggregatorError('SSE timeout', ERROR_CODES.TIMEOUT, undefined, { jobId }))
      }, this.config.timeout)

      // Handle external abort signal
      const abortHandler = () => {
        clearTimeout(timeoutId)
        controller.abort()
        reject(new AIAggregatorError('Request aborted', ERROR_CODES.ABORTED, undefined, { jobId }))
      }
      signal?.addEventListener('abort', abortHandler)

      const cleanup = () => {
        clearTimeout(timeoutId)
        signal?.removeEventListener('abort', abortHandler)
      }

      this.connectSSE(url, controller.signal)
        .then(async (reader) => {
          try {
            const result = await this.processEvents(reader, jobId)
            cleanup()
            resolve(result)
          } catch (error) {
            cleanup()
            reject(error)
          }
        })
        .catch((error) => {
          cleanup()
          if (error instanceof AIAggregatorError) {
            reject(error)
          } else if (error instanceof Error && error.name === 'AbortError') {
            reject(
              new AIAggregatorError(
                signal?.aborted ? 'Request aborted' : 'SSE timeout',
                signal?.aborted ? ERROR_CODES.ABORTED : ERROR_CODES.TIMEOUT,
                undefined,
                { jobId }
              )
            )
          } else {
            reject(
              new AIAggregatorError(
                error instanceof Error ? error.message : 'SSE connection failed',
                ERROR_CODES.SSE_FAILED,
                undefined,
                error
              )
            )
          }
        })
    })
  }

  /**
   * Connect to SSE endpoint and return reader
   */
  private async connectSSE(
    url: string,
    signal: AbortSignal
  ): Promise<ReadableStreamDefaultReader<Uint8Array>> {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        Accept: 'text/event-stream',
        'X-API-Key': this.config.apiKey,
      },
      signal,
    })

    if (!response.ok) {
      throw new AIAggregatorError(
        `SSE connection failed: ${response.status}`,
        ERROR_CODES.SSE_FAILED,
        response.status
      )
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new AIAggregatorError('No response body', ERROR_CODES.SSE_FAILED)
    }

    return reader
  }

  /**
   * Process SSE events and wait for terminal state
   */
  private async processEvents(
    reader: ReadableStreamDefaultReader<Uint8Array>,
    jobId: string
  ): Promise<ChatResult> {
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) {
        throw new AIAggregatorError(
          'SSE connection closed unexpectedly',
          ERROR_CODES.SSE_FAILED,
          undefined,
          { jobId }
        )
      }

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() ?? ''

      for (const line of lines) {
        const event = this.parseLine(line)
        if (!event) continue

        if (event.status === 'completed') {
          return this.eventToResult(event, jobId)
        }

        if (event.status === 'failed') {
          throw new AIAggregatorError(event.error ?? 'Job failed', ERROR_CODES.JOB_FAILED, undefined, {
            jobId,
            event,
          })
        }
      }
    }
  }

  /**
   * Parse a single SSE line
   */
  private parseLine(line: string): SSEEvent | null {
    if (!line.startsWith('data: ')) return null

    try {
      return JSON.parse(line.slice(6)) as SSEEvent
    } catch {
      return null
    }
  }

  /**
   * Convert SSE event to ChatResult
   */
  private eventToResult(event: SSEEvent, jobId: string): ChatResult {
    return {
      content: event.output ?? '',
      toolCalls: event.toolCalls,
      finishReason: event.finishReason ?? 'stop',
      usage: {
        tokensIn: event.tokensIn ?? 0,
        tokensOut: event.tokensOut ?? 0,
        cost: event.cost ?? 0,
      },
      jobId,
      provider: event.provider,
      model: event.model,
    }
  }
}
