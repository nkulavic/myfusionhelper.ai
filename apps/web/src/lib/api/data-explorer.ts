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
  displayName?: string
  sampleValues?: string[]
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

export interface DataQueryResponse {
  records: Record<string, unknown>[]
  totalRecords: number
  page: number
  pageSize: number
  totalPages: number
  hasNextPage: boolean
  hasPrevPage: boolean
  columns: string[]
  schema?: FieldSchema[]
  queryTimeMs: number
  generatedSql?: string
  modelUsed?: string
}

export interface CatalogObjectType {
  objectType: string
  label: string
  icon: string
  recordCount?: number
  connectionId: string
  connectionName: string
  platformId: string
  platformName: string
}

export interface DataCatalogResponse {
  sources: CatalogObjectType[]
}

export interface RecordDetailResponse {
  record: Record<string, unknown>
  objectType: string
  connectionId: string
}

export const dataExplorerApi = {
  getCatalog: () =>
    apiClient.get<DataCatalogResponse>('/data/catalog'),

  query: (params: DataQueryRequest) =>
    apiClient.post<DataQueryResponse>('/data/query', params),

  getRecord: (connectionId: string, objectType: string, recordId: string) =>
    apiClient.get<RecordDetailResponse>(
      `/data/record/${connectionId}/${objectType}/${recordId}`
    ),

  exportRecords: (connectionId: string, objectType: string, format: 'json' | 'csv') =>
    apiClient.post<{ downloadUrl: string }>('/data/export', {
      connectionId,
      objectType,
      format,
    }),
}
