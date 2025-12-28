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
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { formatDate, formatCurrency, truncate } from '@/lib/utils'
import { Eye, MoreHorizontal, RefreshCw, Loader2, Search, X } from 'lucide-react'
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
import { useState, useMemo } from 'react'

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

type StatusFilter = 'ALL' | 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED'
type TypeFilter = 'ALL' | 'TEXT' | 'IMAGE'

export function JobsTable() {
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('ALL')
  const [typeFilter, setTypeFilter] = useState<TypeFilter>('ALL')
  const [searchQuery, setSearchQuery] = useState('')

  // Build filter for API
  const apiFilter = useMemo(() => {
    const filter: { status?: 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED'; type?: 'TEXT' | 'IMAGE' } = {}
    if (statusFilter !== 'ALL') filter.status = statusFilter
    if (typeFilter !== 'ALL') filter.type = typeFilter
    return Object.keys(filter).length > 0 ? filter : undefined
  }, [statusFilter, typeFilter])

  const { data: jobs, isLoading, error, isFetching } = useJobs(apiFilter)
  const queryClient = useQueryClient()

  // Client-side search filtering
  const filteredJobs = useMemo(() => {
    if (!jobs?.edges) return []
    if (!searchQuery.trim()) return jobs.edges

    const query = searchQuery.toLowerCase()
    return jobs.edges.filter(({ node: job }) =>
      job.input.toLowerCase().includes(query) ||
      job.id.toLowerCase().includes(query) ||
      job.provider?.toLowerCase().includes(query)
    )
  }, [jobs?.edges, searchQuery])

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

  const handleClearFilters = () => {
    setStatusFilter('ALL')
    setTypeFilter('ALL')
    setSearchQuery('')
  }

  const hasActiveFilters = statusFilter !== 'ALL' || typeFilter !== 'ALL' || searchQuery.trim() !== ''

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

        {/* Filters and Search */}
        <div className="flex flex-col sm:flex-row gap-3 pt-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search by input, ID, or provider..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>
          <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v as StatusFilter)}>
            <SelectTrigger className="w-full sm:w-[140px]">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">All Status</SelectItem>
              <SelectItem value="PENDING">Pending</SelectItem>
              <SelectItem value="PROCESSING">Processing</SelectItem>
              <SelectItem value="COMPLETED">Completed</SelectItem>
              <SelectItem value="FAILED">Failed</SelectItem>
            </SelectContent>
          </Select>
          <Select value={typeFilter} onValueChange={(v) => setTypeFilter(v as TypeFilter)}>
            <SelectTrigger className="w-full sm:w-[120px]">
              <SelectValue placeholder="Type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">All Types</SelectItem>
              <SelectItem value="TEXT">Text</SelectItem>
              <SelectItem value="IMAGE">Image</SelectItem>
            </SelectContent>
          </Select>
          {hasActiveFilters && (
            <Button variant="ghost" size="sm" onClick={handleClearFilters} className="h-10">
              <X className="mr-1 h-4 w-4" />
              Clear
            </Button>
          )}
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
        ) : filteredJobs.length === 0 ? (
          <div className="py-12 text-center text-sm text-muted-foreground">
            {jobs?.edges.length === 0
              ? 'No jobs yet. Create your first job!'
              : 'No jobs match your filters.'}
          </div>
        ) : (
          <div className="relative overflow-x-auto">
            {searchQuery && (
              <p className="mb-3 text-sm text-muted-foreground">
                Showing {filteredJobs.length} of {jobs?.edges.length} jobs
              </p>
            )}
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
                {filteredJobs.map(({ node: job }) => (
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
