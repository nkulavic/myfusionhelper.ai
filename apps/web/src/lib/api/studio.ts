import { apiClient } from './client'
import type {
  Widget,
  Dashboard,
  ChartType,
  DataSource,
  MetricType,
  WidgetSize,
} from '@/lib/stores/studio-store'

// ---------------------------------------------------------------------------
// API response types (after camelCase auto-conversion by apiClient)
// ---------------------------------------------------------------------------

export interface ApiDashboard {
  dashboardId: string
  accountId: string
  name: string
  description?: string
  widgets: ApiWidget[]
  templateId?: string
  connectionId?: string
  status: string
  createdBy: string
  createdAt: string
  updatedAt: string
}

export interface ApiWidget {
  widgetId: string
  type: string
  title: string
  dataSource: string
  metric: string
  metricField?: string
  dimension: string
  size: string
  order: number
  connectionId?: string
}

export interface ApiTemplate {
  id: string
  name: string
  description: string
  platform: string
  widgets: ApiWidget[]
}

// ---------------------------------------------------------------------------
// Mapping: API → Frontend types
// ---------------------------------------------------------------------------

export function mapWidget(w: ApiWidget): Widget {
  return {
    id: w.widgetId,
    type: w.type as ChartType,
    title: w.title,
    dataSource: w.dataSource as DataSource,
    metric: w.metric as MetricType,
    metricField: w.metricField,
    dimension: w.dimension,
    size: w.size as WidgetSize,
    order: w.order,
    connectionId: w.connectionId,
  }
}

export function mapDashboard(d: ApiDashboard): Dashboard {
  return {
    id: d.dashboardId,
    name: d.name,
    description: d.description,
    widgets: (d.widgets || []).map(mapWidget),
    connectionId: d.connectionId,
    templateId: d.templateId,
    createdAt: d.createdAt,
    updatedAt: d.updatedAt,
  }
}

// ---------------------------------------------------------------------------
// Mapping: Frontend → API (for request bodies)
// ---------------------------------------------------------------------------

function mapWidgetToApi(w: Widget): Record<string, unknown> {
  return {
    widgetId: w.id,
    type: w.type,
    title: w.title,
    dataSource: w.dataSource,
    metric: w.metric,
    metricField: w.metricField,
    dimension: w.dimension,
    size: w.size,
    order: w.order,
    connectionId: w.connectionId,
  }
}

// ---------------------------------------------------------------------------
// Aggregation types (used with postRaw — snake_case)
// ---------------------------------------------------------------------------

export interface AggregateRequest {
  connection_id: string
  object_type: string
  metric: string
  metric_field?: string
  dimension: string
  date_range?: { start: string; end: string }
  limit?: number
}

export interface AggregateResponse {
  results: { label: string; value: number }[]
  total: number
  query_time_ms: number
}

export interface TimeSeriesRequest {
  connection_id: string
  object_type: string
  metric: string
  metric_field?: string
  interval: string
  date_range?: { start: string; end: string }
  date_column?: string
}

export interface TimeSeriesResponse {
  points: { date: string; value: number }[]
  interval: string
  query_time_ms: number
}

// ---------------------------------------------------------------------------
// API client
// ---------------------------------------------------------------------------

export const studioApi = {
  // Dashboard CRUD (uses apiClient with auto key conversion)
  listDashboards: () =>
    apiClient.get<{ dashboards: ApiDashboard[] }>('/studio/dashboards'),

  createDashboard: (data: {
    name: string
    description?: string
    widgets?: Widget[]
    connectionId?: string
    templateId?: string
  }) =>
    apiClient.post<ApiDashboard>('/studio/dashboards', {
      ...data,
      widgets: data.widgets?.map(mapWidgetToApi),
    }),

  getDashboard: (id: string) =>
    apiClient.get<ApiDashboard>(`/studio/dashboards/${id}`),

  updateDashboard: (
    id: string,
    data: { name?: string; description?: string; widgets?: Widget[] },
  ) =>
    apiClient.put<ApiDashboard>(`/studio/dashboards/${id}`, {
      ...data,
      widgets: data.widgets?.map(mapWidgetToApi),
    }),

  deleteDashboard: (id: string) =>
    apiClient.delete(`/studio/dashboards/${id}`),

  // Templates
  listTemplates: () =>
    apiClient.get<{ templates: ApiTemplate[] }>('/studio/templates'),

  applyTemplate: (templateId: string, data: { connectionId: string; name?: string }) =>
    apiClient.post<ApiDashboard>(`/studio/templates/${templateId}/apply`, data),

  // Aggregation (uses postRaw — dynamic data, no key conversion)
  aggregate: (params: AggregateRequest) =>
    apiClient.postRaw<AggregateResponse>('/data/aggregate', params),

  timeSeries: (params: TimeSeriesRequest) =>
    apiClient.postRaw<TimeSeriesResponse>('/data/timeseries', params),
}
