'use client'

// Schema: see schemas.ts > defaultToFieldSchema
import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function DefaultToFieldForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const defaultValue = (config.default as string) || ''
  const toField = (config.toField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Default Value"
        description={`The value to set on the contact. Use {{field_name}} tokens for dynamic values.`}
        placeholder="Value to set (supports {{field_name}} tokens)"
        value={defaultValue}
        onChange={(e) => updateConfig({ default: e.target.value })}
        disabled={disabled}
      />

      <div className="grid gap-2">
        <Label>To Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={toField}
          onChange={(value) => updateConfig({ toField: value })}
          placeholder="Select target field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field where the default value will be saved.
        </p>
      </div>
    </div>
  )
}
