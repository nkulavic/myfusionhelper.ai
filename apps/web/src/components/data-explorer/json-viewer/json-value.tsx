'use client'

import React from 'react'
import { cn } from '@/lib/utils'

export type JsonValueType = 'string' | 'number' | 'boolean' | 'null' | 'undefined'

interface JsonValueProps {
  value: unknown
  type: JsonValueType
  isHighlighted?: boolean
  className?: string
}

const typeColors: Record<JsonValueType, string> = {
  string: 'text-green-600 dark:text-green-400',
  number: 'text-blue-600 dark:text-blue-400',
  boolean: 'text-purple-600 dark:text-purple-400',
  null: 'text-gray-500 dark:text-gray-400',
  undefined: 'text-gray-500 dark:text-gray-400',
}

export function JsonValue({ value, type, isHighlighted = false, className }: JsonValueProps) {
  const colorClass = typeColors[type]

  const formattedValue = React.useMemo(() => {
    if (type === 'string') return `"${value}"`
    if (type === 'null') return 'null'
    if (type === 'undefined') return 'undefined'
    return String(value)
  }, [value, type])

  return (
    <span
      className={cn(
        'font-mono text-sm',
        colorClass,
        isHighlighted && 'bg-yellow-200 dark:bg-yellow-900/40 rounded px-0.5',
        className
      )}
    >
      {formattedValue}
    </span>
  )
}

/** Determine the display type for a primitive value. */
export function getValueType(value: unknown): JsonValueType {
  if (value === null) return 'null'
  if (value === undefined) return 'undefined'
  const t = typeof value
  if (t === 'string') return 'string'
  if (t === 'number') return 'number'
  if (t === 'boolean') return 'boolean'
  return 'null'
}

/** Returns true when the value is a JSON primitive (not an object or array). */
export function isPrimitive(value: unknown): boolean {
  return (
    value === null ||
    value === undefined ||
    typeof value === 'string' ||
    typeof value === 'number' ||
    typeof value === 'boolean'
  )
}
