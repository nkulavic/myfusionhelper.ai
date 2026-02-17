'use client'

import { useCallback } from 'react'
import { ArrowLeft, AlertCircle, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useDataExplorerStore } from '@/lib/stores/data-explorer-store'
import { JsonViewer } from '@/components/data-explorer/json-viewer'
import { useDataRecord } from '@/lib/hooks/use-data-explorer'

export function RecordDetail() {
  const { selection, selectObjectType, selectConnection, selectPlatform } =
    useDataExplorerStore()

  const {
    data: recordData,
    isLoading: loading,
    error: queryError,
  } = useDataRecord(
    selection.connectionId,
    selection.objectType,
    selection.recordId,
  )

  const record = recordData?.record ?? null
  const error = queryError instanceof Error ? queryError.message : queryError ? 'Failed to load record' : null

  const handleBackToObjectType = useCallback(() => {
    if (
      !selection.platformId ||
      !selection.platformName ||
      !selection.connectionId ||
      !selection.connectionName ||
      !selection.objectType ||
      !selection.objectTypeLabel
    )
      return
    selectObjectType(
      selection.platformId,
      selection.platformName,
      selection.connectionId,
      selection.connectionName,
      selection.objectType,
      selection.objectTypeLabel
    )
  }, [
    selection.platformId,
    selection.platformName,
    selection.connectionId,
    selection.connectionName,
    selection.objectType,
    selection.objectTypeLabel,
    selectObjectType,
  ])

  if (loading) {
    return (
      <div className="p-6 space-y-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-6 w-72" />
        <Skeleton className="h-5 w-96" />
        <Skeleton className="h-[400px] rounded-lg" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6 flex flex-col items-center justify-center gap-3 text-center">
        <AlertCircle className="h-10 w-10 text-destructive" />
        <p className="text-sm text-destructive font-medium">Error loading record</p>
        <p className="text-xs text-muted-foreground">{error}</p>
        <Button variant="outline" size="sm" onClick={handleBackToObjectType}>
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to {selection.objectTypeLabel ?? 'list'}
        </Button>
      </div>
    )
  }

  return (
    <ScrollArea className="h-full">
      <div className="p-6 space-y-4">
        {/* Back button */}
        <Button
          variant="ghost"
          size="sm"
          className="text-muted-foreground hover:text-foreground -ml-2"
          onClick={handleBackToObjectType}
        >
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to {selection.objectTypeLabel ?? 'list'}
        </Button>

        {/* Record summary heading */}
        <h2 className="text-2xl font-bold tracking-tight">
          {selection.recordSummary ?? `Record ${selection.recordId}`}
        </h2>

        {/* Breadcrumb */}
        <nav className="flex items-center gap-1 text-sm text-muted-foreground flex-wrap">
          {selection.platformName && (
            <>
              <button
                className="hover:text-foreground transition-colors hover:underline"
                onClick={() => {
                  if (selection.platformId && selection.platformName) {
                    selectPlatform(selection.platformId, selection.platformName)
                  }
                }}
              >
                {selection.platformName}
              </button>
              <ChevronRight className="h-3 w-3" />
            </>
          )}
          {selection.connectionName && (
            <>
              <button
                className="hover:text-foreground transition-colors hover:underline"
                onClick={() => {
                  if (
                    selection.platformId &&
                    selection.platformName &&
                    selection.connectionId &&
                    selection.connectionName
                  ) {
                    selectConnection(
                      selection.platformId,
                      selection.platformName,
                      selection.connectionId,
                      selection.connectionName
                    )
                  }
                }}
              >
                {selection.connectionName}
              </button>
              <ChevronRight className="h-3 w-3" />
            </>
          )}
          {selection.objectTypeLabel && (
            <>
              <button
                className="hover:text-foreground transition-colors hover:underline"
                onClick={handleBackToObjectType}
              >
                {selection.objectTypeLabel}
              </button>
              <ChevronRight className="h-3 w-3" />
            </>
          )}
          <span className="text-foreground font-medium">
            {selection.recordSummary ?? selection.recordId}
          </span>
        </nav>

        {/* JSON Viewer */}
        <JsonViewer
          data={record}
          title={selection.recordSummary ?? `Record ${selection.recordId}`}
          className="min-h-[400px]"
        />
      </div>
    </ScrollArea>
  )
}
