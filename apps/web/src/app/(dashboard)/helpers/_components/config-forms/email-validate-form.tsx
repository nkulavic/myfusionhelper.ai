'use client'

// Schema: see schemas.ts > emailValidateSchema
import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormTextField, FormSwitch } from './form-fields'
import type { ConfigFormProps } from './types'

export function EmailValidateForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const emailField = (config.emailField as string) || ''
  const resultField = (config.resultField as string) || ''
  const checkMx = (config.checkMx as boolean) ?? false
  const validGoal = (config.validGoal as string) || ''
  const invalidGoal = (config.invalidGoal as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label>Email Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={emailField}
          onChange={(value) => updateConfig({ emailField: value })}
          placeholder="Select email field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field containing the email address to validate.
        </p>
      </div>

      <div className="grid gap-2">
        <Label>Result Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={resultField}
          onChange={(value) => updateConfig({ resultField: value })}
          placeholder="Store validation result..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field where the validation result will be stored.
        </p>
      </div>

      <FormSwitch
        label="Check MX Records"
        description="Verify that the email domain has valid MX records for receiving mail."
        checked={checkMx}
        onCheckedChange={(v) => updateConfig({ checkMx: v })}
        disabled={disabled}
      />

      <FormTextField
        label="Valid Goal"
        placeholder="e.g. goal name or ID for valid emails"
        value={validGoal}
        onChange={(e) => updateConfig({ validGoal: e.target.value })}
        disabled={disabled}
        description="The goal to trigger when the email is valid."
      />

      <FormTextField
        label="Invalid Goal"
        placeholder="e.g. goal name or ID for invalid emails"
        value={invalidGoal}
        onChange={(e) => updateConfig({ invalidGoal: e.target.value })}
        disabled={disabled}
        description="The goal to trigger when the email is invalid."
      />
    </div>
  )
}
