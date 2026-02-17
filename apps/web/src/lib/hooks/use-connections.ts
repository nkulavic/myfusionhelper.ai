import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { connectionsApi, type CreateConnectionInput, type UpdateConnectionInput } from '@/lib/api/connections'
import { toast } from 'sonner'

export function useConnections() {
  return useQuery({
    queryKey: ['connections'],
    queryFn: async () => {
      const res = await connectionsApi.list()
      // Backend returns { connections: [...], total: number }
      return Array.isArray(res.data?.connections) ? res.data.connections : Array.isArray(res.data) ? res.data : []
    },
  })
}

export function useConnection(platformId: string, connectionId: string) {
  return useQuery({
    queryKey: ['connections', platformId, connectionId],
    queryFn: async () => {
      const res = await connectionsApi.get(platformId, connectionId)
      return res.data
    },
    enabled: !!platformId && !!connectionId,
  })
}

export function usePlatforms() {
  return useQuery({
    queryKey: ['platforms'],
    queryFn: async () => {
      const res = await connectionsApi.listPlatforms()
      // Backend returns { platforms: [...], total: number } in res.data
      return Array.isArray(res.data?.platforms) ? res.data.platforms : []
    },
  })
}

export function useCreateConnection() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      platformId,
      input,
    }: {
      platformId: string
      input: Omit<CreateConnectionInput, 'platformId'>
    }) => connectionsApi.create(platformId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['connections'] })
      toast.success('Connection created')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create connection')
    },
  })
}

export function useDeleteConnection() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      platformId,
      connectionId,
    }: {
      platformId: string
      connectionId: string
    }) => connectionsApi.delete(platformId, connectionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['connections'] })
      toast.success('Connection deleted')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete connection')
    },
  })
}

export function useTestConnection() {
  return useMutation({
    mutationFn: ({
      platformId,
      connectionId,
    }: {
      platformId: string
      connectionId: string
    }) => connectionsApi.test(platformId, connectionId),
    onSuccess: () => {
      toast.success('Connection test passed')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Connection test failed')
    },
  })
}

export function useConnectionFields(platformId: string, connectionId: string) {
  return useQuery({
    queryKey: ['connection-fields', platformId, connectionId],
    queryFn: async () => {
      const res = await connectionsApi.listConnectionFields(platformId, connectionId)
      return res.data
    },
    enabled: !!platformId && !!connectionId,
    staleTime: 5 * 60 * 1000,
  })
}

export function useConnectionTags(platformId: string, connectionId: string) {
  return useQuery({
    queryKey: ['connection-tags', platformId, connectionId],
    queryFn: async () => {
      const res = await connectionsApi.listConnectionTags(platformId, connectionId)
      return Array.isArray(res.data) ? res.data : []
    },
    enabled: !!platformId && !!connectionId,
    staleTime: 5 * 60 * 1000,
  })
}

export function useStartOAuth() {
  return useMutation({
    mutationFn: (platformId: string) => connectionsApi.startOAuth(platformId),
  })
}
