'use client'

import { useMemo, useCallback } from 'react'
import {
  Users,
  Tag,
  DollarSign,
  FileText,
  Layers,
  AlertCircle,
  RefreshCw,
  Loader2,
  Clock,
  CheckCircle2,
} from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import { useDataExplorerStore } from '@/lib/stores/data-explorer-store'
import { usePlatforms } from '@/lib/hooks/use-connections'
import type { PlatformDefinition } from '@/lib/api/connections'
import { PlatformLogo } from '@/components/platform-logo'
import type { CatalogObjectType } from '@/lib/api/data-explorer'
import { useDataCatalog, useTriggerSync } from '@/lib/hooks/use-data-explorer'

function getObjectTypeIcon(objectType: string) {
  const lower = objectType.toLowerCase()
  if (lower.includes('contact')) return Users
  if (lower.includes('tag')) return Tag
  if (lower.includes('deal') || lower.includes('opportunity')) return DollarSign
  return FileText
}

function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

export function ConnectionOverview() {
  const { selection, selectObjectType } = useDataExplorerStore()
  const { data: catalogData, isLoading: loading, error: queryError } = useDataCatalog()
  const triggerSync = useTriggerSync()

  const catalog = catalogData?.sources ?? []
  const error = queryError instanceof Error ? queryError.message : queryError ? 'Failed to load catalog' : null

  const connectionEntries = useMemo(
    () => catalog.filter((e) => e.connection_id === selection.connectionId),
    [catalog, selection.connectionId]
  )

  // Sync status derived from catalog entries for this connection
  const syncInfo = useMemo(() => {
    const entry = connectionEntries[0]
    if (!entry) return null
    return {
      status: entry.sync_status,
      lastSyncedAt: entry.last_synced_at,
    }
  }, [connectionEntries])

  const isSyncing = syncInfo?.status === 'syncing' || triggerSync.isPending

  const { data: allPlatforms } = usePlatforms()
  const platform = useMemo(
    () =>
      selection.platformId
        ? allPlatforms?.find((p: PlatformDefinition) => p.slug === selection.platformId || p.platformId === selection.platformId)
        : undefined,
    [selection.platformId, allPlatforms]
  )

  const handleSelectObjectType = useCallback(
    (entry: CatalogObjectType) => {
      if (
        !selection.platformId ||
        !selection.platformName ||
        !selection.connectionId ||
        !selection.connectionName
      )
        return
      selectObjectType(
        selection.platformId,
        selection.platformName,
        selection.connectionId,
        selection.connectionName,
        entry.object_type,
        entry.label
      )
    },
    [
      selection.platformId,
      selection.platformName,
      selection.connectionId,
      selection.connectionName,
      selectObjectType,
    ]
  )

  const handleSync = useCallback(() => {
    if (!selection.connectionId || isSyncing) return
    triggerSync.mutate(selection.connectionId)
  }, [selection.connectionId, isSyncing, triggerSync])

  if (loading) {
    return (
      <div className="p-6 space-y-6">
        <div className="flex items-center gap-3">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-6 w-24 rounded-full" />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-28 rounded-lg" />
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6 flex flex-col items-center justify-center gap-3 text-center">
        <AlertCircle className="h-10 w-10 text-destructive" />
        <p className="text-sm text-destructive font-medium">Error loading connection data</p>
        <p className="text-xs text-muted-foreground">{error}</p>
      </div>
    )
  }

  const accentColor = platform?.displayConfig?.color ?? '#6B7280'

  return (
    <ScrollArea className="h-full">
      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between flex-wrap gap-3">
          <div className="flex items-center gap-3 flex-wrap">
            <h2 className="text-2xl font-bold tracking-tight">
              {selection.connectionName ?? 'Connection'}
            </h2>
            {platform && (
              <Badge variant="secondary" className="flex items-center gap-1.5">
                <PlatformLogo definition={platform} size={20} />
                <span>{platform.name}</span>
              </Badge>
            )}
          </div>

          {/* Sync controls */}
          <div className="flex items-center gap-3">
            {/* Sync status */}
            {syncInfo?.lastSyncedAt && (
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                {isSyncing ? (
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                ) : (
                  <CheckCircle2 className="h-3.5 w-3.5 text-green-500" />
                )}
                <Clock className="h-3 w-3" />
                <span>{formatRelativeTime(syncInfo.lastSyncedAt)}</span>
              </div>
            )}

            {/* Sync button */}
            <Button
              variant="outline"
              size="sm"
              onClick={handleSync}
              disabled={isSyncing}
            >
              {isSyncing ? (
                <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />
              ) : (
                <RefreshCw className="mr-1.5 h-3.5 w-3.5" />
              )}
              {isSyncing ? 'Syncing...' : 'Sync Now'}
            </Button>
          </div>
        </div>

        {/* Object type cards */}
        {connectionEntries.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 gap-3 text-center">
            <Layers className="h-10 w-10 text-muted-foreground" />
            <p className="text-sm text-muted-foreground">
              No object types found for this connection.
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {connectionEntries.map((entry) => {
              const Icon = getObjectTypeIcon(entry.object_type)
              return (
                <Card
                  key={entry.object_type}
                  className={cn(
                    'relative overflow-hidden cursor-pointer',
                    'p-4 transition-all duration-200',
                    'hover:shadow-lg hover:-translate-y-0.5'
                  )}
                  onClick={() => handleSelectObjectType(entry)}
                >
                  <div
                    className="absolute left-0 top-0 bottom-0 w-1 rounded-l-lg"
                    style={{ backgroundColor: accentColor }}
                  />
                  <div className="pl-3 space-y-2">
                    <div className="flex items-center gap-2">
                      <Icon className="h-5 w-5 text-muted-foreground" />
                      <span className="font-medium">{entry.label}</span>
                    </div>
                    <p className="text-2xl font-bold tabular-nums">
                      {(entry.record_count ?? 0).toLocaleString()}
                    </p>
                    <div className="flex items-center justify-between">
                      <p className="text-xs text-muted-foreground">records</p>
                      {entry.last_synced_at && (
                        <p className="text-[10px] text-muted-foreground/70">
                          {formatRelativeTime(entry.last_synced_at)}
                        </p>
                      )}
                    </div>
                  </div>
                </Card>
              )
            })}
          </div>
        )}
      </div>
    </ScrollArea>
  )
}
