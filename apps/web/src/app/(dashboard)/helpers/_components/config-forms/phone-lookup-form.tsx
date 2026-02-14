'use client'

// Schema: see schemas.ts > phoneLookupSchema
import { FormSelect, FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

const phoneFields = [
  { value: 'Phone1', label: 'Phone 1' },
  { value: 'Phone2', label: 'Phone 2' },
  { value: 'Phone3', label: 'Phone 3' },
  { value: 'Fax1', label: 'Fax 1' },
  { value: 'Fax2', label: 'Fax 2' },
]

export function PhoneLookupForm({ config, onChange, disabled }: ConfigFormProps) {
  const phoneField = (config.phoneField as string) || 'Phone1'
  const countryCode = (config.countryCode as string) || 'US'
  const validGoal = (config.validGoal as string) || ''
  const invalidGoal = (config.invalidGoal as string) || ''
  const emptyGoal = (config.emptyGoal as string) || ''
  const saveFormattedTo = (config.saveFormattedTo as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Phone Field to Validate"
        value={phoneField}
        onValueChange={(v) => updateConfig({ phoneField: v })}
        options={phoneFields}
        disabled={disabled}
      />

      <FormTextField
        label="Country Code"
        placeholder="e.g. US, GB, CA"
        value={countryCode}
        onChange={(e) => updateConfig({ countryCode: e.target.value.toUpperCase() })}
        disabled={disabled}
        maxLength={2}
        description="ISO country code to validate the phone number against."
      />

      <FormTextField
        label="Valid Goal (API goal name)"
        placeholder="Goal to trigger if valid"
        value={validGoal}
        onChange={(e) => updateConfig({ validGoal: e.target.value.replace(/[^a-zA-Z0-9]/g, '') })}
        disabled={disabled}
        description="API goal to trigger when the phone number is valid. Leave empty to skip."
      />

      <FormTextField
        label="Invalid Goal (API goal name)"
        placeholder="Goal to trigger if invalid"
        value={invalidGoal}
        onChange={(e) => updateConfig({ invalidGoal: e.target.value.replace(/[^a-zA-Z0-9]/g, '') })}
        disabled={disabled}
        description="API goal to trigger when the phone number is invalid. Leave empty to skip."
      />

      <FormTextField
        label="Empty Goal (API goal name)"
        placeholder="Goal to trigger if field is empty"
        value={emptyGoal}
        onChange={(e) => updateConfig({ emptyGoal: e.target.value.replace(/[^a-zA-Z0-9]/g, '') })}
        disabled={disabled}
        description="API goal to trigger when the phone field is empty. Leave empty to skip."
      />

      <FormTextField
        label="Save Formatted Number To"
        placeholder="Field to save formatted phone number"
        value={saveFormattedTo}
        onChange={(e) => updateConfig({ saveFormattedTo: e.target.value })}
        disabled={disabled}
        description="Contact field where the formatted E.164 phone number will be saved. Leave empty to skip."
      />
    </div>
  )
}
