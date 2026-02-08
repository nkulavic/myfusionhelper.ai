'use client'

// Schema: see schemas.ts > assignItSchema
import { FieldPicker } from '@/components/field-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function AssignItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const ownerId = (config.ownerId as string) || ''
  const ownerField = (config.ownerField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Owner ID"
        placeholder="Enter the user/owner ID..."
        value={ownerId}
        onChange={(e) => updateConfig({ ownerId: e.target.value })}
        disabled={disabled}
        description="The ID of the user to assign as the contact owner."
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Owner Field (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={ownerField}
          onChange={(value) => updateConfig({ ownerField: value })}
          placeholder="Select field to store owner ID..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          If set, the owner ID will also be written to this contact field.
        </p>
      </div>
    </div>
  )
}
