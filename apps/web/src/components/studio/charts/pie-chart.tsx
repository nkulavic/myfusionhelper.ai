'use client'

import { useMemo } from 'react'
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { aggregate } from '@/lib/mock-data/studio-mock-data'
import type { Widget, DateRangePreset } from '@/lib/stores/studio-store'
import { CHART_COLORS, chartTooltipStyle } from './theme'

interface PieChartWidgetProps {
  widget: Widget
  dateRange: DateRangePreset
}

export function PieChartWidget({ widget, dateRange }: PieChartWidgetProps) {
  const data = useMemo(
    () =>
      aggregate(
        widget.dataSource,
        widget.metric,
        widget.dimension,
        widget.metricField,
        dateRange
      ).slice(0, 8),
    [widget.dataSource, widget.metric, widget.dimension, widget.metricField, dateRange]
  )

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {widget.title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={data}
                cx="50%"
                cy="50%"
                innerRadius={50}
                outerRadius={85}
                dataKey="value"
                nameKey="label"
                paddingAngle={2}
              >
                {data.map((_, index) => (
                  <Cell
                    key={`cell-${index}`}
                    fill={CHART_COLORS[index % CHART_COLORS.length]}
                  />
                ))}
              </Pie>
              <Tooltip contentStyle={chartTooltipStyle} />
              <Legend
                verticalAlign="bottom"
                iconType="circle"
                iconSize={8}
                formatter={(value: string) => (
                  <span style={{ color: 'hsl(var(--foreground))', fontSize: 12 }}>
                    {value}
                  </span>
                )}
              />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}
