import type { WorkflowExecutionConfig, WorkflowExecutionEvent, IntakeQuestionEvent } from './types'
import { AIAggregatorError } from './types'

type EventHandler<T = unknown> = (data: T) => void
type EventType = 'question' | 'answer_recorded' | 'completed' | 'failed' | 'error' | 'reconnecting'

/**
 * Workflow Execution with SSE event streaming
 *
 * Manages real-time workflow execution with automatic reconnection
 * and intake question/answer handling.
 *
 * @example
 * ```typescript
 * const execution = client.executeWorkflow('workflow-uuid', { input: { key: 'value' } })
 *
 * execution.on('question', async (question) => {
 *   console.log('Question:', question.text)
 *   const userAnswer = await getUserInput()
 *   await execution.answer(userAnswer)
 * })
 *
 * execution.on('completed', (result) => {
 *   console.log('Workflow completed:', result.output)
 * })
 *
 * execution.on('failed', (error) => {
 *   console.error('Workflow failed:', error.error)
 * })
 * ```
 */
export class WorkflowExecution {
  private readonly config: WorkflowExecutionConfig
  private readonly executionId: string
  private eventSource: EventSource | null = null
  private readonly listeners = new Map<EventType, Set<EventHandler>>()
  private currentQuestion: IntakeQuestionEvent | null = null
  private reconnectAttempts = 0
  private readonly maxReconnectAttempts = 5
  private reconnectDelay = 1000

  constructor(executionId: string, config: WorkflowExecutionConfig) {
    this.executionId = executionId
    this.config = config
  }

  /**
   * Connect to SSE event stream for this execution
   */
  connect(): void {
    if (this.eventSource) {
      return // Already connected
    }

    // EventSource doesn't support custom headers, so we pass API key as query param
    const url = `${this.config.baseUrl}/api/workflows/executions/${this.executionId}/events?api_key=${encodeURIComponent(this.config.apiKey)}`

    this.eventSource = new EventSource(url, {
      withCredentials: false,
    })

    // Generic message handler for all event types
    this.eventSource.onmessage = (e) => {
      try {
        const event: WorkflowExecutionEvent = JSON.parse(e.data)
        this.handleEvent(event)
      } catch (error) {
        this.emit('error', new AIAggregatorError(
          'Failed to parse SSE event',
          'parse_error',
          undefined,
          { error }
        ))
      }
    }

    this.eventSource.onerror = (error) => {
      this.handleConnectionError(error)
    }

    this.eventSource.onopen = () => {
      this.reconnectAttempts = 0
      this.reconnectDelay = 1000
    }
  }

  /**
   * Register event listener
   *
   * @param event - Event type to listen for
   * @param handler - Callback function
   *
   * @example
   * ```typescript
   * execution.on('question', (question) => {
   *   console.log('Question:', question.text)
   * })
   * ```
   */
  on<T = unknown>(event: EventType, handler: EventHandler<T>): this {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(handler as EventHandler)
    return this
  }

  /**
   * Remove event listener
   */
  off<T = unknown>(event: EventType, handler: EventHandler<T>): this {
    const handlers = this.listeners.get(event)
    if (handlers) {
      handlers.delete(handler as EventHandler)
    }
    return this
  }

  /**
   * Submit answer to current intake question
   *
   * @param value - Answer value (any JSON-serializable type)
   *
   * @example
   * ```typescript
   * execution.on('question', async (question) => {
   *   if (question.field === 'email') {
   *     await execution.answer('user@example.com')
   *   }
   * })
   * ```
   */
  async answer(value: unknown): Promise<void> {
    if (!this.currentQuestion) {
      throw new AIAggregatorError(
        'No active question to answer',
        'no_question'
      )
    }

    const url = `${this.config.baseUrl}/api/workflows/executions/${this.executionId}/answer`

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.config.apiKey,
        },
        body: JSON.stringify({
          token: this.currentQuestion.token,
          answer: value,
        }),
      })

      if (!response.ok) {
        const error = await response.json().catch(() => ({ error: 'Unknown error' }))
        throw new AIAggregatorError(
          error.error || 'Failed to submit answer',
          'answer_failed',
          response.status,
          error
        )
      }

      // Clear current question after successful submission
      this.currentQuestion = null
    } catch (error) {
      if (error instanceof AIAggregatorError) {
        throw error
      }
      throw new AIAggregatorError(
        'Network error while submitting answer',
        'network_error',
        undefined,
        { error }
      )
    }
  }

  /**
   * Disconnect from SSE stream and cleanup
   */
  disconnect(): void {
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }
    this.listeners.clear()
    this.currentQuestion = null
  }

  /**
   * Get current execution ID
   */
  getExecutionId(): string {
    return this.executionId
  }

  /**
   * Check if there's an active question waiting for answer
   */
  hasActiveQuestion(): boolean {
    return this.currentQuestion !== null
  }

  /**
   * Get current active question (if any)
   */
  getActiveQuestion(): IntakeQuestionEvent | null {
    return this.currentQuestion
  }

  // Private methods

  private handleEvent(event: WorkflowExecutionEvent): void {
    switch (event.type) {
      case 'intake_question':
        this.currentQuestion = event as IntakeQuestionEvent
        this.emit('question', event)
        break

      case 'intake_answer_recorded':
        this.emit('answer_recorded', event)
        break

      case 'execution_completed':
        this.emit('completed', event)
        this.disconnect() // Auto-disconnect on completion
        break

      case 'execution_failed':
        this.emit('failed', event)
        this.disconnect() // Auto-disconnect on failure
        break

      default:
        // Ignore unknown event types
        break
    }
  }

  private emit(event: EventType, data: unknown): void {
    const handlers = this.listeners.get(event)
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(data)
        } catch (error) {
          console.error(`Error in ${event} handler:`, error)
        }
      })
    }
  }

  private handleConnectionError(error: Event): void {
    // Connection lost, attempt reconnect
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.emit('reconnecting', { attempt: this.reconnectAttempts + 1 })

      if (this.eventSource) {
        this.eventSource.close()
        this.eventSource = null
      }

      setTimeout(() => {
        this.reconnectAttempts++
        this.reconnectDelay *= 2 // Exponential backoff
        this.connect()
      }, this.reconnectDelay)
    } else {
      this.emit('error', new AIAggregatorError(
        'Max reconnection attempts exceeded',
        'connection_failed',
        undefined,
        { error }
      ))
      this.disconnect()
    }
  }
}
