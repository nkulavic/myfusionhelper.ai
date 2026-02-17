import { getMockContacts, getMockDeals, getMockTags } from './crm-mock-data'
import type { DateRangePreset } from '../stores/studio-store'

// ---------------------------------------------------------------------------
// Date range helpers
// ---------------------------------------------------------------------------

function getDateRange(preset: DateRangePreset): { start: Date; end: Date } {
  const end = new Date('2024-01-15T00:00:00Z') // matches BASE_DATE in crm-mock-data
  const start = new Date(end)

  switch (preset) {
    case '7d':
      start.setDate(start.getDate() - 7)
      break
    case '30d':
      start.setDate(start.getDate() - 30)
      break
    case '90d':
      start.setDate(start.getDate() - 90)
      break
    case '12m':
      start.setFullYear(start.getFullYear() - 1)
      break
    case 'all':
      start.setFullYear(start.getFullYear() - 3)
      break
  }

  return { start, end }
}

// ---------------------------------------------------------------------------
// Accessor helpers
// ---------------------------------------------------------------------------

type RecordLike = Record<string, unknown>

function getField(record: RecordLike, dimension: string): string {
  // Direct fields
  if (dimension in record) return String(record[dimension] ?? 'Unknown')

  // Custom fields (contacts)
  const cf = record.customFields as RecordLike | undefined
  if (cf && dimension in cf) return String(cf[dimension] ?? 'Unknown')

  return 'Unknown'
}

function getNumericField(record: RecordLike, field: string): number {
  if (field in record) return Number(record[field]) || 0
  const cf = record.customFields as RecordLike | undefined
  if (cf && field in cf) return Number(cf[field]) || 0
  return 0
}

function getDateField(record: RecordLike): Date | null {
  const dateStr = (record.createdAt as string) || (record.lastActivity as string)
  if (!dateStr) return null
  return new Date(dateStr)
}

// ---------------------------------------------------------------------------
// Core aggregation
// ---------------------------------------------------------------------------

export interface AggregateResult {
  label: string
  value: number
}

export function aggregate(
  dataSource: 'contacts' | 'deals' | 'tags',
  metric: 'count' | 'sum' | 'average',
  dimension: string,
  metricField?: string,
  dateRange?: DateRangePreset
): AggregateResult[] {
  const data = getData(dataSource)
  const range = dateRange ? getDateRange(dateRange) : null
  const filtered = range ? filterByDate(data, range) : data

  if (dimension === '_total') {
    return [{ label: 'Total', value: computeMetric(filtered, metric, metricField) }]
  }

  // Group by dimension
  const groups = new Map<string, RecordLike[]>()
  for (const record of filtered) {
    const key = getField(record, dimension)
    if (!groups.has(key)) groups.set(key, [])
    groups.get(key)!.push(record)
  }

  const results: AggregateResult[] = []
  for (const [label, records] of groups) {
    results.push({ label, value: computeMetric(records, metric, metricField) })
  }

  // Sort descending by value
  results.sort((a, b) => b.value - a.value)
  return results
}

// ---------------------------------------------------------------------------
// Time series
// ---------------------------------------------------------------------------

export interface TimeSeriesPoint {
  date: string
  value: number
}

export function timeSeries(
  dataSource: 'contacts' | 'deals' | 'tags',
  metric: 'count' | 'sum' | 'average',
  metricField?: string,
  dateRange?: DateRangePreset
): TimeSeriesPoint[] {
  const data = getData(dataSource)
  const range = dateRange ? getDateRange(dateRange) : getDateRange('12m')
  const filtered = filterByDate(data, range)

  // Determine interval based on range
  const rangeMs = range.end.getTime() - range.start.getTime()
  const useDays = rangeMs < 45 * 24 * 60 * 60 * 1000 // < 45 days → daily, else monthly

  if (useDays) {
    return aggregateByDay(filtered, metric, metricField, range)
  }
  return aggregateByMonth(filtered, metric, metricField, range)
}

function aggregateByMonth(
  data: RecordLike[],
  metric: 'count' | 'sum' | 'average',
  metricField?: string,
  range?: { start: Date; end: Date }
): TimeSeriesPoint[] {
  const buckets = new Map<string, RecordLike[]>()

  // Initialize months
  if (range) {
    const cursor = new Date(range.start)
    cursor.setDate(1)
    while (cursor <= range.end) {
      const key = `${cursor.getFullYear()}-${String(cursor.getMonth() + 1).padStart(2, '0')}`
      buckets.set(key, [])
      cursor.setMonth(cursor.getMonth() + 1)
    }
  }

  for (const record of data) {
    const d = getDateField(record)
    if (!d) continue
    const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
    if (!buckets.has(key)) buckets.set(key, [])
    buckets.get(key)!.push(record)
  }

  const sorted = [...buckets.entries()].sort((a, b) => a[0].localeCompare(b[0]))
  return sorted.map(([date, records]) => ({
    date,
    value: computeMetric(records, metric, metricField),
  }))
}

function aggregateByDay(
  data: RecordLike[],
  metric: 'count' | 'sum' | 'average',
  metricField?: string,
  range?: { start: Date; end: Date }
): TimeSeriesPoint[] {
  const buckets = new Map<string, RecordLike[]>()

  if (range) {
    const cursor = new Date(range.start)
    while (cursor <= range.end) {
      const key = cursor.toISOString().slice(0, 10)
      buckets.set(key, [])
      cursor.setDate(cursor.getDate() + 1)
    }
  }

  for (const record of data) {
    const d = getDateField(record)
    if (!d) continue
    const key = d.toISOString().slice(0, 10)
    if (!buckets.has(key)) buckets.set(key, [])
    buckets.get(key)!.push(record)
  }

  const sorted = [...buckets.entries()].sort((a, b) => a[0].localeCompare(b[0]))
  return sorted.map(([date, records]) => ({
    date,
    value: computeMetric(records, metric, metricField),
  }))
}

