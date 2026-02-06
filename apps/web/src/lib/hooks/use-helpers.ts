import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { Helper, HelperTypeDefinition } from '@myfusionhelper/types'
import { helpersApi, type CreateHelperInput, type UpdateHelperInput, type ExecuteHelperInput, type ListExecutionsParams } from '@/lib/api/helpers'

export function useHelpers() {
  return useQuery({
    queryKey: ['helpers'],
    queryFn: async () => {
      const res = await helpersApi.list()
      // Backend returns { helpers: [...], totalCount: N } inside data
      const data = res.data as unknown
      if (Array.isArray(data)) return data as Helper[]
      if (data && typeof data === 'object' && 'helpers' in data) {
        const nested = (data as { helpers: Helper[] }).helpers
        if (Array.isArray(nested)) return nested
      }
      return (res.data ?? []) as Helper[]
    },
  })
}

export function useHelper(id: string) {
  return useQuery({
    queryKey: ['helpers', id],
    queryFn: async () => {
      const res = await helpersApi.get(id)
      return res.data
    },
    enabled: !!id,
  })
}

export function useCreateHelper() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateHelperInput) => helpersApi.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['helpers'] })
    },
  })
}

export function useUpdateHelper() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateHelperInput }) =>
      helpersApi.update(id, input),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['helpers'] })
      queryClient.invalidateQueries({ queryKey: ['helpers', id] })
    },
  })
}

export function useDeleteHelper() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => helpersApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['helpers'] })
    },
  })
}

export function useExecuteHelper() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: ExecuteHelperInput }) =>
      helpersApi.execute(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['executions'] })
    },
  })
}

export function useHelperTypes() {
  return useQuery({
    queryKey: ['helper-types'],
    queryFn: async () => {
      const res = await helpersApi.listTypes()
      const data = res.data as unknown
      if (data && typeof data === 'object' && 'types' in data) {
        const nested = data as { types: HelperTypeDefinition[]; totalCount: number; categories: string[] }
        return {
          types: Array.isArray(nested.types) ? nested.types : [],
          categories: Array.isArray(nested.categories) ? nested.categories : [],
          totalCount: nested.totalCount ?? 0,
        }
      }
      return { types: [] as HelperTypeDefinition[], categories: [] as string[], totalCount: 0 }
    },
    staleTime: 5 * 60 * 1000, // cache for 5 min -- types don't change often
  })
}

export function useHelperType(type: string) {
  return useQuery({
    queryKey: ['helper-types', type],
    queryFn: async () => {
      const res = await helpersApi.getType(type)
      return res.data as HelperTypeDefinition
    },
    enabled: !!type,
    staleTime: 5 * 60 * 1000,
  })
}

export function useExecutions(params?: ListExecutionsParams) {
  return useQuery({
    queryKey: ['executions', params],
    queryFn: async () => {
      const res = await helpersApi.listExecutions(params)
      return res.data?.executions ?? []
    },
  })
}

export function useExecutionsPaginated(params?: ListExecutionsParams) {
  return useQuery({
    queryKey: ['executions', 'paginated', params],
    queryFn: async () => {
      const res = await helpersApi.listExecutions(params)
      return res.data ?? { executions: [], totalCount: 0, hasMore: false }
    },
  })
}

export function useExecution(id: string) {
  return useQuery({
    queryKey: ['executions', id],
    queryFn: async () => {
      const res = await helpersApi.getExecution(id)
      return res.data
    },
    enabled: !!id,
  })
}
