'use client'

import { useState, useCallback, useRef } from 'react'

interface StreamEvent {
  type: 'chunk' | 'done' | 'error'
  content?: string
  tokensIn?: number
  tokensOut?: number
  cost?: number
  error?: string
}

interface StreamRequest {
  prompt: string
  provider?: string
  model?: string
  maxTokens?: number
}

interface StreamStats {
  tokensIn: number
  tokensOut: number
  cost: number
}

interface UseStreamingReturn {
  content: string
  isStreaming: boolean
  error: string | null
  stats: StreamStats | null
  startStream: (request: StreamRequest) => void
  stopStream: () => void
  reset: () => void
}

const API_URL = process.env.NEXT_PUBLIC_API_URL?.replace('/graphql', '') || 'http://localhost:8080'

export function useStreaming(): UseStreamingReturn {
  const [content, setContent] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [stats, setStats] = useState<StreamStats | null>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

  const reset = useCallback(() => {
    setContent('')
    setError(null)
    setStats(null)
  }, [])

  const stopStream = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      abortControllerRef.current = null
    }
    setIsStreaming(false)
  }, [])

  const startStream = useCallback(async (request: StreamRequest) => {
    // Reset state
    reset()
    setIsStreaming(true)
    setError(null)

    // Create abort controller
    abortControllerRef.current = new AbortController()

    try {
      const response = await fetch(`${API_URL}/stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': 'dev-api-key-12345',
        },
        body: JSON.stringify({
          prompt: request.prompt,
          provider: request.provider,
          model: request.model,
          maxTokens: request.maxTokens,
        }),
        signal: abortControllerRef.current.signal,
      })

      if (!response.ok) {
        const text = await response.text()
        throw new Error(text || `HTTP ${response.status}`)
      }

      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('No response body')
      }

      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })

        // Process complete SSE events
        const lines = buffer.split('\n')
        buffer = lines.pop() || '' // Keep incomplete line in buffer

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const event: StreamEvent = JSON.parse(line.slice(6))

              switch (event.type) {
                case 'chunk':
                  if (event.content) {
                    setContent((prev) => prev + event.content)
                  }
                  break
                case 'done':
                  setStats({
                    tokensIn: event.tokensIn || 0,
                    tokensOut: event.tokensOut || 0,
                    cost: event.cost || 0,
                  })
                  setIsStreaming(false)
                  break
                case 'error':
                  setError(event.error || 'Unknown error')
                  setIsStreaming(false)
                  break
              }
            } catch (e) {
              // Ignore JSON parse errors for incomplete chunks
            }
          }
        }
      }
    } catch (err) {
      if (err instanceof Error) {
        if (err.name === 'AbortError') {
          // User cancelled, not an error
        } else {
          setError(err.message)
        }
      } else {
        setError('Unknown error occurred')
      }
    } finally {
      setIsStreaming(false)
      abortControllerRef.current = null
    }
  }, [reset])

  return {
    content,
    isStreaming,
    error,
    stats,
    startStream,
    stopStream,
    reset,
  }
}
