'use client'

import type { Widget, DateRangePreset } from '@/lib/stores/studio-store'
import { ScorecardChart } from './charts/scorecard'
import { BarChartWidget } from './charts/bar-chart'
import { LineChartWidget } from './charts/line-chart'
import { AreaChartWidget } from './charts/area-chart'
import { PieChartWidget } from './charts/pie-chart'
import { FunnelChartWidget } from './charts/funnel-chart'
import { DataTableWidget } from './charts/data-table'

interface WidgetRendererProps {
  widget: Widget
  dateRange: DateRangePreset
}

export function WidgetRenderer({ widget, dateRange }: WidgetRendererProps) {
  switch (widget.type) {
    case 'scorecard':
      return <ScorecardChart widget={widget} dateRange={dateRange} />
    case 'bar':
      return <BarChartWidget widget={widget} dateRange={dateRange} />
    case 'line':
      return <LineChartWidget widget={widget} dateRange={dateRange} />
    case 'area':
      return <AreaChartWidget widget={widget} dateRange={dateRange} />
    case 'pie':
      return <PieChartWidget widget={widget} dateRange={dateRange} />
    case 'funnel':
      return <FunnelChartWidget widget={widget} dateRange={dateRange} />
    case 'table':
      return <DataTableWidget widget={widget} dateRange={dateRange} />
    default:
      return <div className="p-4 text-sm text-muted-foreground">Unknown widget type</div>
  }
}