// ---------------------------------------------------------------------------
// Scorecard
// ---------------------------------------------------------------------------

export interface ScorecardResult {
  value: number
  formattedValue: string
}

export function scorecard(
  dataSource: 'contacts' | 'deals' | 'tags',
  metric: 'count' | 'sum' | 'average',
  dimension: string,
  metricField?: string,
  dateRange?: DateRangePreset
): ScorecardResult {
  const data = getData(dataSource)
  const range = dateRange ? getDateRange(dateRange) : null
  let filtered = range ? filterByDate(data, range) : data

  // If dimension is a specific value (e.g. status=active), filter to that
  if (dimension !== '_total' && dimension !== '_timeseries' && metricField) {
    filtered = filtered.filter((r) => {
      const val = getField(r, dimension)
      return val.toLowerCase() === metricField.toLowerCase()
    })
    const count = filtered.length
    return { value: count, formattedValue: formatNumber(count) }
  }

  const value = computeMetric(filtered, metric, metricField)
  return { value, formattedValue: formatNumber(value) }
}

// ---------------------------------------------------------------------------
// Funnel data
// ---------------------------------------------------------------------------

export function funnelData(
  dataSource: 'contacts' | 'deals' | 'tags',
  dimension: string,
  dateRange?: DateRangePreset
): AggregateResult[] {
  const results = aggregate(dataSource, 'count', dimension, undefined, dateRange)

  // For deals/stage, use pipeline order
  if (dataSource === 'deals' && dimension === 'stage') {
    const stageOrder = ['Discovery', 'Proposal', 'Negotiation', 'Closed Won', 'Closed Lost']
    const ordered = stageOrder
      .map((stage) => results.find((r) => r.label === stage))
      .filter(Boolean) as AggregateResult[]
    return ordered
  }

  // For contacts lifecycle, use lifecycle order
  if (dataSource === 'contacts' && dimension === 'status') {
    const order = ['active', 'inactive', 'unsubscribed']
    const ordered = order
      .map((s) => results.find((r) => r.label === s))
      .filter(Boolean) as AggregateResult[]
    return ordered
  }

  return results
}

// ---------------------------------------------------------------------------
// Available dimensions per data source
// ---------------------------------------------------------------------------

export interface DimensionOption {
  value: string
  label: string
  supportsTimeSeries?: boolean
}

export const DIMENSIONS: Record<'contacts' | 'deals' | 'tags', DimensionOption[]> = {
  contacts: [
    { value: '_total', label: 'Total' },
    { value: '_timeseries', label: 'Over Time', supportsTimeSeries: true },
    { value: 'status', label: 'Status' },
    { value: 'source', label: 'Source' },
    { value: 'company', label: 'Company' },
    { value: 'title', label: 'Job Title' },
    { value: 'industry', label: 'Industry' },
    { value: 'leadSource', label: 'Lead Source' },
    { value: 'preferredContact', label: 'Preferred Contact' },
  ],
  deals: [
    { value: '_total', label: 'Total' },
    { value: '_timeseries', label: 'Over Time', supportsTimeSeries: true },
    { value: 'stage', label: 'Stage' },
    { value: 'owner', label: 'Owner' },
    { value: 'contactName', label: 'Contact' },
  ],
  tags: [
    { value: '_total', label: 'Total' },
    { value: 'category', label: 'Category' },
    { value: 'name', label: 'Tag Name' },
  ],
}

export interface MetricFieldOption {
  value: string
  label: string
}

export const METRIC_FIELDS: Record<'contacts' | 'deals' | 'tags', MetricFieldOption[]> = {
  contacts: [
    { value: 'score', label: 'Score' },
    { value: 'annualRevenue', label: 'Annual Revenue' },
  ],
  deals: [
    { value: 'value', label: 'Deal Value' },
    { value: 'probability', label: 'Probability' },
  ],
  tags: [{ value: 'contactCount', label: 'Contact Count' }],
}

// ---------------------------------------------------------------------------
// Internals
// ---------------------------------------------------------------------------

function getData(source: 'contacts' | 'deals' | 'tags'): RecordLike[] {
  switch (source) {
    case 'contacts':
      return getMockContacts() as unknown as RecordLike[]
    case 'deals':
      return getMockDeals() as unknown as RecordLike[]
    case 'tags':
      return getMockTags() as unknown as RecordLike[]
  }
}

function filterByDate(data: RecordLike[], range: { start: Date; end: Date }): RecordLike[] {
  return data.filter((record) => {
    const d = getDateField(record)
    if (!d) return true // include records without dates
    return d >= range.start && d <= range.end
  })
}

function computeMetric(
  records: RecordLike[],
  metric: 'count' | 'sum' | 'average',
  field?: string
): number {
  if (metric === 'count') return records.length

  if (!field) return records.length

  const values = records.map((r) => getNumericField(r, field)).filter((v) => !isNaN(v))
  if (values.length === 0) return 0

  const total = values.reduce((sum, v) => sum + v, 0)
  if (metric === 'sum') return Math.round(total)
  if (metric === 'average') return Math.round(total / values.length)
  return total
}

function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
  return n.toLocaleString()
}
