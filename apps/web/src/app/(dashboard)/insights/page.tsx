'use client'

import { useMemo } from 'react'
import Link from 'next/link'
import {
  Sparkles,
  TrendingUp,
  AlertTriangle,
  Lightbulb,
  BarChart3,
  CheckCircle,
  Clock,
  Users,
  Zap,
  Target,
  Award,
  ArrowRight,
  Link2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useHelpers, useExecutions } from '@/lib/hooks/use-helpers'
import { useConnections } from '@/lib/hooks/use-connections'
import { useInsights, type AIInsight } from '@/lib/hooks/use-insights'
import { Skeleton } from '@/components/ui/skeleton'
import { getCRMPlatform } from '@/lib/crm-platforms'
import { PlatformLogo } from '@/components/platform-logo'

function getInsightIcon(insight: AIInsight) {
  switch (insight.type) {
    case 'anomaly':
      return AlertTriangle
    case 'pattern':
      return TrendingUp
    case 'achievement':
      return Award
    case 'suggestion':
    default:
      return Lightbulb
  }
}

function getInsightBorder(insight: AIInsight) {
  switch (insight.severity) {
    case 'critical':
      return 'border-l-destructive'
    case 'warning':
      return 'border-l-warning'
    case 'success':
      return 'border-l-success'
    case 'info':
    default:
      return 'border-l-primary'
  }
}

function getInsightIconColor(insight: AIInsight) {
  switch (insight.severity) {
    case 'critical':
      return 'text-destructive bg-destructive/10'
    case 'warning':
      return 'text-warning bg-warning/10'
    case 'success':
      return 'text-success bg-success/10'
    case 'info':
    default:
      return 'text-primary bg-primary/10'
  }
}

function StatsSkeleton() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {[1, 2, 3, 4].map((i) => (
        <div key={i} className="rounded-lg border bg-card p-5">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="mt-2 h-8 w-20" />
          <Skeleton className="mt-1 h-3 w-28" />
        </div>
      ))}
    </div>
  )
}

