'use client'

// Schema: see schemas.ts > uploadItSchema
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function UploadItForm({ config, onChange, disabled }: ConfigFormProps) {
  const goal = (config.goal as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="API Goal on Upload"
        description="API goal to trigger when a file is uploaded. Leave empty for no goal."
        placeholder="Goal name (alphanumeric only)"
        value={goal}
        onChange={(e) => updateConfig({ goal: e.target.value.replace(/[^a-zA-Z0-9]/g, '') })}
        disabled={disabled}
      />
    </div>
  )
}
