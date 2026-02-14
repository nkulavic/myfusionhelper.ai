'use client'

// Go keys: day, time, save_time_zone, save_lat_lng, save_time_zone_offset, trigger_goal, failed_goal

import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function TimezoneTriggersForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const day = (config.day as string) || ''
  const time = (config.time as string) || ''
  const saveTimeZone = (config.saveTimeZone as string) || ''
  const saveLatLng = (config.saveLatLng as string) || ''
  const saveTimeZoneOffset = (config.saveTimeZoneOffset as string) || ''
  const triggerGoal = (config.triggerGoal as string) || ''
  const failedGoal = (config.failedGoal as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Day"
        placeholder="e.g. Monday"
        value={day}
        onChange={(e) => updateConfig({ day: e.target.value })}
        disabled={disabled}
        description="Day of the week to trigger (e.g. Monday, Tuesday)."
      />

      <FormTextField
        label="Time"
        placeholder="e.g. 09:00"
        value={time}
        onChange={(e) => updateConfig({ time: e.target.value })}
        disabled={disabled}
        description="Time to trigger in the contact's timezone."
      />

      <div className="grid gap-2">
        <Label>Save Time Zone Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={saveTimeZone}
          onChange={(value) => updateConfig({ saveTimeZone: value })}
          placeholder="Field to store the contact's timezone..."
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <Label>Save Lat/Lng Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={saveLatLng}
          onChange={(value) => updateConfig({ saveLatLng: value })}
          placeholder="Field to store latitude/longitude..."
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <Label>Save Time Zone Offset Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={saveTimeZoneOffset}
          onChange={(value) => updateConfig({ saveTimeZoneOffset: value })}
          placeholder="Field to store timezone offset..."
          disabled={disabled}
        />
      </div>

      <FormTextField
        label="Trigger Goal"
        placeholder="Goal name on success"
        value={triggerGoal}
        onChange={(e) => updateConfig({ triggerGoal: e.target.value })}
        disabled={disabled}
        description="API goal to trigger when the timezone matches."
      />

      <FormTextField
        label="Failed Goal"
        placeholder="Goal name on failure"
        value={failedGoal}
        onChange={(e) => updateConfig({ failedGoal: e.target.value })}
        disabled={disabled}
        description="API goal to trigger when timezone lookup fails."
      />
    </div>
  )
}
