'use client'

// Schema: see schemas.ts > combineItSchema
import { FieldPicker } from '@/components/field-picker'
import { FormSwitch, FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function CombineItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
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
        <label className="text-sm font-medium">Fields to Combine</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={sourceFields}
          onChange={(value) => updateConfig({ sourceFields: value })}
          placeholder="Select fields to combine..."
          disabled={disabled}
          multiple
        />
        <p className="text-xs text-muted-foreground">
          Select the fields whose values will be combined together.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Save Combined Data To</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={targetField}
          onChange={(value) => updateConfig({ targetField: value })}
          placeholder="Select target field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field where the combined result will be saved.
        </p>
      </div>

      <FormTextField
        label="Separator"
        description="Character(s) placed between each combined value."
        placeholder="e.g. a space, comma, dash"
        value={separator}
        onChange={(e) => updateConfig({ separator: e.target.value })}
        disabled={disabled}
      />

      <FormSwitch
        label="Skip empty fields"
        description="When enabled, empty source fields are omitted from the combined result."
        checked={skipEmpty}
        onCheckedChange={(v) => updateConfig({ skipEmpty: v })}
        disabled={disabled}
      />
    </div>
  )
}
