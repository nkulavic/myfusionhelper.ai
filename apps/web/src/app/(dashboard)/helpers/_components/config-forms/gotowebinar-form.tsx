'use client'

// Schema: see schemas.ts > gotowebinarSchema
import { TagPicker } from '@/components/tag-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function GotoWebinarForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const organizerKey = (config.organizer_key as string) || ''
  const webinarKey = (config.webinar_key as string) || ''
  const applyTag = (config.apply_tag as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Organizer Key"
        placeholder="Enter your GoToWebinar organizer key"
        value={organizerKey}
        onChange={(e) => updateConfig({ organizer_key: e.target.value })}
        disabled={disabled}
        description="Your GoToWebinar organizer key from your account settings."
        required
      />

      <FormTextField
        label="Webinar Key"
        placeholder="Enter the webinar key"
        value={webinarKey}
        onChange={(e) => updateConfig({ webinar_key: e.target.value })}
        disabled={disabled}
        description="The unique key for the specific webinar you want to register contacts to."
        required
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
