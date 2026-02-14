'use client'

// Schema: see schemas.ts > goalItSchema
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function GoalItForm({ config, onChange, disabled }: ConfigFormProps) {
  const goalName = (config.goalName as string) || ''
  const integration = (config.integration as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="API Goal Name"
        description="The API goal to achieve on the contact record."
        placeholder="Goal name (alphanumeric only)"
        value={goalName}
        onChange={(e) => updateConfig({ goalName: e.target.value.replace(/[^a-zA-Z0-9_]/g, '') })}
        disabled={disabled}
      />
      <FormTextField
        label="Integration Name"
        description="Optional integration name for the goal call."
        placeholder="e.g. myapp"
        value={integration}
        onChange={(e) => updateConfig({ integration: e.target.value })}
        disabled={disabled}
      />
    </div>
  )
}
