'use client'

import { useState } from 'react'
import Link from 'next/link'
import {
  BarChart3,
  Plus,
  Calendar,
  TrendingUp,
  Users,
  Tag,
  Zap,
  FileText,
  Download,
  Clock,
  ArrowRight,
  Search,
} from 'lucide-react'

const reportTemplates = [
  {
    id: 'weekly-summary',
    name: 'Weekly Summary',
    description: 'Overview of all helper activity, contact changes, and key metrics from the past week.',
    icon: Calendar,
    category: 'overview',
  },
  {
    id: 'helper-performance',
    name: 'Helper Performance',
    description: 'Execution success rates, average duration, and error analysis for each helper.',
    icon: Zap,
    category: 'performance',
  },
  {
    id: 'tag-distribution',
    name: 'Tag Distribution',
    description: 'Analysis of tag usage patterns, most/least applied tags, and segmentation coverage.',
    icon: Tag,
    category: 'contacts',
  },
  {
    id: 'contact-growth',
    name: 'Contact Growth',
    description: 'Track new contacts, engagement rates, and pipeline movement over time.',
    icon: Users,
    category: 'contacts',
  },
  {
    id: 'execution-trends',
    name: 'Execution Trends',
    description: 'Daily/weekly execution volume, peak usage times, and processing speed trends.',
    icon: TrendingUp,
    category: 'performance',
  },
  {
    id: 'error-analysis',
    name: 'Error Analysis',
    description: 'Breakdown of failed executions by error type, helper, and connection.',
    icon: FileText,
    category: 'performance',
  },
]

const savedReports = [
  {
    id: 'rpt_001',
    name: 'January 2026 Monthly Summary',
    template: 'Weekly Summary',
    createdAt: '2026-01-31',
    lastRun: '2026-01-31T18:00:00Z',
    schedule: 'monthly',
  },
  {
    id: 'rpt_002',
    name: 'Q4 Helper Performance',
    template: 'Helper Performance',
    createdAt: '2026-01-15',
    lastRun: '2026-01-15T12:00:00Z',
    schedule: 'one-time',
  },
  {
    id: 'rpt_003',
    name: 'Weekly Tag Report',
    template: 'Tag Distribution',
    createdAt: '2026-02-01',
    lastRun: '2026-02-03T09:00:00Z',
    schedule: 'weekly',
  },
]

const categories = ['all', 'overview', 'performance', 'contacts']

export default function ReportsPage() {
  const [activeTab, setActiveTab] = useState<'saved' | 'templates'>('saved')
  const [categoryFilter, setCategoryFilter] = useState('all')
  const [search, setSearch] = useState('')

  const filteredTemplates = reportTemplates.filter((t) => {
    if (categoryFilter !== 'all' && t.category !== categoryFilter) return false
    if (search && !t.name.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Reports</h1>
          <p className="text-muted-foreground">Generate and schedule automated reports</p>
        </div>
        <button className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
          <Plus className="h-4 w-4" />
          New Report
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 rounded-lg border bg-muted p-1">
        <button
          onClick={() => setActiveTab('saved')}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium transition-colors ${
            activeTab === 'saved' ? 'bg-background shadow-sm' : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Saved Reports
        </button>
        <button
          onClick={() => setActiveTab('templates')}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium transition-colors ${
            activeTab === 'templates' ? 'bg-background shadow-sm' : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Templates
        </button>
      </div>

      {activeTab === 'saved' ? (
        <div className="space-y-4">
          {savedReports.length === 0 ? (
            <div className="rounded-lg border bg-card p-12 text-center">
              <BarChart3 className="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">No reports yet</h3>
              <p className="mt-2 text-sm text-muted-foreground">
                Create your first report from a template to start tracking your data.
              </p>
              <button
                onClick={() => setActiveTab('templates')}
                className="mt-4 inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
              >
                Browse Templates
              </button>
            </div>
          ) : (
            <div className="divide-y rounded-lg border bg-card">
              {savedReports.map((report) => (
                <Link
                  key={report.id}
                  href={`/reports/${report.id}`}
                  className="flex items-center gap-4 px-5 py-4 transition-colors hover:bg-accent/50"
                >
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                    <BarChart3 className="h-5 w-5 text-primary" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{report.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {report.template} &middot; Created {report.createdAt}
                    </p>
                  </div>
                  <div className="flex items-center gap-4 flex-shrink-0">
                    <div className="text-right">
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                        report.schedule === 'weekly'
                          ? 'bg-info/10 text-info'
                          : report.schedule === 'monthly'
                          ? 'bg-primary/10 text-primary'
                          : 'bg-muted text-muted-foreground'
                      }`}>
                        {report.schedule}
                      </span>
                    </div>
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
                      <Clock className="h-3 w-3" />
                      {new Date(report.lastRun).toLocaleDateString()}
                    </div>
                    <ArrowRight className="h-4 w-4 text-muted-foreground" />
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      ) : (
        <div className="space-y-4">
          {/* Search and Filter */}
          <div className="flex gap-3">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search templates..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full rounded-md border bg-background py-2 pl-9 pr-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
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
              <div
                key={template.id}
                className="group rounded-lg border bg-card p-5 transition-colors hover:border-primary"
              >
                <div className="mb-3 flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                    <template.icon className="h-5 w-5 text-primary" />
                  </div>
                  <h3 className="font-semibold">{template.name}</h3>
                </div>
                <p className="text-sm text-muted-foreground">{template.description}</p>
                <div className="mt-4 flex items-center gap-2">
                  <button className="inline-flex items-center gap-1 rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground hover:bg-primary/90">
                    Use Template
                  </button>
                  <button className="inline-flex items-center gap-1 rounded-md border px-3 py-1.5 text-xs font-medium hover:bg-accent">
                    Preview
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
