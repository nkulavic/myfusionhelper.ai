'use client'

import Link from 'next/link'
import {
  ArrowRight,
  CheckCircle,
  XCircle,
  Clock,
  Link2,
  Blocks,
  Activity,
  TrendingUp,
  Plus,
} from 'lucide-react'
import { useHelpers, useExecutions } from '@/lib/hooks/use-helpers'
import { useConnections } from '@/lib/hooks/use-connections'
import { usePlanLimits } from '@/lib/hooks/use-plan-limits'
import type { PlatformConnection } from '@myfusionhelper/types'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { FirstStepsTab } from './_components/first-steps-tab'

const quickActions = [
  { label: 'Create Helper', href: '/helpers?view=new', icon: Plus },
  { label: 'Add Connection', href: '/connections', icon: Link2 },
  { label: 'View Executions', href: '/executions', icon: Activity },
  { label: 'Browse Helpers', href: '/helpers', icon: Blocks },
]

function StatsSkeleton() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {[1, 2, 3, 4].map((i) => (
        <div key={i} className="rounded-lg border bg-card p-5">
          <div className="flex items-center justify-between">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-4 w-4" />
          </div>
          <Skeleton className="mt-2 h-8 w-16" />
          <Skeleton className="mt-1 h-3 w-20" />
        </div>
      ))}
    </div>
  )
}

function ExecutionsSkeleton() {
  return (
    <div className="divide-y">
      {[1, 2, 3, 4, 5].map((i) => (
        <div key={i} className="flex items-center gap-4 px-5 py-3">
          <Skeleton className="h-4 w-4 rounded-full" />
          <div className="flex-1">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="mt-1 h-3 w-32" />
          </div>
          <div className="text-right">
            <Skeleton className="h-3 w-12" />
            <Skeleton className="mt-1 h-3 w-16" />
          </div>
        </div>
      ))}
    </div>
  )
}

export default function DashboardPage() {
  const { data: helpers, isLoading: helpersLoading } = useHelpers()
  const { data: connections, isLoading: connectionsLoading } = useConnections()
  const { data: executions, isLoading: executionsLoading } = useExecutions({ limit: 5 })
  const { isTrialing, isTrialExpired } = usePlanLimits()

  const activeHelpers = helpers?.filter((h) => h.status === 'active').length ?? 0
  const activeConnections = connections?.filter((c) => c.status === 'active') ?? []
  const todayExecutions = executions?.length ?? 0
  const successRate =
    executions && executions.length > 0
      ? (
          (executions.filter((e) => e.status === 'completed').length / executions.length) *
          100
        ).toFixed(1)
      : '0'

  const statsLoading = helpersLoading || connectionsLoading || executionsLoading

  const hasConnections = (connections?.length ?? 0) > 0
  const hasHelpers = (helpers?.length ?? 0) > 0
  const hasExecutions = (executions?.length ?? 0) > 0
  const allSetupComplete = hasConnections && hasHelpers && hasExecutions

  const showFirstSteps = (isTrialing || isTrialExpired) && !allSetupComplete
  const defaultTab = showFirstSteps ? 'first-steps' : 'dashboard'

  const stats = [
    {
      label: 'Active Helpers',
      value: String(activeHelpers),
      change: `${helpers?.length ?? 0} total`,
      icon: Blocks,
    },
    {
      label: 'Recent Executions',
      value: String(todayExecutions),
      change: 'Latest runs',
      icon: Activity,
    },
    {
      label: 'Success Rate',
      value: `${successRate}%`,
      change: 'Based on recent',
      icon: TrendingUp,
    },
    {
      label: 'Connections',
      value: String(activeConnections.length),
      change: `${connections?.length ?? 0} total`,
      icon: Link2,
    },
  ]

  return (
    <div className="animate-fade-in-up space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground">Overview of your CRM automation activity</p>
      </div>

      {/* Tabs: First Steps (trial only) + Dashboard */}
      {showFirstSteps ? (
        <Tabs defaultValue={defaultTab}>
          <TabsList>
            <TabsTrigger value="first-steps">First Steps</TabsTrigger>
            <TabsTrigger value="dashboard">Dashboard</TabsTrigger>
          </TabsList>

          <TabsContent value="first-steps" className="mt-6">
            <FirstStepsTab
              connections={connections}
              helpers={helpers}
              executions={executions}
              isLoading={statsLoading}
            />
          </TabsContent>

          <TabsContent value="dashboard" className="mt-6">
            <DashboardContent
              stats={stats}
              statsLoading={statsLoading}
              executions={executions}
              executionsLoading={executionsLoading}
              connections={connections}
              connectionsLoading={connectionsLoading}
            />
          </TabsContent>
        </Tabs>
      ) : (
        <DashboardContent
          stats={stats}
          statsLoading={statsLoading}
          executions={executions}
          executionsLoading={executionsLoading}
          connections={connections}
          connectionsLoading={connectionsLoading}
        />
      )}
    </div>
  )
}

