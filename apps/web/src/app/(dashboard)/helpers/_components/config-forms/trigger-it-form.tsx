'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

export function TriggerItForm({ config, onChange, disabled }: ConfigFormProps) {
  const automationType = (config.automationType as string) || 'campaign'
  const automationId = (config.automationId as string) || ''
  const integration = (config.integration as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="trigger-type">Automation Type</Label>
        <select
          id="trigger-type"
          value={automationType}
          onChange={(e) => updateConfig({ automationType: e.target.value })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          <option value="campaign">Campaign Sequence</option>
          <option value="workflow">Workflow</option>
          <option value="goal">Campaign Goal</option>
        </select>
        <p className="text-xs text-muted-foreground">
          {automationType === 'campaign'
            ? 'Start a campaign sequence for the contact.'
            : automationType === 'workflow'
            ? 'Trigger a workflow or automation rule.'
            : 'Achieve a campaign goal for the contact.'}
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="trigger-automation-id">
          {automationType === 'goal' ? 'Goal Name' : 'Automation ID'}
        </Label>
        <Input
          id="trigger-automation-id"
          placeholder={
            automationType === 'goal'
              ? 'e.g. purchased_product'
              : 'e.g. 12345'
          }
          value={automationId}
          onChange={(e) => updateConfig({ automationId: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          {automationType === 'goal'
            ? 'The goal name as configured in your CRM campaign.'
            : 'The numeric ID of the campaign sequence or workflow.'}
        </p>
      </div>

      {automationType === 'goal' && (
        <div className="grid gap-2">
          <Label htmlFor="trigger-integration">Integration Name (optional)</Label>
          <Input
            id="trigger-integration"
            placeholder="e.g. my_app"
            value={integration}
            onChange={(e) => updateConfig({ integration: e.target.value })}
            disabled={disabled}
          />
          <p className="text-xs text-muted-foreground">
            Optional integration identifier for goal tracking.
          </p>
        </div>
      )}
    </div>
  )
}
