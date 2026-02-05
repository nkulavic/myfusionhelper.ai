'use client'

import React from 'react'
import { JsonNode } from './json-node'
import { ScrollArea } from '@/components/ui/scroll-area'

interface JsonTreeProps {
  data: unknown
  searchQuery?: string
  className?: string
}

export function JsonTree({ data, searchQuery = '', className }: JsonTreeProps) {
  const isObject = React.useMemo(() => {
    return data !== null && typeof data === 'object'
  }, [data])

  const entries = React.useMemo(() => {
    if (!isObject) return []
    return Object.entries(data as Record<string, unknown>)
  }, [isObject, data])

  if (!isObject) {
    return (
      <div className="p-4 text-sm text-muted-foreground">
        Data must be an object or array to display as a tree.
      </div>
    )
  }

  return (
    <ScrollArea className={className}>
      <div className="p-4 font-mono text-sm">
        {entries.map(([key, value], index) => (
          <JsonNode
            key={key}
            name={key}
            value={value}
            path={[key]}
            depth={0}
            isLast={index === entries.length - 1}
            searchQuery={searchQuery}
          />
        ))}
      </div>
    </ScrollArea>
  )
}
