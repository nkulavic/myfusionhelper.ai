import { create } from 'zustand'

// ---------------------------------------------------------------------------
// Types (exported for use by components and hooks)
// ---------------------------------------------------------------------------

export type ChartType = 'scorecard' | 'bar' | 'line' | 'area' | 'pie' | 'funnel' | 'table'
export type DataSource = 'contacts' | 'deals' | 'tags'
export type MetricType = 'count' | 'sum' | 'average'
export type WidgetSize = 'sm' | 'md' | 'lg' | 'full'

export interface Widget {
  id: string
  type: ChartType
  title: string
  dataSource: DataSource
  metric: MetricType
  metricField?: string
  dimension: string
  size: WidgetSize
  order: number
  connectionId?: string
}

export interface Dashboard {
  id: string
  name: string
  description?: string
  widgets: Widget[]
  connectionId?: string
  templateId?: string
  createdAt: string
  updatedAt: string
}

export type DateRangePreset = '7d' | '30d' | '90d' | '12m' | 'all'

// ---------------------------------------------------------------------------
// Store — UI-only state (dashboard data comes from React Query hooks)
// ---------------------------------------------------------------------------

interface StudioState {
  activeDateRange: DateRangePreset
}

interface StudioActions {
  setDateRange: (range: DateRangePreset) => void
}

export const useStudioStore = create<StudioState & StudioActions>()((set) => ({
  activeDateRange: '12m',
  setDateRange: (range) => set({ activeDateRange: range }),
}))
