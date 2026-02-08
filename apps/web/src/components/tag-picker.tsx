'use client'

import { useState, useMemo } from 'react'
import { Check, ChevronsUpDown, X } from 'lucide-react'
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
import { useConnectionTags } from '@/lib/hooks/use-connections'
import type { ConnectionTag } from '@/lib/api/connections'

interface TagPickerProps {
  platformId: string
  connectionId: string
  value: string[]
  onChange: (value: string[]) => void
  placeholder?: string
  disabled?: boolean
}

export function TagPicker({
  platformId,
  connectionId,
  value,
  onChange,
  placeholder = 'Select tags...',
  disabled = false,
}: TagPickerProps) {
  const [open, setOpen] = useState(false)

  const { data: tags, isLoading } = useConnectionTags(platformId, connectionId)

  const grouped = useMemo(() => {
    const map = new Map<string, ConnectionTag[]>()
    if (!tags) return map
    for (const tag of tags) {
      const group = tag.category || 'Tags'
      const existing = map.get(group) ?? []
      existing.push(tag)
      map.set(group, existing)
    }
    return map
  }, [tags])

  const selectedNames = useMemo(() => {
    if (!tags) return []
    return value
      .map((id) => tags.find((t) => t.id === id)?.name ?? id)
  }, [value, tags])

  function handleToggle(tagId: string) {
    if (value.includes(tagId)) {
      onChange(value.filter((id) => id !== tagId))
    } else {
      onChange([...value, tagId])
    }
  }

  function handleRemove(tagId: string) {
    onChange(value.filter((id) => id !== tagId))
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
    <div className="space-y-2">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className={cn(
              'w-full justify-between',
              value.length === 0 && 'text-muted-foreground'
            )}
            disabled={disabled}
          >
            <span className="truncate">
              {value.length > 0
                ? `${value.length} tag${value.length !== 1 ? 's' : ''} selected`
                : placeholder}
            </span>
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[--radix-popover-trigger-width] p-0" align="start">
          <Command>
            <CommandInput placeholder="Search tags..." />
            <CommandList>
              <CommandEmpty>No tags found.</CommandEmpty>
              {Array.from(grouped.entries()).map(([groupName, groupTags]) => (
                <CommandGroup key={groupName} heading={groupName}>
                  {groupTags.map((tag) => {
                    const isSelected = value.includes(tag.id)
                    return (
                      <CommandItem
                        key={tag.id}
                        value={tag.name}
                        onSelect={() => handleToggle(tag.id)}
                      >
                        <Check
                          className={cn(
                            'mr-2 h-4 w-4',
                            isSelected ? 'opacity-100' : 'opacity-0'
                          )}
                        />
                        <span>{tag.name}</span>
                      </CommandItem>
                    )
                  })}
                </CommandGroup>
              ))}
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      {value.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {value.map((tagId) => {
            const tag = tags?.find((t) => t.id === tagId)
            return (
              <Badge key={tagId} variant="secondary" className="gap-1">
                {tag?.name ?? tagId}
                <button
                  type="button"
                  className="ml-0.5 rounded-full outline-none hover:bg-secondary-foreground/20"
                  onClick={() => handleRemove(tagId)}
                >
                  <X className="h-3 w-3" />
                </button>
              </Badge>
            )
          })}
        </div>
      )}
    </div>
  )
}
