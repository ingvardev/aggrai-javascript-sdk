'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Zap, CheckCircle, Clock, DollarSign, Loader2 } from 'lucide-react'
import { useJobs, useUsageStats } from '@/lib/hooks'
import { formatCurrency, formatNumber } from '@/lib/utils'

export function StatsCards() {
  const { data: jobs, isLoading: jobsLoading } = useJobs()
  const { data: usageStats, isLoading: usageLoading } = useUsageStats()

  const isLoading = jobsLoading || usageLoading

  const totalJobs = jobs?.pageInfo.totalCount || 0
  const completedJobs = jobs?.edges.filter((e) => e.node.status === 'COMPLETED').length || 0
  const successRate = totalJobs > 0 ? ((completedJobs / Math.min(totalJobs, 20)) * 100).toFixed(1) : '0'

  const stats = [
    {
      title: 'Total Jobs',
      value: isLoading ? '...' : formatNumber(totalJobs),
      description: 'All time',
      icon: Zap,
    },
    {
      title: 'Completed',
      value: isLoading ? '...' : formatNumber(usageStats.totalJobs),
      description: `${successRate}% success rate`,
      icon: CheckCircle,
    },
    {
      title: 'Total Tokens',
      value: isLoading ? '...' : formatNumber(usageStats.totalTokensIn + usageStats.totalTokensOut),
      description: `In: ${formatNumber(usageStats.totalTokensIn)} / Out: ${formatNumber(usageStats.totalTokensOut)}`,
      icon: Clock,
    },
    {
      title: 'Total Cost',
      value: isLoading ? '...' : formatCurrency(usageStats.totalCost),
      description: 'All time',
      icon: DollarSign,
    },
  ]

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {stats.map((stat, index) => (
        <Card
          key={stat.title}
          className="hover-lift hover-glow animate-fade-in"
          style={{ animationDelay: `${index * 50}ms` }}
        >
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{stat.title}</CardTitle>
            {isLoading ? (
              <Loader2 className="h-4 w-4 text-muted-foreground animate-spin" />
            ) : (
              <stat.icon className="h-4 w-4 text-muted-foreground" />
            )}
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stat.value}</div>
            <p className="text-xs text-muted-foreground">{stat.description}</p>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
