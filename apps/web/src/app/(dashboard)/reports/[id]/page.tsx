'use client'

import { useMemo } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import {
  ArrowLeft,
  BarChart3,
  Clock,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  Activity,
  CheckCircle,
  XCircle,
  Blocks,
  Users,
  Zap,
} from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { useReportStats } from '@/lib/hooks/use-reports'
import { useHelpers } from '@/lib/hooks/use-helpers'

const reportMeta: Record<string, { title: string; description: string }> = {
  'execution-overview': {
    title: 'Execution Overview',
    description: 'Live dashboard of all execution activity and key metrics.',
  },
  'helper-performance': {
    title: 'Helper Performance',
    description: 'Execution success rates, duration, and error analysis per helper.',
  },
  'execution-trends': {
    title: 'Execution Trends',
    description: 'Daily execution volume and processing trends over the last 30 days.',
  },
  'error-analysis': {
    title: 'Error Analysis',
    description: 'Breakdown of failed executions by error type and affected helpers.',
  },
  'contact-activity': {
    title: 'Contact Activity',
    description: 'Unique contacts processed and execution patterns.',
  },
  'helper-catalog': {
    title: 'Helper Catalog',
    description: 'Summary of all configured helpers and their usage.',
  },
}

function LoadingSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Skeleton className="h-10 w-10" />
        <div className="flex-1">
          <Skeleton className="h-7 w-48" />
          <Skeleton className="mt-1 h-4 w-72" />
        </div>
      </div>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="rounded-lg border bg-card p-4">
            <Skeleton className="h-4 w-20" />
            <Skeleton className="mt-2 h-7 w-16" />
          </div>
        ))}
      </div>
      <Skeleton className="h-64 w-full rounded-lg" />
    </div>
  )
}

