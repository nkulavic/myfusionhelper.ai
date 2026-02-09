'use client'

// Hook It V4 (Multi-Goal)
// Go keys: goals (array of strings), integration (string), stop_on_error (boolean), goal_prefix (string), goal_suffix (string)

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FormTextField, FormSwitch, DynamicList, AddItemRow, InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

export function HookItV4Form({ config, onChange, disabled }: ConfigFormProps) {
  const goals = (config.goals as string[]) || []
  const integration = (config.integration as string) || 'myfusionhelper'
  const stopOnError = config.stopOnError !== undefined ? (config.stopOnError as boolean) : false
  const goalPrefix = (config.goalPrefix as string) || ''
  const goalSuffix = (config.goalSuffix as string) || ''

  const [newGoal, setNewGoal] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <InfoBanner>
        Fires multiple goals sequentially for each webhook event. Perfect for triggering a series of campaigns or sequences. For example: fire goals ["welcome_email", "sales_sequence", "follow_up"] in order.
      </InfoBanner>

      <DynamicList<string>
        label="Goals to Achieve"
        description="List of goal names to achieve sequentially. Goals are triggered in the order listed."
        items={goals}
        onItemsChange={(items) => updateConfig({ goals: items })}
        disabled={disabled}
        renderAddForm={(onAdd) => (
          <AddItemRow
            onAdd={() => {
              const goal = newGoal.trim()
              if (!goal) return
              onAdd(goal)
              setNewGoal('')
            }}
            disabled={disabled}
            canAdd={!!newGoal.trim()}
          >
            <Input
              placeholder="Goal name"
              value={newGoal}
              onChange={(e) => setNewGoal(e.target.value)}
              disabled={disabled}
              className="flex-1"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  const goal = newGoal.trim()
                  if (!goal) return
                  onAdd(goal)
                  setNewGoal('')
                }
              }}
            />
          </AddItemRow>
        )}
        renderItem={(goal) => (
          <span className="font-mono text-xs">
            <span className="font-medium">{goal}</span>
          </span>
        )}
      />

      <FormTextField
        label="Integration Name"
        placeholder="myfusionhelper"
        value={integration}
        onChange={(e) => updateConfig({ integration: e.target.value })}
        disabled={disabled}
        description="Integration name for goal calls. Defaults to 'myfusionhelper'."
      />

      <FormTextField
        label="Goal Prefix (Optional)"
        placeholder="e.g., webhook_"
        value={goalPrefix}
        onChange={(e) => updateConfig({ goalPrefix: e.target.value })}
        disabled={disabled}
        description="Prefix to prepend to all goal names (e.g., 'webhook_' makes 'goal1' → 'webhook_goal1')."
      />

      <FormTextField
        label="Goal Suffix (Optional)"
        placeholder="e.g., _event"
        value={goalSuffix}
        onChange={(e) => updateConfig({ goalSuffix: e.target.value })}
        disabled={disabled}
        description="Suffix to append to all goal names (e.g., '_event' makes 'goal1' → 'goal1_event')."
      />

      <FormSwitch
        label="Stop on Error"
        description="Stop processing remaining goals if one fails. If disabled, continues to next goal on error."
        checked={stopOnError}
        onCheckedChange={(checked) => updateConfig({ stopOnError: checked })}
        disabled={disabled}
      />
    </div>
  )
}
