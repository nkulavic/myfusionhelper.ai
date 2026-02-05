'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import {
  ArrowLeft,
  BarChart3,
  Calendar,
  Download,
  RefreshCw,
  Clock,
  TrendingUp,
  TrendingDown,
  Share2,
  MoreHorizontal,
} from 'lucide-react'

const mockReport = {
  id: 'rpt_001',
  name: 'January 2026 Monthly Summary',
  template: 'Weekly Summary',
  description: 'Overview of all helper activity and key metrics for January 2026.',
  createdAt: '2026-01-31',
  lastRun: '2026-01-31T18:00:00Z',
  schedule: 'monthly',
  dateRange: { start: '2026-01-01', end: '2026-01-31' },
  metrics: {
    totalExecutions: 45_892,
    successRate: 98.7,
    contactsProcessed: 12_847,
    avgDuration: 124,
    errorsCount: 596,
    uniqueHelpers: 14,
  },
  topHelpers: [
    { name: 'Tag It', executions: 12_450, successRate: 99.2 },
    { name: 'Copy It', executions: 8_732, successRate: 98.9 },
    { name: 'Date Calculator', executions: 6_421, successRate: 97.8 },
    { name: 'Format It', executions: 5_843, successRate: 99.1 },
    { name: 'Score It', executions: 4_210, successRate: 96.5 },
  ],
  dailyExecutions: [
    { date: 'Jan 1', count: 1200 },
    { date: 'Jan 5', count: 1450 },
    { date: 'Jan 10', count: 1800 },
    { date: 'Jan 15', count: 1650 },
    { date: 'Jan 20', count: 2100 },
    { date: 'Jan 25', count: 1900 },
    { date: 'Jan 31', count: 2200 },
  ],
}

export default function ReportDetailPage() {
  const params = useParams()
  const report = mockReport

  const maxExec = Math.max(...report.dailyExecutions.map((d) => d.count))

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link
          href="/reports"
          className="rounded-md p-2 hover:bg-accent"
        >
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">{report.name}</h1>
          <p className="text-muted-foreground">{report.description}</p>
        </div>
        <div className="flex items-center gap-2">
          <button className="inline-flex items-center gap-2 rounded-md border px-3 py-2 text-sm font-medium hover:bg-accent">
            <RefreshCw className="h-4 w-4" />
            Refresh
          </button>
          <button className="inline-flex items-center gap-2 rounded-md border px-3 py-2 text-sm font-medium hover:bg-accent">
            <Download className="h-4 w-4" />
            Export
          </button>
          <button className="inline-flex items-center gap-2 rounded-md border px-3 py-2 text-sm font-medium hover:bg-accent">
            <Share2 className="h-4 w-4" />
            Share
          </button>
        </div>
      </div>

      {/* Report Meta */}
      <div className="flex items-center gap-6 text-sm text-muted-foreground">
        <span className="flex items-center gap-1">
          <Calendar className="h-4 w-4" />
          {report.dateRange.start} to {report.dateRange.end}
        </span>
        <span className="flex items-center gap-1">
          <Clock className="h-4 w-4" />
          Last run: {new Date(report.lastRun).toLocaleString()}
        </span>
        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
          report.schedule === 'monthly'
            ? 'bg-primary/10 text-primary'
            : 'bg-info/10 text-info'
        }`}>
          {report.schedule}
        </span>
      </div>

      {/* KPI Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
        {[
          { label: 'Total Executions', value: report.metrics.totalExecutions.toLocaleString(), icon: BarChart3 },
          { label: 'Success Rate', value: `${report.metrics.successRate}%`, icon: TrendingUp },
          { label: 'Contacts Processed', value: report.metrics.contactsProcessed.toLocaleString(), icon: null },
          { label: 'Avg Duration', value: `${report.metrics.avgDuration}ms`, icon: Clock },
          { label: 'Errors', value: report.metrics.errorsCount.toLocaleString(), icon: TrendingDown },
          { label: 'Active Helpers', value: report.metrics.uniqueHelpers.toString(), icon: null },
        ].map((kpi) => (
          <div key={kpi.label} className="rounded-lg border bg-card p-4">
            <p className="text-xs text-muted-foreground">{kpi.label}</p>
            <p className="mt-1 text-xl font-bold">{kpi.value}</p>
          </div>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Execution Trend Chart (simple bar chart) */}
        <div className="rounded-lg border bg-card p-5">
          <h3 className="mb-4 font-semibold">Execution Trend</h3>
          <div className="flex items-end gap-2" style={{ height: 200 }}>
            {report.dailyExecutions.map((day) => (
              <div key={day.date} className="flex flex-1 flex-col items-center gap-1">
                <span className="text-xs text-muted-foreground">{day.count}</span>
                <div
                  className="w-full rounded-t bg-primary"
                  style={{ height: `${(day.count / maxExec) * 160}px` }}
                />
                <span className="text-xs text-muted-foreground">{day.date}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Top Helpers */}
        <div className="rounded-lg border bg-card p-5">
          <h3 className="mb-4 font-semibold">Top Helpers</h3>
          <div className="space-y-3">
            {report.topHelpers.map((helper, i) => (
              <div key={helper.name} className="flex items-center gap-3">
                <span className="flex h-7 w-7 items-center justify-center rounded-full bg-muted text-xs font-medium">
                  {i + 1}
                </span>
                <div className="flex-1">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium">{helper.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {helper.executions.toLocaleString()} runs
                    </p>
                  </div>
                  <div className="mt-1 h-2 w-full rounded-full bg-muted">
                    <div
                      className="h-full rounded-full bg-primary"
                      style={{
                        width: `${(helper.executions / report.topHelpers[0].executions) * 100}%`,
                      }}
                    />
                  </div>
                </div>
                <span className="text-xs text-success">{helper.successRate}%</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
