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
import { ArrowLeft, Copy, RefreshCw } from 'lucide-react'

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
  // Mock job data
  const job = {
    id: jobId,
    type: 'TEXT',
    input: 'Explain quantum computing in simple terms that a beginner can understand. Include examples of real-world applications.',
    status: 'COMPLETED',
    result: 'Quantum computing is a revolutionary approach to computation that harnesses the principles of quantum mechanics. Unlike classical computers that use bits (0s and 1s), quantum computers use quantum bits or "qubits" that can exist in multiple states simultaneously through a phenomenon called superposition.\n\nImagine a coin spinning in the air - before it lands, it\'s effectively both heads and tails at the same time. This is similar to how qubits work. Additionally, qubits can be "entangled," meaning the state of one qubit instantly affects another, regardless of distance.\n\nReal-world applications include:\n1. Drug discovery - simulating molecular interactions\n2. Cryptography - breaking and creating secure codes\n3. Financial modeling - complex risk analysis\n4. Climate modeling - more accurate predictions\n5. Optimization problems - logistics and supply chains',
    provider: 'stub-provider',
    tokensIn: 25,
    tokensOut: 180,
    cost: 0.00205,
    createdAt: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 29 * 60 * 1000).toISOString(),
    startedAt: new Date(Date.now() - 30 * 60 * 1000 + 100).toISOString(),
    finishedAt: new Date(Date.now() - 29 * 60 * 1000).toISOString(),
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
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
          variant={statusVariant[job.status as keyof typeof statusVariant]}
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
                  onClick={() => copyToClipboard(job.result)}
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </CardHeader>
              <CardContent>
                <p className="text-sm whitespace-pre-wrap">{job.result}</p>
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
                <span className="text-sm font-medium">{job.provider}</span>
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
