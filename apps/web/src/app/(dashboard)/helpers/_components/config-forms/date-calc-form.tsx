'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

const operations = [
  { value: 'add_days', label: 'Add Days', description: 'Add a number of days to the date' },
  { value: 'subtract_days', label: 'Subtract Days', description: 'Subtract days from the date' },
  { value: 'add_months', label: 'Add Months', description: 'Add months to the date' },
  { value: 'subtract_months', label: 'Subtract Months', description: 'Subtract months from the date' },
  { value: 'diff_days', label: 'Difference in Days', description: 'Calculate days between two dates' },
  { value: 'set_now', label: 'Set to Now', description: 'Set the field to the current date/time' },
  { value: 'format', label: 'Format Date', description: 'Reformat the date to a specific pattern' },
]

const dateFormats = [
  { value: 'YYYY-MM-DD', label: 'YYYY-MM-DD (2025-01-15)' },
  { value: 'MM/DD/YYYY', label: 'MM/DD/YYYY (01/15/2025)' },
  { value: 'DD/MM/YYYY', label: 'DD/MM/YYYY (15/01/2025)' },
  { value: 'MMM DD, YYYY', label: 'MMM DD, YYYY (Jan 15, 2025)' },
  { value: 'MMMM DD, YYYY', label: 'MMMM DD, YYYY (January 15, 2025)' },
]

export function DateCalcForm({ config, onChange, disabled }: ConfigFormProps) {
  const operation = (config.operation as string) || 'add_days'
  const dateField = (config.dateField as string) || ''
  const amount = (config.amount as number) ?? 0
  const targetField = (config.targetField as string) || ''
  const secondDateField = (config.secondDateField as string) || ''
  const dateFormat = (config.dateFormat as string) || 'YYYY-MM-DD'

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const needsAmount = ['add_days', 'subtract_days', 'add_months', 'subtract_months'].includes(operation)
  const needsSecondDate = operation === 'diff_days'
  const needsFormat = operation === 'format'

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="date-operation">Operation</Label>
        <select
          id="date-operation"
          value={operation}
          onChange={(e) => updateConfig({ operation: e.target.value })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          {operations.map((op) => (
            <option key={op.value} value={op.value}>
              {op.label}
            </option>
          ))}
        </select>
        <p className="text-xs text-muted-foreground">
          {operations.find((o) => o.value === operation)?.description}
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="date-field">Date Field</Label>
        <Input
          id="date-field"
          placeholder="e.g. _DateCreated, custom_date_1"
          value={dateField}
          onChange={(e) => updateConfig({ dateField: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field containing the date to operate on.
        </p>
      </div>

      {needsAmount && (
        <div className="grid gap-2">
          <Label htmlFor="date-amount">Amount</Label>
          <Input
            id="date-amount"
            type="number"
            placeholder="0"
            value={amount}
            onChange={(e) => updateConfig({ amount: parseInt(e.target.value, 10) || 0 })}
            disabled={disabled}
          />
          <p className="text-xs text-muted-foreground">
            Number of {operation.includes('month') ? 'months' : 'days'} to add or subtract.
          </p>
        </div>
      )}

      {needsSecondDate && (
        <div className="grid gap-2">
          <Label htmlFor="date-second">Second Date Field</Label>
          <Input
            id="date-second"
            placeholder="e.g. _LastPurchaseDate"
            value={secondDateField}
            onChange={(e) => updateConfig({ secondDateField: e.target.value })}
            disabled={disabled}
          />
          <p className="text-xs text-muted-foreground">
            The second date field to compare against. Result is stored as number of days.
          </p>
        </div>
      )}

      {needsFormat && (
        <div className="grid gap-2">
          <Label htmlFor="date-format">Output Format</Label>
          <select
            id="date-format"
            value={dateFormat}
            onChange={(e) => updateConfig({ dateFormat: e.target.value })}
            disabled={disabled}
            className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
          >
            {dateFormats.map((f) => (
              <option key={f.value} value={f.value}>
                {f.label}
              </option>
            ))}
          </select>
        </div>
      )}

      <div className="grid gap-2">
        <Label htmlFor="date-target">Target Field (optional)</Label>
        <Input
          id="date-target"
          placeholder="Leave empty to update the source field"
          value={targetField}
          onChange={(e) => updateConfig({ targetField: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          If set, the result is stored in this field instead of overwriting the source.
        </p>
      </div>
    </div>
  )
}
