'use client'

// Schema: see schemas.ts > clearItSchema
import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import type { ConfigFormProps } from './types'

export function ClearItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const fields = (config.fields as string[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label>Fields to Clear</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={fields}
          onChange={(value) => updateConfig({ fields: value })}
          placeholder="Select fields to clear..."
          disabled={disabled}
          multiple
        />
        <p className="text-xs text-muted-foreground">
          All selected fields will be emptied (cleared) on the contact.
        </p>
      </div>
    </div>
  )
}
