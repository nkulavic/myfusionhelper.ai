'use client'

// Schema: see schemas.ts > optInSchema
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function OptInForm({ config, onChange, disabled }: ConfigFormProps) {
  const emailField = (config.emailField as string) || 'email'
  const reason = (config.reason as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Email Field"
        description="The contact field containing the email address to opt in/out."
        placeholder="e.g. email"
        value={emailField}
        onChange={(e) => updateConfig({ emailField: e.target.value })}
        disabled={disabled}
      />
      <FormTextField
        label="Opt-In Reason"
        description="Reason recorded for the opt-in action. Leave empty for default."
        placeholder="e.g. Newsletter signup"
        value={reason}
        onChange={(e) => updateConfig({ reason: e.target.value })}
        disabled={disabled}
      />
    </div>
  )
}
