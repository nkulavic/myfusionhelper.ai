'use client'

// Schema: see schemas.ts > zoomWebinarParticipantSchema
import { TagPicker } from '@/components/tag-picker'
import { FormCheckbox, FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function ZoomWebinarParticipantForm({
  config,
  onChange,
  disabled,
  platformId,
  connectionId,
}: ConfigFormProps) {
  const webinarId = (config.webinarId as string) || ''
  const attendedPercent = (config.attendedPercent as number) ?? 0
  const durationMinutes = (config.durationMinutes as number) ?? 0
  const tagPrefix = (config.tagPrefix as string) || 'Webinar'
  const highEngagementThreshold = (config.highEngagementThreshold as number) ?? 75
  const mediumEngagementThreshold = (config.mediumEngagementThreshold as number) ?? 50
  const applyAttendanceTag = (config.applyAttendanceTag as boolean) ?? true
  const applyEngagementTags = (config.applyEngagementTags as boolean) ?? true
  const applyDurationTag = (config.applyDurationTag as boolean) ?? false
  const customTags = (config.customTags as string[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Webinar ID"
        placeholder="Zoom webinar ID (optional)"
        description="Leave empty to process all webinar attendees"
        value={webinarId}
        onChange={(e) => updateConfig({ webinarId: e.target.value })}
        disabled={disabled}
      />

      <div className="grid grid-cols-2 gap-3">
        <FormTextField
          label="Attended Percent"
          type="number"
          placeholder="0"
          description="Percent of webinar attended (0-100)"
          value={attendedPercent.toString()}
          onChange={(e) => updateConfig({ attendedPercent: parseFloat(e.target.value) || 0 })}
          disabled={disabled}
          min={0}
          max={100}
        />

        <FormTextField
          label="Duration (Minutes)"
          type="number"
          placeholder="0"
          description="Minutes participant attended"
          value={durationMinutes.toString()}
          onChange={(e) => updateConfig({ durationMinutes: parseFloat(e.target.value) || 0 })}
          disabled={disabled}
          min={0}
        />
      </div>

      <FormTextField
        label="Tag Prefix"
        placeholder="Webinar"
        description="Prefix added to all generated tags (e.g., 'Webinar Attended')"
        value={tagPrefix}
        onChange={(e) => updateConfig({ tagPrefix: e.target.value })}
        disabled={disabled}
      />

      <div className="grid grid-cols-2 gap-3">
        <FormTextField
          label="High Engagement Threshold (%)"
          type="number"
          placeholder="75"
          description="Minimum % for high engagement tag"
          value={highEngagementThreshold.toString()}
          onChange={(e) =>
            updateConfig({ highEngagementThreshold: parseFloat(e.target.value) || 75 })
          }
          disabled={disabled}
          min={0}
          max={100}
        />

        <FormTextField
          label="Medium Engagement Threshold (%)"
          type="number"
          placeholder="50"
          description="Minimum % for medium engagement tag"
          value={mediumEngagementThreshold.toString()}
          onChange={(e) =>
            updateConfig({ mediumEngagementThreshold: parseFloat(e.target.value) || 50 })
          }
          disabled={disabled}
          min={0}
          max={100}
        />
      </div>

      <div className="space-y-3">
        <FormCheckbox
          label="Apply 'Attended' Tag"
          description="Tag contacts who attended the webinar"
          checked={applyAttendanceTag}
          onCheckedChange={(checked) => updateConfig({ applyAttendanceTag: checked })}
          disabled={disabled}
        />

        <FormCheckbox
          label="Apply Engagement Level Tags"
          description="Tag based on attendance percentage (High/Medium/Low Engagement)"
          checked={applyEngagementTags}
          onCheckedChange={(checked) => updateConfig({ applyEngagementTags: checked })}
          disabled={disabled}
        />

        <FormCheckbox
          label="Apply Duration Tags"
          description="Tag based on attendance duration buckets (&lt;15 min, 15-30 min, 30-60 min, 60+ min)"
          checked={applyDurationTag}
          onCheckedChange={(checked) => updateConfig({ applyDurationTag: checked })}
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Custom Tags</label>
        <p className="text-xs text-muted-foreground">
          Additional tags to apply to all participants
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
