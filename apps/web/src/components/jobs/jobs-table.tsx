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
import { Eye, MoreHorizontal, RefreshCw } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

const jobs = [
  {
    id: '550e8400-e29b-41d4-a716-446655440001',
    type: 'TEXT',
    input: 'Explain quantum computing in simple terms that a beginner can understand',
    status: 'COMPLETED',
    provider: 'stub-provider',
    tokensIn: 15,
    tokensOut: 250,
    cost: 0.0025,
    createdAt: new Date().toISOString(),
  },
  {
    id: '550e8400-e29b-41d4-a716-446655440002',
    type: 'TEXT',
    input: 'Generate a product description for a smartwatch with health features',
    status: 'PROCESSING',
    provider: 'stub-provider',
    tokensIn: 12,
    tokensOut: 0,
    cost: 0,
    createdAt: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
  },
  {
    id: '550e8400-e29b-41d4-a716-446655440003',
    type: 'IMAGE',
    input: 'A futuristic cityscape with flying cars and neon lights',
    status: 'COMPLETED',
    provider: 'stub-provider',
    tokensIn: 0,
    tokensOut: 0,
    cost: 0.02,
    createdAt: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
  },
  {
    id: '550e8400-e29b-41d4-a716-446655440004',
    type: 'TEXT',
    input: 'Translate the following text to French',
    status: 'FAILED',
    provider: 'stub-provider',
    tokensIn: 8,
    tokensOut: 0,
    cost: 0,
    createdAt: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
  },
  {
    id: '550e8400-e29b-41d4-a716-446655440005',
    type: 'TEXT',
    input: 'Write a haiku about programming and debugging code',
    status: 'PENDING',
    provider: null,
    tokensIn: 0,
    tokensOut: 0,
    cost: 0,
    createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
  },
]

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
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>All Jobs</CardTitle>
            <CardDescription>
              A list of all your AI processing jobs
            </CardDescription>
          </div>
          <Button variant="outline" size="sm">
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
        </div>
      </CardHeader>
      <CardContent>
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
                <th className="pb-3 text-right font-medium text-muted-foreground">
                  Cost
                </th>
                <th className="pb-3 text-left font-medium text-muted-foreground">
                  Created
                </th>
                <th className="pb-3 text-right font-medium text-muted-foreground">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              {jobs.map((job) => (
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
                    <Badge variant={typeVariant[job.type as keyof typeof typeVariant]}>
                      {job.type}
                    </Badge>
                  </td>
                  <td className="py-3">
                    <Badge variant={statusVariant[job.status as keyof typeof statusVariant]}>
                      {job.status}
                    </Badge>
                  </td>
                  <td className="py-3 text-muted-foreground">
                    {job.provider || '-'}
                  </td>
                  <td className="py-3 text-right tabular-nums">
                    {job.tokensIn + job.tokensOut}
                  </td>
                  <td className="py-3 text-right tabular-nums">
                    {formatCurrency(job.cost)}
                  </td>
                  <td className="py-3 text-muted-foreground">
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
      </CardContent>
    </Card>
  )
}
