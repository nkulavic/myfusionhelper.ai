'use client'

import { useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { funnelData } from '@/lib/mock-data/studio-mock-data'
import type { Widget, DateRangePreset } from '@/lib/stores/studio-store'
import { CHART_COLORS } from './theme'

interface FunnelChartWidgetProps {
  widget: Widget
  dateRange: DateRangePreset
}

export function FunnelChartWidget({ widget, dateRange }: FunnelChartWidgetProps) {
  const data = useMemo(
    () => funnelData(widget.dataSource, widget.dimension, dateRange),
    [widget.dataSource, widget.dimension, dateRange]
  )

  const maxValue = data.length > 0 ? data[0].value : 1

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {widget.title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-2">
          {data.map((item, index) => {
            const widthPercent = Math.max((item.value / maxValue) * 100, 20)
            const conversionRate =
              index > 0 && data[index - 1].value > 0
                ? ((item.value / data[index - 1].value) * 100).toFixed(0)
                : null

            return (
              <div key={item.label} className="flex items-center gap-3">
                <div className="flex flex-1 flex-col items-center">
                  {conversionRate && (
                    <div className="mb-0.5 text-[10px] text-muted-foreground">
                      {conversionRate}%
                    </div>
                  )}
                  <div
                    className="flex items-center justify-center rounded-md py-2.5 text-sm font-medium transition-all"
                    style={{
                      width: `${widthPercent}%`,
                      backgroundColor: CHART_COLORS[index % CHART_COLORS.length],
                      color: 'white',
                      minWidth: 80,
                    }}
                  >
                    {item.label}
                  </div>
                </div>
                <span className="w-12 text-right text-sm font-mono text-muted-foreground">
                  {item.value}
                </span>
              </div>
            )
          })}
        </div>
      </CardContent>
    </Card>
  )
}
