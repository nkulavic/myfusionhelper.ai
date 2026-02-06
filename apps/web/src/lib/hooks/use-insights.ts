import { useQuery } from '@tanstack/react-query'
import type { Helper, HelperExecution, PlatformConnection } from '@myfusionhelper/types'

// ---------- Types ----------

export interface AIInsight {
  id: string
  type: 'suggestion' | 'pattern' | 'anomaly' | 'achievement'
  severity: 'info' | 'warning' | 'critical' | 'success'
  title: string
  description: string
  action?: string
  href?: string
  metric?: string
  metricLabel?: string
}

export interface TagDistribution {
  name: string
  count: number
  percentage: number
}

export interface HelperEfficiency {
  helperId: string
  name: string
  successRate: number
  avgDurationMs: number
  executionCount: number
}

export interface InsightsData {
  aiInsights: AIInsight[]
  tagDistribution: TagDistribution[]
  helperEfficiency: HelperEfficiency[]
  engagementScore: number
  automationRoi: { timeSavedHours: number; executionsAutomated: number }
  contactGrowth: { total: number; newThisWeek: number; percentChange: number }
}

// ---------- Mock AI insights generator ----------

function generateAIInsights(
  helpers: Helper[],
  executions: HelperExecution[],
  connections: PlatformConnection[]
): AIInsight[] {
  const insights: AIInsight[] = []

  // Anomaly: high failure rate
  const totalExec = executions.length
  const failedExec = executions.filter((e) => e.status === 'failed').length
  const failRate = totalExec > 0 ? (failedExec / totalExec) * 100 : 0

  if (failRate > 10 && totalExec > 5) {
    insights.push({
      id: 'anomaly-fail-rate',
      type: 'anomaly',
      severity: failRate > 25 ? 'critical' : 'warning',
      title: `Execution failure rate is ${failRate.toFixed(0)}%`,
      description: `${failedExec} of ${totalExec} recent executions failed. This is above the recommended threshold of 5%. Check helper configurations and connection health.`,
      action: 'Review failed executions',
      href: '/executions',
      metric: `${failRate.toFixed(0)}%`,
      metricLabel: 'Failure Rate',
    })
  }

  // Pattern: most active helper
  const helperExecCounts = new Map<string, number>()
  for (const exec of executions) {
    helperExecCounts.set(exec.helperId, (helperExecCounts.get(exec.helperId) || 0) + 1)
  }

  if (helperExecCounts.size > 0) {
    const topHelperId = Array.from(helperExecCounts.entries()).sort((a, b) => b[1] - a[1])[0]
    const topHelper = helpers.find((h) => h.helperId === topHelperId[0])
    if (topHelper) {
      insights.push({
        id: 'pattern-top-helper',
        type: 'pattern',
        severity: 'info',
        title: `"${topHelper.name}" is your most active helper`,
        description: `It has run ${topHelperId[1]} times recently. Consider setting up automated triggers to maximize its impact.`,
        action: 'View helper',
        href: '/helpers',
        metric: String(topHelperId[1]),
        metricLabel: 'Executions',
      })
    }
  }

  // Suggestion: unused helpers
  const executedHelperIds = new Set(executions.map((e) => e.helperId))
  const unusedHelpers = helpers.filter((h) => h.enabled && !executedHelperIds.has(h.helperId))
  if (unusedHelpers.length > 0) {
    insights.push({
      id: 'suggestion-unused-helpers',
      type: 'suggestion',
      severity: 'info',
      title: `${unusedHelpers.length} enabled helper${unusedHelpers.length > 1 ? 's have' : ' has'} never run`,
      description: `You have enabled helpers that haven't been executed yet: ${unusedHelpers.slice(0, 3).map((h) => h.name).join(', ')}${unusedHelpers.length > 3 ? '...' : ''}. Run them to start automating.`,
      action: 'View helpers',
      href: '/helpers',
    })
  }

  // Suggestion: disabled helpers
  const disabledHelpers = helpers.filter((h) => !h.enabled && h.status !== 'deleted')
  if (disabledHelpers.length > 0) {
    insights.push({
      id: 'suggestion-disabled',
      type: 'suggestion',
      severity: 'info',
      title: `${disabledHelpers.length} helper${disabledHelpers.length > 1 ? 's are' : ' is'} disabled`,
      description: 'Enable these helpers to start processing contacts automatically.',
      action: 'Manage helpers',
      href: '/helpers',
    })
  }

  // Anomaly: connection errors
  const errorConns = connections.filter((c) => c.status === 'error')
  if (errorConns.length > 0) {
    insights.push({
      id: 'anomaly-conn-error',
      type: 'anomaly',
      severity: 'critical',
      title: `${errorConns.length} CRM connection${errorConns.length > 1 ? 's' : ''} in error state`,
      description: `The following connection${errorConns.length > 1 ? 's need' : ' needs'} attention: ${errorConns.map((c) => c.name).join(', ')}. Reconnect or update credentials.`,
      action: 'Fix connections',
      href: '/connections',
    })
  }

  // Achievement: perfect success rate
  if (totalExec > 10 && failRate === 0) {
    insights.push({
      id: 'achievement-perfect',
      type: 'achievement',
      severity: 'success',
      title: 'Perfect execution record',
      description: `All ${totalExec} recent executions completed successfully. Your automation setup is running smoothly.`,
      metric: '100%',
      metricLabel: 'Success',
    })
  }

  // Pattern: avg duration insight
  const durations = executions.filter((e) => e.durationMs > 0).map((e) => e.durationMs)
  if (durations.length > 5) {
    const avgMs = durations.reduce((s, d) => s + d, 0) / durations.length
    const slowExecs = durations.filter((d) => d > avgMs * 2).length
    if (slowExecs > 0) {
      insights.push({
        id: 'pattern-slow-execs',
        type: 'pattern',
        severity: 'warning',
        title: `${slowExecs} execution${slowExecs > 1 ? 's' : ''} took over 2x average duration`,
        description: `Average duration is ${Math.round(avgMs)}ms, but some executions are significantly slower. This may indicate API rate limits or large data sets.`,
        action: 'View executions',
        href: '/executions',
        metric: `${Math.round(avgMs)}ms`,
        metricLabel: 'Avg Duration',
      })
    }
  }

  // Suggestion: no connections
  if (connections.length === 0) {
    insights.push({
      id: 'suggestion-no-conn',
      type: 'suggestion',
      severity: 'info',
      title: 'Connect your first CRM',
      description: 'Add a CRM connection to unlock the full power of automation helpers.',
      action: 'Add connection',
      href: '/connections',
    })
  }

  // Suggestion: no helpers
  if (helpers.length === 0) {
    insights.push({
      id: 'suggestion-no-helpers',
      type: 'suggestion',
      severity: 'info',
      title: 'Create your first helper',
      description: 'Browse 62 automation helpers across 7 categories to start automating your CRM workflow.',
      action: 'Browse helpers',
      href: '/helpers',
    })
  }

  // Always return at least one insight
  if (insights.length === 0) {
    insights.push({
      id: 'suggestion-explore',
      type: 'suggestion',
      severity: 'info',
      title: 'Explore new automation helpers',
      description: 'Discover helpers that can tag contacts, sync data, trigger workflows, and more.',
      action: 'Browse catalog',
      href: '/helpers',
    })
  }

  return insights
}

