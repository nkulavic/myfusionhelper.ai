'use client'

// Schema: see schemas.ts > emailFieldSchema
import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormSelect } from './form-fields'
import type { ConfigFormProps } from './types'

const emailFields = [
  { value: 'Email', label: 'Email (Primary)' },
  { value: 'EmailAddress2', label: 'Email Address 2' },
  { value: 'EmailAddress3', label: 'Email Address 3' },
]

export function EmailFieldForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const emailField = (config.emailField as string) || 'Email'
  const saveTo = (config.saveTo as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Email Address Field"
        description="The email address field to use for the lookup."
        value={emailField}
        onValueChange={(v) => updateConfig({ emailField: v })}
        options={emailFields}
        disabled={disabled}
      />

      <div className="grid gap-2">
        <Label>Save Result To</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={saveTo}
          onChange={(value) => updateConfig({ saveTo: value })}
          placeholder="Select field to save the result..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field where the lookup result will be saved.
        </p>
      </div>
    </div>
  )
}
