'use client'

import Link from 'next/link'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Loader2, Wifi } from 'lucide-react'
import { formatDate, truncate } from '@/lib/utils'
import { useRecentJobs } from '@/lib/hooks'
import { useJobSubscription } from '@/lib/hooks/use-subscriptions'

const statusVariant = {
  PENDING: 'secondary',
  PROCESSING: 'warning',
  COMPLETED: 'success',
  FAILED: 'destructive',
} as const

export function RecentJobs() {
  const { data: jobs, isLoading, error } = useRecentJobs()

  // Subscribe to real-time job updates (updates cache automatically)
  useJobSubscription({ enabled: true })

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Recent Jobs</CardTitle>
        <Link
          href="/jobs"
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          View all →
        </Link>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            Failed to load jobs. Is the API running?
          </div>
        ) : jobs?.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            No jobs yet. Create your first job!
          </div>
        ) : (
          <div className="space-y-4">
            {jobs?.map((job) => (
              <Link
                key={job.id}
                href={`/jobs/${job.id}`}
                className="flex items-center justify-between rounded-lg border p-3 transition-colors hover:bg-accent"
              >
                <div className="flex-1 min-w-0 mr-4">
                  <p className="text-sm font-medium truncate">
                    {truncate(job.input, 50)}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {formatDate(job.createdAt)}
                  </p>
                </div>
                <Badge variant={statusVariant[job.status]}>
                  {job.status.toLowerCase()}
                </Badge>
              </Link>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
