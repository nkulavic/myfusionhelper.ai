'use client'

// Schema: see schemas.ts > dateCalcSchema
import { FieldPicker } from '@/components/field-picker'
import { FormSelect, FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

const operations = [
  { value: 'add_days', label: 'Add Days', description: 'Add a number of days to the date' },
  { value: 'subtract_days', label: 'Subtract Days', description: 'Subtract days from the date' },
  { value: 'add_months', label: 'Add Months', description: 'Add months to the date' },
  { value: 'subtract_months', label: 'Subtract Months', description: 'Subtract months from the date' },
  { value: 'add_years', label: 'Add Years', description: 'Add years to the date' },
  { value: 'subtract_years', label: 'Subtract Years', description: 'Subtract years from the date' },
  { value: 'set_now', label: 'Set to Now', description: 'Set the field to the current date/time' },
  { value: 'diff_days', label: 'Difference in Days', description: 'Calculate days between two dates' },
  { value: 'format', label: 'Format Date', description: 'Reformat the date to a specific pattern' },
]

const dateFormats = [
  { value: 'YYYY-MM-DD', label: 'YYYY-MM-DD (2025-01-15)' },
  { value: 'MM/DD/YYYY', label: 'MM/DD/YYYY (01/15/2025)' },
  { value: 'DD/MM/YYYY', label: 'DD/MM/YYYY (15/01/2025)' },
  { value: 'MMM DD, YYYY', label: 'MMM DD, YYYY (Jan 15, 2025)' },
  { value: 'MMMM DD, YYYY', label: 'MMMM DD, YYYY (January 15, 2025)' },
]

export function DateCalcForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const operation = (config.operation as string) || 'add_days'
  const field = (config.field as string) || ''
  const amount = (config.amount as number) ?? 0
  const targetField = (config.targetField as string) || ''
  const compareField = (config.compareField as string) || ''
  const outputFormat = (config.outputFormat as string) || 'YYYY-MM-DD'

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const needsAmount = ['add_days', 'subtract_days', 'add_months', 'subtract_months', 'add_years', 'subtract_years'].includes(operation)
  const needsCompareField = operation === 'diff_days'
  const needsFormat = operation === 'format'

  return (
    <div className="space-y-4">
      <FormSelect
        label="Operation"
        description={operations.find((o) => o.value === operation)?.description}
        value={operation}
        onValueChange={(v) => updateConfig({ operation: v })}
        options={operations}
        disabled={disabled}
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={field}
          onChange={(value) => updateConfig({ field: value })}
          placeholder="Select date field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field containing the date to operate on.
        </p>
      </div>

      {needsAmount && (
        <FormTextField
          label="Amount"
          type="number"
          placeholder="0"
          value={amount}
          onChange={(e) => updateConfig({ amount: parseInt(e.target.value, 10) || 0 })}
          disabled={disabled}
          description={`Number of ${operation.includes('year') ? 'years' : operation.includes('month') ? 'months' : 'days'} to add or subtract.`}
        />
      )}

      {needsCompareField && (
        <div className="grid gap-2">
          <label className="text-sm font-medium">Compare Field</label>
          <FieldPicker
            platformId={platformId ?? ''}
            connectionId={connectionId ?? ''}
            value={compareField}
            onChange={(value) => updateConfig({ compareField: value })}
            placeholder="Select second date field..."
            disabled={disabled}
          />
          <p className="text-xs text-muted-foreground">
            The second date field to compare against. Result is stored as number of days.
          </p>
        </div>
      )}

      {needsFormat && (
        <FormSelect
          label="Output Format"
          value={outputFormat}
          onValueChange={(v) => updateConfig({ outputFormat: v })}
          options={dateFormats}
          disabled={disabled}
        />
      )}

      <div className="grid gap-2">
        <label className="text-sm font-medium">Target Field (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={targetField}
          onChange={(value) => updateConfig({ targetField: value })}
          placeholder="Leave empty to update the source field"
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          If set, the result is stored in this field instead of overwriting the source.
        </p>
      </div>
    </div>
  )
}
