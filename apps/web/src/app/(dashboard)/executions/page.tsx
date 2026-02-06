'use client'

import { useState, useMemo } from 'react'
import Link from 'next/link'
import {
  Search,
  CheckCircle,
  XCircle,
  Clock,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  Blocks,
  TrendingUp,
  AlertTriangle,
  Activity,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useExecutionsPaginated } from '@/lib/hooks/use-helpers'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'

const PAGE_SIZE = 20

export default function ExecutionsPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [nextToken, setNextToken] = useState<string | undefined>(undefined)
  const [tokenHistory, setTokenHistory] = useState<(string | undefined)[]>([undefined])

  const { data, isLoading, error } = useExecutionsPaginated({
    status: statusFilter !== 'all' ? statusFilter : undefined,
    nextToken,
    limit: PAGE_SIZE,
  })

  const executions = data?.executions ?? []
  const hasMore = data?.hasMore ?? false

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-success" />
      case 'failed':
        return <XCircle className="h-4 w-4 text-destructive" />
      case 'running':
        return <RefreshCw className="h-4 w-4 animate-spin text-info animate-status-pulse" />
      case 'pending':
        return <Clock className="h-4 w-4 text-warning" />
      default:
        return <Clock className="h-4 w-4 text-muted-foreground" />
    }
  }

  const filteredExecutions = useMemo(() => {
    if (!executions) return []
    if (!searchQuery) return executions
    return executions.filter(
      (exec) =>
        exec.executionId.toLowerCase().includes(searchQuery.toLowerCase()) ||
        exec.helperId.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (exec.contactId && exec.contactId.toLowerCase().includes(searchQuery.toLowerCase()))
    )
  }, [executions, searchQuery])

  const stats = useMemo(() => {
    if (!executions || executions.length === 0) {
      return { total: 0, successRate: '0', avgDuration: 0, failed: 0 }
    }
    const completed = executions.filter((e) => e.status === 'completed').length
    const failed = executions.filter((e) => e.status === 'failed').length
    const withDuration = executions.filter((e) => e.durationMs && e.durationMs > 0)
    const avgDuration = withDuration.length > 0
      ? Math.round(withDuration.reduce((sum, e) => sum + (e.durationMs || 0), 0) / withDuration.length)
      : 0
    return {
      total: executions.length,
      successRate: executions.length > 0 ? ((completed / executions.length) * 100).toFixed(1) : '0',
      avgDuration,
      failed,
    }
  }, [executions])

  return (
    <div className="animate-fade-in-up space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Executions</h1>
        <p className="text-muted-foreground">View and monitor helper execution history</p>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-4">
        {isLoading ? (
          [1, 2, 3, 4].map((i) => (
            <div key={i} className="rounded-lg border bg-card p-4">
              <Skeleton className="h-4 w-20" />
              <Skeleton className="mt-2 h-7 w-16" />
              <Skeleton className="mt-1 h-3 w-24" />
            </div>
          ))
        ) : (
          <>
            <div className="card-hover rounded-lg border bg-card p-4">
              <div className="flex items-center justify-between">
                <p className="text-sm text-muted-foreground">Total</p>
                <TrendingUp className="h-4 w-4 text-muted-foreground" />
              </div>
              <p className="mt-1 text-2xl font-bold">{stats.total.toLocaleString()}</p>
            </div>
            <div className="card-hover rounded-lg border bg-card p-4">
              <div className="flex items-center justify-between">
                <p className="text-sm text-muted-foreground">Success Rate</p>
                <CheckCircle className="h-4 w-4 text-success" />
              </div>
              <p className="mt-1 text-2xl font-bold text-success">{stats.successRate}%</p>
            </div>
            <div className="card-hover rounded-lg border bg-card p-4">
              <div className="flex items-center justify-between">
                <p className="text-sm text-muted-foreground">Avg Duration</p>
                <Clock className="h-4 w-4 text-muted-foreground" />
              </div>
              <p className="mt-1 text-2xl font-bold">{stats.avgDuration}ms</p>
            </div>
            <div className="card-hover rounded-lg border bg-card p-4">
              <div className="flex items-center justify-between">
                <p className="text-sm text-muted-foreground">Failed</p>
                <AlertTriangle className="h-4 w-4 text-destructive" />
              </div>
              <p className="mt-1 text-2xl font-bold text-destructive">{stats.failed}</p>
            </div>
          </>
        )}
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search by helper, contact, or execution ID..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setNextToken(undefined); setTokenHistory([undefined]) }}
          className="h-10 rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <option value="all">All Status</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
          <option value="running">Running</option>
          <option value="pending">Pending</option>
        </select>
      </div>

      {/* Table */}
      {isLoading ? (
        <div className="rounded-lg border">
          <div className="border-b bg-muted/50 p-4">
            <Skeleton className="h-4 w-full max-w-md" />
          </div>
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="flex items-center gap-4 border-b p-4 last:border-0">
              <Skeleton className="h-4 w-4 rounded-full" />
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-4 flex-1 max-w-[120px]" />
              <Skeleton className="h-4 flex-1 max-w-[160px]" />
              <Skeleton className="h-4 w-12" />
              <Skeleton className="h-4 w-16" />
            </div>
          ))}
        </div>
      ) : filteredExecutions.length > 0 ? (
        <div className="skeleton-fade-enter rounded-lg border overflow-x-auto">
          <table className="w-full min-w-[640px]">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="p-4 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">Status</th>
                <th className="p-4 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">Execution ID</th>
                <th className="p-4 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">Helper</th>
                <th className="p-4 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">Contact</th>
                <th className="p-4 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">Duration</th>
                <th className="p-4 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">Time</th>
              </tr>
            </thead>
            <tbody>
              {filteredExecutions.map((execution) => (
                <tr
                  key={execution.executionId}
                  className="border-b last:border-0 hover:bg-muted/50"
                >
                  <td className="p-4">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(execution.status)}
                      <span
                        className={cn(
                          'text-xs font-medium capitalize',
                          execution.status === 'completed' && 'text-success',
                          execution.status === 'failed' && 'text-destructive',
                          execution.status === 'running' && 'text-info',
                          execution.status === 'pending' && 'text-warning'
                        )}
                      >
                        {execution.status}
                      </span>
                    </div>
                  </td>
                  <td className="p-4">
                    <Link
                      href={`/executions/${execution.executionId}`}
                      className="font-mono text-xs text-primary hover:underline"
                    >
                      {execution.executionId}
                    </Link>
                  </td>
                  <td className="p-4">
                    <div className="flex items-center gap-2">
                      <Blocks className="h-3.5 w-3.5 text-muted-foreground" />
                      <span className="text-sm font-medium">{execution.helperId}</span>
                    </div>
                  </td>
                  <td className="p-4 font-mono text-xs text-muted-foreground">
                    {execution.contactId || '-'}
                  </td>
                  <td className="p-4">
                    <span className="font-mono text-xs text-muted-foreground">
                      {execution.durationMs ? `${execution.durationMs}ms` : '-'}
                    </span>
                  </td>
                  <td className="p-4 text-xs text-muted-foreground">
                    {new Date(execution.startedAt).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
          <Activity className="mb-4 h-12 w-12 text-muted-foreground/50" />
          <h3 className="mb-1 font-semibold">No executions found</h3>
          <p className="max-w-sm text-sm text-muted-foreground">
            {error
              ? 'Unable to load executions. Please check your connection and try again.'
              : 'When your CRM automations trigger a helper, every execution will appear here with status, duration, and contact details.'}
          </p>
        </div>
      )}

      {/* Pagination */}
      {(tokenHistory.length > 1 || hasMore) && (
        <div className="flex items-center justify-end gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              const prev = tokenHistory.slice(0, -1)
              setTokenHistory(prev)
              setNextToken(prev[prev.length - 1])
            }}
            disabled={tokenHistory.length <= 1}
          >
            <ChevronLeft className="h-4 w-4" />
            Previous
          </Button>
          <span className="text-sm text-muted-foreground">Page {tokenHistory.length}</span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              if (data?.nextToken) {
                setTokenHistory([...tokenHistory, data.nextToken])
                setNextToken(data.nextToken)
              }
            }}
            disabled={!hasMore}
          >
            Next
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  )
}
