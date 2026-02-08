'use client'

// Schema: see schemas.ts > contactUpdaterSchema
import { FormTextField, InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

export function ContactUpdaterForm({ config, onChange, disabled }: ConfigFormProps) {
  const updateKey = (config.updateKey as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Update Key"
        description="A private key that must be included in the HTTP POST parameters for the update to be accepted."
        placeholder="Enter a private security key"
        value={updateKey}
        onChange={(e) => updateConfig({ updateKey: e.target.value })}
        disabled={disabled}
      />

      <InfoBanner>
        Send HTTP POST requests with <code className="text-[11px]">contact_update[FieldName]=value</code> pairs
        to update contact fields. Include the update key for authentication.
      </InfoBanner>
    </div>
  )
}
