'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { useUsageSummary } from '@/lib/hooks'
import { BarChart3, TrendingUp, DollarSign, Zap, Activity } from 'lucide-react'

export default function UsagePage() {
  const { data: usageSummary, isLoading } = useUsageSummary()

  const totalCost = usageSummary?.reduce((acc, u) => acc + u.totalCost, 0) ?? 0
  const totalTokensIn = usageSummary?.reduce((acc, u) => acc + u.totalTokensIn, 0) ?? 0
  const totalTokensOut = usageSummary?.reduce((acc, u) => acc + u.totalTokensOut, 0) ?? 0
  const totalJobs = usageSummary?.reduce((acc, u) => acc + u.jobCount, 0) ?? 0

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Usage Analytics</h1>
        <p className="text-muted-foreground">
          Monitor your API usage and costs across all providers
        </p>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Cost</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <>
                <div className="text-2xl font-bold">${totalCost.toFixed(4)}</div>
                <p className="text-xs text-muted-foreground">This month</p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Jobs</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <>
                <div className="text-2xl font-bold">{totalJobs.toLocaleString()}</div>
                <p className="text-xs text-muted-foreground">Processed requests</p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Input Tokens</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <>
                <div className="text-2xl font-bold">{totalTokensIn.toLocaleString()}</div>
                <p className="text-xs text-muted-foreground">Tokens sent</p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Output Tokens</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <>
                <div className="text-2xl font-bold">{totalTokensOut.toLocaleString()}</div>
                <p className="text-xs text-muted-foreground">Tokens received</p>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Usage by Provider */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BarChart3 className="h-5 w-5" />
            Usage by Provider
          </CardTitle>
          <CardDescription>
            Breakdown of usage and costs per AI provider
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="flex items-center justify-between">
                  <Skeleton className="h-6 w-32" />
                  <Skeleton className="h-6 w-24" />
                </div>
              ))}
            </div>
          ) : usageSummary && usageSummary.length > 0 ? (
            <div className="space-y-4">
              {usageSummary.map((provider) => {
                const percentage = totalCost > 0 ? (provider.totalCost / totalCost) * 100 : 0
                return (
                  <div key={provider.provider} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="font-medium capitalize">{provider.provider}</span>
                        <span className="text-sm text-muted-foreground">
                          {provider.jobCount} jobs
                        </span>
                      </div>
                      <div className="text-right">
                        <span className="font-medium">${provider.totalCost.toFixed(4)}</span>
                        <span className="ml-2 text-sm text-muted-foreground">
                          ({percentage.toFixed(1)}%)
                        </span>
                      </div>
                    </div>
                    <div className="h-2 w-full overflow-hidden rounded-full bg-secondary">
                      <div
                        className="h-full bg-primary transition-all"
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                    <div className="flex justify-between text-xs text-muted-foreground">
                      <span>{provider.totalTokensIn.toLocaleString()} input tokens</span>
                      <span>{provider.totalTokensOut.toLocaleString()} output tokens</span>
                    </div>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <BarChart3 className="h-12 w-12 text-muted-foreground/50" />
              <h3 className="mt-4 text-lg font-medium">No usage data yet</h3>
              <p className="mt-2 text-sm text-muted-foreground">
                Start processing jobs to see your usage analytics here
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Usage Tips */}
      <Card>
        <CardHeader>
          <CardTitle>Cost Optimization Tips</CardTitle>
          <CardDescription>
            Ways to reduce your API costs
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ul className="space-y-2 text-sm text-muted-foreground">
            <li className="flex items-start gap-2">
              <span className="text-primary">•</span>
              Use shorter prompts when possible to reduce input token costs
            </li>
            <li className="flex items-start gap-2">
              <span className="text-primary">•</span>
              Set max_tokens limits to control output length and costs
            </li>
            <li className="flex items-start gap-2">
              <span className="text-primary">•</span>
              Consider using Ollama for development and testing (free local inference)
            </li>
            <li className="flex items-start gap-2">
              <span className="text-primary">•</span>
              Use the auto provider selection to get the best price/performance ratio
            </li>
          </ul>
        </CardContent>
      </Card>
    </div>
  )
}