function InsightsSkeleton() {
  return (
    <div className="grid gap-4 lg:grid-cols-2">
      {[1, 2, 3, 4].map((i) => (
        <div key={i} className="rounded-lg border bg-card p-4">
          <div className="flex gap-3">
            <Skeleton className="h-9 w-9 rounded-lg" />
            <div className="flex-1 space-y-2">
              <Skeleton className="h-5 w-48" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-3/4" />
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}

function ChartsSkeleton() {
  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <div className="rounded-lg border bg-card p-5">
        <Skeleton className="h-5 w-40 mb-4" />
        <Skeleton className="h-48 w-full rounded" />
      </div>
      <div className="rounded-lg border bg-card p-5">
        <Skeleton className="h-5 w-40 mb-4" />
        <Skeleton className="h-48 w-full rounded" />
      </div>
    </div>
  )
}

// Simple donut chart using SVG
function DonutChart({
  data,
}: {
  data: { name: string; percentage: number; color: string }[]
}) {
  const total = data.reduce((s, d) => s + d.percentage, 0)
  let accumulated = 0

  return (
    <svg viewBox="0 0 200 200" className="h-48 w-48">
      {data.map((segment) => {
        const startAngle = (accumulated / total) * 360
        const segmentAngle = (segment.percentage / total) * 360
        accumulated += segment.percentage

        const startRad = ((startAngle - 90) * Math.PI) / 180
        const endRad = ((startAngle + segmentAngle - 90) * Math.PI) / 180

        const x1 = 100 + 80 * Math.cos(startRad)
        const y1 = 100 + 80 * Math.sin(startRad)
        const x2 = 100 + 80 * Math.cos(endRad)
        const y2 = 100 + 80 * Math.sin(endRad)

        const largeArc = segmentAngle > 180 ? 1 : 0

        return (
          <path
            key={segment.name}
            d={`M 100 100 L ${x1} ${y1} A 80 80 0 ${largeArc} 1 ${x2} ${y2} Z`}
            fill={segment.color}
            stroke="hsl(var(--background))"
            strokeWidth="2"
          />
        )
      })}
      {/* Inner circle for donut hole */}
      <circle cx="100" cy="100" r="50" fill="hsl(var(--card))" />
    </svg>
  )
}

const DONUT_COLORS = [
  'hsl(var(--primary))',
  'hsl(var(--chart-2, 220 70% 55%))',
  'hsl(var(--chart-3, 142 60% 45%))',
  'hsl(var(--chart-4, 38 92% 55%))',
  'hsl(var(--chart-5, 0 70% 55%))',
  'hsl(var(--info, 217 91% 60%))',
  'hsl(var(--warning, 38 92% 50%))',
  'hsl(var(--muted-foreground))',
]

export default function InsightsPage() {
  const { data: helpers, isLoading: helpersLoading } = useHelpers()
  const { data: executions, isLoading: executionsLoading } = useExecutions({ limit: 100 })
  const { data: connections, isLoading: connectionsLoading } = useConnections()

  const isLoading = helpersLoading || executionsLoading || connectionsLoading

  const { data: insights, isLoading: insightsLoading } = useInsights(helpers, executions, connections)

  // Basic metrics from real data
  const metrics = useMemo(() => {
    const totalHelpers = helpers?.length ?? 0
    const activeHelpers = helpers?.filter((h) => h.status === 'active' && h.enabled)?.length ?? 0
    const totalExec = executions?.length ?? 0
    const completed = executions?.filter((e) => e.status === 'completed')?.length ?? 0
    const failed = executions?.filter((e) => e.status === 'failed')?.length ?? 0
    const successRate = totalExec > 0 ? ((completed / totalExec) * 100).toFixed(1) : '0'
    const withDuration = executions?.filter((e) => e.durationMs > 0) ?? []
    const avgDuration = withDuration.length > 0
      ? Math.round(withDuration.reduce((sum, e) => sum + e.durationMs, 0) / withDuration.length)
      : 0

    return { totalHelpers, activeHelpers, totalExec, completed, failed, successRate, avgDuration }
  }, [helpers, executions])

  // Execution daily trend (last 14 days)
  const dailyTrend = useMemo(() => {
    if (!executions) return []
    const buckets = new Map<string, { date: string; total: number; completed: number; failed: number }>()
    const now = new Date()
    for (let i = 13; i >= 0; i--) {
      const d = new Date(now)
      d.setDate(d.getDate() - i)
      const key = d.toISOString().slice(0, 10)
      buckets.set(key, { date: key, total: 0, completed: 0, failed: 0 })
    }
    for (const exec of executions) {
      const key = exec.startedAt.slice(0, 10)
      const bucket = buckets.get(key)
      if (bucket) {
        bucket.total++
        if (exec.status === 'completed') bucket.completed++
        if (exec.status === 'failed') bucket.failed++
      }
    }
    return Array.from(buckets.values())
  }, [executions])

  return (
    <div className="animate-fade-in-up space-y-8">
      {/* Header */}
      <div>
        <div className="flex items-center gap-2">
          <Sparkles className="h-6 w-6 text-primary" />
          <h1 className="text-2xl font-bold">Insights</h1>
        </div>
        <p className="mt-1 text-muted-foreground">
          AI-powered analytics and recommendations for your CRM automations
        </p>
      </div>

      {/* KPI Overview */}
      {isLoading ? (
        <StatsSkeleton />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {/* Engagement Score */}
          <div className="rounded-lg border bg-card p-5">
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">Engagement Score</p>
              <Target className="h-4 w-4 text-primary" />
            </div>
            <p className={cn(
              'mt-1 text-3xl font-bold',
              (insights?.engagementScore ?? 0) >= 75 ? 'text-success' :
              (insights?.engagementScore ?? 0) >= 50 ? 'text-primary' : 'text-warning'
            )}>
              {insights?.engagementScore ?? 0}
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">out of 100</p>
          </div>

          {/* Automation ROI */}
          <div className="rounded-lg border bg-card p-5">
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">Time Saved</p>
              <Clock className="h-4 w-4 text-primary" />
            </div>
            <p className="mt-1 text-3xl font-bold">
              {insights?.automationRoi.timeSavedHours ?? 0}h
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {insights?.automationRoi.executionsAutomated ?? 0} automated runs
            </p>
          </div>

          {/* Contact Activity */}
          <div className="rounded-lg border bg-card p-5">
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">Contacts Processed</p>
              <Users className="h-4 w-4 text-primary" />
            </div>
            <p className="mt-1 text-3xl font-bold">
              {(insights?.contactGrowth.total ?? 0).toLocaleString()}
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {insights?.contactGrowth.newThisWeek ?? 0} this week
            </p>
          </div>

          {/* Success Rate */}
          <div className="rounded-lg border bg-card p-5">
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">Success Rate</p>
              <CheckCircle className={cn('h-4 w-4', Number(metrics.successRate) >= 95 ? 'text-success' : 'text-warning')} />
            </div>
            <p className={cn(
              'mt-1 text-3xl font-bold',
              Number(metrics.successRate) >= 95 ? 'text-success' :
              Number(metrics.successRate) >= 80 ? 'text-warning' : 'text-destructive'
            )}>
              {metrics.successRate}%
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {metrics.completed} of {metrics.totalExec} executions
            </p>
          </div>
        </div>
      )}

      {/* AI Insights */}
      <div>
        <div className="mb-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-primary" />
            <h2 className="text-lg font-semibold">AI-Powered Insights</h2>
          </div>
          <span className="text-xs text-muted-foreground">
            {insights?.aiInsights.length ?? 0} active insights
          </span>
        </div>
        {isLoading || insightsLoading ? (
          <InsightsSkeleton />
        ) : insights && insights.aiInsights.length > 0 ? (
          <div className="grid gap-4 lg:grid-cols-2">
            {insights.aiInsights.map((insight) => {
              const Icon = getInsightIcon(insight)
              return (
                <div
                  key={insight.id}
                  className={cn(
                    'rounded-lg border border-l-4 bg-card p-4 transition-colors hover:bg-accent/30',
                    getInsightBorder(insight)
                  )}
                >
                  <div className="flex items-start gap-3">
                    <div className={cn('rounded-lg p-2', getInsightIconColor(insight))}>
                      <Icon className="h-4 w-4" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-2">
                        <h3 className="font-semibold text-sm">{insight.title}</h3>
                        {insight.metric && (
                          <div className="flex-shrink-0 text-right">
                            <p className="text-lg font-bold leading-none">{insight.metric}</p>
                            {insight.metricLabel && (
                              <p className="text-[10px] text-muted-foreground">{insight.metricLabel}</p>
                            )}
                          </div>
                        )}
                      </div>
                      <p className="mt-1 text-xs text-muted-foreground leading-relaxed">
                        {insight.description}
                      </p>
                      {insight.action && insight.href && (
                        <Link
                          href={insight.href}
                          className="mt-2 inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline"
                        >
                          {insight.action}
                          <ArrowRight className="h-3 w-3" />
                        </Link>
                      )}
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
            <Sparkles className="mb-3 h-8 w-8 text-muted-foreground/50" />
            <p className="text-sm text-muted-foreground">
              Add connections and helpers to generate AI insights
            </p>
          </div>
        )}
      </div>

      {/* Trend Visualizations */}
      {isLoading || insightsLoading ? (
        <ChartsSkeleton />
      ) : (
        <div className="grid gap-6 lg:grid-cols-2">
          {/* Execution Activity Chart */}
          <div className="rounded-lg border bg-card p-5">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold">Execution Activity (14 days)</h3>
              <Link href="/reports/execution-trends" className="text-xs text-primary hover:underline">
                Full Report
              </Link>
            </div>
            {dailyTrend.length > 0 && dailyTrend.some((d) => d.total > 0) ? (
              <>
                <div className="flex items-end gap-1" style={{ height: 180 }}>
                  {(() => {
                    const maxVal = Math.max(...dailyTrend.map((d) => d.total), 1)
                    return dailyTrend.map((day) => {
                      const completedH = day.total > 0 ? ((day.total - day.failed) / maxVal) * 160 : 0
                      const failedH = day.total > 0 ? (day.failed / maxVal) * 160 : 0
                      return (
                        <div key={day.date} className="flex flex-1 flex-col items-center group relative">
                          <div className="w-full flex flex-col-reverse">
                            <div
                              className="w-full bg-primary/70 hover:bg-primary transition-colors rounded-t"
                              style={{ height: `${Math.max(completedH, 1)}px` }}
                            />
                            {failedH > 0 && (
                              <div
                                className="w-full bg-destructive/60"
                                style={{ height: `${failedH}px` }}
                              />
                            )}
                          </div>
                          <div className="absolute bottom-full mb-1 hidden group-hover:block rounded bg-popover px-2 py-1 text-xs shadow-md border whitespace-nowrap z-10">
                            <p className="font-medium">
                              {new Date(day.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                            </p>
                            <p>{day.total} total</p>
                            {day.failed > 0 && <p className="text-destructive">{day.failed} failed</p>}
                          </div>
                        </div>
                      )
                    })
                  })()}
                </div>
                <div className="flex justify-between mt-2 text-xs text-muted-foreground">
                  <span>{new Date(dailyTrend[0].date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</span>
                  <span>{new Date(dailyTrend[dailyTrend.length - 1].date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</span>
                </div>
                <div className="flex items-center gap-4 mt-2 text-xs text-muted-foreground">
                  <span className="flex items-center gap-1">
                    <span className="h-2 w-2 rounded-sm bg-primary/70" /> Completed
                  </span>
                  <span className="flex items-center gap-1">
                    <span className="h-2 w-2 rounded-sm bg-destructive/60" /> Failed
                  </span>
                </div>
              </>
            ) : (
              <div className="flex items-center justify-center h-[180px] text-sm text-muted-foreground">
                No execution data in the last 14 days
              </div>
            )}
          </div>

          {/* Helper Efficiency Comparison */}
          <div className="rounded-lg border bg-card p-5">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold">Helper Efficiency</h3>
              <Link href="/reports/helper-performance" className="text-xs text-primary hover:underline">
                Full Report
              </Link>
            </div>
            {insights && insights.helperEfficiency.length > 0 ? (
              <div className="space-y-3">
                {insights.helperEfficiency.slice(0, 5).map((helper) => (
                  <div key={helper.helperId} className="space-y-1">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2 min-w-0">
                        <Zap className="h-3 w-3 text-muted-foreground flex-shrink-0" />
                        <span className="text-sm font-medium truncate">{helper.name}</span>
                      </div>
                      <div className="flex items-center gap-3 flex-shrink-0">
                        <span className="text-xs text-muted-foreground">
                          {helper.executionCount} runs
                        </span>
                        <span className={cn(
                          'text-xs font-medium w-12 text-right',
                          helper.successRate >= 95 ? 'text-success' :
                          helper.successRate >= 80 ? 'text-warning' : 'text-destructive'
                        )}>
                          {helper.successRate}%
                        </span>
                      </div>
                    </div>
                    <div className="h-2 w-full rounded-full bg-muted overflow-hidden">
                      <div className="h-full flex">
                        <div
                          className="bg-success/70 rounded-l-full"
                          style={{
                            width: `${helper.successRate}%`,
                          }}
                        />
                        {helper.successRate < 100 && (
                          <div
                            className="bg-destructive/40"
                            style={{
                              width: `${100 - helper.successRate}%`,
                            }}
                          />
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex items-center justify-center h-[180px] text-sm text-muted-foreground">
                Execute helpers to see efficiency data
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tag Distribution + Connection Health */}
      {!isLoading && !insightsLoading && (
        <div className="grid gap-6 lg:grid-cols-2">
          {/* Tag Distribution */}
          <div className="rounded-lg border bg-card p-5">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold">Tag Distribution</h3>
              <Link href="/data-explorer" className="text-xs text-primary hover:underline">
                Explore Data
              </Link>
            </div>
            {insights && insights.tagDistribution.length > 0 ? (
              <div className="flex items-center gap-6">
                <DonutChart
                  data={insights.tagDistribution.map((t, i) => ({
                    name: t.name,
                    percentage: t.percentage,
                    color: DONUT_COLORS[i % DONUT_COLORS.length],
                  }))}
                />
                <div className="flex-1 space-y-2">
                  {insights.tagDistribution.slice(0, 6).map((tag, i) => (
                    <div key={tag.name} className="flex items-center gap-2">
                      <span
                        className="h-2.5 w-2.5 rounded-full flex-shrink-0"
                        style={{ backgroundColor: DONUT_COLORS[i % DONUT_COLORS.length] }}
                      />
                      <span className="text-xs flex-1 truncate">{tag.name}</span>
                      <span className="text-xs text-muted-foreground">{tag.count.toLocaleString()}</span>
                    </div>
                  ))}
                  {insights.tagDistribution.length > 6 && (
                    <p className="text-xs text-muted-foreground pl-5">
                      +{insights.tagDistribution.length - 6} more
                    </p>
                  )}
                </div>
              </div>
            ) : (
              <div className="flex items-center justify-center h-[180px] text-sm text-muted-foreground">
                No tag data available
              </div>
            )}
          </div>

          {/* Connection Health */}
          <div className="rounded-lg border bg-card p-5">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold">Connection Health</h3>
              <Link href="/connections" className="text-xs text-primary hover:underline">
                Manage
              </Link>
            </div>
            {connections && connections.length > 0 ? (
              <div className="space-y-3">
                {connections.map((conn) => {
                  const platform = getCRMPlatform(conn.platformId)
                  return (
                    <div
                      key={conn.connectionId}
                      className="flex items-center gap-3 rounded-md border p-3"
                    >
                      <div className={cn(
                        'h-2.5 w-2.5 rounded-full flex-shrink-0',
                        conn.status === 'active' ? 'bg-success' :
                        conn.status === 'error' ? 'bg-destructive animate-pulse' : 'bg-warning'
                      )} />
                      {platform && (
                        <PlatformLogo platform={platform} size={24} />
                      )}
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium truncate">{conn.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {conn.lastConnected
                            ? `Last active ${new Date(conn.lastConnected).toLocaleDateString()}`
                            : platform?.name || conn.platformId}
                        </p>
                      </div>
                      <span className={cn(
                        'text-xs font-medium px-2 py-0.5 rounded-full',
                        conn.status === 'active' ? 'bg-success/10 text-success' :
                        conn.status === 'error' ? 'bg-destructive/10 text-destructive' :
                        'bg-warning/10 text-warning'
                      )}>
                        {conn.status === 'active' ? 'Healthy' :
                         conn.status === 'error' ? 'Error' : 'Disconnected'}
                      </span>
                    </div>
                  )
                })}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <Link2 className="mb-3 h-8 w-8 text-muted-foreground/50" />
                <p className="text-sm text-muted-foreground">No CRM connections</p>
                <Link
                  href="/connections"
                  className="mt-2 inline-flex items-center gap-1 text-xs text-primary hover:underline"
                >
                  Add your first connection
                  <ArrowRight className="h-3 w-3" />
                </Link>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Quick Reports */}
      <div>
        <div className="mb-4 flex items-center gap-2">
          <BarChart3 className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold">Quick Reports</h2>
        </div>
        <div className="grid gap-4 sm:grid-cols-3">
          <Link
            href="/reports/execution-overview"
            className="flex flex-col items-center rounded-lg border bg-card p-6 text-center transition-colors hover:border-primary hover:shadow-sm"
          >
            <BarChart3 className="h-8 w-8 text-muted-foreground" />
            <span className="mt-2 font-medium">Execution Overview</span>
            <span className="text-xs text-muted-foreground">Full analytics dashboard</span>
          </Link>
          <Link
            href="/reports/helper-performance"
            className="flex flex-col items-center rounded-lg border bg-card p-6 text-center transition-colors hover:border-primary hover:shadow-sm"
          >
            <CheckCircle className="h-8 w-8 text-muted-foreground" />
            <span className="mt-2 font-medium">Helper Performance</span>
            <span className="text-xs text-muted-foreground">Success rates and duration</span>
          </Link>
          <Link
            href="/reports/error-analysis"
            className="flex flex-col items-center rounded-lg border bg-card p-6 text-center transition-colors hover:border-primary hover:shadow-sm"
          >
            <AlertTriangle className="h-8 w-8 text-muted-foreground" />
            <span className="mt-2 font-medium">Error Analysis</span>
            <span className="text-xs text-muted-foreground">Failure patterns and causes</span>
          </Link>
        </div>
      </div>

      {/* AI Chat CTA */}
      <div className="rounded-lg border bg-gradient-to-r from-primary/10 to-primary/5 p-6">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-primary-foreground">
            <Sparkles className="h-5 w-5" />
          </div>
          <div className="flex-1">
            <h3 className="font-semibold">Ask AI about your data</h3>
            <p className="text-sm text-muted-foreground">
              Try: &quot;Which helpers have the highest failure rate?&quot; or &quot;Show contacts without tags&quot;
            </p>
          </div>
          <Link
            href="/data-explorer"
            className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            Explore Data
          </Link>
        </div>
      </div>
    </div>
  )
}
