'use client'

// Schema: see schemas.ts > mathItSchema
import { FieldPicker } from '@/components/field-picker'
import { FormSelect, FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

const operations = [
  { value: 'add', label: 'Add (+)', description: 'Add a value to the field' },
  { value: 'subtract', label: 'Subtract (-)', description: 'Subtract a value from the field' },
  { value: 'multiply', label: 'Multiply (*)', description: 'Multiply the field by a value' },
  { value: 'divide', label: 'Divide (/)', description: 'Divide the field by a value' },
  { value: 'round', label: 'Round', description: 'Round to specified decimal places' },
  { value: 'ceil', label: 'Ceiling', description: 'Round up to the nearest integer' },
  { value: 'floor', label: 'Floor', description: 'Round down to the nearest integer' },
  { value: 'abs', label: 'Absolute', description: 'Convert to absolute value' },
  { value: 'percent', label: 'Percent', description: 'Calculate percentage of the value' },
]

export function MathItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const field = (config.field as string) || ''
  const targetField = (config.targetField as string) || ''
  const operation = (config.operation as string) || 'add'
  const operand = (config.operand as number) ?? 0
  const decimalPlaces = (config.decimalPlaces as number) ?? 2

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const needsOperand = !['abs', 'ceil', 'floor'].includes(operation)
  const showDecimals = operation === 'round'

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={field}
          onChange={(value) => updateConfig({ field: value })}
          placeholder="Select numeric field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The numeric field to read the value from.
        </p>
      </div>

      <FormSelect
        label="Operation"
        description={operations.find((o) => o.value === operation)?.description}
        value={operation}
        onValueChange={(v) => updateConfig({ operation: v })}
        options={operations}
        disabled={disabled}
      />

      {needsOperand && (
        <FormTextField
          label={showDecimals ? 'Decimal Places' : 'Operand'}
          type="number"
          placeholder="0"
          value={showDecimals ? decimalPlaces : operand}
          onChange={(e) => {
            const num = parseFloat(e.target.value) || 0
            updateConfig(showDecimals ? { decimalPlaces: num } : { operand: num })
          }}
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
          placeholder="Leave empty to update source field"
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          If provided, the result is stored here instead of overwriting the source field.
        </p>
      </div>
    </div>
  )
}
