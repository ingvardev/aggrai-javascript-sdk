import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { AIAggregator, AIAggregatorError } from '../index'

describe('AIAggregator', () => {
  let client: AIAggregator
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    client = new AIAggregator({
      baseUrl: 'https://api.test.com',
      apiKey: 'test-key',
      pollingInterval: 10,
      maxPollingAttempts: 5,
    })

    fetchMock = vi.fn()
    global.fetch = fetchMock
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('chat', () => {
    it('should create chat job and wait for result', async () => {
      // POST /api/chat - create job
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ jobId: 'job-123', status: 'pending' }),
      })

      // GET /api/chat/job-123 - poll status (completed)
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () =>
          Promise.resolve({
            id: 'job-123',
            status: 'completed',
            output: 'Hello back!',
            tokensIn: 10,
            tokensOut: 20,
            cost: 0.001,
            provider: 'openai',
            model: 'gpt-4o-mini',
          }),
      })

      const result = await client.chat({
        messages: [{ role: 'user', content: 'Hello!' }],
      })

      expect(result.content).toBe('Hello back!')
      expect(result.usage.tokensIn).toBe(10)
      expect(result.usage.tokensOut).toBe(20)
      expect(result.usage.cost).toBe(0.001)
      expect(result.jobId).toBe('job-123')

      // Check POST /api/chat was called
      expect(fetchMock).toHaveBeenCalledWith(
        'https://api.test.com/api/chat',
        expect.objectContaining({
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-API-Key': 'test-key',
          },
        })
      )
    })

    it('should throw on error response from /api/chat', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: () => Promise.resolve({ error: 'unauthorized', message: 'Invalid API key' }),
      })

      await expect(client.chat({ prompt: 'Hello' })).rejects.toThrow(AIAggregatorError)
    })
  })

  describe('chatAsync', () => {
    it('should return job ID without waiting', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ jobId: 'job-456', status: 'pending' }),
      })

      const response = await client.chatAsync({
        prompt: 'Hello',
        provider: 'claude',
      })

      expect(response.jobId).toBe('job-456')
      expect(response.status).toBe('pending')
      expect(fetchMock).toHaveBeenCalledTimes(1)
    })
  })

  describe('getJobStatus', () => {
    it('should get job status by ID', async () => {
      const job = {
        id: 'job-123',
        status: 'processing',
        tokensIn: 10,
      }

      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(job),
      })

      const result = await client.getJobStatus('job-123')

      expect(result.id).toBe('job-123')
      expect(result.status).toBe('processing')

      expect(fetchMock).toHaveBeenCalledWith(
        'https://api.test.com/api/chat/job-123',
        expect.objectContaining({ method: 'GET' })
      )
    })
  })

  describe('waitForJob', () => {
    it('should poll until job is completed', async () => {
      // First poll: pending
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: 'job-123', status: 'pending' }),
      })

      // Second poll: processing
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: 'job-123', status: 'processing' }),
      })

      // Third poll: completed
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () =>
          Promise.resolve({
            id: 'job-123',
            status: 'completed',
            output: 'Done!',
            tokensIn: 10,
            tokensOut: 20,
            cost: 0.001,
          }),
      })

      const result = await client.waitForJob('job-123')

      expect(result.content).toBe('Done!')
      expect(result.usage.tokensIn).toBe(10)
      expect(fetchMock).toHaveBeenCalledTimes(3)
    })

    it('should throw on job failure', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () =>
          Promise.resolve({
            id: 'job-123',
            status: 'failed',
            error: 'Provider error',
          }),
      })

      await expect(client.waitForJob('job-123')).rejects.toThrow('Provider error')
    })

    it('should throw on timeout', async () => {
      // Always return pending
      fetchMock.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ id: 'job-123', status: 'pending' }),
      })

      await expect(client.waitForJob('job-123')).rejects.toThrow('Job polling timeout')
    })
  })

  describe('cancelJob', () => {
    it('should cancel a job', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({}),
      })

      await client.cancelJob('job-123')

      expect(fetchMock).toHaveBeenCalledWith(
        'https://api.test.com/api/chat/job-123',
        expect.objectContaining({ method: 'DELETE' })
      )
    })
  })

  describe('with tools', () => {
    it('should send tools and receive tool calls', async () => {
      // Create job
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ jobId: 'job-789', status: 'pending' }),
      })

      // Get job - completed with tool calls
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () =>
          Promise.resolve({
            id: 'job-789',
            status: 'completed',
            output: '',
            finishReason: 'tool_calls',
            toolCalls: [
              {
                id: 'call_123',
                type: 'function',
                function: {
                  name: 'get_weather',
                  arguments: '{"location": "Paris"}',
                },
              },
            ],
          }),
      })

      const result = await client.chat({
        messages: [{ role: 'user', content: "What's the weather in Paris?" }],
        tools: [
          {
            type: 'function',
            function: {
              name: 'get_weather',
              description: 'Get weather',
              parameters: { type: 'object', properties: { location: { type: 'string' } } },
            },
          },
        ],
      })

      expect(result.finishReason).toBe('tool_calls')
      expect(result.toolCalls).toHaveLength(1)
      expect(result.toolCalls?.[0].function.name).toBe('get_weather')
    })
  })
})
