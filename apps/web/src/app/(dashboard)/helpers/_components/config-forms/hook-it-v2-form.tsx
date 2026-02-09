'use client'

// Hook It V2 (Tag Router)
// Go keys: event_tag_map (object), event_field (string), default_tag_id (string)

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FormTextField, DynamicList, AddItemRow, InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

interface EventTagMapping {
  event: string
  tagId: string
}

export function HookItV2Form({ config, onChange, disabled }: ConfigFormProps) {
  const eventField = (config.eventField as string) || 'event'
  const defaultTagId = (config.defaultTagId as string) || ''

  // Convert object to array for DynamicList
  const eventTagMap = (config.eventTagMap as Record<string, string>) || {}
  const mappings: EventTagMapping[] = Object.entries(eventTagMap).map(([event, tagId]) => ({
    event,
    tagId,
  }))

  const [newEvent, setNewEvent] = useState('')
  const [newTagId, setNewTagId] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const handleMappingsChange = (items: EventTagMapping[]) => {
    const newMap: Record<string, string> = {}
    for (const item of items) {
      newMap[item.event] = item.tagId
    }
    updateConfig({ eventTagMap: newMap })
  }

  return (
    <div className="space-y-4">
      <InfoBanner>
        Routes webhook events to apply different tags based on event type. For example: "order.created" → apply "New Order" tag, "order.canceled" → apply "Canceled Order" tag.
      </InfoBanner>

      <DynamicList<EventTagMapping>
        label="Event → Tag Mappings"
        description="Map event names to tag IDs that should be applied when that event occurs."
        items={mappings}
        onItemsChange={handleMappingsChange}
        disabled={disabled}
        renderAddForm={(onAdd) => (
          <AddItemRow
            onAdd={() => {
              const event = newEvent.trim()
              const tagId = newTagId.trim()
              if (!event || !tagId) return
              onAdd({ event, tagId })
              setNewEvent('')
              setNewTagId('')
            }}
            disabled={disabled}
            canAdd={!!newEvent.trim() && !!newTagId.trim()}
          >
            <Input
              placeholder="Event name (e.g., order.created)"
              value={newEvent}
              onChange={(e) => setNewEvent(e.target.value)}
              disabled={disabled}
              className="flex-1"
            />
            <Input
              placeholder="Tag ID"
              value={newTagId}
              onChange={(e) => setNewTagId(e.target.value)}
              disabled={disabled}
              className="flex-1"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  const event = newEvent.trim()
                  const tagId = newTagId.trim()
                  if (!event || !tagId) return
                  onAdd({ event, tagId })
                  setNewEvent('')
                  setNewTagId('')
                }
              }}
            />
          </AddItemRow>
        )}
        renderItem={(entry) => (
          <span className="font-mono text-xs">
            <span className="font-medium">{entry.event}</span>{' '}
            <span className="text-muted-foreground">→ Tag: {entry.tagId}</span>
          </span>
        )}
      />

      <FormTextField
        label="Event Field Name"
        placeholder="event"
        value={eventField}
        onChange={(e) => updateConfig({ eventField: e.target.value })}
        disabled={disabled}
        description="Field name in webhook payload containing the event type (defaults to 'event')."
      />

      <FormTextField
        label="Default Tag ID (Optional)"
        placeholder="Leave empty for no default"
        value={defaultTagId}
        onChange={(e) => updateConfig({ defaultTagId: e.target.value })}
        disabled={disabled}
        description="Tag ID to apply if event type doesn't match any mapping above."
      />
    </div>
  )
}
