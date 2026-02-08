'use client'

// Schema: see schemas.ts > webinarSchema
import { FieldPicker } from '@/components/field-picker'
import { TagPicker } from '@/components/tag-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function WebinarForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const webinarId = (config.webinarId as string) || ''
  const scheduleId = (config.scheduleId as string) || ''
  const firstNameField = (config.firstNameField as string) || ''
  const lastNameField = (config.lastNameField as string) || ''
  const emailField = (config.emailField as string) || ''
  const phoneField = (config.phoneField as string) || ''
  const registeredTags = (config.registeredTags as string[]) || []
  const attendedTags = (config.attendedTags as string[]) || []
  const joinUrlField = (config.joinUrlField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3">
        <FormTextField
          label="Webinar/Event ID"
          placeholder="Webinar ID"
          value={webinarId}
          onChange={(e) => updateConfig({ webinarId: e.target.value })}
          disabled={disabled}
        />
        <FormTextField
          label="Schedule/Session ID"
          placeholder="Schedule ID (optional)"
          value={scheduleId}
          onChange={(e) => updateConfig({ scheduleId: e.target.value })}
          disabled={disabled}
        />
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div className="grid gap-2">
          <label className="text-sm font-medium">First Name Field</label>
          <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={firstNameField} onChange={(value) => updateConfig({ firstNameField: value })} placeholder="First name..." disabled={disabled} />
        </div>
        <div className="grid gap-2">
          <label className="text-sm font-medium">Last Name Field</label>
          <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={lastNameField} onChange={(value) => updateConfig({ lastNameField: value })} placeholder="Last name..." disabled={disabled} />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div className="grid gap-2">
          <label className="text-sm font-medium">Email Field</label>
          <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={emailField} onChange={(value) => updateConfig({ emailField: value })} placeholder="Email..." disabled={disabled} />
        </div>
        <div className="grid gap-2">
          <label className="text-sm font-medium">Phone Field</label>
          <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={phoneField} onChange={(value) => updateConfig({ phoneField: value })} placeholder="Phone (optional)..." disabled={disabled} />
        </div>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Join URL Field</label>
        <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={joinUrlField} onChange={(value) => updateConfig({ joinUrlField: value })} placeholder="Store join URL..." disabled={disabled} />
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Registered Tags</label>
        <TagPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={registeredTags} onChange={(value) => updateConfig({ registeredTags: value })} placeholder="Tags on registration..." disabled={disabled} />
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Attended Tags</label>
        <TagPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={attendedTags} onChange={(value) => updateConfig({ attendedTags: value })} placeholder="Tags on attendance..." disabled={disabled} />
      </div>
    </div>
  )
}
