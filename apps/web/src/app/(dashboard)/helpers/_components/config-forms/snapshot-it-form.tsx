'use client'

// Schema: see schemas.ts > snapshotItSchema
import { FormSwitch } from './form-fields'
import type { ConfigFormProps } from './types'

export function SnapshotItForm({ config, onChange, disabled }: ConfigFormProps) {
  const includeTags = (config.includeTags as boolean) ?? true
  const includeCustomFields = (config.includeCustomFields as boolean) ?? true

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormSwitch
        label="Include tags"
        checked={includeTags}
        onCheckedChange={(v) => updateConfig({ includeTags: v })}
        disabled={disabled}
      />
      <FormSwitch
        label="Include custom fields"
        checked={includeCustomFields}
        onCheckedChange={(v) => updateConfig({ includeCustomFields: v })}
        disabled={disabled}
      />
    </div>
  )
}
