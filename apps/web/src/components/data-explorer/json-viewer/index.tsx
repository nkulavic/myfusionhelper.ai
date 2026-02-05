'use client'

import React, { useState, useCallback } from 'react'
import {
  Copy,
  Check,
  ChevronsDown,
  ChevronsUp,
  FileJson,
  Search,
  X,
  Download,
} from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import { useDataExplorerStore } from '@/lib/stores/data-explorer-store'
import { JsonTree } from './json-tree'

// ---------------------------------------------------------------------------
// Props
// ---------------------------------------------------------------------------

interface JsonViewerProps {
  data: unknown
  title?: string
  className?: string
  showHeader?: boolean
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function JsonViewer({
  data,
  title = 'JSON Data',
  className,
  showHeader = true,
}: JsonViewerProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [copied, setCopied] = useState(false)
  const { expandedNodes, collapseAll } = useDataExplorerStore()

  // -- Search ---------------------------------------------------------------

  const handleSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchQuery(e.target.value)
    },
    []
  )

  const handleSearchClear = useCallback(() => {
    setSearchQuery('')
  }, [])

  // -- Copy entire JSON -----------------------------------------------------

  const handleCopy = useCallback(async () => {
    try {
      const jsonString = JSON.stringify(data, null, 2)
      await navigator.clipboard.writeText(jsonString)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (error) {
      console.error('Failed to copy JSON:', error)
    }
  }, [data])

  // -- Download as file -----------------------------------------------------

  const handleDownload = useCallback(() => {
    try {
      const jsonString = JSON.stringify(data, null, 2)
      const blob = new Blob([jsonString], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${title.toLowerCase().replace(/\s+/g, '-')}-${Date.now()}.json`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to download JSON:', error)
    }
  }, [data, title])

  // -- Expand / Collapse all ------------------------------------------------

  const handleExpandAll = useCallback(() => {
    const getAllPaths = (obj: unknown, currentPath: string[] = []): string[] => {
      if (obj === null || typeof obj !== 'object') return []
      const paths: string[] = []
      Object.entries(obj as Record<string, unknown>).forEach(([key, value]) => {
        const newPath = [...currentPath, key]
        paths.push(newPath.join('.'))
        if (value !== null && typeof value === 'object') {
          paths.push(...getAllPaths(value, newPath))
        }
      })
      return paths
    }

    const allPaths = getAllPaths(data)
    const store = useDataExplorerStore.getState()
    allPaths.forEach((path) => {
      if (!store.expandedNodes.includes(path)) {
        store.expandNode(path)
      }
    })
  }, [data])

  const handleCollapseAll = useCallback(() => {
    collapseAll()
  }, [collapseAll])

  // -- Data size label ------------------------------------------------------

  const dataSize = React.useMemo(() => {
    try {
      const jsonString = JSON.stringify(data)
      const bytes = new Blob([jsonString]).size
      if (bytes < 1024) return `${bytes} B`
      if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
      return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
    } catch {
      return 'Unknown'
    }
  }, [data])

  // -- Raw JSON string ------------------------------------------------------

  const rawJson = React.useMemo(() => {
    try {
      return JSON.stringify(data, null, 2)
    } catch {
      return 'Unable to serialize data'
    }
  }, [data])

  // -------------------------------------------------------------------------
  // Render
  // -------------------------------------------------------------------------

  return (
    <Card className={cn('flex flex-col h-full', className)}>
      {showHeader && (
        <>
          {/* Header */}
          <div className="p-4 space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <FileJson className="h-5 w-5 text-muted-foreground" />
                <h3 className="font-semibold text-lg">{title}</h3>
                <span className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
                  {dataSize}
                </span>
              </div>

              {/* Action buttons */}
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleExpandAll}
                  title="Expand all nodes"
                >
                  <ChevronsDown className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleCollapseAll}
                  title="Collapse all nodes"
                >
                  <ChevronsUp className="h-4 w-4" />
                </Button>
                <Separator orientation="vertical" className="h-6 mx-1" />
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleCopy}
                  title="Copy JSON"
                >
                  {copied ? (
                    <Check className="h-4 w-4 text-green-600" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleDownload}
                  title="Download JSON"
                >
                  <Download className="h-4 w-4" />
                </Button>
              </div>
            </div>

            {/* Search bar */}
            <div className="relative flex items-center gap-2">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={searchQuery}
                  onChange={handleSearchChange}
                  placeholder="Search in JSON..."
                  className="pl-9 pr-9"
                />
                {searchQuery && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="absolute right-1 top-1/2 h-7 w-7 -translate-y-1/2"
                    onClick={handleSearchClear}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                )}
              </div>
            </div>
          </div>

          <Separator />
        </>
      )}

      {/* Tabs: Tree / Raw */}
      <Tabs defaultValue="tree" className="flex-1 flex flex-col min-h-0">
        <TabsList className="mx-4 mt-3 w-fit">
          <TabsTrigger value="tree">Tree</TabsTrigger>
          <TabsTrigger value="raw">Raw</TabsTrigger>
        </TabsList>

        <TabsContent value="tree" className="flex-1 min-h-0">
          <JsonTree
            data={data}
            searchQuery={searchQuery}
            className="h-full"
          />
        </TabsContent>

        <TabsContent value="raw" className="flex-1 min-h-0">
          <ScrollArea className="h-full">
            <pre className="p-4 font-mono text-sm whitespace-pre-wrap break-all text-foreground">
              {rawJson}
            </pre>
          </ScrollArea>
        </TabsContent>
      </Tabs>
    </Card>
  )
}

// Re-export child components for individual use
export { JsonTree } from './json-tree'
export { JsonNode } from './json-node'
export { JsonValue, getValueType, isPrimitive } from './json-value'
