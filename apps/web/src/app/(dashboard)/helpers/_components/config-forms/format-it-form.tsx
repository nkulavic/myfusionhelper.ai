'use client'

// Schema: see schemas.ts > formatItSchema
import { FieldPicker } from '@/components/field-picker'
import { FormSelect } from './form-fields'
import type { ConfigFormProps } from './types'

const formatOptions = [
  { value: 'uppercase', label: 'UPPERCASE', description: 'Convert to all uppercase' },
  { value: 'lowercase', label: 'lowercase', description: 'Convert to all lowercase' },
  { value: 'title_case', label: 'Title Case', description: 'Capitalize first letter of each word' },
  { value: 'trim', label: 'Trim whitespace', description: 'Remove leading and trailing spaces' },
  { value: 'trim_uppercase', label: 'Trim + UPPERCASE', description: 'Trim whitespace then convert to uppercase' },
  { value: 'trim_lowercase', label: 'Trim + lowercase', description: 'Trim whitespace then convert to lowercase' },
  { value: 'trim_title_case', label: 'Trim + Title Case', description: 'Trim whitespace then capitalize each word' },
]

export function FormatItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const field = (config.field as string) || ''
  const format = (config.format as string) || 'title_case'
  const targetField = (config.targetField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={field}
          onChange={(value) => updateConfig({ field: value })}
          placeholder="Select field to format..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field to apply formatting to.
        </p>
      </div>

      <FormSelect
        label="Format Type"
        description={formatOptions.find((o) => o.value === format)?.description}
        value={format}
        onValueChange={(v) => updateConfig({ format: v })}
        options={formatOptions}
        disabled={disabled}
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Target Field (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={targetField}
          onChange={(value) => updateConfig({ targetField: value })}
          placeholder="Leave empty to update the source field"
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          If set, the formatted result is stored here instead of overwriting the source field.
        </p>
      </div>
    </div>
  )
}
