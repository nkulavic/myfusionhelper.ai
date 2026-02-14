'use client'

import { useState, useMemo } from 'react'
import { Check, ChevronsUpDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { useConnectionFields } from '@/lib/hooks/use-connections'
import type { ConnectionField } from '@/lib/api/connections'

interface FieldPickerSingleProps {
  platformId: string
  connectionId: string
  value: string
  onChange: (value: string) => void
  multiple?: false
  filterType?: string
  placeholder?: string
  disabled?: boolean
}

interface FieldPickerMultiProps {
  platformId: string
  connectionId: string
  value: string[]
  onChange: (value: string[]) => void
  multiple: true
  filterType?: string
  placeholder?: string
  disabled?: boolean
}

type FieldPickerProps = FieldPickerSingleProps | FieldPickerMultiProps

export function FieldPicker(props: FieldPickerProps) {
  const {
    platformId,
    connectionId,
    filterType,
    placeholder = 'Select field...',
    disabled = false,
    multiple,
  } = props
  const [open, setOpen] = useState(false)

  const { data, isLoading } = useConnectionFields(platformId, connectionId)

  const allFields = useMemo(() => {
    if (!data) return []
    const combined = [
      ...(data.standardFields ?? []),
      ...(data.customFields ?? []),
    ]
    if (filterType) {
      return combined.filter(
        (f) => f.fieldType.toLowerCase() === filterType.toLowerCase()
      )
    }
    return combined
  }, [data, filterType])

  const grouped = useMemo(() => {
    const map = new Map<string, ConnectionField[]>()
    for (const field of allFields) {
      const group = field.groupName || 'Other'
      const existing = map.get(group) ?? []
      existing.push(field)
      map.set(group, existing)
    }
    return map
  }, [allFields])

  const selectedValues = multiple ? props.value : props.value ? [props.value] : []

  const selectedLabels = useMemo(() => {
    return selectedValues
      .map((key) => allFields.find((f) => f.key === key)?.label ?? key)
  }, [selectedValues, allFields])

  function handleSelect(fieldKey: string) {
    if (multiple) {
      const current = props.value
      if (current.includes(fieldKey)) {
        props.onChange(current.filter((k) => k !== fieldKey))
      } else {
        props.onChange([...current, fieldKey])
      }
    } else {
      props.onChange(fieldKey)
      setOpen(false)
    }
  }

  if (isLoading) {
    return <Skeleton className="h-10 w-full rounded-md" />
  }

  if (!platformId || !connectionId) {
    return (
      <Button
        variant="outline"
        className="w-full justify-between text-muted-foreground"
        disabled
      >
        Select a connection first
      </Button>
    )
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn(
            'w-full justify-between',
            selectedValues.length === 0 && 'text-muted-foreground'
          )}
          disabled={disabled}
        >
          <span className="truncate">
            {selectedLabels.length > 0
              ? multiple
                ? `${selectedLabels.length} field${selectedLabels.length !== 1 ? 's' : ''} selected`
                : selectedLabels[0]
              : placeholder}
          </span>
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[--radix-popover-trigger-width] p-0" align="start">
        <Command>
          <CommandInput placeholder="Search fields..." />
          <CommandList>
            <CommandEmpty>No fields found.</CommandEmpty>
            {Array.from(grouped.entries()).map(([groupName, fields]) => (
              <CommandGroup key={groupName} heading={groupName}>
                {fields.map((field) => {
                  const isSelected = selectedValues.includes(field.key)
                  return (
                    <CommandItem
                      key={field.key}
                      value={`${field.label} ${field.key}`}
                      onSelect={() => handleSelect(field.key)}
                    >
                      <Check
                        className={cn(
                          'mr-2 h-4 w-4',
                          isSelected ? 'opacity-100' : 'opacity-0'
                        )}
                      />
                      <span className="flex-1">{field.label}</span>
                      <Badge variant="outline" className="ml-2 text-[10px] px-1.5">
                        {field.fieldType}
                      </Badge>
                    </CommandItem>
                  )
                })}
              </CommandGroup>
            ))}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
