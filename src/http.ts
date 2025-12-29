/**
 * HTTP client utilities with retry and error handling
 */

import { AIAggregatorError } from './types'
import { ERROR_CODES, DEFAULT_CONFIG } from './constants'

export interface RequestOptions {
  method: string
  path: string
  body?: unknown
  signal?: AbortSignal
  timeout?: number
  retries?: number
}

export interface HttpClientConfig {
  baseUrl: string
  apiKey: string
  timeout: number
}

/**
 * HTTP client with retry logic and proper error handling
 */
export class HttpClient {
  constructor(private config: HttpClientConfig) {}

  /**
   * Make an HTTP request with automatic retry and timeout
   */
  async request<T>(options: RequestOptions): Promise<T> {
    const {
      method,
      path,
      body,
      signal,
      timeout = this.config.timeout,
      retries = DEFAULT_CONFIG.MAX_RETRIES,
    } = options

    const url = `${this.config.baseUrl}${path}`

    let lastError: Error | undefined
    let attempt = 0

    while (attempt < retries) {
      attempt++

      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), timeout)

      // Combine with external signal if provided
      if (signal?.aborted) {
        throw new AIAggregatorError('Request aborted', ERROR_CODES.ABORTED)
      }

      const abortHandler = () => controller.abort()
      signal?.addEventListener('abort', abortHandler)

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
        signal?.removeEventListener('abort', abortHandler)

        if (!response.ok) {
          const errorData = await this.parseErrorResponse(response)
          throw new AIAggregatorError(
            errorData.message,
            errorData.code,
            response.status,
            errorData.details
          )
        }

        return response.json()
      } catch (error) {
        clearTimeout(timeoutId)
        signal?.removeEventListener('abort', abortHandler)

        if (error instanceof AIAggregatorError) {
          // Don't retry client errors (4xx)
          if (error.status && error.status >= 400 && error.status < 500) {
            throw error
          }
          lastError = error
        } else if (error instanceof Error) {
          if (error.name === 'AbortError') {
            if (signal?.aborted) {
              throw new AIAggregatorError('Request aborted', ERROR_CODES.ABORTED)
            }
            throw new AIAggregatorError('Request timeout', ERROR_CODES.TIMEOUT)
          }
          lastError = new AIAggregatorError(
            error.message,
            ERROR_CODES.NETWORK_ERROR,
            undefined,
            error
          )
        }

        // Exponential backoff before retry
        if (attempt < retries) {
          await this.sleep(DEFAULT_CONFIG.RETRY_BASE_DELAY * Math.pow(2, attempt - 1))
        }
      }
    }

    throw lastError ?? new AIAggregatorError('Request failed', ERROR_CODES.REQUEST_FAILED)
  }

  /**
   * Parse error response from server
   */
  private async parseErrorResponse(response: Response): Promise<{
    message: string
    code: string
    details?: unknown
  }> {
    let errorData: unknown
    try {
      errorData = await response.json()
    } catch {
      errorData = await response.text()
    }

    const message =
      typeof errorData === 'object' && errorData !== null && 'message' in errorData
        ? String((errorData as { message: unknown }).message)
        : `Request failed with status ${response.status}`

    const code =
      typeof errorData === 'object' && errorData !== null && 'error' in errorData
        ? String((errorData as { error: unknown }).error)
        : ERROR_CODES.REQUEST_FAILED

    return { message, code, details: errorData }
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms))
  }
}
