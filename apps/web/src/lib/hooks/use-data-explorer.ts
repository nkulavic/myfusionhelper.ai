import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  dataExplorerApi,
  type DataQueryRequest,
  type DataCatalogResponse,
  type DataQueryResponse,
  type RecordDetailResponse,
  type SchemaResponse,
} from '@/lib/api/data-explorer'
import { toast } from 'sonner'

export function useDataCatalog() {
  return useQuery<DataCatalogResponse>({
    queryKey: ['data-catalog'],
    queryFn: () => dataExplorerApi.getCatalog(),
  })
}

export function useDataQuery(params: DataQueryRequest | null) {
  return useQuery<DataQueryResponse>({
    queryKey: ['data-query', params],
    queryFn: () => dataExplorerApi.query(params!),
    enabled: !!params?.connectionId && !!params?.objectType,
  })
}

export function useDataRecord(
  connectionId: string | null,
  objectType: string | null,
  recordId: string | null,
) {
  return useQuery<RecordDetailResponse>({
    queryKey: ['data-record', connectionId, objectType, recordId],
    queryFn: () => dataExplorerApi.getRecord(connectionId!, objectType!, recordId!),
    enabled: !!connectionId && !!objectType && !!recordId,
  })
}

export function useDataSchema(connectionId: string | null, objectType: string | null) {
  return useQuery<SchemaResponse>({
    queryKey: ['data-schema', connectionId, objectType],
    queryFn: () => dataExplorerApi.getSchema(connectionId!, objectType!),
    enabled: !!connectionId && !!objectType,
  })
}

export function useTriggerSync() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (connectionId: string) => dataExplorerApi.triggerSync(connectionId),
    onSuccess: () => {
      toast.success('Sync triggered', {
        description: 'Data sync has been queued and will begin shortly.',
      })
      // Invalidate catalog so sync_status updates on next fetch
      queryClient.invalidateQueries({ queryKey: ['data-catalog'] })
    },
    onError: (err) => {
      toast.error('Failed to trigger sync', {
        description: err instanceof Error ? err.message : 'An error occurred',
      })
    },
  })
}