function generateHelperEfficiency(
  helpers: Helper[],
  executions: HelperExecution[]
): HelperEfficiency[] {
  const helperMap = new Map<string, { total: number; completed: number; totalDuration: number; withDuration: number }>()

  for (const exec of executions) {
    let entry = helperMap.get(exec.helperId)
    if (!entry) {
      entry = { total: 0, completed: 0, totalDuration: 0, withDuration: 0 }
      helperMap.set(exec.helperId, entry)
    }
    entry.total++
    if (exec.status === 'completed') entry.completed++
    if (exec.durationMs > 0) {
      entry.totalDuration += exec.durationMs
      entry.withDuration++
    }
  }

  return Array.from(helperMap.entries())
    .map(([helperId, stats]) => {
      const helper = helpers.find((h) => h.helperId === helperId)
      return {
        helperId,
        name: helper?.name || helperId,
        successRate: stats.total > 0 ? Math.round((stats.completed / stats.total) * 1000) / 10 : 0,
        avgDurationMs: stats.withDuration > 0 ? Math.round(stats.totalDuration / stats.withDuration) : 0,
        executionCount: stats.total,
      }
    })
    .sort((a, b) => b.executionCount - a.executionCount)
    .slice(0, 8)
}

// Mock tag distribution (until backend provides real tag analytics)
const MOCK_TAG_DISTRIBUTION: TagDistribution[] = [
  { name: 'Customer', count: 4250, percentage: 28.5 },
  { name: 'Hot Lead', count: 2100, percentage: 14.1 },
  { name: 'Newsletter', count: 1890, percentage: 12.7 },
  { name: 'VIP', count: 1340, percentage: 9.0 },
  { name: 'Prospect', count: 1180, percentage: 7.9 },
  { name: 'Event Attendee', count: 920, percentage: 6.2 },
  { name: 'Warm Lead', count: 870, percentage: 5.8 },
  { name: 'Other', count: 2350, percentage: 15.8 },
]

