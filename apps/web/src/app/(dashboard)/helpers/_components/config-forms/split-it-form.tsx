'use client'

// Go keys: mode (enum: tag|goal), option_a, option_b, state_field
import { FieldPicker } from '@/components/field-picker'
import { FormTextField, FormRadioGroup } from './form-fields'
import type { ConfigFormProps } from './types'

export function SplitItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const mode = (config.mode as string) || 'tag'
  const optionA = (config.optionA as string) || ''
  const optionB = (config.optionB as string) || ''
  const stateField = (config.stateField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormRadioGroup
        label="Split Mode"
        value={mode}
        onValueChange={(value) => updateConfig({ mode: value })}
        options={[
          { value: 'tag', label: 'Tag' },
          { value: 'goal', label: 'Goal' },
        ]}
        disabled={disabled}
        description="Choose whether to split contacts using tags or goals."
      />

      <div className="grid grid-cols-2 gap-4">
        <FormTextField
          label={mode === 'tag' ? 'Option A Tag ID' : 'Option A Goal'}
          placeholder={mode === 'tag' ? 'Tag ID for group A' : 'Goal name for group A'}
          value={optionA}
          onChange={(e) => updateConfig({ optionA: e.target.value })}
          disabled={disabled}
        />
        <FormTextField
          label={mode === 'tag' ? 'Option B Tag ID' : 'Option B Goal'}
          placeholder={mode === 'tag' ? 'Tag ID for group B' : 'Goal name for group B'}
          value={optionB}
          onChange={(e) => updateConfig({ optionB: e.target.value })}
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">State Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={stateField}
          onChange={(value) => updateConfig({ stateField: value })}
          placeholder="Select field to store split state..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          A contact field used to remember which group (A or B) this contact was assigned to.
        </p>
      </div>
    </div>
  )
}
