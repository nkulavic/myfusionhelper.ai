'use client'

// Go keys: basic_match, to_stage, opportunity_count (enum: first|all), found_goal, not_found_goal
import { FormTextField, FormSelect } from './form-fields'
import type { ConfigFormProps } from './types'

export function StageItForm({ config, onChange, disabled }: ConfigFormProps) {
  const basicMatch = (config.basicMatch as string) || ''
  const toStage = (config.toStage as string) || ''
  const opportunityCount = (config.opportunityCount as string) || 'first'
  const foundGoal = (config.foundGoal as string) || ''
  const notFoundGoal = (config.notFoundGoal as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Basic Match"
        placeholder="Match expression for opportunities"
        value={basicMatch}
        onChange={(e) => updateConfig({ basicMatch: e.target.value })}
        disabled={disabled}
        description="A basic match string to find the target opportunity."
      />

      <FormTextField
        label="To Stage"
        placeholder="Target stage name or ID"
        value={toStage}
        onChange={(e) => updateConfig({ toStage: e.target.value })}
        disabled={disabled}
        description="The stage to move the matched opportunity to."
      />

      <FormSelect
        label="Opportunity Count"
        description="Whether to update the first matching opportunity or all matches."
        value={opportunityCount}
        onValueChange={(v) => updateConfig({ opportunityCount: v })}
        options={[
          { value: 'first', label: 'First match only' },
          { value: 'all', label: 'All matches' },
        ]}
        disabled={disabled}
      />

      <FormTextField
        label="Found Goal"
        placeholder="API goal name (optional)"
        value={foundGoal}
        onChange={(e) => updateConfig({ foundGoal: e.target.value })}
        disabled={disabled}
        description="API goal to trigger when a matching opportunity is found. Leave blank to skip."
      />

      <FormTextField
        label="Not Found Goal"
        placeholder="API goal name (optional)"
        value={notFoundGoal}
        onChange={(e) => updateConfig({ notFoundGoal: e.target.value })}
        disabled={disabled}
        description="API goal to trigger when no matching opportunity is found. Leave blank to skip."
      />
    </div>
  )
}
