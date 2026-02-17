import { apiClient } from './client'

// Types

export type FilterOperator = 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte'
  | 'contains' | 'startswith' | 'in' | 'between' | 'daterange'

export interface FilterCondition {
  column: string
  operator: FilterOperator
  value: string | number | string[]
  value2?: string | number
}

export interface FieldSchema {
  name: string
  type: string  // 'string' | 'number' | 'date' | 'boolean' | 'json'
  display_name?: string
  sample_values?: string[]
}

export interface DataQueryRequest {
  connectionId: string
  objectType: string
  page?: number
  pageSize?: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
  filterConditions?: FilterCondition[]
  nlQuery?: string
  model?: string
  search?: string
}

// Response from Go backend (snake_case — no key transformation applied)
export interface DataQueryResponse {
  records: Record<string, unknown>[]
  total_records: number
  page: number
  page_size: number
  total_pages: number
  has_next_page: boolean
  has_prev_page: boolean
  columns: string[]
  schema?: FieldSchema[]
  query_time_ms: number
  generated_sql?: string
  model_used?: string
}

export interface CatalogObjectType {
  object_type: string
  label: string
  icon: string
  record_count?: number
  connection_id: string
  connection_name: string
  platform_id: string
  platform_name: string
  last_synced_at?: string
  sync_status?: string
}

export interface DataCatalogResponse {
  sources: CatalogObjectType[]
}

export interface RecordDetailResponse {
  record: Record<string, unknown>
  object_type: string
  connection_id: string
}

// Use raw methods to skip key transformation — data explorer records have
// dynamic column names (from parquet) that must stay as-is (e.g. "first_name").
export const dataExplorerApi = {
  getCatalog: () =>
    apiClient.getRaw<DataCatalogResponse>('/data/catalog'),

  query: (params: DataQueryRequest) =>
    apiClient.postRaw<DataQueryResponse>('/data/query', {
      connection_id: params.connectionId,
      object_type: params.objectType,
      page: params.page,
      page_size: params.pageSize,
      sort_by: params.sortBy,
      sort_order: params.sortOrder,
      filter_conditions: params.filterConditions,
      search: params.search,
    }),

  getRecord: (connectionId: string, objectType: string, recordId: string) =>
    apiClient.getRaw<RecordDetailResponse>(
      `/data/record/${connectionId}/${objectType}/${recordId}`
    ),

  exportRecords: (connectionId: string, objectType: string, format: 'json' | 'csv') =>
    apiClient.postRaw<Blob>('/data/export', {
      connection_id: connectionId,
      object_type: objectType,
      format,
    }),

  triggerSync: (connectionId: string) =>
    apiClient.postRaw<{ connection_id: string; status: string }>('/data/sync', {
      connection_id: connectionId,
    }),
}
