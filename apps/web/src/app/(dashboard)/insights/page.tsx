'use client'

import { useMemo } from 'react'
import Link from 'next/link'
import {
  Sparkles,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  Lightbulb,
  BarChart3,
  CheckCircle,
  Loader2,
} from 'lucide-react'
import { useHelpers, useExecutions } from '@/lib/hooks/use-helpers'
import { useConnections } from '@/lib/hooks/use-connections'
import { Skeleton } from '@/components/ui/skeleton'

export default function InsightsPage() {
  const { data: helpers, isLoading: helpersLoading } = useHelpers()
  const { data: executions, isLoading: executionsLoading } = useExecutions({ limit: 100 })
  const { data: connections, isLoading: connectionsLoading } = useConnections()

  const isLoading = helpersLoading || executionsLoading || connectionsLoading

  const metrics = useMemo(() => {
    const totalHelpers = helpers?.length ?? 0
    const activeHelpers = helpers?.filter((h) => h.status === 'active')?.length ?? 0
    const totalExecutions = executions?.length ?? 0
    const completedExecutions = executions?.filter((e) => e.status === 'completed')?.length ?? 0
    const failedExecutions = executions?.filter((e) => e.status === 'failed')?.length ?? 0
    const successRate = totalExecutions > 0 ? ((completedExecutions / totalExecutions) * 100).toFixed(1) : '0'
    const avgDurationMs = executions?.length
      ? Math.round(
          executions.filter((e) => e.durationMs && e.durationMs > 0)
            .reduce((sum, e) => sum + (e.durationMs || 0), 0) /
            Math.max(executions.filter((e) => e.durationMs && e.durationMs > 0).length, 1)
        )
      : 0

    return [
      { label: 'Active Helpers', value: String(activeHelpers), sub: `${totalHelpers} total` },
      { label: 'Executions', value: String(totalExecutions), sub: `${completedExecutions} succeeded` },
      { label: 'Success Rate', value: `${successRate}%`, sub: `${failedExecutions} failed`, good: Number(successRate) >= 95 },
      { label: 'Avg Duration', value: avgDurationMs > 1000 ? `${(avgDurationMs / 1000).toFixed(1)}s` : `${avgDurationMs}ms`, sub: 'per execution' },
    ]
  }, [helpers, executions])

  const insights = useMemo(() => {
    const items: Array<{
      id: string
      type: string
      title: string
      description: string
      action: string
      href: string
      icon: typeof AlertTriangle
      color: string
    }> = []

    const totalHelpers = helpers?.length ?? 0
    const activeHelpers = helpers?.filter((h) => h.status === 'active')?.length ?? 0
    const disabledHelpers = totalHelpers - activeHelpers
    const totalConnections = connections?.length ?? 0
    const errorConnections = connections?.filter((c) => c.status === 'error')?.length ?? 0
    const failedExecutions = executions?.filter((e) => e.status === 'failed')?.length ?? 0
    const totalExecutions = executions?.length ?? 0

    if (errorConnections > 0) {
      items.push({
        id: 'conn-error',
        type: 'alert',
        title: `${errorConnections} connection${errorConnections > 1 ? 's' : ''} need attention`,
        description: 'Some CRM connections are in an error state. Check credentials and reconnect.',
        action: 'Fix connections',
        href: '/connections',
        icon: AlertTriangle,
        color: 'text-destructive',
      })
    }

    if (failedExecutions > 0 && totalExecutions > 0) {
      const failRate = ((failedExecutions / totalExecutions) * 100).toFixed(0)
      items.push({
        id: 'fail-rate',
        type: 'alert',
        title: `${failRate}% execution failure rate`,
        description: `${failedExecutions} of ${totalExecutions} recent executions failed. Review error logs to identify the issue.`,
        action: 'View executions',
        href: '/executions',
        icon: AlertTriangle,
        color: 'text-warning',
      })
    }

    if (disabledHelpers > 0) {
      items.push({
        id: 'disabled',
        type: 'opportunity',
        title: `${disabledHelpers} helper${disabledHelpers > 1 ? 's' : ''} disabled`,
        description: 'You have configured helpers that are currently disabled. Enable them to start automating.',
        action: 'View helpers',
        href: '/helpers',
        icon: Lightbulb,
        color: 'text-info',
      })
    }

    if (totalConnections === 0) {
      items.push({
        id: 'no-conn',
        type: 'suggestion',
        title: 'Connect your first CRM',
        description: 'Add a CRM connection to start using helpers with your contact data.',
        action: 'Add connection',
        href: '/connections',
        icon: Lightbulb,
        color: 'text-info',
      })
    }

    if (totalHelpers === 0) {
      items.push({
        id: 'no-helpers',
        type: 'suggestion',
        title: 'Create your first helper',
        description: 'Set up an automation helper to start processing contacts automatically.',
        action: 'Browse helpers',
        href: '/helpers',
        icon: Lightbulb,
        color: 'text-info',
      })
    }

    if (totalExecutions > 10) {
      const completedCount = executions?.filter((e) => e.status === 'completed')?.length ?? 0
      if (completedCount === totalExecutions) {
        items.push({
          id: 'all-green',
          type: 'trend',
          title: 'All executions succeeded',
          description: 'Your recent helpers are running without errors. Great job!',
          action: 'View details',
          href: '/executions',
          icon: TrendingUp,
          color: 'text-success',
        })
      }
    }

    // Always show at least one suggestion
    if (items.length === 0) {
      items.push({
        id: 'explore',
        type: 'suggestion',
        title: 'Explore helper catalog',
        description: 'Browse 60+ automation helpers across 7 categories to find new ways to automate your workflow.',
        action: 'Browse catalog',
        href: '/helpers',
        icon: Sparkles,
        color: 'text-primary',
      })
    }

    return items
  }, [helpers, executions, connections])

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold">Insights</h1>
        <p className="text-muted-foreground">Analytics and recommendations for your automations</p>
      </div>

      {/* Metrics */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {isLoading
          ? [1, 2, 3, 4].map((i) => (
              <div key={i} className="rounded-lg border bg-card p-4">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="mt-2 h-7 w-16" />
                <Skeleton className="mt-1 h-3 w-20" />
              </div>
            ))
          : metrics.map((metric) => (
              <div key={metric.label} className="rounded-lg border bg-card p-4">
                <p className="text-sm text-muted-foreground">{metric.label}</p>
                <p className="mt-1 text-2xl font-bold">{metric.value}</p>
                <p className="mt-0.5 text-xs text-muted-foreground">{metric.sub}</p>
              </div>
            ))}
      </div>

      {/* AI Insights */}
      <div>
        <div className="mb-4 flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold">Recommendations</h2>
        </div>
        {isLoading ? (
          <div className="grid gap-4 lg:grid-cols-2">
            {[1, 2].map((i) => (
              <div key={i} className="rounded-lg border bg-card p-4">
                <div className="flex gap-3">
                  <Skeleton className="h-9 w-9 rounded-lg" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-5 w-48" />
                    <Skeleton className="h-4 w-full" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="grid gap-4 lg:grid-cols-2">
            {insights.map((insight) => (
              <div key={insight.id} className="rounded-lg border bg-card p-4">
                <div className="mb-3 flex items-start gap-3">
                  <div className={`rounded-lg bg-muted p-2 ${insight.color}`}>
                    <insight.icon className="h-5 w-5" />
                  </div>
                  <div className="flex-1">
                    <h3 className="font-semibold">{insight.title}</h3>
                    <p className="mt-1 text-sm text-muted-foreground">{insight.description}</p>
                  </div>
                </div>
                <Link href={insight.href} className="text-sm font-medium text-primary hover:underline">
                  {insight.action} &rarr;
                </Link>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Quick Reports */}
      <div>
        <div className="mb-4 flex items-center gap-2">
          <BarChart3 className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold">Quick Reports</h2>
        </div>
        <div className="grid gap-4 sm:grid-cols-3">
          <Link
            href="/reports"
            className="flex flex-col items-center rounded-lg border bg-card p-6 text-center hover:border-primary hover:shadow-md"
          >
            <BarChart3 className="h-8 w-8 text-muted-foreground" />
            <span className="mt-2 font-medium">Weekly Summary</span>
            <span className="text-sm text-muted-foreground">Last 7 days overview</span>
          </Link>
          <Link
            href="/reports"
            className="flex flex-col items-center rounded-lg border bg-card p-6 text-center hover:border-primary hover:shadow-md"
          >
            <CheckCircle className="h-8 w-8 text-muted-foreground" />
            <span className="mt-2 font-medium">Helper Performance</span>
            <span className="text-sm text-muted-foreground">Execution analytics</span>
          </Link>
          <Link
            href="/connections"
            className="flex flex-col items-center rounded-lg border bg-card p-6 text-center hover:border-primary hover:shadow-md"
          >
            <TrendingUp className="h-8 w-8 text-muted-foreground" />
            <span className="mt-2 font-medium">Connection Health</span>
            <span className="text-sm text-muted-foreground">CRM connection status</span>
          </Link>
        </div>
      </div>

      {/* AI Chat Prompt */}
      <div className="rounded-lg border bg-gradient-to-r from-primary/10 to-primary/5 p-6">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-primary-foreground">
            <Sparkles className="h-5 w-5" />
          </div>
          <div className="flex-1">
            <h3 className="font-semibold">Ask AI anything about your data</h3>
            <p className="text-sm text-muted-foreground">
              Try: &quot;Show me contacts who haven&apos;t been emailed in 30 days&quot;
            </p>
          </div>
          <button className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
            Open Chat
          </button>
        </div>
      </div>
    </div>
  )
}
