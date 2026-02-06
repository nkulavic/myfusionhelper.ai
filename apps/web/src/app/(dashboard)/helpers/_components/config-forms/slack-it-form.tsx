'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
      <div className="grid gap-2">
        <Label htmlFor="slack-webhook">Slack Webhook URL</Label>
        <Input
          id="slack-webhook"
          type="url"
          placeholder="https://hooks.slack.com/services/..."
          value={webhookUrl}
          onChange={(e) => updateConfig({ webhookUrl: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Create an incoming webhook in your Slack workspace settings.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="slack-channel">Channel (optional override)</Label>
        <Input
          id="slack-channel"
          placeholder="#general or @username"
          value={channel}
          onChange={(e) => updateConfig({ channel: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Override the webhook&apos;s default channel. Leave empty to use the webhook default.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="slack-username">Bot Name (optional)</Label>
        <Input
          id="slack-username"
          placeholder="e.g. MyFusion Helper"
          value={username}
          onChange={(e) => updateConfig({ username: e.target.value })}
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <Label htmlFor="slack-message">Message Template</Label>
        <textarea
          id="slack-message"
          rows={4}
          placeholder={'New contact: {{first_name}} {{last_name}} ({{email}})'}
          value={message}
          onChange={(e) => updateConfig({ message: e.target.value })}
          disabled={disabled}
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        />
        <p className="text-xs text-muted-foreground">
          Use {'{{field_name}}'} tokens to include contact data. Supports Slack markdown formatting.
        </p>
      </div>
    </div>
  )
}
