import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { helpersApi, type CreateHelperInput, type UpdateHelperInput, type ExecuteHelperInput, type ListExecutionsParams } from '@/lib/api/helpers'

export function useHelpers() {
  return useQuery({
    queryKey: ['helpers'],
    queryFn: async () => {
      const res = await helpersApi.list()
      return res.data ?? []
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
      return res.data ?? { executions: [], total_count: 0, has_more: false }
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
