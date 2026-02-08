'use client'

// Go keys: engagement_type (enum: opens|clicks|sends|all), lookback_days,
//   thresholds {highly_engaged, engaged},
//   tags {highly_engaged_tag, engaged_tag, disengaged_tag},
//   score_field, last_engagement_field

import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormSelect, FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

const engagementTypeOptions = [
  { value: 'opens', label: 'Opens' },
  { value: 'clicks', label: 'Clicks' },
  { value: 'sends', label: 'Sends' },
  { value: 'all', label: 'All' },
]

export function EmailEngagementForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const engagementType = (config.engagementType as string) || 'all'
  const lookbackDays = (config.lookbackDays as number) ?? 90
  const thresholds = (config.thresholds as Record<string, number>) || {}
  const tags = (config.tags as Record<string, string>) || {}
  const scoreField = (config.scoreField as string) || ''
  const lastEngagementField = (config.lastEngagementField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const updateThresholds = (key: string, value: number) => {
    updateConfig({ thresholds: { ...thresholds, [key]: value } })
  }

  const updateTags = (key: string, value: string) => {
    updateConfig({ tags: { ...tags, [key]: value } })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Engagement Type"
        value={engagementType}
        onValueChange={(value) => updateConfig({ engagementType: value })}
        options={engagementTypeOptions}
        disabled={disabled}
        description="Which type of email engagement to measure."
      />

      <FormTextField
        label="Lookback Days"
        type="number"
        min={1}
        value={String(lookbackDays)}
        onChange={(e) => updateConfig({ lookbackDays: Number(e.target.value) || 0 })}
        disabled={disabled}
        description="Number of days to look back for engagement data."
      />

      {/* Thresholds section */}
      <div className="space-y-3 rounded-lg border p-4">
        <Label className="text-sm font-semibold">Thresholds</Label>
        <p className="text-xs text-muted-foreground">
          Minimum engagement counts to qualify for each tier.
        </p>
        <FormTextField
          label="Highly Engaged"
          type="number"
          min={0}
          value={String(thresholds.highlyEngaged ?? 10)}
          onChange={(e) => updateThresholds('highlyEngaged', Number(e.target.value) || 0)}
          disabled={disabled}
          description="Minimum engagements for the highly engaged tier."
        />
        <FormTextField
          label="Engaged"
          type="number"
          min={0}
          value={String(thresholds.engaged ?? 3)}
          onChange={(e) => updateThresholds('engaged', Number(e.target.value) || 0)}
          disabled={disabled}
          description="Minimum engagements for the engaged tier."
        />
      </div>

      {/* Tags section */}
      <div className="space-y-3 rounded-lg border p-4">
        <Label className="text-sm font-semibold">Tags</Label>
        <p className="text-xs text-muted-foreground">
          Tags applied to contacts based on their engagement tier.
        </p>
        <FormTextField
          label="Highly Engaged Tag"
          placeholder="e.g. highly-engaged"
          value={tags.highlyEngagedTag ?? ''}
          onChange={(e) => updateTags('highlyEngagedTag', e.target.value)}
          disabled={disabled}
        />
        <FormTextField
          label="Engaged Tag"
          placeholder="e.g. engaged"
          value={tags.engagedTag ?? ''}
          onChange={(e) => updateTags('engagedTag', e.target.value)}
          disabled={disabled}
        />
        <FormTextField
          label="Disengaged Tag"
          placeholder="e.g. disengaged"
          value={tags.disengagedTag ?? ''}
          onChange={(e) => updateTags('disengagedTag', e.target.value)}
          disabled={disabled}
        />
      </div>

      {/* Field pickers */}
      <div className="grid gap-2">
        <Label>Score Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={scoreField}
          onChange={(value) => updateConfig({ scoreField: value })}
          placeholder="Field to store engagement score..."
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <Label>Last Engagement Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={lastEngagementField}
          onChange={(value) => updateConfig({ lastEngagementField: value })}
          placeholder="Field to store last engagement date..."
          disabled={disabled}
        />
      </div>
    </div>
  )
}