// ---------- Hook ----------

export function useInsights(
  helpers: Helper[] | undefined,
  executions: HelperExecution[] | undefined,
  connections: PlatformConnection[] | undefined
) {
  return useQuery({
    queryKey: ['insights', helpers?.length, executions?.length, connections?.length],
    queryFn: async (): Promise<InsightsData> => {
      const h = helpers ?? []
      const e = executions ?? []
      const c = connections ?? []

      const aiInsights = generateAIInsights(h, e, c)
      const helperEfficiency = generateHelperEfficiency(h, e)

      // Compute engagement score (0-100) based on activity
      const activeHelpers = h.filter((x) => x.enabled).length
      const totalHelpers = h.length
      const totalExec = e.length
      const successRate = totalExec > 0
        ? e.filter((x) => x.status === 'completed').length / totalExec
        : 0
      const activeConns = c.filter((x) => x.status === 'active').length

      const engagementScore = Math.min(100, Math.round(
        (activeHelpers > 0 ? 25 : 0) +
        (activeConns > 0 ? 25 : 0) +
        (successRate * 30) +
        Math.min(20, totalExec * 0.2)
      ))

      // Automation ROI
      const completedExec = e.filter((x) => x.status === 'completed').length
      const avgDuration = e.filter((x) => x.durationMs > 0)
      const avgMs = avgDuration.length > 0
        ? avgDuration.reduce((s, x) => s + x.durationMs, 0) / avgDuration.length
        : 0
      // Estimate: each automated execution saves ~2 minutes of manual work
      const timeSavedHours = Math.round((completedExec * 2) / 60 * 10) / 10

      // Contact growth (mock data augmented with real execution counts)
      const uniqueContacts = new Set(e.filter((x) => x.contactId).map((x) => x.contactId)).size
      const recentExecs = e.filter((x) => {
        const d = new Date(x.startedAt)
        const weekAgo = new Date()
        weekAgo.setDate(weekAgo.getDate() - 7)
        return d >= weekAgo
      })
      const recentContacts = new Set(recentExecs.filter((x) => x.contactId).map((x) => x.contactId)).size

      return {
        aiInsights,
        tagDistribution: MOCK_TAG_DISTRIBUTION,
        helperEfficiency,
        engagementScore,
        automationRoi: {
          timeSavedHours,
          executionsAutomated: completedExec,
        },
        contactGrowth: {
          total: uniqueContacts || 0,
          newThisWeek: recentContacts || 0,
          percentChange: uniqueContacts > 0 ? Math.round((recentContacts / uniqueContacts) * 100) : 0,
        },
      }
    },
    enabled: helpers !== undefined || executions !== undefined || connections !== undefined,
    staleTime: 60 * 1000,
  })
}
