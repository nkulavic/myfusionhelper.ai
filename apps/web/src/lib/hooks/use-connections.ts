import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { connectionsApi, type CreateConnectionInput, type UpdateConnectionInput } from '@/lib/api/connections'

export function useConnections() {
  return useQuery({
    queryKey: ['connections'],
    queryFn: async () => {
      const res = await connectionsApi.list()
      return res.data ?? []
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
      return res.data ?? []
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
  })
}

export function useStartOAuth() {
  return useMutation({
    mutationFn: (platformId: string) => connectionsApi.startOAuth(platformId),
  })
}
