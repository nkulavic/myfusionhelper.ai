'use client'

// Schema: see schemas.ts > groupItSchema
import { FieldPicker } from '@/components/field-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function GroupItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const field = (config.field as string) || ''
  const tagPrefix = (config.tagPrefix as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Group By Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={field}
          onChange={(value) => updateConfig({ field: value })}
          placeholder="Select field to group by..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Contacts sharing the same value in this field will be grouped together with a tag.
        </p>
      </div>

      <FormTextField
        label="Tag Prefix (optional)"
        placeholder="e.g. group_"
        value={tagPrefix}
        onChange={(e) => updateConfig({ tagPrefix: e.target.value })}
        disabled={disabled}
        description="Optional prefix to prepend to auto-created tag names."
      />
    </div>
  )
}
