'use client'

// Schema: see schemas.ts > copyItSchema
import { FieldPicker } from '@/components/field-picker'
import { FormSwitch } from './form-fields'
import type { ConfigFormProps } from './types'

export function CopyItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const sourceField = (config.sourceField as string) || ''
  const targetField = (config.targetField as string) || ''
  const overwrite = (config.overwrite as boolean) ?? true

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Source Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={sourceField}
          onChange={(value) => updateConfig({ sourceField: value })}
          placeholder="Select source field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field to copy the value from.
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
          The field to copy the value to.
        </p>
      </div>

      <FormSwitch
        label="Overwrite existing value in target field"
        description="When disabled, the value will only be copied if the target field is empty."
        checked={overwrite}
        onCheckedChange={(v) => updateConfig({ overwrite: v })}
        disabled={disabled}
      />
    </div>
  )
}
