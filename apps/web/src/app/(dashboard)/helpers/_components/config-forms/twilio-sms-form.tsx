'use client'

// Go keys: account_sid, auth_token, from_number, message_template, to_field
import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormTextField, FormTextArea } from './form-fields'
import type { ConfigFormProps } from './types'

export function TwilioSmsForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const accountSid = (config.accountSid as string) || ''
  const authToken = (config.authToken as string) || ''
  const fromNumber = (config.fromNumber as string) || ''
  const messageTemplate = (config.messageTemplate as string) || ''
  const toField = (config.toField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Account SID"
        placeholder="Your Twilio Account SID"
        value={accountSid}
        onChange={(e) => updateConfig({ accountSid: e.target.value })}
        disabled={disabled}
        description="Your Twilio account SID."
      />

      <FormTextField
        label="Auth Token"
        placeholder="Your Twilio Auth Token"
        value={authToken}
        onChange={(e) => updateConfig({ authToken: e.target.value })}
        disabled={disabled}
        type="password"
        description="Your Twilio auth token for authentication."
      />

      <FormTextField
        label="From Number"
        placeholder="+1234567890"
        value={fromNumber}
        onChange={(e) => updateConfig({ fromNumber: e.target.value })}
        disabled={disabled}
        description="Your Twilio phone number to send from."
      />

      <div className="grid gap-2">
        <Label>To Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={toField}
          onChange={(value) => updateConfig({ toField: value })}
          placeholder="Select phone number field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field containing the recipient phone number.
        </p>
      </div>

      <FormTextArea
        label="Message Template"
        placeholder="Hi ~Contact.FirstName~, ..."
        value={messageTemplate}
        onChange={(e) => updateConfig({ messageTemplate: e.target.value })}
        disabled={disabled}
        rows={4}
        description="Supports merge fields. Max 1600 characters."
      />
    </div>
  )
}
