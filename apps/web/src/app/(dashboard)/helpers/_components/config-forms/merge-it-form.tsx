'use client'

// Go keys: source_fields ([]string), target_field, separator, skip_empty
import { FieldPicker } from '@/components/field-picker'
import { FormTextField, FormSwitch } from './form-fields'
import type { ConfigFormProps } from './types'

export function MergeItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const sourceFields = (config.sourceFields as string[]) || []
  const targetField = (config.targetField as string) || ''
  const separator = (config.separator as string) ?? ' '
  const skipEmpty = (config.skipEmpty as boolean) ?? true

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Source Fields</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={sourceFields}
          onChange={(value) => updateConfig({ sourceFields: value })}
          placeholder="Select fields to merge..."
          disabled={disabled}
          multiple
        />
        <p className="text-xs text-muted-foreground">
          Select the fields whose values will be merged together.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Target Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={targetField}
          onChange={(value) => updateConfig({ targetField: value })}
          placeholder="Select target field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field where the merged result will be saved.
        </p>
      </div>

      <FormTextField
        label="Separator"
        placeholder="e.g. a space, comma, dash"
        value={separator}
        onChange={(e) => updateConfig({ separator: e.target.value })}
        disabled={disabled}
        description="The character(s) placed between each merged value."
      />

      <FormSwitch
        label="Skip empty values"
        description="When enabled, empty source field values will be omitted from the merged result."
        checked={skipEmpty}
        onCheckedChange={(v) => updateConfig({ skipEmpty: v })}
        disabled={disabled}
      />
    </div>
  )
}
