'use client'

// Schema: see schemas.ts > notifyMeSchema
import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { FormSelect, FormTextField, FormTextArea } from './form-fields'
import type { ConfigFormProps } from './types'

const channelOptions = [
  { value: 'email', label: 'Email' },
  { value: 'slack', label: 'Slack' },
  { value: 'webhook', label: 'Webhook' },
]

export function NotifyMeForm({ config, onChange, disabled }: ConfigFormProps) {
  const channel = (config.channel as string) || 'email'
  const recipient = (config.recipient as string) || ''
  const subject = (config.subject as string) || ''
  const message = (config.message as string) || ''
  const includeFields = (config.includeFields as string[]) || []
  const [newField, setNewField] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addField = () => {
    const trimmed = newField.trim()
    if (!trimmed || includeFields.includes(trimmed)) return
    updateConfig({ includeFields: [...includeFields, trimmed] })
    setNewField('')
  }

  const removeField = (field: string) => {
    updateConfig({ includeFields: includeFields.filter((f) => f !== field) })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Notification Channel"
        value={channel}
        onValueChange={(value) => updateConfig({ channel: value })}
        options={channelOptions}
        disabled={disabled}
      />

      <FormTextField
        label="Recipient"
        placeholder={channel === 'email' ? 'you@company.com' : channel === 'slack' ? '#channel or @username' : 'https://hooks.example.com/notify'}
        value={recipient}
        onChange={(e) => updateConfig({ recipient: e.target.value })}
        disabled={disabled}
        description="The recipient for the notification."
      />

      {channel === 'email' && (
        <FormTextField
          label="Subject"
          placeholder='e.g. "New lead: {{first_name}} {{last_name}}"'
          value={subject}
          onChange={(e) => updateConfig({ subject: e.target.value })}
          disabled={disabled}
          description="You can use contact field tokens like {{first_name}}."
        />
      )}

      <FormTextArea
        label="Message"
        rows={4}
        placeholder='e.g. "Contact {{first_name}} {{last_name}} ({{email}}) completed the onboarding flow."'
        value={message}
        onChange={(e) => updateConfig({ message: e.target.value })}
        disabled={disabled}
        description="Use {{field_name}} tokens to include contact data in the notification."
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Include Fields</label>
        <div className="flex gap-2">
          <Input
            placeholder="Enter a field name to include..."
            value={newField}
            onChange={(e) => setNewField(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault()
                addField()
              }
            }}
            disabled={disabled}
            className="flex-1"
          />
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={addField}
            disabled={disabled || !newField.trim()}
            aria-label="Add field"
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>
        {includeFields.length > 0 ? (
          <div className="flex flex-wrap gap-1.5 mt-1">
            {includeFields.map((field) => (
              <span
                key={field}
                className="inline-flex items-center gap-1 rounded-md bg-muted px-2 py-0.5 text-xs font-mono"
              >
                {field}
                {!disabled && (
                  <button
                    type="button"
                    onClick={() => removeField(field)}
                    className="ml-0.5 rounded p-0.5 hover:bg-accent"
                    aria-label={`Remove field ${field}`}
                  >
                    <X className="h-3 w-3" />
                  </button>
                )}
              </span>
            ))}
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">
            Add contact field names to include their values in the notification payload.
          </p>
        )}
      </div>
    </div>
  )
}
