'use client'

import { useState, useMemo } from 'react'
import {
  BarChart3,
  LineChart as LineChartIcon,
  PieChart as PieChartIcon,
  AreaChart as AreaChartIcon,
  Hash,
  Filter,
  Table,
} from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { ChartType, DataSource, MetricType, WidgetSize } from '@/lib/stores/studio-store'
import { DIMENSIONS, METRIC_FIELDS } from '@/lib/mock-data/studio-mock-data'
import { cn } from '@/lib/utils'

interface AddWidgetModalProps {
  open: boolean
  onClose: () => void
  onAdd: (widget: {
    type: ChartType
    title: string
    dataSource: DataSource
    metric: MetricType
    metricField?: string
    dimension: string
    size: WidgetSize
  }) => void
}

const chartTypes: { type: ChartType; label: string; icon: typeof BarChart3; description: string }[] =
  [
    { type: 'scorecard', label: 'Scorecard', icon: Hash, description: 'Single KPI number' },
    { type: 'bar', label: 'Bar Chart', icon: BarChart3, description: 'Compare categories' },
    { type: 'line', label: 'Line Chart', icon: LineChartIcon, description: 'Trends over time' },
    { type: 'area', label: 'Area Chart', icon: AreaChartIcon, description: 'Volume over time' },
    { type: 'pie', label: 'Pie Chart', icon: PieChartIcon, description: 'Proportions' },
    { type: 'funnel', label: 'Funnel', icon: Filter, description: 'Stage progression' },
    { type: 'table', label: 'Table', icon: Table, description: 'Top-N data table' },
  ]

const dataSources: { value: DataSource; label: string }[] = [
  { value: 'contacts', label: 'Contacts' },
  { value: 'deals', label: 'Deals' },
  { value: 'tags', label: 'Tags' },
]

const sizes: { value: WidgetSize; label: string }[] = [
  { value: 'sm', label: 'Small (1/4)' },
  { value: 'md', label: 'Medium (1/2)' },
  { value: 'lg', label: 'Large (2/3)' },
  { value: 'full', label: 'Full Width' },
]

export function AddWidgetModal({ open, onClose, onAdd }: AddWidgetModalProps) {
  const [chartType, setChartType] = useState<ChartType>('bar')
  const [dataSource, setDataSource] = useState<DataSource>('contacts')
  const [metric, setMetric] = useState<MetricType>('count')
  const [metricField, setMetricField] = useState<string>('')
  const [dimension, setDimension] = useState<string>('')
  const [title, setTitle] = useState('')
  const [size, setSize] = useState<WidgetSize>('md')

  const isTimeSeries = chartType === 'line' || chartType === 'area'
  const isScorecard = chartType === 'scorecard'

  const availableDimensions = useMemo(() => {
    const dims = DIMENSIONS[dataSource] || []
    if (isTimeSeries) return dims.filter((d) => d.supportsTimeSeries)
    if (isScorecard) return dims.filter((d) => d.value === '_total')
    if (chartType === 'funnel') {
      // Funnel works best with ordered dimensions
      return dims.filter((d) => d.value !== '_total' && d.value !== '_timeseries')
    }
    return dims.filter((d) => d.value !== '_timeseries')
  }, [dataSource, isTimeSeries, isScorecard, chartType])

  const availableMetricFields = METRIC_FIELDS[dataSource] || []

  // Auto-set dimension when options change
  const effectiveDimension = dimension || (availableDimensions[0]?.value ?? '')

  // Auto-generate title
  const autoTitle = useMemo(() => {
    const ds = dataSources.find((d) => d.value === dataSource)?.label ?? dataSource
    const dim = DIMENSIONS[dataSource]?.find((d) => d.value === effectiveDimension)?.label ?? ''
    const met = metric === 'count' ? '' : ` (${metric}${metricField ? ` ${metricField}` : ''})`

    if (isScorecard) {
      if (metric === 'count') return `Total ${ds}`
      return `${metric === 'sum' ? 'Total' : 'Avg'} ${metricField || ds}`
    }
    if (isTimeSeries) return `${ds}${met} Over Time`
    return `${ds} by ${dim}${met}`
  }, [dataSource, effectiveDimension, metric, metricField, isScorecard, isTimeSeries])

  const handleAdd = () => {
    onAdd({
      type: chartType,
      title: title || autoTitle,
      dataSource,
      metric,
      metricField: metric !== 'count' ? metricField : undefined,
      dimension: effectiveDimension || '_total',
      size: isScorecard ? 'sm' : size,
    })
    // Reset
    setChartType('bar')
    setDataSource('contacts')
    setMetric('count')
    setMetricField('')
    setDimension('')
    setTitle('')
    setSize('md')
    onClose()
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Add Widget</DialogTitle>
        </DialogHeader>

        <div className="space-y-5 py-2">
          {/* Chart Type */}
          <div className="space-y-2">
            <Label>Chart Type</Label>
            <div className="grid grid-cols-4 gap-2 sm:grid-cols-7">
              {chartTypes.map((ct) => (
                <button
                  key={ct.type}
                  onClick={() => {
                    setChartType(ct.type)
                    setDimension('')
                  }}
                  className={cn(
                    'flex flex-col items-center gap-1 rounded-lg border p-2 text-center transition-colors',
                    chartType === ct.type
                      ? 'border-primary bg-primary/5'
                      : 'hover:border-primary/50'
                  )}
                >
                  <ct.icon className="h-5 w-5" />
                  <span className="text-[10px] leading-tight">{ct.label}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Data Source */}
          <div className="space-y-2">
            <Label>Data Source</Label>
            <Select
              value={dataSource}
              onValueChange={(v) => {
                setDataSource(v as DataSource)
                setDimension('')
                setMetricField('')
              }}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {dataSources.map((ds) => (
                  <SelectItem key={ds.value} value={ds.value}>
                    {ds.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Dimension — hidden for scorecard & timeseries (auto-selected) */}
          {!isScorecard && !isTimeSeries && (
            <div className="space-y-2">
              <Label>Group By</Label>
              <Select
                value={effectiveDimension}
                onValueChange={setDimension}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select dimension..." />
                </SelectTrigger>
                <SelectContent>
                  {availableDimensions.map((d) => (
                    <SelectItem key={d.value} value={d.value}>
                      {d.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {/* Metric */}
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label>Metric</Label>
              <Select value={metric} onValueChange={(v) => setMetric(v as MetricType)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="count">Count</SelectItem>
                  <SelectItem value="sum">Sum</SelectItem>
                  <SelectItem value="average">Average</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {metric !== 'count' && (
              <div className="space-y-2">
                <Label>Field</Label>
                <Select value={metricField} onValueChange={setMetricField}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select field..." />
                  </SelectTrigger>
                  <SelectContent>
                    {availableMetricFields.map((f) => (
                      <SelectItem key={f.value} value={f.value}>
                        {f.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
          </div>

          {/* Size — hidden for scorecards */}
          {!isScorecard && (
            <div className="space-y-2">
              <Label>Widget Size</Label>
              <div className="flex gap-2">
                {sizes.map((s) => (
                  <button
                    key={s.value}
                    onClick={() => setSize(s.value)}
                    className={cn(
                      'flex-1 rounded-md border px-3 py-1.5 text-xs font-medium transition-colors',
                      size === s.value
                        ? 'border-primary bg-primary/5'
                        : 'hover:border-primary/50'
                    )}
                  >
                    {s.label}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Title */}
          <div className="space-y-2">
            <Label>Widget Title</Label>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={autoTitle}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleAdd}>Add Widget</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
