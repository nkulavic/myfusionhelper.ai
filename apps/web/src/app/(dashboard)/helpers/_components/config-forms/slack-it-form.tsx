'use client'

// Schema: see schemas.ts > slackItSchema

import { FormTextField, FormTextArea } from './form-fields'
import type { ConfigFormProps } from './types'

export function SlackItForm({ config, onChange, disabled }: ConfigFormProps) {
  const channel = (config.channel as string) || ''
  const message = (config.message as string) || ''
  const webhookUrl = (config.webhookUrl as string) || ''
  const username = (config.username as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Slack Webhook URL"
        type="url"
        placeholder="https://hooks.slack.com/services/..."
        value={webhookUrl}
        onChange={(e) => updateConfig({ webhookUrl: e.target.value })}
        disabled={disabled}
        description="Create an incoming webhook in your Slack workspace settings."
      />

      <FormTextField
        label="Channel (optional override)"
        placeholder="#general or @username"
        value={channel}
        onChange={(e) => updateConfig({ channel: e.target.value })}
        disabled={disabled}
        description="Override the webhook's default channel. Leave empty to use the webhook default."
      />

      <FormTextField
        label="Bot Name (optional)"
        placeholder="e.g. MyFusion Helper"
        value={username}
        onChange={(e) => updateConfig({ username: e.target.value })}
        disabled={disabled}
      />

      <FormTextArea
        label="Message Template"
        rows={4}
        placeholder={'New contact: {{first_name}} {{last_name}} ({{email}})'}
        value={message}
        onChange={(e) => updateConfig({ message: e.target.value })}
        disabled={disabled}
        description="Use {{field_name}} tokens to include contact data. Supports Slack markdown formatting."
      />
    </div>
  )
}