function DashboardContent({
  stats,
  statsLoading,
  executions,
  executionsLoading,
  connections,
  connectionsLoading,
}: {
  stats: { label: string; value: string; change: string; icon: React.ComponentType<{ className?: string }> }[]
  statsLoading: boolean
  executions: ReturnType<typeof useExecutions>['data']
  executionsLoading: boolean
  connections: ReturnType<typeof useConnections>['data']
  connectionsLoading: boolean
}) {
  return (
    <div className="space-y-8">
      {/* Stats Grid */}
      {statsLoading ? (
        <StatsSkeleton />
      ) : (
        <div className="animate-stagger-in grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {stats.map((stat) => (
            <div key={stat.label} className="card-hover rounded-lg border bg-card p-5">
              <div className="flex items-center justify-between">
                <p className="text-sm text-muted-foreground">{stat.label}</p>
                <stat.icon className="h-4 w-4 text-muted-foreground" />
              </div>
              <p className="mt-2 text-2xl font-bold">{stat.value}</p>
              <p className="mt-1 text-xs text-muted-foreground">{stat.change}</p>
            </div>
          ))}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Recent Executions */}
        <div className="rounded-lg border bg-card lg:col-span-2">
          <div className="flex items-center justify-between border-b px-5 py-4">
            <h2 className="font-semibold">Recent Executions</h2>
            <Link
              href="/executions"
              className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
            >
              View all
              <ArrowRight className="h-3 w-3" />
            </Link>
          </div>
          {executionsLoading ? (
            <ExecutionsSkeleton />
          ) : executions && executions.length > 0 ? (
            <div className="skeleton-fade-enter divide-y">
              {executions.map((exec) => (
                <Link
                  key={exec.executionId}
                  href={`/executions/${exec.executionId}`}
                  className="flex items-center gap-4 px-5 py-3 transition-colors hover:bg-accent/50"
                >
                  <div className="flex-shrink-0">
                    {exec.status === 'completed' && (
                      <CheckCircle className="h-4 w-4 text-success" />
                    )}
                    {exec.status === 'failed' && (
                      <XCircle className="h-4 w-4 text-destructive" />
                    )}
                    {(exec.status === 'running' || exec.status === 'pending') && (
                      <Clock className="h-4 w-4 animate-spin text-info" />
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">{exec.helperId}</p>
                    <p className="truncate text-xs text-muted-foreground">
                      {exec.contactId || 'No contact'}
                    </p>
                  </div>
                  <div className="flex-shrink-0 text-right">
                    <p className="text-xs font-mono text-muted-foreground">
                      {exec.durationMs ? `${exec.durationMs}ms` : '-'}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(exec.startedAt).toLocaleTimeString()}
                    </p>
                  </div>
                </Link>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Activity className="mb-3 h-8 w-8 text-muted-foreground/50" />
              <p className="text-sm font-medium text-muted-foreground">No executions yet</p>
              <p className="mt-1 max-w-xs text-xs text-muted-foreground">
                Once you configure and trigger a helper from your CRM, every execution will appear
                here with status, timing, and contact details.
              </p>
            </div>
          )}
        </div>

        {/* Quick Actions + Connection Health */}
        <div className="space-y-4">
          <div className="rounded-lg border bg-card p-5">
            <h2 className="mb-4 font-semibold">Quick Actions</h2>
            <div className="space-y-2">
              {quickActions.map((action) => (
                <Link
                  key={action.label}
                  href={action.href}
                  className="flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium transition-all hover:bg-accent active:scale-[0.98]"
                >
                  <action.icon className="h-4 w-4 text-primary" />
                  {action.label}
                  <ArrowRight className="ml-auto h-3 w-3 text-muted-foreground" />
                </Link>
              ))}
            </div>
          </div>

          {/* Connection Health */}
          <div className="rounded-lg border bg-card p-5">
            <h2 className="mb-4 font-semibold">Connection Health</h2>
            {connectionsLoading ? (
              <div className="space-y-3">
                {[1, 2].map((i) => (
                  <div key={i} className="flex items-center gap-3">
                    <Skeleton className="h-2 w-2 rounded-full" />
                    <Skeleton className="h-4 flex-1" />
                    <Skeleton className="h-3 w-16" />
                  </div>
                ))}
              </div>
            ) : connections && connections.length > 0 ? (
              <div className="space-y-3">
                {connections.map((conn: PlatformConnection) => (
                  <div key={conn.connectionId} className="flex items-center gap-3">
                    <div
                      className={`h-2 w-2 rounded-full ${
                        conn.status === 'active'
                          ? 'bg-success'
                          : conn.status === 'error'
                            ? 'bg-destructive'
                            : 'bg-warning'
                      }`}
                    />
                    <span className="flex-1 text-sm">{conn.name}</span>
                    <span
                      className={`text-xs ${
                        conn.status === 'active'
                          ? 'text-success'
                          : conn.status === 'error'
                            ? 'text-destructive'
                            : 'text-warning'
                      }`}
                    >
                      {conn.status === 'active'
                        ? 'Healthy'
                        : conn.status === 'error'
                          ? 'Error'
                          : 'Warning'}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="py-4 text-center">
                <p className="text-sm font-medium text-muted-foreground">No connections yet</p>
                <p className="mt-1 text-xs text-muted-foreground">
                  Connect your CRM to start using helpers
                </p>
                <Link
                  href="/connections"
                  className="mt-2 inline-flex items-center gap-1 text-xs text-primary hover:underline"
                >
                  Add your first connection
                  <ArrowRight className="h-3 w-3" />
                </Link>
              </div>
            )}
            {connections && connections.length > 0 && (
              <Link
                href="/connections"
                className="mt-4 inline-flex items-center gap-1 text-xs text-primary hover:underline"
              >
                Manage connections
                <ArrowRight className="h-3 w-3" />
              </Link>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
