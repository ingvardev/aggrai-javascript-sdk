'use client'

import { useQuery } from '@tanstack/react-query'
import { graphqlClient, UsageSummary, USAGE_SUMMARY_QUERY } from '@/lib/api'
import { useUsageSubscription } from './use-subscriptions'

interface UsageSummaryResponse {
  usageSummary: UsageSummary[]
}

export function useUsageSummary() {
  const query = useQuery({
    queryKey: ['usage', 'summary'],
    queryFn: async () => {
      const data = await graphqlClient.request<UsageSummaryResponse>(USAGE_SUMMARY_QUERY)
      return data.usageSummary
    },
    // No need for refetchInterval - subscription handles real-time updates
  })

  // Subscribe to real-time usage updates
  useUsageSubscription({ enabled: true })

  return query
}

// Compute totals from usage summary
export function useUsageStats() {
  const { data: usageSummary, ...rest } = useUsageSummary()

  const stats = usageSummary?.reduce(
    (acc, usage) => ({
      totalJobs: acc.totalJobs + usage.jobCount,
      totalTokensIn: acc.totalTokensIn + usage.totalTokensIn,
      totalTokensOut: acc.totalTokensOut + usage.totalTokensOut,
      totalCost: acc.totalCost + usage.totalCost,
    }),
    { totalJobs: 0, totalTokensIn: 0, totalTokensOut: 0, totalCost: 0 }
  ) || { totalJobs: 0, totalTokensIn: 0, totalTokensOut: 0, totalCost: 0 }

  return { data: stats, ...rest }
}
