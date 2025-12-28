'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { graphqlClient, Job, JOBS_QUERY, JOB_QUERY, CREATE_JOB_MUTATION } from '@/lib/api'

interface JobsResponse {
  jobs: {
    edges: Array<{ node: Job; cursor: string }>
    pageInfo: {
      totalCount: number
      hasNextPage: boolean
      hasPreviousPage: boolean
    }
  }
}

interface JobResponse {
  job: Job
}

interface CreateJobResponse {
  createJob: Job
}

interface JobsFilter {
  status?: 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED'
  type?: 'TEXT' | 'IMAGE'
}

interface PaginationInput {
  limit?: number
  offset?: number
}

export function useJobs(filter?: JobsFilter, pagination?: PaginationInput) {
  return useQuery({
    queryKey: ['jobs', filter, pagination],
    queryFn: async () => {
      const data = await graphqlClient.request<JobsResponse>(JOBS_QUERY, {
        filter,
        pagination: pagination || { limit: 20, offset: 0 },
      })
      return data.jobs
    },
    // No polling - use WebSocket subscriptions for real-time updates
    // See useJobSubscription() hook
  })
}

export function useJob(id: string) {
  return useQuery({
    queryKey: ['job', id],
    queryFn: async () => {
      const data = await graphqlClient.request<JobResponse>(JOB_QUERY, { id })
      return data.job
    },
    enabled: !!id,
    // No polling - use WebSocket subscriptions for real-time updates
    // See useJobStatusSubscription() hook
  })
}

export function useCreateJob() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (input: { type: 'TEXT' | 'IMAGE'; input: string }) => {
      const data = await graphqlClient.request<CreateJobResponse>(
        CREATE_JOB_MUTATION,
        { input }
      )
      return data.createJob
    },
    onSuccess: () => {
      // Invalidate jobs query to refetch
      queryClient.invalidateQueries({ queryKey: ['jobs'] })
    },
  })
}

// Helper to get recent jobs (last 5)
export function useRecentJobs() {
  return useQuery({
    queryKey: ['jobs', 'recent'],
    queryFn: async () => {
      const data = await graphqlClient.request<JobsResponse>(JOBS_QUERY, {
        pagination: { limit: 5, offset: 0 },
      })
      return data.jobs.edges.map((e) => e.node)
    },
    // No polling - use WebSocket subscriptions for real-time updates
  })
}
