'use client'

// Go keys: automation_id (single string)
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function TriggerItForm({ config, onChange, disabled }: ConfigFormProps) {
  const automationId = (config.automationId as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Automation ID"
        placeholder="Enter the automation ID to trigger"
        value={automationId}
        onChange={(e) => updateConfig({ automationId: e.target.value })}
        disabled={disabled}
        description="The ID of the automation to trigger when this helper runs."
      />
    </div>
  )
}
