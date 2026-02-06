import { NextRequest, NextResponse } from 'next/server'
import { cookies } from 'next/headers'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'https://api.myfusionhelper.ai'

interface BackendExecution {
  execution_id: string
  helper_id: string
  contact_id?: string
  status: string
  started_at: string
  completed_at?: string
  duration_ms?: number
  error?: string
}

interface BackendExecutionsResponse {
  data?: {
    executions: BackendExecution[]
    total_count: number
    has_more: boolean
    next_token?: string
  }
}

interface DailyBucket {
  date: string
  total: number
  completed: number
  failed: number
}

interface HelperStat {
  helperId: string
  total: number
  completed: number
  failed: number
  avgDurationMs: number
}

async function fetchAllExecutions(token: string, limit = 200): Promise<BackendExecution[]> {
  const allExecutions: BackendExecution[] = []
  let nextToken: string | undefined

  // Fetch up to a few pages to get a decent dataset
  for (let page = 0; page < 5; page++) {
    const params = new URLSearchParams({ limit: String(limit) })
    if (nextToken) params.set('next_token', nextToken)

    const res = await fetch(`${API_BASE_URL}/executions?${params}`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    })

    if (!res.ok) break

    const body = (await res.json()) as BackendExecutionsResponse
    if (body.data?.executions) {
      allExecutions.push(...body.data.executions)
    }

    if (!body.data?.has_more || !body.data?.next_token) break
    nextToken = body.data.next_token
  }

  return allExecutions
}

function computeStats(executions: BackendExecution[]) {
  const total = executions.length
  const completed = executions.filter((e) => e.status === 'completed').length
  const failed = executions.filter((e) => e.status === 'failed').length
  const running = executions.filter((e) => e.status === 'running').length
  const pending = executions.filter((e) => e.status === 'pending').length

  const withDuration = executions.filter((e) => e.duration_ms && e.duration_ms > 0)
  const avgDurationMs =
    withDuration.length > 0
      ? Math.round(withDuration.reduce((sum, e) => sum + (e.duration_ms || 0), 0) / withDuration.length)
      : 0

  const successRate = total > 0 ? Math.round((completed / total) * 1000) / 10 : 0

  // Unique contacts processed
  const uniqueContacts = new Set(executions.filter((e) => e.contact_id).map((e) => e.contact_id)).size

  // Unique helpers used
  const uniqueHelpers = new Set(executions.map((e) => e.helper_id)).size

  // Daily buckets (last 30 days)
  const dailyMap = new Map<string, DailyBucket>()
  const now = new Date()
  for (let i = 29; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    const key = d.toISOString().slice(0, 10)
    dailyMap.set(key, { date: key, total: 0, completed: 0, failed: 0 })
  }

  for (const exec of executions) {
    const dateKey = exec.started_at.slice(0, 10)
    const bucket = dailyMap.get(dateKey)
    if (bucket) {
      bucket.total++
      if (exec.status === 'completed') bucket.completed++
      if (exec.status === 'failed') bucket.failed++
    }
  }

  const dailyTrend = Array.from(dailyMap.values())

  // Per-helper stats
  const helperMap = new Map<string, { total: number; completed: number; failed: number; totalDuration: number; withDuration: number }>()
  for (const exec of executions) {
    let entry = helperMap.get(exec.helper_id)
    if (!entry) {
      entry = { total: 0, completed: 0, failed: 0, totalDuration: 0, withDuration: 0 }
      helperMap.set(exec.helper_id, entry)
    }
    entry.total++
    if (exec.status === 'completed') entry.completed++
    if (exec.status === 'failed') entry.failed++
    if (exec.duration_ms && exec.duration_ms > 0) {
      entry.totalDuration += exec.duration_ms
      entry.withDuration++
    }
  }

  const topHelpers: HelperStat[] = Array.from(helperMap.entries())
    .map(([helperId, stats]) => ({
      helperId,
      total: stats.total,
      completed: stats.completed,
      failed: stats.failed,
      avgDurationMs: stats.withDuration > 0 ? Math.round(stats.totalDuration / stats.withDuration) : 0,
    }))
    .sort((a, b) => b.total - a.total)
    .slice(0, 10)

  // Error breakdown (group by error message prefix)
  const errorMap = new Map<string, number>()
  for (const exec of executions) {
    if (exec.status === 'failed' && exec.error) {
      const errorKey = exec.error.length > 60 ? exec.error.slice(0, 60) + '...' : exec.error
      errorMap.set(errorKey, (errorMap.get(errorKey) || 0) + 1)
    }
  }

  const errorBreakdown = Array.from(errorMap.entries())
    .map(([error, count]) => ({ error, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 10)

  return {
    summary: {
      total,
      completed,
      failed,
      running,
      pending,
      successRate,
      avgDurationMs,
      uniqueContacts,
      uniqueHelpers,
    },
    dailyTrend,
    topHelpers,
    errorBreakdown,
  }
}

export async function GET(request: NextRequest) {
  try {
    // Get auth token from request headers (forwarded by client)
    const authHeader = request.headers.get('authorization')
    const token = authHeader?.replace('Bearer ', '')

    if (!token) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const executions = await fetchAllExecutions(token)
    const stats = computeStats(executions)

    return NextResponse.json(stats)
  } catch (error) {
    console.error('Error generating report stats:', error)
    return NextResponse.json(
      { error: 'Failed to generate report stats' },
      { status: 500 }
    )
  }
}
