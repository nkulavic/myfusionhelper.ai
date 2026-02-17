'use client'

import { useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { scorecard } from '@/lib/mock-data/studio-mock-data'
import type { Widget, DateRangePreset } from '@/lib/stores/studio-store'

interface ScorecardChartProps {
  widget: Widget
  dateRange: DateRangePreset
}

export function ScorecardChart({ widget, dateRange }: ScorecardChartProps) {
  const result = useMemo(
    () =>
      scorecard(
        widget.dataSource,
        widget.metric,
        widget.dimension,
        widget.metricField,
        dateRange
      ),
    [widget.dataSource, widget.metric, widget.dimension, widget.metricField, dateRange]
  )

  const prefix = widget.metric === 'sum' && widget.metricField === 'value' ? '$' : ''

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {widget.title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-bold">
          {prefix}
          {result.formattedValue}
        </div>
        <p className="mt-1 text-xs text-muted-foreground capitalize">
          {widget.metric === 'count' ? 'records' : widget.metricField ?? ''}
        </p>
      </CardContent>
    </Card>
  )
}
