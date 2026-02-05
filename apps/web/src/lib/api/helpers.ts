import { apiClient } from './client'
import type { Helper, HelperExecution, HelperTypeDefinition } from '@myfusionhelper/types'

export interface CreateHelperInput {
  name: string
  helperType: string
  category: string
  connectionId: string
  config: Record<string, unknown>
}

export interface UpdateHelperInput {
  name?: string
  config?: Record<string, unknown>
  connectionId?: string
  enabled?: boolean
}

export interface ExecuteHelperInput {
  contactId: string
  overrides?: Record<string, unknown>
}

export interface ListExecutionsParams {
  helperId?: string
  status?: string
  limit?: number
  nextToken?: string
}

export const helpersApi = {
  list: () => apiClient.get<Helper[]>('/helpers'),

  get: (id: string) => apiClient.get<Helper>(`/helpers/${id}`),

  create: (input: CreateHelperInput) =>
    apiClient.post<Helper>('/helpers', input),

  update: (id: string, input: UpdateHelperInput) =>
    apiClient.put<Helper>(`/helpers/${id}`, input),

  delete: (id: string) => apiClient.delete<void>(`/helpers/${id}`),

  execute: (id: string, input: ExecuteHelperInput) =>
    apiClient.post<{ executionId: string }>(`/helpers/${id}/execute`, input),

  enable: (id: string) =>
    apiClient.put<Helper>(`/helpers/${id}`, { enabled: true }),

  disable: (id: string) =>
    apiClient.put<Helper>(`/helpers/${id}`, { enabled: false }),

  listTypes: () =>
    apiClient.get<{ types: HelperTypeDefinition[]; total_count: number; categories: string[] }>('/helpers/types'),

  getType: (type: string) =>
    apiClient.get<HelperTypeDefinition>(`/helpers/types/${type}`),

  listExecutions: (params?: ListExecutionsParams) => {
    const searchParams = new URLSearchParams()
    if (params?.helperId) searchParams.set('helper_id', params.helperId)
    if (params?.status) searchParams.set('status', params.status)
    if (params?.limit) searchParams.set('limit', String(params.limit))
    if (params?.nextToken) searchParams.set('next_token', params.nextToken)
    const qs = searchParams.toString()
    return apiClient.get<{ executions: HelperExecution[]; total_count: number; next_token?: string; has_more: boolean }>(`/executions${qs ? `?${qs}` : ''}`)
  },

  getExecution: (id: string) =>
    apiClient.get<HelperExecution>(`/executions/${id}`),
}
