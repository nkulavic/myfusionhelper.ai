'use client'

// Schema: see schemas.ts > advanceMathSchema
import { FieldPicker } from '@/components/field-picker'
import { FormSelect, FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

const operations = [
  { value: 'power', label: 'Power (^)', description: 'Raise source to the power of operand/second field' },
  { value: 'sqrt', label: 'Square Root (√)', description: 'Calculate square root of source value' },
  { value: 'abs', label: 'Absolute Value (|x|)', description: 'Convert to absolute value' },
  { value: 'round', label: 'Round', description: 'Round to nearest integer' },
  { value: 'ceil', label: 'Ceiling', description: 'Round up to nearest integer' },
  { value: 'floor', label: 'Floor', description: 'Round down to nearest integer' },
  { value: 'min', label: 'Minimum', description: 'Get smaller of source and operand/second field' },
  { value: 'max', label: 'Maximum', description: 'Get larger of source and operand/second field' },
]

export function AdvanceMathForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const sourceField = (config.sourceField as string) || ''
  const targetField = (config.targetField as string) || ''
  const operation = (config.operation as string) || 'sqrt'
  const operand = (config.operand as number) ?? undefined
  const secondField = (config.secondField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  // Operations that require a second value: power, min, max
  const needsSecondValue = ['power', 'min', 'max'].includes(operation)
  // Operations that don't need anything: sqrt, abs, round, ceil, floor
  const needsNothing = ['sqrt', 'abs', 'round', 'ceil', 'floor'].includes(operation)

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Source Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={sourceField}
          onChange={(value) => updateConfig({ sourceField: value })}
          placeholder="Select numeric field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The numeric field containing the value to perform the operation on.
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

      {needsSecondValue && (
        <>
          <FormTextField
            label="Operand (number)"
            type="number"
            placeholder="0"
            value={operand ?? ''}
            onChange={(e) => {
              const num = e.target.value ? parseFloat(e.target.value) : undefined
              updateConfig({ operand: num })
            }}
            disabled={disabled}
            description="Use a fixed number for the operation"
          />

          <div className="grid gap-2">
            <label className="text-sm font-medium">Second Field (alternative)</label>
            <FieldPicker
              platformId={platformId ?? ''}
              connectionId={connectionId ?? ''}
              value={secondField}
              onChange={(value) => updateConfig({ secondField: value })}
              placeholder="Or select a field..."
              disabled={disabled}
            />
            <p className="text-xs text-muted-foreground">
              Use a field value instead of a fixed operand. Either operand or second field is required for {operation}.
            </p>
          </div>
        </>
      )}

      {needsNothing && (
        <div className="rounded-lg border border-muted bg-muted/20 p-3">
          <p className="text-sm text-muted-foreground">
            This operation does not require additional parameters.
          </p>
        </div>
      )}

      <div className="grid gap-2">
        <label className="text-sm font-medium">Target Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={targetField}
          onChange={(value) => updateConfig({ targetField: value })}
          placeholder="Select field to store result..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field where the computed result will be stored.
        </p>
      </div>
    </div>
  )
}
