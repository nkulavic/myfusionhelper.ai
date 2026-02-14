'use client'

// Schema: see schemas.ts > searchItSchema
import { FieldPicker } from '@/components/field-picker'
import { TagPicker } from '@/components/tag-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function SearchItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const savedSearchId = (config.savedSearchId as string) || ''
  const resultField = (config.resultField as string) || ''
  const matchTags = (config.matchTags as string[]) || []
  const goalName = (config.goalName as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Saved Search ID"
        placeholder="Saved search ID"
        value={savedSearchId}
        onChange={(e) => updateConfig({ savedSearchId: e.target.value })}
        disabled={disabled}
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Result Count Field</label>
        <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={resultField} onChange={(value) => updateConfig({ resultField: value })} placeholder="Store match count..." disabled={disabled} />
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Match Tags</label>
        <TagPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={matchTags} onChange={(value) => updateConfig({ matchTags: value })} placeholder="Tags for matching contacts..." disabled={disabled} />
      </div>

      <FormTextField
        label="API Goal"
        placeholder="Goal on match (optional)"
        value={goalName}
        onChange={(e) => updateConfig({ goalName: e.target.value })}
        disabled={disabled}
      />
    </div>
  )
}
