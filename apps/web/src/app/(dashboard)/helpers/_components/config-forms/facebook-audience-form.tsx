'use client'

// Schema: see schemas.ts > facebookAudienceSchema
import { FormTextField, FormRadioGroup } from './form-fields'
import type { ConfigFormProps } from './types'

const actionOptions = [
  { value: 'add', label: 'Add to Audience' },
  { value: 'remove', label: 'Remove from Audience' },
]

export function FacebookAudienceForm({ config, onChange, disabled }: ConfigFormProps) {
  const audienceId = (config.audienceId as string) || ''
  const action = (config.action as string) || 'add'

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Custom Audience ID"
        description="The Facebook Custom Audience ID to add or remove contacts from."
        placeholder="Facebook Custom Audience ID"
        value={audienceId}
        onChange={(e) => updateConfig({ audienceId: e.target.value })}
        disabled={disabled}
      />
      <FormRadioGroup
        label="Action"
        value={action}
        onValueChange={(v) => updateConfig({ action: v })}
        options={actionOptions}
        disabled={disabled}
      />
    </div>
  )
}
