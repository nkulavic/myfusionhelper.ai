import { useQuery } from '@tanstack/react-query'
import { getAccessToken } from '@/lib/auth-client'

export interface ReportSummary {
  total: number
  completed: number
  failed: number
  running: number
  pending: number
  successRate: number
  avgDurationMs: number
  uniqueContacts: number
  uniqueHelpers: number
}

export interface DailyBucket {
  date: string
  total: number
  completed: number
  failed: number
}

export interface HelperStat {
  helperId: string
  total: number
  completed: number
  failed: number
  avgDurationMs: number
}

export interface ErrorEntry {
  error: string
  count: number
}

export interface ReportStats {
  summary: ReportSummary
  dailyTrend: DailyBucket[]
  topHelpers: HelperStat[]
  errorBreakdown: ErrorEntry[]
}

async function fetchReportStats(): Promise<ReportStats> {
  const token = getAccessToken()
  const res = await fetch('/api/reports/stats', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })

  if (!res.ok) {
    throw new Error(`Failed to fetch report stats: ${res.status}`)
  }

  return res.json()
}

export function useReportStats() {
  return useQuery({
    queryKey: ['report-stats'],
    queryFn: fetchReportStats,
    staleTime: 60 * 1000, // 1 minute
    retry: 1,
  })
}
