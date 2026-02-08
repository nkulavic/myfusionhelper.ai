'use client'

// Schema: see schemas.ts > hubspotWorkflowTriggerSchema
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function HubspotWorkflowTriggerForm({ config, onChange, disabled }: ConfigFormProps) {
  const workflowId = (config.workflowId as string) || ''
  const enrollmentEmail = (config.enrollmentEmail as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Workflow ID"
        description="The ID of the HubSpot workflow to trigger."
        placeholder="HubSpot workflow ID"
        value={workflowId}
        onChange={(e) => updateConfig({ workflowId: e.target.value })}
        disabled={disabled}
      />
      <FormTextField
        label="Enrollment Email"
        description="Optional email for workflow enrollment context."
        placeholder="contact@example.com (optional)"
        value={enrollmentEmail}
        onChange={(e) => updateConfig({ enrollmentEmail: e.target.value })}
        disabled={disabled}
      />
    </div>
  )
}
