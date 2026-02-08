'use client'

// Go keys: source_field, target_field, from_timezone, to_timezone, output_format
import { FieldPicker } from '@/components/field-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function WhenIsItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const sourceField = (config.sourceField as string) || ''
  const targetField = (config.targetField as string) || ''
  const fromTimezone = (config.fromTimezone as string) || 'UTC'
  const toTimezone = (config.toTimezone as string) || ''
  const outputFormat = (config.outputFormat as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Source Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={sourceField}
          onChange={(value) => updateConfig({ sourceField: value })}
          placeholder="Select source date/time field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field containing the date/time value to convert.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Target Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={targetField}
          onChange={(value) => updateConfig({ targetField: value })}
          placeholder="Select target field for converted value..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field where the converted date/time will be saved.
        </p>
      </div>

      <FormTextField
        label="From Timezone"
        placeholder="e.g. UTC, America/New_York"
        value={fromTimezone}
        onChange={(e) => updateConfig({ fromTimezone: e.target.value })}
        disabled={disabled}
        description="The timezone of the source value. Defaults to UTC."
      />

      <FormTextField
        label="To Timezone"
        placeholder="e.g. America/Los_Angeles, Europe/London"
        value={toTimezone}
        onChange={(e) => updateConfig({ toTimezone: e.target.value })}
        disabled={disabled}
        description="The target timezone to convert the value into."
      />

      <FormTextField
        label="Output Format"
        placeholder="e.g. 2006-01-02 15:04:05"
        value={outputFormat}
        onChange={(e) => updateConfig({ outputFormat: e.target.value })}
        disabled={disabled}
        description="Go time format string for the output. Leave blank for default."
      />
    </div>
  )
}
