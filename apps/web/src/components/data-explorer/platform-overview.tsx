'use client'

import { useMemo, useCallback } from 'react'
import { Link2, Layers, Hash, AlertCircle } from 'lucide-react'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useDataExplorerStore } from '@/lib/stores/data-explorer-store'
import { usePlatforms } from '@/lib/hooks/use-connections'
import type { PlatformDefinition } from '@/lib/api/connections'
import { PlatformLogo } from '@/components/platform-logo'
import { useDataCatalog } from '@/lib/hooks/use-data-explorer'

export function PlatformOverview() {
  const { selection, selectConnection } = useDataExplorerStore()
  const { data: catalogData, isLoading: loading, error: queryError } = useDataCatalog()

  const catalog = catalogData?.sources ?? []
  const error = queryError instanceof Error ? queryError.message : queryError ? 'Failed to load catalog' : null

  const platformEntries = useMemo(
    () => catalog.filter((e) => e.platform_id === selection.platformId),
    [catalog, selection.platformId]
  )

  const { data: allPlatforms } = usePlatforms()
  const platform = useMemo(
    () =>
      selection.platformId
        ? allPlatforms?.find((p: PlatformDefinition) => p.slug === selection.platformId || p.platformId === selection.platformId)
        : undefined,
    [selection.platformId, allPlatforms]
  )

  const stats = useMemo(() => {
    const connectionIds = new Set(platformEntries.map((e) => e.connection_id))
    const objectTypes = new Set(platformEntries.map((e) => e.object_type))
    const totalRecords = platformEntries.reduce((sum, e) => sum + (e.record_count ?? 0), 0)
    return {
      connections: connectionIds.size,
      objectTypes: objectTypes.size,
      totalRecords,
    }
  }, [platformEntries])

  const connections = useMemo(() => {
    const map = new Map<string, { id: string; name: string; objectTypeCount: number }>()
    for (const entry of platformEntries) {
      const existing = map.get(entry.connection_id)
      if (existing) {
        existing.objectTypeCount += 1
      } else {
        map.set(entry.connection_id, {
          id: entry.connection_id,
          name: entry.connection_name,
          objectTypeCount: 1,
        })
      }
    }
    return Array.from(map.values())
  }, [platformEntries])

  const handleExploreConnection = useCallback(
    (connectionId: string, connectionName: string) => {
      if (!selection.platformId || !selection.platformName) return
      selectConnection(
        selection.platformId,
        selection.platformName,
        connectionId,
        connectionName
      )
    },
    [selection.platformId, selection.platformName, selectConnection]
  )

  if (loading) {
    return (
      <div className="p-6 space-y-6">
        <div className="flex items-center gap-4">
          <Skeleton className="h-16 w-16 rounded-md" />
          <Skeleton className="h-8 w-48" />
        </div>
        <div className="grid grid-cols-2 gap-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-24 rounded-lg" />
          ))}
        </div>
        <Skeleton className="h-32 rounded-lg" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6 flex flex-col items-center justify-center gap-3 text-center">
        <AlertCircle className="h-10 w-10 text-destructive" />
        <p className="text-sm text-destructive font-medium">Error loading platform data</p>
        <p className="text-xs text-muted-foreground">{error}</p>
      </div>
    )
  }

  const accentColor = platform?.displayConfig?.color ?? '#6B7280'

  return (
    <ScrollArea className="h-full">
      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center gap-4">
          {platform && <PlatformLogo definition={platform} size={64} />}
          <h2 className="text-2xl font-bold tracking-tight">
            {selection.platformName ?? 'Platform'}
          </h2>
        </div>

        {/* Stats grid */}
        <div className="grid grid-cols-2 gap-4">
          <Card className="relative overflow-hidden">
            <div
              className="absolute left-0 top-0 bottom-0 w-1 rounded-l-lg"
              style={{ backgroundColor: accentColor }}
            />
            <CardHeader className="pb-2 pl-5">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Link2 className="h-4 w-4" />
                Total Connections
              </CardTitle>
            </CardHeader>
            <CardContent className="pl-5">
              <p className="text-3xl font-bold">{stats.connections}</p>
            </CardContent>
          </Card>

          <Card className="relative overflow-hidden">
            <div
              className="absolute left-0 top-0 bottom-0 w-1 rounded-l-lg"
              style={{ backgroundColor: accentColor }}
            />
            <CardHeader className="pb-2 pl-5">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Layers className="h-4 w-4" />
                Total Object Types
              </CardTitle>
            </CardHeader>
            <CardContent className="pl-5">
              <p className="text-3xl font-bold">{stats.objectTypes}</p>
            </CardContent>
          </Card>

          <Card className="relative overflow-hidden col-span-2">
            <div
              className="absolute left-0 top-0 bottom-0 w-1 rounded-l-lg"
              style={{ backgroundColor: accentColor }}
            />
            <CardHeader className="pb-2 pl-5">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Hash className="h-4 w-4" />
                Total Records
              </CardTitle>
            </CardHeader>
            <CardContent className="pl-5">
              <p className="text-3xl font-bold">
                {stats.totalRecords.toLocaleString()}
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Connection list */}
        <div className="space-y-3">
          <h3 className="text-lg font-semibold">Connections</h3>
          {connections.length === 0 ? (
            <p className="text-sm text-muted-foreground">No connections found for this platform.</p>
          ) : (
            <div className="space-y-2">
              {connections.map((conn) => (
                <Card
                  key={conn.id}
                  className="flex items-center justify-between p-4 hover:shadow-md transition-shadow"
                >
                  <div className="space-y-1">
                    <p className="font-medium">{conn.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {conn.objectTypeCount} object type{conn.objectTypeCount !== 1 ? 's' : ''}
                    </p>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleExploreConnection(conn.id, conn.name)}
                  >
                    Explore
                  </Button>
                </Card>
              ))}
            </div>
          )}
        </div>
      </div>
    </ScrollArea>
  )
}
