'use client'

// Go keys: hook_action, goal_prefix, actions (array of {event, goal_name})

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FormTextField, DynamicList, AddItemRow } from './form-fields'
import type { ConfigFormProps } from './types'

interface ActionEntry {
  event: string
  goalName: string
}

export function HookItForm({ config, onChange, disabled }: ConfigFormProps) {
  const hookAction = (config.hookAction as string) || ''
  const goalPrefix = (config.goalPrefix as string) || ''
  const actions = (config.actions as ActionEntry[]) || []
  const [newEvent, setNewEvent] = useState('')
  const [newGoalName, setNewGoalName] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Hook Action"
        placeholder="e.g. tag_applied"
        value={hookAction}
        onChange={(e) => updateConfig({ hookAction: e.target.value })}
        disabled={disabled}
        description="The action type this hook responds to."
      />

      <FormTextField
        label="Goal Prefix"
        placeholder="e.g. hook_"
        value={goalPrefix}
        onChange={(e) => updateConfig({ goalPrefix: e.target.value })}
        disabled={disabled}
        description="Prefix prepended to goal names when triggering."
      />

      <DynamicList<ActionEntry>
        label="Actions"
        description="Event and goal name pairs that define what this hook triggers."
        items={actions}
        onItemsChange={(items) => updateConfig({ actions: items })}
        disabled={disabled}
        renderAddForm={(onAdd) => (
          <AddItemRow
            onAdd={() => {
              const event = newEvent.trim()
              const goalName = newGoalName.trim()
              if (!event || !goalName) return
              onAdd({ event, goalName })
              setNewEvent('')
              setNewGoalName('')
            }}
            disabled={disabled}
            canAdd={!!newEvent.trim() && !!newGoalName.trim()}
          >
            <Input
              placeholder="Event name"
              value={newEvent}
              onChange={(e) => setNewEvent(e.target.value)}
              disabled={disabled}
              className="flex-1"
            />
            <Input
              placeholder="Goal name"
              value={newGoalName}
              onChange={(e) => setNewGoalName(e.target.value)}
              disabled={disabled}
              className="flex-1"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  const event = newEvent.trim()
                  const goalName = newGoalName.trim()
                  if (!event || !goalName) return
                  onAdd({ event, goalName })
                  setNewEvent('')
                  setNewGoalName('')
                }
              }}
            />
          </AddItemRow>
        )}
        renderItem={(entry) => (
          <span className="font-mono">
            <span className="font-medium">{entry.event}</span>{' '}
            <span className="text-muted-foreground">→ {entry.goalName}</span>
          </span>
        )}
      />
    </div>
  )
}
