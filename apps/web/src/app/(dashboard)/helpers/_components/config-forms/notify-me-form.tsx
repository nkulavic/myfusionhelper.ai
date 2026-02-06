'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

export function NotifyMeForm({ config, onChange, disabled }: ConfigFormProps) {
  const channel = (config.channel as string) || 'email'
  const recipient = (config.recipient as string) || ''
  const subject = (config.subject as string) || ''
  const message = (config.message as string) || ''
  const webhookUrl = (config.webhookUrl as string) || ''
  const slackChannel = (config.slackChannel as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="notify-channel">Notification Channel</Label>
        <select
          id="notify-channel"
          value={channel}
          onChange={(e) => updateConfig({ channel: e.target.value })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          <option value="email">Email</option>
          <option value="slack">Slack</option>
          <option value="webhook">Webhook</option>
        </select>
      </div>

      {channel === 'email' && (
        <>
          <div className="grid gap-2">
            <Label htmlFor="notify-recipient">Recipient Email</Label>
            <Input
              id="notify-recipient"
              type="email"
              placeholder="you@company.com"
              value={recipient}
              onChange={(e) => updateConfig({ recipient: e.target.value })}
              disabled={disabled}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="notify-subject">Email Subject</Label>
            <Input
              id="notify-subject"
              placeholder='e.g. "New lead: {{first_name}} {{last_name}}"'
              value={subject}
              onChange={(e) => updateConfig({ subject: e.target.value })}
              disabled={disabled}
            />
            <p className="text-xs text-muted-foreground">
              You can use contact field tokens like {'{{first_name}}'}.
            </p>
          </div>
        </>
      )}

      {channel === 'slack' && (
        <div className="grid gap-2">
          <Label htmlFor="notify-slack">Slack Channel</Label>
          <Input
            id="notify-slack"
            placeholder="#general or @username"
            value={slackChannel}
            onChange={(e) => updateConfig({ slackChannel: e.target.value })}
            disabled={disabled}
          />
        </div>
      )}

      {channel === 'webhook' && (
        <div className="grid gap-2">
          <Label htmlFor="notify-webhook">Webhook URL</Label>
          <Input
            id="notify-webhook"
            type="url"
            placeholder="https://hooks.example.com/notify"
            value={webhookUrl}
            onChange={(e) => updateConfig({ webhookUrl: e.target.value })}
            disabled={disabled}
          />
        </div>
      )}

      <div className="grid gap-2">
        <Label htmlFor="notify-message">Message Template</Label>
        <textarea
          id="notify-message"
          rows={4}
          placeholder='e.g. "Contact {{first_name}} {{last_name}} ({{email}}) completed the onboarding flow."'
          value={message}
          onChange={(e) => updateConfig({ message: e.target.value })}
          disabled={disabled}
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        />
        <p className="text-xs text-muted-foreground">
          Use {'{{field_name}}'} tokens to include contact data in the notification.
        </p>
      </div>
    </div>
  )
}
