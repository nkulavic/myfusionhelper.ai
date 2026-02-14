'use client'

import React, { useState, useCallback } from 'react'
import { ChevronRight, ChevronDown, Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { JsonValue, getValueType, isPrimitive } from './json-value'
import { useDataExplorerStore } from '@/lib/stores/data-explorer-store'

interface JsonNodeProps {
  name: string
  value: unknown
  path: string[]
  depth?: number
  isLast?: boolean
  searchQuery?: string
}

export function JsonNode({
  name,
  value,
  path,
  depth = 0,
  isLast = false,
  searchQuery = '',
}: JsonNodeProps) {
  const nodeId = path.join('.')
  const { expandedNodes, toggleNodeExpansion } = useDataExplorerStore()

  const isExpanded = expandedNodes.includes(nodeId)
  const [copied, setCopied] = useState(false)

  const handleToggle = useCallback(() => {
    toggleNodeExpansion(nodeId)
  }, [nodeId, toggleNodeExpansion])

  const handleCopy = useCallback(
    async (e: React.MouseEvent) => {
      e.stopPropagation()
      try {
        await navigator.clipboard.writeText(JSON.stringify(value, null, 2))
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
      } catch (error) {
        console.error('Failed to copy:', error)
      }
    },
    [value]
  )

  const isObject = value !== null && typeof value === 'object'
  const isArray = Array.isArray(value)
  const primitive = isPrimitive(value)

  // Highlight search matches on key or value
  const isHighlighted = React.useMemo(() => {
    if (!searchQuery) return false
    const q = searchQuery.toLowerCase()
    return (
      name.toLowerCase().includes(q) ||
      String(value).toLowerCase().includes(q)
    )
  }, [name, value, searchQuery])

  const entries = React.useMemo(() => {
    if (!isObject || primitive) return []
    return Object.entries(value as Record<string, unknown>)
  }, [isObject, primitive, value])

  // Collapsed preview text, e.g. "Array(3)" or "{2 keys}"
  const preview = React.useMemo(() => {
    if (primitive) return null
    if (isArray) return `Array(${(value as unknown[]).length})`
    if (isObject) return `{${entries.length} ${entries.length === 1 ? 'key' : 'keys'}}`
    return null
  }, [primitive, isArray, isObject, value, entries])

  return (
    <div className={cn('relative', depth > 0 && 'ml-4')}>
      <div
        className={cn(
          'group flex items-center gap-2 py-1 rounded hover:bg-accent/50 transition-colors',
          isHighlighted && 'bg-yellow-100 dark:bg-yellow-900/20'
        )}
      >
        {/* Expand / Collapse toggle */}
        {!primitive ? (
          <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5 p-0 hover:bg-accent"
            onClick={handleToggle}
          >
            {isExpanded ? (
              <ChevronDown className="h-3 w-3" />
            ) : (
              <ChevronRight className="h-3 w-3" />
            )}
          </Button>
        ) : (
          <div className="w-5" />
        )}

        {/* Property name */}
        <span className="font-mono text-sm font-medium text-foreground">
          {name}:
        </span>

        {/* Inline value or collapsed preview */}
        {primitive ? (
          <JsonValue
            value={value}
            type={getValueType(value)}
            isHighlighted={isHighlighted}
          />
        ) : (
          <span className="font-mono text-sm text-muted-foreground">{preview}</span>
        )}

        {/* Copy button (visible on hover) */}
        <Button
          variant="ghost"
          size="icon"
          className="h-5 w-5 p-0 ml-auto opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={handleCopy}
        >
          {copied ? (
            <Check className="h-3 w-3 text-success" />
          ) : (
            <Copy className="h-3 w-3" />
          )}
        </Button>
      </div>

      {/* Nested children */}
      {!primitive && isExpanded && (
        <div className="border-l border-border ml-2 pl-2">
          {entries.map(([key, val], index) => (
            <JsonNode
              key={`${nodeId}.${key}`}
              name={key}
              value={val}
              path={[...path, key]}
              depth={depth + 1}
              isLast={index === entries.length - 1}
              searchQuery={searchQuery}
            />
          ))}
        </div>
      )}
    </div>
  )
}
