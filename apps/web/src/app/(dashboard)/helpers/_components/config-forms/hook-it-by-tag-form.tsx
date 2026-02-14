'use client'

// Hook It By Tag (Conditional)
// Go keys: required_tags (array), forbidden_tags (array), match_mode (string), goal_name (string), integration (string), apply_tag_on_success (string), apply_tag_on_skip (string)

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import {
  FormTextField,
  FormSelect,
  DynamicList,
  AddItemRow,
  InfoBanner,
} from './form-fields'
import type { ConfigFormProps } from './types'

export function HookItByTagForm({ config, onChange, disabled }: ConfigFormProps) {
  const requiredTags = (config.requiredTags as string[]) || []
  const forbiddenTags = (config.forbiddenTags as string[]) || []
  const matchMode = (config.matchMode as string) || 'all'
  const goalName = (config.goalName as string) || ''
  const integration = (config.integration as string) || 'myfusionhelper'
  const applyTagOnSuccess = (config.applyTagOnSuccess as string) || ''
  const applyTagOnSkip = (config.applyTagOnSkip as string) || ''

  const [newRequiredTag, setNewRequiredTag] = useState('')
  const [newForbiddenTag, setNewForbiddenTag] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <InfoBanner>
        Conditionally process webhooks based on contact tags. Only processes if contact has required tags and lacks forbidden tags. Useful for filtering webhooks to specific contact segments.
      </InfoBanner>

      <DynamicList<string>
        label="Required Tags"
        description="Tag IDs the contact must have for webhook to process."
        items={requiredTags}
        onItemsChange={(items) => updateConfig({ requiredTags: items })}
        disabled={disabled}
        renderAddForm={(onAdd) => (
          <AddItemRow
            onAdd={() => {
              const tag = newRequiredTag.trim()
              if (!tag) return
              onAdd(tag)
              setNewRequiredTag('')
            }}
            disabled={disabled}
            canAdd={!!newRequiredTag.trim()}
          >
            <Input
              placeholder="Tag ID"
              value={newRequiredTag}
              onChange={(e) => setNewRequiredTag(e.target.value)}
              disabled={disabled}
              className="flex-1"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  const tag = newRequiredTag.trim()
                  if (!tag) return
                  onAdd(tag)
                  setNewRequiredTag('')
                }
              }}
            />
          </AddItemRow>
        )}
        renderItem={(tag) => (
          <span className="font-mono text-xs">
            <span className="font-medium">Tag: {tag}</span>
          </span>
        )}
      />

      <FormSelect
        label="Required Tags Match Mode"
        description="For required tags: 'all' = must have all tags, 'any' = must have at least one tag."
        value={matchMode}
        onValueChange={(value) => updateConfig({ matchMode: value })}
        options={[
          { value: 'all', label: 'All (contact must have all required tags)' },
          { value: 'any', label: 'Any (contact must have at least one required tag)' },
        ]}
        disabled={disabled}
      />

      <DynamicList<string>
        label="Forbidden Tags"
        description="Tag IDs the contact must NOT have for webhook to process."
        items={forbiddenTags}
        onItemsChange={(items) => updateConfig({ forbiddenTags: items })}
        disabled={disabled}
        renderAddForm={(onAdd) => (
          <AddItemRow
            onAdd={() => {
              const tag = newForbiddenTag.trim()
              if (!tag) return
              onAdd(tag)
              setNewForbiddenTag('')
            }}
            disabled={disabled}
            canAdd={!!newForbiddenTag.trim()}
          >
            <Input
              placeholder="Tag ID"
              value={newForbiddenTag}
              onChange={(e) => setNewForbiddenTag(e.target.value)}
              disabled={disabled}
              className="flex-1"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  const tag = newForbiddenTag.trim()
                  if (!tag) return
                  onAdd(tag)
                  setNewForbiddenTag('')
                }
              }}
            />
          </AddItemRow>
        )}
        renderItem={(tag) => (
          <span className="font-mono text-xs">
            <span className="font-medium">Tag: {tag}</span>
          </span>
        )}
      />

      <FormTextField
        label="Goal Name (Optional)"
        placeholder="e.g., webhook_processed"
        value={goalName}
        onChange={(e) => updateConfig({ goalName: e.target.value })}
        disabled={disabled}
        description="Goal to achieve if tag conditions are met and webhook is processed."
      />

      <FormTextField
        label="Integration Name"
        placeholder="myfusionhelper"
        value={integration}
        onChange={(e) => updateConfig({ integration: e.target.value })}
        disabled={disabled}
        description="Integration name for goal calls. Defaults to 'myfusionhelper'."
      />

      <FormTextField
        label="Success Tag ID (Optional)"
        placeholder="Tag ID"
        value={applyTagOnSuccess}
        onChange={(e) => updateConfig({ applyTagOnSuccess: e.target.value })}
        disabled={disabled}
        description="Tag ID to apply if webhook is processed successfully (tag conditions met)."
      />

      <FormTextField
        label="Skip Tag ID (Optional)"
        placeholder="Tag ID"
        value={applyTagOnSkip}
        onChange={(e) => updateConfig({ applyTagOnSkip: e.target.value })}
        disabled={disabled}
        description="Tag ID to apply if webhook is skipped due to tag conditions not being met."
      />
    </div>
  )
}
