'use client'

// Schema: see schemas.ts > countTagsSchema
import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function CountTagsForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const targetField = (config.targetField as string) || ''
  const category = (config.category as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Tag Category Filter"
        description="Only count tags in this category. Leave empty to count all tags."
        placeholder="e.g. score (leave empty to count all)"
        value={category}
        onChange={(e) => updateConfig({ category: e.target.value })}
        disabled={disabled}
      />
      <div className="grid gap-2">
        <Label>Store Count In</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={targetField}
          onChange={(value) => updateConfig({ targetField: value })}
          placeholder="Select field for count..."
          filterType="number"
          disabled={disabled}
        />
      </div>
    </div>
  )
}
