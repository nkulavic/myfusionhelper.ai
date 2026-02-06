'use client'

import { useState, useMemo } from 'react'
import Link from 'next/link'
import {
  BarChart3,
  Calendar,
  TrendingUp,
  Users,
  Tag,
  Zap,
  FileText,
  Clock,
  ArrowRight,
  Search,
  Activity,
  AlertTriangle,
  CheckCircle,
  Blocks,
  RefreshCw,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { useReportStats } from '@/lib/hooks/use-reports'
import { useHelpers } from '@/lib/hooks/use-helpers'
import { cn } from '@/lib/utils'

const reportTemplates = [
  {
    id: 'execution-overview',
    name: 'Execution Overview',
    description: 'Live dashboard of total executions, success rates, and processing volume.',
    icon: BarChart3,
    category: 'overview',
    href: '/reports/execution-overview',
  },
  {
    id: 'helper-performance',
    name: 'Helper Performance',
    description: 'Execution success rates, average duration, and error analysis for each helper.',
    icon: Zap,
    category: 'performance',
    href: '/reports/helper-performance',
  },
  {
    id: 'execution-trends',
    name: 'Execution Trends',
    description: 'Daily execution volume, peak usage, and processing speed trends over 30 days.',
    icon: TrendingUp,
    category: 'performance',
    href: '/reports/execution-trends',
  },
  {
    id: 'error-analysis',
    name: 'Error Analysis',
    description: 'Breakdown of failed executions by error type and affected helpers.',
    icon: AlertTriangle,
    category: 'performance',
    href: '/reports/error-analysis',
  },
  {
    id: 'contact-activity',
    name: 'Contact Activity',
    description: 'Track unique contacts processed, execution frequency, and engagement patterns.',
    icon: Users,
    category: 'contacts',
    href: '/reports/contact-activity',
  },
  {
    id: 'helper-catalog',
    name: 'Helper Catalog',
    description: 'Summary of all configured helpers, their status, and usage statistics.',
    icon: Blocks,
    category: 'overview',
    href: '/reports/helper-catalog',
  },
]

const categories = ['all', 'overview', 'performance', 'contacts']

function StatsSkeleton() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {[1, 2, 3, 4].map((i) => (
        <div key={i} className="rounded-lg border bg-card p-4">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="mt-2 h-7 w-16" />
          <Skeleton className="mt-1 h-3 w-20" />
        </div>
      ))}
    </div>
  )
}

