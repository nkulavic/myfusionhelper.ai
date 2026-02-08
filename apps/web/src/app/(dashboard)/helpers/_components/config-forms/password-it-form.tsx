'use client'

// Schema: see schemas.ts > passwordItSchema
import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormTextField, FormSwitch } from './form-fields'
import type { ConfigFormProps } from './types'

export function PasswordItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const length = (config.length as number) ?? 12
  const includeSpecial = (config.includeSpecial as boolean) ?? false
  const overwrite = (config.overwrite as boolean) ?? false
  const targetField = (config.targetField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Password Length"
        description="Number of characters in the generated password."
        type="number"
        min={1}
        placeholder="12"
        value={length}
        onChange={(e) => updateConfig({ length: parseInt(e.target.value, 10) || 12 })}
        disabled={disabled}
      />

      <FormSwitch
        label="Include special characters"
        description="Include symbols like !@#$%^&*() in the generated password."
        checked={includeSpecial}
        onCheckedChange={(v) => updateConfig({ includeSpecial: v })}
        disabled={disabled}
      />

      <FormSwitch
        label="Overwrite existing value"
        description="If enabled, overwrite any existing data in the target field."
        checked={overwrite}
        onCheckedChange={(v) => updateConfig({ overwrite: v })}
        disabled={disabled}
      />

      <div className="grid gap-2">
        <Label>Save To Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={targetField}
          onChange={(value) => updateConfig({ targetField: value })}
          placeholder="Select field to save password..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field where the generated password will be stored.
        </p>
      </div>
    </div>
  )
}
