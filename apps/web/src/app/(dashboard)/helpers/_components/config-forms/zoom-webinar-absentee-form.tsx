'use client'

// Schema: see schemas.ts > zoomWebinarAbsenteeSchema
import { TagPicker } from '@/components/tag-picker'
import { FormCheckbox, FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function ZoomWebinarAbsenteeForm({
  config,
  onChange,
  disabled,
  platformId,
  connectionId,
}: ConfigFormProps) {
  const webinarId = (config.webinarId as string) || ''
  const tagPrefix = (config.tagPrefix as string) || 'Webinar'
  const applyNoShowTag = (config.applyNoShowTag as boolean) ?? true
  const applyRegisteredTag = (config.applyRegisteredTag as boolean) ?? true
  const customTags = (config.customTags as string[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Webinar ID"
        placeholder="Zoom webinar ID (optional)"
        description="Leave empty to apply tags regardless of webinar ID"
        value={webinarId}
        onChange={(e) => updateConfig({ webinarId: e.target.value })}
        disabled={disabled}
      />

      <FormTextField
        label="Tag Prefix"
        placeholder="Webinar"
        description="Prefix added to all generated tags (e.g., 'Webinar No Show')"
        value={tagPrefix}
        onChange={(e) => updateConfig({ tagPrefix: e.target.value })}
        disabled={disabled}
      />

      <div className="space-y-3">
        <FormCheckbox
          label="Apply 'No Show' Tag"
          description="Tag contacts who registered but did not attend"
          checked={applyNoShowTag}
          onCheckedChange={(checked) => updateConfig({ applyNoShowTag: checked })}
          disabled={disabled}
        />

        <FormCheckbox
          label="Apply 'Registered' Tag"
          description="Tag contacts who registered for the webinar"
          checked={applyRegisteredTag}
          onCheckedChange={(checked) => updateConfig({ applyRegisteredTag: checked })}
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Custom Tags</label>
        <p className="text-xs text-muted-foreground">
          Additional tags to apply to all absentees
        </p>
        <TagPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={customTags}
          onChange={(value) => updateConfig({ customTags: value })}
          placeholder="Select custom tags..."
          disabled={disabled}
        />
      </div>
    </div>
  )
}
