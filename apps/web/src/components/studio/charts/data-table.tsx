'use client'

import { useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { aggregate } from '@/lib/mock-data/studio-mock-data'
import type { Widget, DateRangePreset } from '@/lib/stores/studio-store'

interface DataTableWidgetProps {
  widget: Widget
  dateRange: DateRangePreset
}

export function DataTableWidget({ widget, dateRange }: DataTableWidgetProps) {
  const data = useMemo(
    () =>
      aggregate(
        widget.dataSource,
        widget.metric,
        widget.dimension,
        widget.metricField,
        dateRange
      ).slice(0, 20),
    [widget.dataSource, widget.metric, widget.dimension, widget.metricField, dateRange]
  )

  const metricLabel =
    widget.metric === 'count'
      ? 'Count'
      : widget.metric === 'sum'
        ? `Sum (${widget.metricField ?? ''})`
        : `Avg (${widget.metricField ?? ''})`

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {widget.title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b">
                <th className="pb-2 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  #
                </th>
                <th className="pb-2 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {widget.dimension === '_total' ? 'Item' : widget.dimension}
                </th>
                <th className="pb-2 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {metricLabel}
                </th>
                <th className="pb-2 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  %
                </th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {data.map((row, i) => {
                const total = data.reduce((sum, r) => sum + r.value, 0)
                const pct = total > 0 ? ((row.value / total) * 100).toFixed(1) : '0'
                return (
                  <tr key={row.label} className="hover:bg-muted/50">
                    <td className="py-2 text-sm text-muted-foreground">{i + 1}</td>
                    <td className="py-2 text-sm font-medium">{row.label}</td>
                    <td className="py-2 text-right font-mono text-sm">
                      {row.value.toLocaleString()}
                    </td>
                    <td className="py-2 text-right text-sm text-muted-foreground">{pct}%</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
          {data.length === 0 && (
            <p className="py-8 text-center text-sm text-muted-foreground">No data</p>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
