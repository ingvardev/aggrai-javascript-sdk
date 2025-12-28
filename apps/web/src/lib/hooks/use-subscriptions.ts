'use client'

import { useEffect, useCallback, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { getWsClient } from '@/lib/ws-client'
import { Job } from '@/lib/api'

// GraphQL subscription queries
const JOB_UPDATED_SUBSCRIPTION = `
  subscription JobUpdated {
    jobUpdated {
      id
      tenantId
      type
      input
      status
      result
      error
      provider
      tokensIn
      tokensOut
      cost
      createdAt
      updatedAt
      startedAt
      finishedAt
    }
  }
`

const JOB_STATUS_CHANGED_SUBSCRIPTION = `
  subscription JobStatusChanged($jobId: ID!) {
    jobStatusChanged(jobId: $jobId) {
      id
      tenantId
      type
      input
      status
      result
      error
      provider
      tokensIn
      tokensOut
      cost
      createdAt
      updatedAt
      startedAt
      finishedAt
    }
  }
`

interface SubscriptionOptions {
  onJobUpdate?: (job: Job) => void
  enabled?: boolean
}

/**
 * Hook to subscribe to all job updates for the current tenant.
 * Automatically updates the React Query cache when jobs change.
 */
export function useJobSubscription(options: SubscriptionOptions = {}) {
  const { onJobUpdate, enabled = true } = options
  const queryClient = useQueryClient()
  const unsubscribeRef = useRef<(() => void) | null>(null)

  const updateJobInCache = useCallback(
    (job: Job) => {
      // Update the jobs list cache (matches all queries starting with ['jobs', ...])
      queryClient.setQueriesData<{
        edges?: Array<{ node: Job; cursor: string }>
        pageInfo?: { totalCount: number; hasNextPage: boolean; hasPreviousPage: boolean }
      }>({ queryKey: ['jobs'] }, (oldData) => {
        if (!oldData?.edges) return oldData

        const existingIndex = oldData.edges.findIndex((edge) => edge.node.id === job.id)

        if (existingIndex >= 0) {
          // Update existing job
          const newEdges = [...oldData.edges]
          newEdges[existingIndex] = { ...newEdges[existingIndex], node: job }
          return { ...oldData, edges: newEdges }
        } else {
          // New job - add to the beginning
          return {
            ...oldData,
            edges: [{ node: job, cursor: job.id }, ...oldData.edges],
            pageInfo: oldData.pageInfo ? {
              ...oldData.pageInfo,
              totalCount: oldData.pageInfo.totalCount + 1,
            } : undefined,
          }
        }
      })

      // Update the individual job cache (useJob returns Job directly, not { job: Job })
      queryClient.setQueryData(['job', job.id], job)

      // Call the callback if provided
      onJobUpdate?.(job)
    },
    [queryClient, onJobUpdate]
  )

  useEffect(() => {
    if (!enabled) return

    const client = getWsClient()

    const unsubscribe = client.subscribe<{ jobUpdated: Job }>(
      { query: JOB_UPDATED_SUBSCRIPTION },
      {
        next: (data) => {
          if (data.data?.jobUpdated) {
            console.log('[Subscription] Job update received:', data.data.jobUpdated)
            updateJobInCache(data.data.jobUpdated)
          }
        },
        error: (err) => {
          console.error('[Subscription] Error:', err)
        },
        complete: () => {
          console.log('[Subscription] Completed')
        },
      }
    )

    unsubscribeRef.current = unsubscribe

    return () => {
      unsubscribe()
    }
  }, [enabled, updateJobInCache])

  const unsubscribe = useCallback(() => {
    unsubscribeRef.current?.()
  }, [])

  return { unsubscribe }
}

/**
 * Hook to subscribe to updates for a specific job.
 */
export function useJobStatusSubscription(
  jobId: string,
  options: SubscriptionOptions = {}
) {
  const { onJobUpdate, enabled = true } = options
  const queryClient = useQueryClient()
  const unsubscribeRef = useRef<(() => void) | null>(null)

  useEffect(() => {
    if (!enabled || !jobId) return

    const client = getWsClient()

    const unsubscribe = client.subscribe<{ jobStatusChanged: Job }>(
      {
        query: JOB_STATUS_CHANGED_SUBSCRIPTION,
        variables: { jobId },
      },
      {
        next: (data) => {
          if (data.data?.jobStatusChanged) {
            const job = data.data.jobStatusChanged

            // Update the individual job cache (useJob returns Job directly)
            queryClient.setQueryData(['job', job.id], job)

            // Also update in the jobs list
            queryClient.setQueriesData<{
              edges?: Array<{ node: Job; cursor: string }>
              pageInfo?: { totalCount: number; hasNextPage: boolean; hasPreviousPage: boolean }
            }>({ queryKey: ['jobs'] }, (oldData) => {
              if (!oldData?.edges) return oldData

              const existingIndex = oldData.edges.findIndex((edge) => edge.node.id === job.id)
              if (existingIndex >= 0) {
                const newEdges = [...oldData.edges]
                newEdges[existingIndex] = { ...newEdges[existingIndex], node: job }
                return { ...oldData, edges: newEdges }
              }
              return oldData
            })

            onJobUpdate?.(job)
          }
        },
        error: (err) => {
          console.error('[Subscription] Job status error:', err)
        },
        complete: () => {
          console.log('[Subscription] Job status completed')
        },
      }
    )

    unsubscribeRef.current = unsubscribe

    return () => {
      unsubscribe()
    }
  }, [enabled, jobId, queryClient, onJobUpdate])

  const unsubscribe = useCallback(() => {
    unsubscribeRef.current?.()
  }, [])

  return { unsubscribe }
}

// Usage subscription query
const USAGE_UPDATED_SUBSCRIPTION = `
  subscription UsageUpdated {
    usageUpdated {
      provider
      totalTokensIn
      totalTokensOut
      totalCost
      jobCount
    }
  }
`

interface UsageSummary {
  provider: string
  totalTokensIn: number
  totalTokensOut: number
  totalCost: number
  jobCount: number
}

interface UsageSubscriptionOptions {
  onUsageUpdate?: (usage: UsageSummary[]) => void
  enabled?: boolean
}

/**
 * Hook to subscribe to usage summary updates for the current tenant.
 * Automatically updates the React Query cache when usage changes.
 */
export function useUsageSubscription(options: UsageSubscriptionOptions = {}) {
  const { onUsageUpdate, enabled = true } = options
  const queryClient = useQueryClient()
  const unsubscribeRef = useRef<(() => void) | null>(null)

  useEffect(() => {
    if (!enabled) return

    const client = getWsClient()

    const unsubscribe = client.subscribe<{ usageUpdated: UsageSummary[] }>(
      { query: USAGE_UPDATED_SUBSCRIPTION },
      {
        next: (data) => {
          if (data.data?.usageUpdated) {
            console.log('[Subscription] Usage update received:', data.data.usageUpdated)

            // Update the usage summary cache
            queryClient.setQueryData(['usage', 'summary'], data.data.usageUpdated)

            // Call the callback if provided
            onUsageUpdate?.(data.data.usageUpdated)
          }
        },
        error: (err) => {
          console.error('[Subscription] Usage error:', err)
        },
        complete: () => {
          console.log('[Subscription] Usage completed')
        },
      }
    )

    unsubscribeRef.current = unsubscribe

    return () => {
      unsubscribe()
    }
  }, [enabled, queryClient, onUsageUpdate])

  const unsubscribe = useCallback(() => {
    unsubscribeRef.current?.()
  }, [])

  return { unsubscribe }
}
