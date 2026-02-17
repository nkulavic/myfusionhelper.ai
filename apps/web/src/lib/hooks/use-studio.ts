import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { studioApi, mapDashboard } from '@/lib/api/studio'
import type { Widget, Dashboard } from '@/lib/stores/studio-store'
import { toast } from 'sonner'

// ---------------------------------------------------------------------------
// Dashboard queries
// ---------------------------------------------------------------------------

export function useDashboards() {
  return useQuery<Dashboard[]>({
    queryKey: ['studio-dashboards'],
    queryFn: async () => {
      const res = await studioApi.listDashboards()
      const dashboards = res.data?.dashboards || []
      return dashboards.map(mapDashboard)
    },
  })
}

export function useDashboard(id: string | null) {
  return useQuery<Dashboard>({
    queryKey: ['studio-dashboard', id],
    queryFn: async () => {
      const res = await studioApi.getDashboard(id!)
      return mapDashboard(res.data!)
    },
    enabled: !!id,
  })
}

// ---------------------------------------------------------------------------
// Dashboard mutations
// ---------------------------------------------------------------------------

export function useCreateDashboard() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: { name: string; description?: string }) =>
      studioApi.createDashboard(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio-dashboards'] })
    },
    onError: (err) => {
      toast.error('Failed to create dashboard', {
        description: err instanceof Error ? err.message : 'An error occurred',
      })
    },
  })
}

export function useUpdateDashboard() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string
      data: { name?: string; description?: string; widgets?: Widget[] }
    }) => studioApi.updateDashboard(id, data),
    onMutate: async ({ id, data }) => {
      // Optimistic update for the single dashboard query
      await queryClient.cancelQueries({ queryKey: ['studio-dashboard', id] })
      const prev = queryClient.getQueryData<Dashboard>(['studio-dashboard', id])
      if (prev) {
        queryClient.setQueryData<Dashboard>(['studio-dashboard', id], {
          ...prev,
          ...data,
          updatedAt: new Date().toISOString(),
        })
      }
      return { prev }
    },
    onError: (err, { id }, context) => {
      if (context?.prev) {
        queryClient.setQueryData(['studio-dashboard', id], context.prev)
      }
      toast.error('Failed to update dashboard', {
        description: err instanceof Error ? err.message : 'An error occurred',
      })
    },
    onSettled: (_, __, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['studio-dashboard', id] })
      queryClient.invalidateQueries({ queryKey: ['studio-dashboards'] })
    },
  })
}

export function useDeleteDashboard() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => studioApi.deleteDashboard(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio-dashboards'] })
      toast.success('Dashboard deleted')
    },
    onError: (err) => {
      toast.error('Failed to delete dashboard', {
        description: err instanceof Error ? err.message : 'An error occurred',
      })
    },
  })
}

// ---------------------------------------------------------------------------
// Template queries & mutations
// ---------------------------------------------------------------------------

export function useTemplates() {
  return useQuery({
    queryKey: ['studio-templates'],
    queryFn: async () => {
      const res = await studioApi.listTemplates()
      return res.data?.templates || []
    },
  })
}

export function useApplyTemplate() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: { templateId: string; connectionId: string; name?: string }) =>
      studioApi.applyTemplate(data.templateId, {
        connectionId: data.connectionId,
        name: data.name,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio-dashboards'] })
      toast.success('Dashboard created from template')
    },
    onError: (err) => {
      toast.error('Failed to apply template', {
        description: err instanceof Error ? err.message : 'An error occurred',
      })
    },
  })
}
