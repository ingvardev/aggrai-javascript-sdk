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
import { formatDate, formatCurrency, formatNumber } from '@/lib/utils'
import { ArrowLeft, Copy, RefreshCw, Loader2, AlertCircle, Wifi } from 'lucide-react'
import { useJob } from '@/lib/hooks'
import { useJobStatusSubscription } from '@/lib/hooks/use-subscriptions'
import { toast } from 'sonner'

interface JobDetailsProps {
  jobId: string
}

const statusVariant = {
  PENDING: 'secondary',
  PROCESSING: 'warning',
  COMPLETED: 'success',
  FAILED: 'destructive',
} as const

export function JobDetails({ jobId }: JobDetailsProps) {
  const { data: job, isLoading, error } = useJob(jobId)

  // Subscribe to real-time updates for this specific job
  useJobStatusSubscription(jobId, {
    enabled: !!jobId,
    onJobUpdate: (updatedJob) => {
      if (updatedJob.status === 'COMPLETED') {
        toast.success('Job completed!', {
          description: 'The AI has finished processing your request.',
        })
      } else if (updatedJob.status === 'FAILED') {
        toast.error('Job failed', {
          description: updatedJob.error || 'An error occurred',
        })
      }
    },
  })

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-24">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error || !job) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" asChild>
            <Link href="/jobs">
              <ArrowLeft className="h-4 w-4" />
            </Link>
          </Button>
          <h1 className="text-2xl font-semibold tracking-tight">Job Not Found</h1>
        </div>
        <Card>
          <CardContent className="flex items-center gap-3 py-8">
            <AlertCircle className="h-5 w-5 text-destructive" />
            <p className="text-muted-foreground">
              Could not load job details. The job may not exist or the API is unavailable.
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link href="/jobs">
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div className="flex-1">
          <h1 className="text-2xl font-semibold tracking-tight">Job Details</h1>
          <p className="text-sm text-muted-foreground font-mono">
            {job.id}
          </p>
        </div>
        <Badge
          variant={statusVariant[job.status]}
          className="text-sm"
        >
          {job.status}
        </Badge>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Main content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Input */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <div>
                <CardTitle className="text-base">Input</CardTitle>
                <CardDescription>The prompt sent to the AI</CardDescription>
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => copyToClipboard(job.input)}
              >
                <Copy className="h-4 w-4" />
              </Button>
            </CardHeader>
            <CardContent>
              <p className="text-sm whitespace-pre-wrap">{job.input}</p>
            </CardContent>
          </Card>

          {/* Result */}
          {job.result && (
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0">
                <div>
                  <CardTitle className="text-base">Result</CardTitle>
                  <CardDescription>AI generated response</CardDescription>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => copyToClipboard(job.result!)}
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </CardHeader>
              <CardContent>
                <p className="text-sm whitespace-pre-wrap">{job.result}</p>
              </CardContent>
            </Card>
          )}

          {/* Error */}
          {job.error && (
            <Card className="border-destructive">
              <CardHeader>
                <CardTitle className="text-base text-destructive">Error</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-destructive whitespace-pre-wrap">{job.error}</p>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Metadata */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Metadata</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Type</span>
                <Badge variant="outline">{job.type}</Badge>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Provider</span>
                <span className="text-sm font-medium">{job.provider || '-'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Created</span>
                <span className="text-sm">{formatDate(job.createdAt)}</span>
              </div>
              {job.startedAt && (
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Started</span>
                  <span className="text-sm">{formatDate(job.startedAt)}</span>
                </div>
              )}
              {job.finishedAt && (
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Finished</span>
                  <span className="text-sm">{formatDate(job.finishedAt)}</span>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Usage */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Usage</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Input Tokens</span>
                <span className="text-sm font-medium tabular-nums">
                  {formatNumber(job.tokensIn)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Output Tokens</span>
                <span className="text-sm font-medium tabular-nums">
                  {formatNumber(job.tokensOut)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Total Tokens</span>
                <span className="text-sm font-medium tabular-nums">
                  {formatNumber(job.tokensIn + job.tokensOut)}
                </span>
              </div>
              <div className="border-t pt-4">
                <div className="flex justify-between">
                  <span className="text-sm font-medium">Cost</span>
                  <span className="text-sm font-bold tabular-nums">
                    {formatCurrency(job.cost)}
                  </span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Actions */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Button variant="outline" className="w-full">
                <RefreshCw className="mr-2 h-4 w-4" />
                Retry Job
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
