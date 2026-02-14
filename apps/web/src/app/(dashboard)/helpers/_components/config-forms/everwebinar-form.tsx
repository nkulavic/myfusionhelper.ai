'use client'

// Schema: see schemas.ts > everwebinarSchema
import { TagPicker } from '@/components/tag-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function EverWebinarForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const webinarId = (config.webinar_id as string) || ''
  const schedule = (config.schedule as string) || ''
  const applyTag = (config.apply_tag as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Webinar ID"
        placeholder="Enter your EverWebinar ID"
        value={webinarId}
        onChange={(e) => updateConfig({ webinar_id: e.target.value })}
        disabled={disabled}
        description="The unique ID for your EverWebinar automated webinar."
        required
      />

      <FormTextField
        label="Schedule"
        placeholder="Enter schedule identifier (optional)"
        value={schedule}
        onChange={(e) => updateConfig({ schedule: e.target.value })}
        disabled={disabled}
        description="Optional schedule identifier for the webinar session."
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">
          Apply Tag After Registration <span className="text-muted-foreground">(optional)</span>
        </label>
        <TagPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={applyTag ? [applyTag] : []}
          onChange={(value) => updateConfig({ apply_tag: value[0] || '' })}
          placeholder="Select tag to apply..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Tag to apply to the contact after successful webinar registration.
        </p>
      </div>
    </div>
  )
}
