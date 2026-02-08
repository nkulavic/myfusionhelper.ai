'use client'

// Schema: see schemas.ts > nameParseItSchema
import { FieldPicker } from '@/components/field-picker'
import type { ConfigFormProps } from './types'

export function NameParseItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const sourceField = (config.sourceField as string) || ''
  const firstNameField = (config.firstNameField as string) || ''
  const lastNameField = (config.lastNameField as string) || ''
  const suffixField = (config.suffixField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Source Field (Full Name)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={sourceField}
          onChange={(value) => updateConfig({ sourceField: value })}
          placeholder="Select field containing the full name..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field containing the full name to parse into components.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Save First Name To</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={firstNameField}
          onChange={(value) => updateConfig({ firstNameField: value })}
          placeholder="Select field for first name..."
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Save Last Name To</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={lastNameField}
          onChange={(value) => updateConfig({ lastNameField: value })}
          placeholder="Select field for last name..."
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Save Suffix To</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={suffixField}
          onChange={(value) => updateConfig({ suffixField: value })}
          placeholder="Select field or leave empty to skip..."
          disabled={disabled}
        />
      </div>
    </div>
  )
}
