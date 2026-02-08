'use client'

// Schema: see schemas.ts > queryItSchema
import { TagPicker } from '@/components/tag-picker'
import { FormTextField, FormSelect } from './form-fields'
import type { ConfigFormProps } from './types'

const actionTypes = [
  { value: 'tag', label: 'Apply Tags' },
  { value: 'goal', label: 'Achieve Goal' },
  { value: 'helper', label: 'Run Helper' },
]

export function QueryItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const savedSearchId = (config.savedSearchId as string) || ''
  const actionType = (config.actionType as string) || 'tag'
  const goalName = (config.goalName as string) || ''
  const actionTags = (config.actionTags as string[]) || []
  const batchSize = (config.batchSize as number) ?? 100

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Saved Search ID"
        placeholder="Saved search or report ID"
        value={savedSearchId}
        onChange={(e) => updateConfig({ savedSearchId: e.target.value })}
        disabled={disabled}
      />

      <FormSelect
        label="Action Type"
        value={actionType}
        onValueChange={(value) => updateConfig({ actionType: value })}
        options={actionTypes}
        disabled={disabled}
      />

      {actionType === 'tag' && (
        <div className="grid gap-2">
          <label className="text-sm font-medium">Tags to Apply</label>
          <TagPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={actionTags} onChange={(value) => updateConfig({ actionTags: value })} placeholder="Select tags..." disabled={disabled} />
        </div>
      )}

      {actionType === 'goal' && (
        <FormTextField
          label="Goal Name"
          placeholder="API goal name"
          value={goalName}
          onChange={(e) => updateConfig({ goalName: e.target.value })}
          disabled={disabled}
        />
      )}

      <FormTextField
        label="Batch Size"
        type="number"
        min={1}
        max={1000}
        value={batchSize}
        onChange={(e) => updateConfig({ batchSize: parseInt(e.target.value, 10) || 100 })}
        disabled={disabled}
        description="Number of contacts to process per batch."
      />
    </div>
  )
}
