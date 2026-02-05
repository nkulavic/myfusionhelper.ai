'use client'

import { useEffect, useState, useMemo, useCallback } from 'react'
import {
  Users,
  Tag,
  DollarSign,
  FileText,
  Layers,
  AlertCircle,
} from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import { useDataExplorerStore } from '@/lib/stores/data-explorer-store'
import { getCRMPlatform } from '@/lib/crm-platforms'
import { PlatformLogo } from '@/components/platform-logo'

interface CatalogEntry {
  platformId: string
  platformName: string
  connectionId: string
  connectionName: string
  objectType: string
  objectTypeLabel: string
  recordCount: number
}

function getObjectTypeIcon(objectType: string) {
  const lower = objectType.toLowerCase()
  if (lower.includes('contact')) return Users
  if (lower.includes('tag')) return Tag
  if (lower.includes('deal') || lower.includes('opportunity')) return DollarSign
  return FileText
}

export function ConnectionOverview() {
  const { selection, selectObjectType } = useDataExplorerStore()
  const [catalog, setCatalog] = useState<CatalogEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function fetchCatalog() {
      setLoading(true)
      setError(null)
      try {
        const res = await fetch('/api/data/catalog')
        if (!res.ok) throw new Error(`Failed to fetch catalog: ${res.status}`)
        const data = await res.json()
        if (!cancelled) {
          setCatalog(data)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load catalog')
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    fetchCatalog()
    return () => { cancelled = true }
  }, [])

  const connectionEntries = useMemo(
    () => catalog.filter((e) => e.connectionId === selection.connectionId),
    [catalog, selection.connectionId]
  )

  const platform = useMemo(
    () => (selection.platformId ? getCRMPlatform(selection.platformId) : undefined),
    [selection.platformId]
  )

  const handleSelectObjectType = useCallback(
    (entry: CatalogEntry) => {
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
        entry.objectType,
        entry.objectTypeLabel
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

  const accentColor = platform?.color ?? '#6B7280'

  return (
    <ScrollArea className="h-full">
      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center gap-3 flex-wrap">
          <h2 className="text-2xl font-bold tracking-tight">
            {selection.connectionName ?? 'Connection'}
          </h2>
          {platform && (
            <Badge variant="secondary" className="flex items-center gap-1.5">
              <PlatformLogo platform={platform} size={20} />
              <span>{platform.name}</span>
            </Badge>
          )}
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
              const Icon = getObjectTypeIcon(entry.objectType)
              return (
                <Card
                  key={entry.objectType}
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
                      <span className="font-medium">{entry.objectTypeLabel}</span>
                    </div>
                    <p className="text-2xl font-bold tabular-nums">
                      {(entry.recordCount ?? 0).toLocaleString()}
                    </p>
                    <p className="text-xs text-muted-foreground">records</p>
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
