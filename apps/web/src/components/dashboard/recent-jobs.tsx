import Link from 'next/link'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { formatDate, truncate } from '@/lib/utils'

const recentJobs = [
  {
    id: '1',
    input: 'Explain quantum computing in simple terms',
    status: 'completed',
    createdAt: new Date().toISOString(),
  },
  {
    id: '2',
    input: 'Generate a product description for a smartwatch',
    status: 'processing',
    createdAt: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
  },
  {
    id: '3',
    input: 'Translate this text to French: Hello world',
    status: 'completed',
    createdAt: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
  },
  {
    id: '4',
    input: 'Summarize the latest news about AI',
    status: 'failed',
    createdAt: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
  },
  {
    id: '5',
    input: 'Write a haiku about programming',
    status: 'pending',
    createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
  },
]

const statusVariant = {
  pending: 'secondary',
  processing: 'warning',
  completed: 'success',
  failed: 'destructive',
} as const

export function RecentJobs() {
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
        <div className="space-y-4">
          {recentJobs.map((job) => (
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
              <Badge variant={statusVariant[job.status as keyof typeof statusVariant]}>
                {job.status}
              </Badge>
            </Link>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
