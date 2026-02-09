'use client'

// Schema: see schemas.ts > countdownTimerSchema

import { FieldPicker } from '@/components/field-picker'
import { FormSelect, FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

const timerModeOptions = [
  { value: 'dynamic', label: 'Dynamic (from now)' },
  { value: 'standard', label: 'Fixed Date' },
  { value: 'evergreen', label: 'Evergreen (persistent)' },
  { value: 'contact_field', label: 'Contact Field' },
]

const timerModeDescriptions: Record<string, string> = {
  dynamic: 'Timer expires after a set duration from when the helper executes',
  standard: 'Timer counts down to a specific fixed date and time',
  evergreen: 'Per-contact persistent timer stored in DynamoDB, reuses existing if not expired',
  contact_field: 'Timer reads expiry date from a contact custom field',
}

export function CountdownTimerForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const timerMode = (config.timerMode as string) || 'dynamic'
  const expireDate = (config.expireDate as string) || ''
  const durationHours = (config.durationHours as number) ?? 24
  const timerField = (config.timerField as string) || ''
  const saveUrlField = (config.saveUrlField as string) || ''
  const jwtSecret = (config.jwtSecret as string) || ''
  const timerUrlBase = (config.timerUrlBase as string) || 'https://app.myfusionhelper.ai/timer'

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Timer Mode"
        value={timerMode}
        onValueChange={(value) => updateConfig({ timerMode: value })}
        options={timerModeOptions}
        description={timerModeDescriptions[timerMode]}
        disabled={disabled}
      />

      {timerMode === 'dynamic' && (
        <FormTextField
          label="Duration (hours)"
          type="number"
          min={0}
          step={0.5}
          value={durationHours}
          onChange={(e) => updateConfig({ durationHours: parseFloat(e.target.value) || 24 })}
          disabled={disabled}
          description="Hours from now until timer expires"
        />
      )}

      {timerMode === 'standard' && (
        <FormTextField
          label="Expiry Date/Time"
          type="datetime-local"
          value={expireDate}
          onChange={(e) => updateConfig({ expireDate: e.target.value })}
          disabled={disabled}
          description="Fixed date/time when timer expires (ISO 8601 format)"
        />
      )}

      {timerMode === 'evergreen' && (
        <FormTextField
          label="Duration (hours)"
          type="number"
          min={0}
          step={0.5}
          value={durationHours}
          onChange={(e) => updateConfig({ durationHours: parseFloat(e.target.value) || 24 })}
          disabled={disabled}
          description="Duration for new timers (reuses existing timer if not expired)"
        />
      )}

      {timerMode === 'contact_field' && (
        <div className="grid gap-2">
          <label className="text-sm font-medium">Expiry Date Field</label>
          <FieldPicker
            platformId={platformId ?? ''}
            connectionId={connectionId ?? ''}
            value={timerField}
            onChange={(value) => updateConfig({ timerField: value })}
            placeholder="Select date field..."
            filterType="date"
            disabled={disabled}
          />
          <p className="text-xs text-muted-foreground">
            Contact field containing the expiry date/time (ISO 8601 format)
          </p>
        </div>
      )}

      <FormTextField
        label="JWT Secret"
        type="password"
        placeholder="Enter secure secret key"
        value={jwtSecret}
        onChange={(e) => updateConfig({ jwtSecret: e.target.value })}
        disabled={disabled}
        description="Secret key for JWT signing (required for secure timer URLs)"
      />

      <FormTextField
        label="Timer URL Base"
        type="url"
        placeholder="https://app.myfusionhelper.ai/timer"
        value={timerUrlBase}
        onChange={(e) => updateConfig({ timerUrlBase: e.target.value })}
        disabled={disabled}
        description="Base URL for timer display page"
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Save Timer URL To (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={saveUrlField}
          onChange={(value) => updateConfig({ saveUrlField: value })}
          placeholder="Select field to save timer URL..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          CRM field to save the generated timer URL (optional)
        </p>
      </div>
    </div>
  )
}
