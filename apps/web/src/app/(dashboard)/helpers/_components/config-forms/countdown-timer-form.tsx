'use client'

// Schema: see schemas.ts > countdownTimerSchema

import { FieldPicker } from '@/components/field-picker'
import { FormSelect, FormTextField, FormColorPicker, FormSwitch } from './form-fields'
import type { ConfigFormProps } from './types'

const timerTypeOptions = [
  { value: 'standard', label: 'Fixed Date' },
  { value: 'contact_field', label: 'Contact Field' },
  { value: 'evergreen', label: 'Evergreen' },
]

const timerTypeDescriptions: Record<string, string> = {
  standard: 'Count down to a specific date and time',
  contact_field: 'Count down to a date stored in a contact field',
  evergreen: 'Add time from when the timer is viewed',
}

export function CountdownTimerForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const timerType = (config.timerType as string) || 'standard'
  const endTime = (config.endTime as string) || ''
  const contactField = (config.contactField as string) || ''
  const addDays = (config.addDays as number) ?? 0
  const addHours = (config.addHours as number) ?? 0
  const addMinutes = (config.addMinutes as number) ?? 0
  const backgroundColor = (config.backgroundColor as string) || '#000000'
  const digitColor = (config.digitColor as string) || '#FFFFFF'
  const labelColor = (config.labelColor as string) || '#CCCCCC'
  const transparentBg = (config.transparentBg as boolean) ?? false

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Timer Type"
        value={timerType}
        onValueChange={(value) => updateConfig({ timerType: value })}
        options={timerTypeOptions}
        description={timerTypeDescriptions[timerType]}
        disabled={disabled}
      />

      {timerType === 'standard' && (
        <FormTextField
          label="End Date/Time"
          type="datetime-local"
          value={endTime}
          onChange={(e) => updateConfig({ endTime: e.target.value })}
          disabled={disabled}
        />
      )}

      {timerType === 'contact_field' && (
        <div className="grid gap-2">
          <label className="text-sm font-medium">Date Field</label>
          <FieldPicker
            platformId={platformId ?? ''}
            connectionId={connectionId ?? ''}
            value={contactField}
            onChange={(value) => updateConfig({ contactField: value })}
            placeholder="Select date field..."
            filterType="date"
            disabled={disabled}
          />
        </div>
      )}

      {timerType === 'evergreen' && (
        <div className="grid grid-cols-3 gap-3">
          <FormTextField
            label="Days"
            type="number"
            min={0}
            value={addDays}
            onChange={(e) => updateConfig({ addDays: parseInt(e.target.value, 10) || 0 })}
            disabled={disabled}
          />
          <FormTextField
            label="Hours"
            type="number"
            min={0}
            max={23}
            value={addHours}
            onChange={(e) => updateConfig({ addHours: parseInt(e.target.value, 10) || 0 })}
            disabled={disabled}
          />
          <FormTextField
            label="Minutes"
            type="number"
            min={0}
            max={59}
            value={addMinutes}
            onChange={(e) => updateConfig({ addMinutes: parseInt(e.target.value, 10) || 0 })}
            disabled={disabled}
          />
        </div>
      )}

      <div className="grid grid-cols-3 gap-3">
        <FormColorPicker
          label="Background"
          value={backgroundColor}
          onChange={(value) => updateConfig({ backgroundColor: value })}
          disabled={disabled || transparentBg}
        />
        <FormColorPicker
          label="Digits"
          value={digitColor}
          onChange={(value) => updateConfig({ digitColor: value })}
          disabled={disabled}
        />
        <FormColorPicker
          label="Labels"
          value={labelColor}
          onChange={(value) => updateConfig({ labelColor: value })}
          disabled={disabled}
        />
      </div>

      <FormSwitch
        label="Transparent background"
        checked={transparentBg}
        onCheckedChange={(checked) => updateConfig({ transparentBg: checked })}
        disabled={disabled}
      />
    </div>
  )
}