export default function ReportDetailPage() {
  const params = useParams()
  const reportId = params.id as string

  const { data: stats, isLoading, error } = useReportStats()
  const { data: helpers } = useHelpers()

  const meta = reportMeta[reportId] || {
    title: 'Report',
    description: 'Execution analytics report.',
  }

  const helperNameMap = useMemo(() => {
    const map = new Map<string, string>()
    if (helpers) {
      for (const h of helpers) {
        map.set(h.helperId, h.name)
      }
    }
    return map
  }, [helpers])

  if (isLoading) return <LoadingSkeleton />

  if (error || !stats) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <Link href="/reports" className="rounded-md p-2 hover:bg-accent">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold">{meta.title}</h1>
            <p className="text-muted-foreground">{meta.description}</p>
          </div>
        </div>
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
          <Activity className="mb-4 h-12 w-12 text-muted-foreground/50" />
          <h3 className="mb-1 font-semibold">Unable to load report data</h3>
          <p className="text-sm text-muted-foreground">
            The executions endpoint may not be available yet. Execute some helpers first.
          </p>
        </div>
      </div>
    )
  }

  const { summary, dailyTrend, topHelpers, errorBreakdown } = stats

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href="/reports" className="rounded-md p-2 hover:bg-accent">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">{meta.title}</h1>
          <p className="text-muted-foreground">{meta.description}</p>
        </div>
      </div>

      {/* KPI Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
        <div className="rounded-lg border bg-card p-4">
          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground">Total Executions</p>
            <BarChart3 className="h-3.5 w-3.5 text-muted-foreground" />
          </div>
          <p className="mt-1 text-xl font-bold">{summary.total.toLocaleString()}</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground">Success Rate</p>
            <TrendingUp className="h-3.5 w-3.5 text-success" />
          </div>
          <p className={cn('mt-1 text-xl font-bold', summary.successRate >= 95 ? 'text-success' : summary.successRate >= 80 ? 'text-warning' : 'text-destructive')}>
            {summary.successRate}%
          </p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground">Contacts Processed</p>
            <Users className="h-3.5 w-3.5 text-muted-foreground" />
          </div>
          <p className="mt-1 text-xl font-bold">{summary.uniqueContacts.toLocaleString()}</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground">Avg Duration</p>
            <Clock className="h-3.5 w-3.5 text-muted-foreground" />
          </div>
          <p className="mt-1 text-xl font-bold">{summary.avgDurationMs}ms</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground">Errors</p>
            <TrendingDown className="h-3.5 w-3.5 text-destructive" />
          </div>
          <p className={cn('mt-1 text-xl font-bold', summary.failed > 0 ? 'text-destructive' : '')}>
            {summary.failed.toLocaleString()}
          </p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground">Active Helpers</p>
            <Blocks className="h-3.5 w-3.5 text-muted-foreground" />
          </div>
          <p className="mt-1 text-xl font-bold">{summary.uniqueHelpers}</p>
        </div>
      </div>

      {/* Report-specific content */}
      {(reportId === 'execution-overview' || reportId === 'execution-trends') && (
        <div className="grid gap-6 lg:grid-cols-2">
          {/* Execution Trend Chart */}
          <div className="rounded-lg border bg-card p-5 lg:col-span-2">
            <h3 className="mb-4 font-semibold">Daily Execution Volume (Last 30 Days)</h3>
            {dailyTrend.length > 0 ? (
              <>
                <div className="flex items-end gap-0.5" style={{ height: 220 }}>
                  {(() => {
                    const maxVal = Math.max(...dailyTrend.map((d) => d.total), 1)
                    return dailyTrend.map((day) => {
                      const failedHeight = day.total > 0 ? (day.failed / maxVal) * 200 : 0
                      const completedHeight = day.total > 0 ? ((day.total - day.failed) / maxVal) * 200 : 0
                      return (
                        <div key={day.date} className="flex flex-1 flex-col items-center group relative">
                          <div className="w-full flex flex-col-reverse">
                            <div
                              className="w-full bg-primary/80 hover:bg-primary transition-colors rounded-t"
                              style={{ height: `${Math.max(completedHeight, 1)}px` }}
                            />
                            {failedHeight > 0 && (
                              <div
                                className="w-full bg-destructive/70"
                                style={{ height: `${failedHeight}px` }}
                              />
                            )}
                          </div>
                          <div className="absolute bottom-full mb-1 hidden group-hover:block rounded bg-popover px-2 py-1 text-xs shadow-md border whitespace-nowrap z-10">
                            <p className="font-medium">
                              {new Date(day.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                            </p>
                            <p>{day.total} total</p>
                            <p className="text-success">{day.completed} completed</p>
                            {day.failed > 0 && <p className="text-destructive">{day.failed} failed</p>}
                          </div>
                        </div>
                      )
                    })
                  })()}
                </div>
                <div className="flex justify-between mt-2 text-xs text-muted-foreground">
                  <span>
                    {new Date(dailyTrend[0].date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                  </span>
                  <span>
                    {new Date(dailyTrend[dailyTrend.length - 1].date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                  </span>
                </div>
                <div className="flex items-center gap-4 mt-3 text-xs text-muted-foreground">
                  <span className="flex items-center gap-1">
                    <span className="h-2.5 w-2.5 rounded-sm bg-primary/80" /> Completed
                  </span>
                  <span className="flex items-center gap-1">
                    <span className="h-2.5 w-2.5 rounded-sm bg-destructive/70" /> Failed
                  </span>
                </div>
              </>
            ) : (
              <div className="flex items-center justify-center h-[200px] text-sm text-muted-foreground">
                No execution data for the last 30 days
              </div>
            )}
          </div>
        </div>
      )}

      {(reportId === 'execution-overview' || reportId === 'helper-performance' || reportId === 'helper-catalog') && (
        <div className="rounded-lg border bg-card p-5">
          <h3 className="mb-4 font-semibold">Helper Performance</h3>
          {topHelpers.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b">
                    <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">#</th>
                    <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">Helper</th>
                    <th className="pb-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">Executions</th>
                    <th className="pb-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">Completed</th>
                    <th className="pb-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">Failed</th>
                    <th className="pb-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">Success Rate</th>
                    <th className="pb-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">Avg Duration</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {topHelpers.map((helper, i) => {
                    const name = helperNameMap.get(helper.helperId) || helper.helperId
                    const successRate = helper.total > 0
                      ? Math.round((helper.completed / helper.total) * 1000) / 10
                      : 0
                    return (
                      <tr key={helper.helperId} className="hover:bg-muted/50">
                        <td className="py-3 text-sm text-muted-foreground">{i + 1}</td>
                        <td className="py-3">
                          <div className="flex items-center gap-2">
                            <Zap className="h-3.5 w-3.5 text-muted-foreground" />
                            <span className="text-sm font-medium">{name}</span>
                          </div>
                        </td>
                        <td className="py-3 text-right font-mono text-sm">{helper.total.toLocaleString()}</td>
                        <td className="py-3 text-right font-mono text-sm text-success">{helper.completed.toLocaleString()}</td>
                        <td className="py-3 text-right font-mono text-sm text-destructive">{helper.failed > 0 ? helper.failed.toLocaleString() : '-'}</td>
                        <td className="py-3 text-right">
                          <span className={cn(
                            'font-mono text-sm',
                            successRate >= 95 ? 'text-success' : successRate >= 80 ? 'text-warning' : 'text-destructive'
                          )}>
                            {successRate}%
                          </span>
                        </td>
                        <td className="py-3 text-right font-mono text-sm text-muted-foreground">
                          {helper.avgDurationMs > 0 ? `${helper.avgDurationMs}ms` : '-'}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
              No helper execution data available
            </div>
          )}
        </div>
      )}

      {(reportId === 'execution-overview' || reportId === 'error-analysis') && errorBreakdown.length > 0 && (
        <div className="rounded-lg border bg-card p-5">
          <h3 className="mb-4 font-semibold">Error Breakdown</h3>
          <div className="space-y-3">
            {errorBreakdown.map((entry) => {
              const percentage = summary.failed > 0
                ? Math.round((entry.count / summary.failed) * 100)
                : 0
              return (
                <div key={entry.error} className="space-y-1.5">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 min-w-0 flex-1">
                      <XCircle className="h-3.5 w-3.5 text-destructive flex-shrink-0" />
                      <p className="text-sm truncate">{entry.error}</p>
                    </div>
                    <span className="text-xs font-medium text-muted-foreground ml-2 flex-shrink-0">
                      {entry.count}x ({percentage}%)
                    </span>
                  </div>
                  <div className="h-1.5 w-full rounded-full bg-muted">
                    <div
                      className="h-full rounded-full bg-destructive/60"
                      style={{ width: `${percentage}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {reportId === 'contact-activity' && (
        <div className="grid gap-6 lg:grid-cols-2">
          <div className="rounded-lg border bg-card p-5">
            <h3 className="mb-4 font-semibold">Contact Processing Summary</h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Unique Contacts Processed</span>
                <span className="text-lg font-bold">{summary.uniqueContacts.toLocaleString()}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Total Executions</span>
                <span className="text-lg font-bold">{summary.total.toLocaleString()}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Avg Executions per Contact</span>
                <span className="text-lg font-bold">
                  {summary.uniqueContacts > 0
                    ? (summary.total / summary.uniqueContacts).toFixed(1)
                    : '-'}
                </span>
              </div>
            </div>
          </div>
          <div className="rounded-lg border bg-card p-5">
            <h3 className="mb-4 font-semibold">Processing Performance</h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Average Duration</span>
                <span className="text-lg font-bold">{summary.avgDurationMs}ms</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Success Rate</span>
                <span className={cn('text-lg font-bold', summary.successRate >= 95 ? 'text-success' : 'text-warning')}>
                  {summary.successRate}%
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Active Helpers</span>
                <span className="text-lg font-bold">{summary.uniqueHelpers}</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
