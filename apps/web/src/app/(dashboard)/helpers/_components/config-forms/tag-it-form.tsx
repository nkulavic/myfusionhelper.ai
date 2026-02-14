'use client'

// Schema: see schemas.ts > tagItSchema
import { TagPicker } from '@/components/tag-picker'
import { FormSelect } from './form-fields'
import type { ConfigFormProps } from './types'

const actionOptions = [
  { value: 'apply', label: 'Apply tags' },
  { value: 'remove', label: 'Remove tags' },
]

export function TagItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const action = (config.action as string) || 'apply'
  const tagIds = (config.tagIds as string[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Action"
        description={
          action === 'apply'
            ? 'Add these tags to the contact.'
            : 'Remove these tags from the contact.'
        }
        value={action}
        onValueChange={(v) => updateConfig({ action: v })}
        options={actionOptions}
        disabled={disabled}
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Tags</label>
        <TagPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={tagIds}
          onChange={(value) => updateConfig({ tagIds: value })}
          placeholder="Select tags..."
          disabled={disabled}
        />
      </div>
    </div>
  )
}
