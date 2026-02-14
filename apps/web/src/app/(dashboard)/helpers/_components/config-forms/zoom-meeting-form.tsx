'use client'

// Schema: see schemas.ts > zoomMeetingSchema

import { FieldPicker } from '@/components/field-picker'
import { FormSelect, FormTextField, FormSwitch } from './form-fields'
import type { ConfigFormProps } from './types'

const actionOptions = [
  { value: 'create', label: 'Create New Meeting' },
  { value: 'register', label: 'Register for Existing Meeting' },
]

const registrationTypeOptions = [
  { value: '1', label: 'Once' },
  { value: '2', label: 'Each Occurrence' },
  { value: '3', label: 'Series' },
]

const autoRecordingOptions = [
  { value: 'none', label: 'No Recording' },
  { value: 'local', label: 'Local Recording' },
  { value: 'cloud', label: 'Cloud Recording' },
]

export function ZoomMeetingForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const action = (config.action as string) || 'create'
  const userId = (config.userId as string) || ''
  const meetingId = (config.meetingId as string) || ''
  const topic = (config.topic as string) || ''
  const startTime = (config.startTime as string) || ''
  const duration = (config.duration as number) ?? 60
  const timezone = (config.timezone as string) || 'UTC'
  const password = (config.password as string) || ''
  const nameField = (config.nameField as string) || 'FirstName'
  const emailField = (config.emailField as string) || 'Email'
  const saveJoinUrlTo = (config.saveJoinUrlTo as string) || ''
  const registrationType = String((config.registrationType as number) ?? 1)
  const autoRecording = (config.autoRecording as string) || 'none'

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Action"
        value={action}
        onValueChange={(value) => updateConfig({ action: value })}
        options={actionOptions}
        description={
          action === 'create'
            ? 'Create a new Zoom meeting and register the contact'
            : 'Register contact for an existing meeting'
        }
        disabled={disabled}
      />

      {action === 'create' && (
        <>
          <FormTextField
            label="Zoom User ID"
            placeholder="user@example.com or user ID"
            value={userId}
            onChange={(e) => updateConfig({ userId: e.target.value })}
            disabled={disabled}
            description="Email address or Zoom user ID of the meeting host (required for create action)"
          />

          <FormTextField
            label="Meeting Topic"
            placeholder="Weekly Team Meeting"
            value={topic}
            onChange={(e) => updateConfig({ topic: e.target.value })}
            disabled={disabled}
          />

          <FormTextField
            label="Start Time"
            type="datetime-local"
            value={startTime}
            onChange={(e) => updateConfig({ startTime: e.target.value })}
            disabled={disabled}
            description="Meeting start time (ISO 8601 format)"
          />

          <div className="grid grid-cols-2 gap-3">
            <FormTextField
              label="Duration (minutes)"
              type="number"
              min={1}
              value={duration}
              onChange={(e) => updateConfig({ duration: parseInt(e.target.value, 10) || 60 })}
              disabled={disabled}
            />

            <FormTextField
              label="Timezone"
              placeholder="UTC, America/Los_Angeles, etc."
              value={timezone}
              onChange={(e) => updateConfig({ timezone: e.target.value })}
              disabled={disabled}
            />
          </div>

          <FormTextField
            label="Meeting Password (optional)"
            type="password"
            placeholder="Optional meeting password"
            value={password}
            onChange={(e) => updateConfig({ password: e.target.value })}
            disabled={disabled}
          />

          <FormSelect
            label="Auto Recording"
            value={autoRecording}
            onValueChange={(value) => updateConfig({ autoRecording: value })}
            options={autoRecordingOptions}
            disabled={disabled}
          />
        </>
      )}

      {action === 'register' && (
        <FormTextField
          label="Meeting ID"
          placeholder="123456789"
          value={meetingId}
          onChange={(e) => updateConfig({ meetingId: e.target.value })}
          disabled={disabled}
          description="Existing Zoom meeting ID to register the contact for"
        />
      )}

      <div className="grid grid-cols-2 gap-3">
        <div className="grid gap-2">
          <label className="text-sm font-medium">Name Field</label>
          <FieldPicker
            platformId={platformId ?? ''}
            connectionId={connectionId ?? ''}
            value={nameField}
            onChange={(value) => updateConfig({ nameField: value })}
            placeholder="Select name field..."
            disabled={disabled}
          />
        </div>

        <div className="grid gap-2">
          <label className="text-sm font-medium">Email Field</label>
          <FieldPicker
            platformId={platformId ?? ''}
            connectionId={connectionId ?? ''}
            value={emailField}
            onChange={(value) => updateConfig({ emailField: value })}
            placeholder="Select email field..."
            filterType="email"
            disabled={disabled}
          />
        </div>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Save Join URL To (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={saveJoinUrlTo}
          onChange={(value) => updateConfig({ saveJoinUrlTo: value })}
          placeholder="Select field to save join URL..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          CRM field to save the meeting join URL (optional)
        </p>
      </div>

      <FormSelect
        label="Registration Type"
        value={registrationType}
        onValueChange={(value) => updateConfig({ registrationType: parseInt(value, 10) })}
        options={registrationTypeOptions}
        description="How registrants can join: once, each occurrence, or entire series"
        disabled={disabled}
      />
    </div>
  )
}
