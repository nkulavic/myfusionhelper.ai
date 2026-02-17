'use client'

import React, { useState, useEffect, useMemo, useCallback } from 'react'
import {
  Search,
  X,
  ChevronRight,
  ChevronDown,
  ChevronsDownUp,
  ChevronsUpDown,
  Database,
  Loader2,
  Users,
  Tag,
  SlidersHorizontal,
  Target,
  Handshake,
  ShoppingCart,
  Package,
  Megaphone,
  Mail,
  GitBranch,
  Calendar,
  Workflow,
  FileText,
  MessageCircle,
  List,
  Zap,
  Layers,
  CheckSquare,
  Layout,
  Building,
  Ticket,
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { useDataExplorerStore } from '@/lib/stores/data-explorer-store'
import { PlatformLogo } from '@/components/platform-logo'
import { usePlatforms } from '@/lib/hooks/use-connections'
import type { PlatformDefinition } from '@/lib/api/connections'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface CatalogObjectType {
  objectType: string
  label: string
  icon: string
  recordCount?: number
  connectionId: string
  connectionName: string
  platformId: string
  platformName: string
}

interface ObjectTypeNode {
  objectType: string
  label: string
  icon: string
  recordCount?: number
}

interface ConnectionNode {
  connectionId: string
  connectionName: string
  platformId: string
  platformName: string
  objectTypes: ObjectTypeNode[]
}

interface PlatformNode {
  platformId: string
  platformName: string
  connections: ConnectionNode[]
}

// ---------------------------------------------------------------------------
// Icon mapping
// ---------------------------------------------------------------------------

const iconMap: Record<string, React.ComponentType<{ className?: string }>> = {
  users: Users,
  tag: Tag,
  sliders: SlidersHorizontal,
  target: Target,
  handshake: Handshake,
  'shopping-cart': ShoppingCart,
  package: Package,
  megaphone: Megaphone,
  mail: Mail,
  'git-branch': GitBranch,
  calendar: Calendar,
  workflow: Workflow,
  'file-text': FileText,
  'message-circle': MessageCircle,
  list: List,
  zap: Zap,
  layers: Layers,
  'check-square': CheckSquare,
  layout: Layout,
  building: Building,
  ticket: Ticket,
}

function getIconComponent(iconName: string): React.ComponentType<{ className?: string }> {
  return iconMap[iconName] ?? Database
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function formatRecordCount(count: number): string {
  return count.toLocaleString('en-US')
}

function buildTree(sources: CatalogObjectType[]): PlatformNode[] {
  const platformMap = new Map<string, PlatformNode>()

  for (const source of sources) {
    let platform = platformMap.get(source.platformId)
    if (!platform) {
      platform = {
        platformId: source.platformId,
        platformName: source.platformName,
        connections: [],
      }
      platformMap.set(source.platformId, platform)
    }

    let connection = platform.connections.find(
      (c) => c.connectionId === source.connectionId
    )
    if (!connection) {
      connection = {
        connectionId: source.connectionId,
        connectionName: source.connectionName,
        platformId: source.platformId,
        platformName: source.platformName,
        objectTypes: [],
      }
      platform.connections.push(connection)
    }

    connection.objectTypes.push({
      objectType: source.objectType,
      label: source.label,
      icon: source.icon,
      recordCount: source.recordCount,
    })
  }

  return Array.from(platformMap.values())
}

function filterTree(tree: PlatformNode[], query: string): PlatformNode[] {
  if (!query.trim()) return tree

  const lowerQuery = query.toLowerCase()

  return tree
    .map((platform) => {
      const platformMatches = platform.platformName
        .toLowerCase()
        .includes(lowerQuery)

      const filteredConnections = platform.connections
        .map((connection) => {
          const connectionMatches = connection.connectionName
            .toLowerCase()
            .includes(lowerQuery)

          const filteredObjectTypes = connection.objectTypes.filter((ot) =>
            ot.label.toLowerCase().includes(lowerQuery)
          )

          // Include all object types if platform or connection matches
          if (platformMatches || connectionMatches) {
            return connection
          }

          // Otherwise only include matching object types
          if (filteredObjectTypes.length > 0) {
            return { ...connection, objectTypes: filteredObjectTypes }
          }

          return null
        })
        .filter(Boolean) as ConnectionNode[]

      if (platformMatches && filteredConnections.length === 0) {
        // Platform name matches but no connections matched - show all connections
        return platform
      }

      if (filteredConnections.length > 0) {
        return { ...platform, connections: filteredConnections }
      }

      return null
    })
    .filter(Boolean) as PlatformNode[]
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function HierarchicalNav() {
  const [catalog, setCatalog] = useState<CatalogObjectType[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const {
    selection,
    expandedNodes,
    searchQuery,
    toggleNodeExpansion,
    expandNode,
    collapseAll,
    selectPlatform,
    selectConnection,
    selectObjectType,
    setSearchQuery,
  } = useDataExplorerStore()

  // -- Fetch catalog on mount -------------------------------------------------

  useEffect(() => {
    let cancelled = false

    async function fetchCatalog() {
      try {
        setLoading(true)
        setError(null)
        const res = await fetch('/api/data/catalog')
        if (!res.ok) throw new Error(`Failed to fetch catalog: ${res.status}`)
        const data = await res.json()
        if (!cancelled) {
          setCatalog(data.sources ?? [])
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load catalog')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    fetchCatalog()
    return () => {
      cancelled = true
    }
  }, [])

  // -- Build and filter tree --------------------------------------------------

  const tree = useMemo(() => buildTree(catalog), [catalog])

  const filteredTree = useMemo(
    () => filterTree(tree, searchQuery),
    [tree, searchQuery]
  )

  // -- Collect all node IDs for expand all ------------------------------------

  const allNodeIds = useMemo(() => {
    const ids: string[] = []
    for (const platform of filteredTree) {
      ids.push(`platform:${platform.platformId}`)
      for (const connection of platform.connections) {
        ids.push(`connection:${connection.connectionId}`)
      }
    }
    return ids
  }, [filteredTree])

  // -- Handlers ---------------------------------------------------------------

  const handleExpandAll = useCallback(() => {
    for (const nodeId of allNodeIds) {
      expandNode(nodeId)
    }
  }, [allNodeIds, expandNode])

  const handleCollapseAll = useCallback(() => {
    collapseAll()
  }, [collapseAll])

  const handleSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchQuery(e.target.value)
    },
    [setSearchQuery]
  )

  const handleSearchClear = useCallback(() => {
    setSearchQuery('')
  }, [setSearchQuery])

  const handlePlatformClick = useCallback(
    (platform: PlatformNode) => {
      const nodeId = `platform:${platform.platformId}`
      toggleNodeExpansion(nodeId)
      selectPlatform(platform.platformId, platform.platformName)
    },
    [toggleNodeExpansion, selectPlatform]
  )

  const handleConnectionClick = useCallback(
    (connection: ConnectionNode) => {
      const nodeId = `connection:${connection.connectionId}`
      toggleNodeExpansion(nodeId)
      selectConnection(
        connection.platformId,
        connection.platformName,
        connection.connectionId,
        connection.connectionName
      )
    },
    [toggleNodeExpansion, selectConnection]
  )

  const handleObjectTypeClick = useCallback(
    (connection: ConnectionNode, objectType: ObjectTypeNode) => {
      selectObjectType(
        connection.platformId,
        connection.platformName,
        connection.connectionId,
        connection.connectionName,
        objectType.objectType,
        objectType.label
      )
    },
    [selectObjectType]
  )

  // -- Active state helpers ---------------------------------------------------

  const isPlatformActive = useCallback(
    (platformId: string) => {
      return (
        selection.level === 'platform' &&
        selection.platformId === platformId
      )
    },
    [selection]
  )

  const isConnectionActive = useCallback(
    (connectionId: string) => {
      return (
        selection.level === 'connection' &&
        selection.connectionId === connectionId
      )
    },
    [selection]
  )

  const isObjectTypeActive = useCallback(
    (connectionId: string, objectType: string) => {
      return (
        selection.level === 'objectType' &&
        selection.connectionId === connectionId &&
        selection.objectType === objectType
      )
    },
    [selection]
  )

  // ---------------------------------------------------------------------------
  // Render
  // ---------------------------------------------------------------------------

  return (
    <div className="flex flex-col h-full">
      {/* Header: Search + Expand/Collapse */}
      <div className="p-3 space-y-2 border-b">
        {/* Search input */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={searchQuery}
            onChange={handleSearchChange}
            placeholder="Search objects..."
            className="pl-9 pr-9 h-9"
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

        {/* Expand / Collapse buttons */}
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            className="h-7 px-2 text-xs text-muted-foreground"
            onClick={handleExpandAll}
          >
            <ChevronsUpDown className="h-3.5 w-3.5 mr-1" />
            Expand All
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="h-7 px-2 text-xs text-muted-foreground"
            onClick={handleCollapseAll}
          >
            <ChevronsDownUp className="h-3.5 w-3.5 mr-1" />
            Collapse All
          </Button>
        </div>
      </div>

      {/* Tree content */}
      <ScrollArea className="flex-1">
        <div className="py-1">
          {loading && (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
              <span className="ml-2 text-sm text-muted-foreground">
                Loading catalog...
              </span>
            </div>
          )}

          {error && (
            <div className="px-4 py-8 text-center">
              <p className="text-sm text-destructive">{error}</p>
              <Button
                variant="ghost"
                size="sm"
                className="mt-2"
                onClick={() => window.location.reload()}
              >
                Retry
              </Button>
            </div>
          )}

          {!loading && !error && filteredTree.length === 0 && (
            <div className="px-4 py-8 text-center">
              <Database className="h-8 w-8 mx-auto mb-2 text-muted-foreground/50" />
              <p className="text-sm text-muted-foreground">
                {searchQuery
                  ? 'No results matching your search'
                  : 'No data sources available'}
              </p>
            </div>
          )}

          {!loading &&
            !error &&
            filteredTree.map((platform) => (
              <PlatformRow
                key={platform.platformId}
                platform={platform}
                isExpanded={expandedNodes.includes(
                  `platform:${platform.platformId}`
                )}
                isActive={isPlatformActive(platform.platformId)}
                expandedNodes={expandedNodes}
                isConnectionActive={isConnectionActive}
                isObjectTypeActive={isObjectTypeActive}
                onPlatformClick={handlePlatformClick}
                onConnectionClick={handleConnectionClick}
                onObjectTypeClick={handleObjectTypeClick}
              />
            ))}
        </div>
      </ScrollArea>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Platform Row
// ---------------------------------------------------------------------------

interface PlatformRowProps {
  platform: PlatformNode
  isExpanded: boolean
  isActive: boolean
  expandedNodes: string[]
  isConnectionActive: (connectionId: string) => boolean
  isObjectTypeActive: (connectionId: string, objectType: string) => boolean
  onPlatformClick: (platform: PlatformNode) => void
  onConnectionClick: (connection: ConnectionNode) => void
  onObjectTypeClick: (connection: ConnectionNode, objectType: ObjectTypeNode) => void
}

const PlatformRow = React.memo(function PlatformRow({
  platform,
  isExpanded,
  isActive,
  expandedNodes,
  isConnectionActive,
  isObjectTypeActive,
  onPlatformClick,
  onConnectionClick,
  onObjectTypeClick,
}: PlatformRowProps) {
  const { data: allPlatforms } = usePlatforms()
  const apiPlatform = allPlatforms?.find(
    (p: PlatformDefinition) => p.slug === platform.platformId || p.platformId === platform.platformId
  )

  return (
    <div>
      {/* Platform trigger row */}
      <button
        type="button"
        className={cn(
          'flex items-center w-full gap-2 px-3 py-2 text-left text-sm font-medium',
          'hover:bg-accent/30 transition-colors cursor-pointer',
          isActive &&
            'bg-accent/50 text-accent-foreground border-l-2 border-accent'
        )}
        onClick={() => onPlatformClick(platform)}
      >
        {isExpanded ? (
          <ChevronDown className="h-4 w-4 shrink-0 text-muted-foreground" />
        ) : (
          <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
        )}

        {apiPlatform ? (
          <PlatformLogo definition={apiPlatform} size={24} />
        ) : (
          <div
            className="flex items-center justify-center rounded-md bg-muted text-muted-foreground font-bold"
            style={{ width: 24, height: 24, fontSize: 10 }}
          >
            {platform.platformName.charAt(0).toUpperCase()}
          </div>
        )}

        <span className="truncate flex-1">{platform.platformName}</span>

        <Badge
          variant="secondary"
          className="ml-auto text-[10px] px-1.5 py-0 h-5 font-normal"
        >
          {platform.connections.length}
        </Badge>
      </button>

      {/* Connections */}
      {isExpanded && (
        <div>
          {platform.connections.map((connection) => (
            <ConnectionRow
              key={connection.connectionId}
              connection={connection}
              platformColor={apiPlatform?.displayConfig?.color ?? '#888'}
              isExpanded={expandedNodes.includes(
                `connection:${connection.connectionId}`
              )}
              isActive={isConnectionActive(connection.connectionId)}
              isObjectTypeActive={isObjectTypeActive}
              onConnectionClick={onConnectionClick}
              onObjectTypeClick={onObjectTypeClick}
            />
          ))}
        </div>
      )}
    </div>
  )
})

// ---------------------------------------------------------------------------
// Connection Row
// ---------------------------------------------------------------------------

interface ConnectionRowProps {
  connection: ConnectionNode
  platformColor: string
  isExpanded: boolean
  isActive: boolean
  isObjectTypeActive: (connectionId: string, objectType: string) => boolean
  onConnectionClick: (connection: ConnectionNode) => void
  onObjectTypeClick: (connection: ConnectionNode, objectType: ObjectTypeNode) => void
}

const ConnectionRow = React.memo(function ConnectionRow({
  connection,
  platformColor,
  isExpanded,
  isActive,
  isObjectTypeActive,
  onConnectionClick,
  onObjectTypeClick,
}: ConnectionRowProps) {
  return (
    <div>
      {/* Connection trigger row */}
      <button
        type="button"
        className={cn(
          'flex items-center w-full gap-2 pl-7 pr-3 py-1.5 text-left text-sm',
          'hover:bg-accent/30 transition-colors cursor-pointer',
          isActive &&
            'bg-accent/50 text-accent-foreground border-l-2 border-accent'
        )}
        onClick={() => onConnectionClick(connection)}
      >
        {isExpanded ? (
          <ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        ) : (
          <ChevronRight className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        )}

        <span
          className="shrink-0 rounded-full"
          style={{
            width: 8,
            height: 8,
            backgroundColor: platformColor,
          }}
        />

        <span className="truncate flex-1">{connection.connectionName}</span>
      </button>

      {/* Object types */}
      {isExpanded && (
        <div>
          {connection.objectTypes.map((objectType) => (
            <ObjectTypeRow
              key={objectType.objectType}
              objectType={objectType}
              connectionId={connection.connectionId}
              isActive={isObjectTypeActive(
                connection.connectionId,
                objectType.objectType
              )}
              onClick={() => onObjectTypeClick(connection, objectType)}
            />
          ))}
        </div>
      )}
    </div>
  )
})

// ---------------------------------------------------------------------------
// Object Type Row (leaf)
// ---------------------------------------------------------------------------

interface ObjectTypeRowProps {
  objectType: ObjectTypeNode
  connectionId: string
  isActive: boolean
  onClick: () => void
}

const ObjectTypeRow = React.memo(function ObjectTypeRow({
  objectType,
  isActive,
  onClick,
}: ObjectTypeRowProps) {
  const IconComponent = getIconComponent(objectType.icon)

  return (
    <button
      type="button"
      className={cn(
        'flex items-center w-full gap-2 pl-14 pr-3 py-1.5 text-left text-sm',
        'hover:bg-accent/30 transition-colors cursor-pointer',
        isActive &&
          'bg-accent/50 text-accent-foreground border-l-2 border-accent'
      )}
      onClick={onClick}
    >
      <IconComponent className="h-4 w-4 shrink-0 text-muted-foreground" />
      <span className="truncate flex-1">{objectType.label}</span>
      {objectType.recordCount != null && (
        <span className="text-xs text-muted-foreground tabular-nums ml-auto">
          {formatRecordCount(objectType.recordCount)}
        </span>
      )}
    </button>
  )
})
