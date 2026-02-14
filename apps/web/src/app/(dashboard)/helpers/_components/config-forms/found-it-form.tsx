'use client'

// Go keys: check_field, found_tag_id (string), not_found_tag_id (string), found_goal, not_found_goal
import { FieldPicker } from '@/components/field-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function FoundItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const checkField = (config.checkField as string) || ''
  const foundTagId = (config.foundTagId as string) || ''
  const notFoundTagId = (config.notFoundTagId as string) || ''
  const foundGoal = (config.foundGoal as string) || ''
  const notFoundGoal = (config.notFoundGoal as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Field to Check</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={checkField}
          onChange={(value) => updateConfig({ checkField: value })}
          placeholder="Select field to check..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field to check for a value.
        </p>
      </div>

      <FormTextField
        label="Found Tag ID"
        placeholder="Tag ID to apply when value is found"
        value={foundTagId}
        onChange={(e) => updateConfig({ foundTagId: e.target.value })}
        disabled={disabled}
        description="The tag ID to apply when the field has a value. Leave blank to skip."
      />

      <FormTextField
        label="Not Found Tag ID"
        placeholder="Tag ID to apply when value is not found"
        value={notFoundTagId}
        onChange={(e) => updateConfig({ notFoundTagId: e.target.value })}
        disabled={disabled}
        description="The tag ID to apply when the field is empty. Leave blank to skip."
      />

      <FormTextField
        label="Found Goal"
        placeholder="API goal name"
        value={foundGoal}
        onChange={(e) => updateConfig({ foundGoal: e.target.value })}
        disabled={disabled}
        description="API goal to trigger when the field has a value. Leave blank to skip."
      />

      <FormTextField
        label="Not Found Goal"
        placeholder="API goal name"
        value={notFoundGoal}
        onChange={(e) => updateConfig({ notFoundGoal: e.target.value })}
        disabled={disabled}
        description="API goal to trigger when the field is empty. Leave blank to skip."
      />
    </div>
  )
}
