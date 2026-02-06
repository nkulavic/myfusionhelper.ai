'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

const operations = [
  { value: 'add', label: 'Add (+)', description: 'Add a value to the field' },
  { value: 'subtract', label: 'Subtract (-)', description: 'Subtract a value from the field' },
  { value: 'multiply', label: 'Multiply (*)', description: 'Multiply the field by a value' },
  { value: 'divide', label: 'Divide (/)', description: 'Divide the field by a value' },
  { value: 'round', label: 'Round', description: 'Round to specified decimal places' },
  { value: 'abs', label: 'Absolute', description: 'Convert to absolute value' },
  { value: 'min', label: 'Minimum', description: 'Set to value if current is higher' },
  { value: 'max', label: 'Maximum', description: 'Set to value if current is lower' },
]

export function MathItForm({ config, onChange, disabled }: ConfigFormProps) {
  const sourceField = (config.sourceField as string) || ''
  const targetField = (config.targetField as string) || ''
  const operation = (config.operation as string) || 'add'
  const value = (config.value as number) ?? 0
  const decimalPlaces = (config.decimalPlaces as number) ?? 2

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const needsValue = !['abs'].includes(operation)
  const showDecimals = operation === 'round'

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="math-source">Source Field</Label>
        <Input
          id="math-source"
          placeholder="e.g. _LeadScore, TotalPurchases"
          value={sourceField}
          onChange={(e) => updateConfig({ sourceField: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The numeric field to read the value from.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="math-operation">Operation</Label>
        <select
          id="math-operation"
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

      {needsValue && (
        <div className="grid gap-2">
          <Label htmlFor="math-value">
            {showDecimals ? 'Decimal Places' : 'Value'}
          </Label>
          <Input
            id="math-value"
            type="number"
            placeholder="0"
            value={showDecimals ? decimalPlaces : value}
            onChange={(e) => {
              const num = parseFloat(e.target.value) || 0
              updateConfig(showDecimals ? { decimalPlaces: num } : { value: num })
            }}
            disabled={disabled}
          />
        </div>
      )}

      <div className="grid gap-2">
        <Label htmlFor="math-target">Target Field (optional)</Label>
        <Input
          id="math-target"
          placeholder="Leave empty to update source field"
          value={targetField}
          onChange={(e) => updateConfig({ targetField: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          If provided, the result is stored here instead of overwriting the source field.
        </p>
      </div>
    </div>
  )
}