export default function ReportsPage() {
  const [activeTab, setActiveTab] = useState<'overview' | 'templates'>('overview')
  const [categoryFilter, setCategoryFilter] = useState('all')
  const [search, setSearch] = useState('')

  const { data: reportStats, isLoading: statsLoading, error: statsError } = useReportStats()
  const { data: helpers } = useHelpers()

  const filteredTemplates = reportTemplates.filter((t) => {
    if (categoryFilter !== 'all' && t.category !== categoryFilter) return false
    if (search && !t.name.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  // Compute helper name map
  const helperNameMap = useMemo(() => {
    const map = new Map<string, string>()
    if (helpers) {
      for (const h of helpers) {
        map.set(h.helperId, h.name)
      }
    }
    return map
  }, [helpers])

  const summary = reportStats?.summary

  return (
    <div className="animate-fade-in-up space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold">Reports</h1>
        <p className="text-muted-foreground">Execution analytics and performance insights</p>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 rounded-lg border bg-muted p-1">
        <Button
          variant="ghost"
          onClick={() => setActiveTab('overview')}
          className={`flex-1 ${
            activeTab === 'overview'
              ? 'bg-background shadow-sm hover:bg-background'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Overview
        </Button>
        <Button
          variant="ghost"
          onClick={() => setActiveTab('templates')}
          className={`flex-1 ${
            activeTab === 'templates'
              ? 'bg-background shadow-sm hover:bg-background'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Report Types
        </Button>
      </div>

      {activeTab === 'overview' ? (
        <div className="space-y-6">
          {/* KPI Stats */}
          {statsLoading ? (
            <StatsSkeleton />
          ) : statsError ? (
            <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
              <AlertTriangle className="mb-3 h-8 w-8 text-muted-foreground/50" />
              <p className="text-sm text-muted-foreground">
                Unable to load execution stats. The executions endpoint may not be available yet.
              </p>
            </div>
          ) : summary ? (
            <>
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <div className="rounded-lg border bg-card p-4">
                  <div className="flex items-center justify-between">
                    <p className="text-sm text-muted-foreground">Total Executions</p>
                    <Activity className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <p className="mt-1 text-2xl font-bold">{summary.total.toLocaleString()}</p>
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {summary.running > 0 ? `${summary.running} running` : 'All complete'}
                  </p>
                </div>
                <div className="rounded-lg border bg-card p-4">
                  <div className="flex items-center justify-between">
                    <p className="text-sm text-muted-foreground">Success Rate</p>
                    <CheckCircle className="h-4 w-4 text-success" />
                  </div>
                  <p className={cn('mt-1 text-2xl font-bold', summary.successRate >= 95 ? 'text-success' : summary.successRate >= 80 ? 'text-warning' : 'text-destructive')}>
                    {summary.successRate}%
                  </p>
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {summary.completed.toLocaleString()} completed
                  </p>
                </div>
                <div className="rounded-lg border bg-card p-4">
                  <div className="flex items-center justify-between">
                    <p className="text-sm text-muted-foreground">Avg Duration</p>
                    <Clock className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <p className="mt-1 text-2xl font-bold">{summary.avgDurationMs}ms</p>
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {summary.uniqueContacts.toLocaleString()} contacts processed
                  </p>
                </div>
                <div className="rounded-lg border bg-card p-4">
                  <div className="flex items-center justify-between">
                    <p className="text-sm text-muted-foreground">Failed</p>
                    <AlertTriangle className="h-4 w-4 text-destructive" />
                  </div>
                  <p className={cn('mt-1 text-2xl font-bold', summary.failed > 0 ? 'text-destructive' : '')}>
                    {summary.failed.toLocaleString()}
                  </p>
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {summary.uniqueHelpers} active helpers
                  </p>
                </div>
              </div>

              <div className="grid gap-6 lg:grid-cols-2">
                {/* Daily Trend Chart */}
                <div className="rounded-lg border bg-card p-5">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="font-semibold">Execution Trend (30 days)</h3>
                    <Link
                      href="/reports/execution-trends"
                      className="text-xs text-primary hover:underline"
                    >
                      View Details
                    </Link>
                  </div>
                  {reportStats.dailyTrend.length > 0 ? (
                    <div className="flex items-end gap-0.5" style={{ height: 180 }}>
                      {(() => {
                        const maxVal = Math.max(...reportStats.dailyTrend.map((d) => d.total), 1)
                        return reportStats.dailyTrend.map((day) => (
                          <div key={day.date} className="flex flex-1 flex-col items-center group relative">
                            <div
                              className="w-full rounded-t bg-primary/80 hover:bg-primary transition-colors min-h-[2px]"
                              style={{ height: `${Math.max((day.total / maxVal) * 160, 2)}px` }}
                            />
                            <div className="absolute bottom-full mb-1 hidden group-hover:block rounded bg-popover px-2 py-1 text-xs shadow-md border whitespace-nowrap z-10">
                              <p className="font-medium">{new Date(day.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</p>
                              <p>{day.total} total</p>
                              {day.failed > 0 && <p className="text-destructive">{day.failed} failed</p>}
                            </div>
                          </div>
                        ))
                      })()}
                    </div>
                  ) : (
                    <div className="flex items-center justify-center h-[180px] text-sm text-muted-foreground">
                      No execution data yet
                    </div>
                  )}
                  <div className="flex justify-between mt-2 text-xs text-muted-foreground">
                    <span>
                      {reportStats.dailyTrend.length > 0
                        ? new Date(reportStats.dailyTrend[0].date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
                        : ''}
                    </span>
                    <span>
                      {reportStats.dailyTrend.length > 0
                        ? new Date(reportStats.dailyTrend[reportStats.dailyTrend.length - 1].date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
                        : ''}
                    </span>
                  </div>
                </div>

                {/* Top Helpers */}
                <div className="rounded-lg border bg-card p-5">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="font-semibold">Top Helpers</h3>
                    <Link
                      href="/reports/helper-performance"
                      className="text-xs text-primary hover:underline"
                    >
                      View All
                    </Link>
                  </div>
                  {reportStats.topHelpers.length > 0 ? (
                    <div className="space-y-3">
                      {reportStats.topHelpers.slice(0, 5).map((helper, i) => {
                        const name = helperNameMap.get(helper.helperId) || helper.helperId
                        const successRate = helper.total > 0
                          ? Math.round((helper.completed / helper.total) * 1000) / 10
                          : 0
                        return (
                          <div key={helper.helperId} className="flex items-center gap-3">
                            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-muted text-xs font-medium">
                              {i + 1}
                            </span>
                            <div className="flex-1 min-w-0">
                              <div className="flex items-center justify-between">
                                <p className="text-sm font-medium truncate">{name}</p>
                                <p className="text-xs text-muted-foreground ml-2 flex-shrink-0">
                                  {helper.total.toLocaleString()} runs
                                </p>
                              </div>
                              <div className="mt-1 h-1.5 w-full rounded-full bg-muted">
                                <div
                                  className="h-full rounded-full bg-primary"
                                  style={{
                                    width: `${(helper.total / reportStats.topHelpers[0].total) * 100}%`,
                                  }}
                                />
                              </div>
                            </div>
                            <span className={cn(
                              'text-xs flex-shrink-0',
                              successRate >= 95 ? 'text-success' : successRate >= 80 ? 'text-warning' : 'text-destructive'
                            )}>
                              {successRate}%
                            </span>
                          </div>
                        )
                      })}
                    </div>
                  ) : (
                    <div className="flex items-center justify-center h-[180px] text-sm text-muted-foreground">
                      No helper data yet
                    </div>
                  )}
                </div>
              </div>

              {/* Error Breakdown */}
              {reportStats.errorBreakdown.length > 0 && (
                <div className="rounded-lg border bg-card p-5">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="font-semibold">Recent Errors</h3>
                    <Link
                      href="/reports/error-analysis"
                      className="text-xs text-primary hover:underline"
                    >
                      View Details
                    </Link>
                  </div>
                  <div className="divide-y">
                    {reportStats.errorBreakdown.slice(0, 5).map((entry) => (
                      <div key={entry.error} className="flex items-center gap-3 py-2.5">
                        <AlertTriangle className="h-4 w-4 text-destructive flex-shrink-0" />
                        <p className="text-sm truncate flex-1">{entry.error}</p>
                        <span className="text-xs font-medium text-destructive flex-shrink-0">
                          {entry.count}x
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </>
          ) : (
            <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
              <Activity className="mb-4 h-12 w-12 text-muted-foreground/50" />
              <h3 className="mb-1 font-semibold">No execution data</h3>
              <p className="text-sm text-muted-foreground">
                Execute a helper to see analytics here.
              </p>
            </div>
          )}
        </div>
      ) : (
        <div className="space-y-4">
          {/* Search and Filter */}
          <div className="flex gap-3">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                type="text"
                placeholder="Search report types..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>
            <div className="flex gap-1 rounded-md border bg-background p-1">
              {categories.map((cat) => (
                <button
                  key={cat}
                  onClick={() => setCategoryFilter(cat)}
                  className={`rounded px-3 py-1 text-xs font-medium capitalize transition-colors ${
                    categoryFilter === cat
                      ? 'bg-primary text-primary-foreground'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  {cat}
                </button>
              ))}
            </div>
          </div>

          {/* Template Grid */}
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {filteredTemplates.map((template) => (
              <Link
                key={template.id}
                href={template.href}
                className="group rounded-lg border bg-card p-5 transition-colors hover:border-primary"
              >
                <div className="mb-3 flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                    <template.icon className="h-5 w-5 text-primary" />
                  </div>
                  <h3 className="font-semibold">{template.name}</h3>
                </div>
                <p className="text-sm text-muted-foreground">{template.description}</p>
                <div className="mt-4 flex items-center gap-1 text-sm text-primary">
                  View Report
                  <ArrowRight className="h-3 w-3" />
                </div>
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
