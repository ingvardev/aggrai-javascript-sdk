'use client'

import Link from 'next/link'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { formatDate, formatCurrency, truncate } from '@/lib/utils'
import { Eye, MoreHorizontal, RefreshCw, Loader2, Wifi, WifiOff } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useJobs } from '@/lib/hooks'
import { useJobSubscription } from '@/lib/hooks/use-subscriptions'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { useState } from 'react'

const statusVariant = {
  PENDING: 'secondary',
  PROCESSING: 'warning',
  COMPLETED: 'success',
  FAILED: 'destructive',
} as const

const typeVariant = {
  TEXT: 'outline',
  IMAGE: 'default',
} as const

export function JobsTable() {
  const { data: jobs, isLoading, error, isFetching } = useJobs()
  const queryClient = useQueryClient()
  const [isConnected, setIsConnected] = useState(true)

  // Subscribe to real-time job updates
  useJobSubscription({
    enabled: true,
    onJobUpdate: (job) => {
      toast.success(`Job ${job.status.toLowerCase()}`, {
        description: `${truncate(job.input, 30)}`,
      })
    },
  })

  const handleRefresh = () => {
    queryClient.invalidateQueries({ queryKey: ['jobs'] })
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>All Jobs</CardTitle>
            <CardDescription>
              A list of all your AI processing jobs
              {jobs?.pageInfo.totalCount ? ` (${jobs.pageInfo.totalCount} total)` : ''}
            </CardDescription>
          </div>
          <Button variant="outline" size="sm" onClick={handleRefresh} disabled={isFetching}>
            <RefreshCw className={`mr-2 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="py-12 text-center text-sm text-muted-foreground">
            Failed to load jobs. Is the API running at localhost:8080?
          </div>
        ) : jobs?.edges.length === 0 ? (
          <div className="py-12 text-center text-sm text-muted-foreground">
            No jobs yet. Create your first job!
          </div>
        ) : (
          <div className="relative overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="pb-3 text-left font-medium text-muted-foreground">
                    Input
                  </th>
                  <th className="pb-3 text-left font-medium text-muted-foreground">
                    Type
                  </th>
                  <th className="pb-3 text-left font-medium text-muted-foreground">
                    Status
                  </th>
                  <th className="pb-3 text-left font-medium text-muted-foreground">
                    Provider
                  </th>
                  <th className="pb-3 text-right font-medium text-muted-foreground">
                    Tokens
                  </th>
                  <th className="pb-3 text-center font-medium text-muted-foreground">
                    Cost
                  </th>
                  <th className="pb-3 text-center font-medium text-muted-foreground">
                    Created
                  </th>
                  <th className="pb-3 text-right font-medium text-muted-foreground">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {jobs?.edges.map(({ node: job }) => (
                  <tr key={job.id} className="border-b last:border-0">
                    <td className="py-3 pr-4">
                      <div className="max-w-[200px]">
                        <p className="truncate font-medium">
                          {truncate(job.input, 40)}
                        </p>
                        <p className="text-xs text-muted-foreground truncate">
                          {job.id.slice(0, 8)}...
                        </p>
                      </div>
                    </td>
                    <td className="py-3">
                      <Badge variant={typeVariant[job.type]}>
                        {job.type}
                      </Badge>
                    </td>
                    <td className="py-3">
                      <Badge variant={statusVariant[job.status]}>
                        {job.status}
                      </Badge>
                    </td>
                    <td className="py-3 text-muted-foreground">
                      {job.provider || '-'}
                    </td>
                    <td className="py-3 text-right tabular-nums">
                      {job.tokensIn + job.tokensOut}
                    </td>
                    <td className="py-3 text-center tabular-nums">
                      {formatCurrency(job.cost)}
                    </td>
                    <td className="py-3 text-center text-muted-foreground">
                      {formatDate(job.createdAt)}
                    </td>
                    <td className="py-3 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem asChild>
                            <Link href={`/jobs/${job.id}`}>
                              <Eye className="mr-2 h-4 w-4" />
                              View Details
                            </Link>
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
