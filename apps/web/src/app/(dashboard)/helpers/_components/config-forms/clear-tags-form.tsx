'use client'

// Schema: see schemas.ts > clearTagsSchema
import { TagPicker } from '@/components/tag-picker'
import { FormSelect, FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

const modeOptions = [
  { value: 'all', label: 'Remove all tags' },
  { value: 'specific', label: 'Remove specific tags' },
  { value: 'prefix', label: 'Remove tags matching a prefix' },
  { value: 'category', label: 'Remove tags in a category' },
]

export function ClearTagsForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const mode = (config.mode as string) || 'all'
  const tagIds = (config.tagIds as string[]) || []
  const prefix = (config.prefix as string) || ''
  const category = (config.category as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Clear Mode"
        description={
          mode === 'all'
            ? 'Remove every tag from the contact.'
            : mode === 'specific'
            ? 'Remove only the selected tags from the contact.'
            : mode === 'prefix'
            ? 'Remove tags whose name starts with a specific prefix.'
            : 'Remove tags belonging to a specific category.'
        }
        value={mode}
        onValueChange={(v) => updateConfig({ mode: v })}
        options={modeOptions}
        disabled={disabled}
      />

      {mode === 'specific' && (
        <div className="grid gap-2">
          <label className="text-sm font-medium">Tags to Remove</label>
          <TagPicker
            platformId={platformId ?? ''}
            connectionId={connectionId ?? ''}
            value={tagIds}
            onChange={(value) => updateConfig({ tagIds: value })}
            placeholder="Select tags to remove..."
            disabled={disabled}
          />
          <p className="text-xs text-muted-foreground">
            Only these specific tags will be removed from the contact.
          </p>
        </div>
      )}

      {mode === 'prefix' && (
        <FormTextField
          label="Tag Prefix"
          placeholder="e.g. campaign_"
          value={prefix}
          onChange={(e) => updateConfig({ prefix: e.target.value })}
          disabled={disabled}
          description="All tags starting with this prefix will be removed."
        />
      )}

      {mode === 'category' && (
        <FormTextField
          label="Tag Category"
          placeholder="e.g. marketing"
          value={category}
          onChange={(e) => updateConfig({ category: e.target.value })}
          disabled={disabled}
          description="All tags in this category will be removed."
        />
      )}

      {mode === 'all' && (
        <div className="rounded-md border border-warning/30 bg-warning/5 p-3">
          <p className="text-xs text-warning">
            Warning: This will remove ALL tags from the contact. Use with caution.
          </p>
        </div>
      )}
    </div>
  )
}
