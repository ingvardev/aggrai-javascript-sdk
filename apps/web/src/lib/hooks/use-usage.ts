'use client'

import { useQuery } from '@tanstack/react-query'
import { graphqlClient, UsageSummary, USAGE_SUMMARY_QUERY } from '@/lib/api'

interface UsageSummaryResponse {
  usageSummary: UsageSummary[]
}

export function useUsageSummary() {
  return useQuery({
    queryKey: ['usage', 'summary'],
    queryFn: async () => {
      const data = await graphqlClient.request<UsageSummaryResponse>(USAGE_SUMMARY_QUERY)
      return data.usageSummary
    },
    refetchInterval: 30000,
  })
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
