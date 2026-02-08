'use client'

// Schema: see schemas.ts > mailItSchema

import { FormTextField, FormRadioGroup, FormTextArea } from './form-fields'
import type { ConfigFormProps } from './types'

export function MailItForm({ config, onChange, disabled }: ConfigFormProps) {
  const toField = (config.toField as string) || 'Email'
  const fromName = (config.fromName as string) || ''
  const fromEmail = (config.fromEmail as string) || ''
  const replyTo = (config.replyTo as string) || ''
  const subjectTemplate = (config.subjectTemplate as string) || ''
  const bodyTemplate = (config.bodyTemplate as string) || ''
  const contentType = (config.contentType as string) || 'text/html'

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="To Field"
        placeholder="e.g. Email"
        value={toField}
        onChange={(e) => updateConfig({ toField: e.target.value })}
        disabled={disabled}
        description="Contact field containing the recipient email address."
      />
      <div className="grid grid-cols-2 gap-3">
        <FormTextField
          label="From Name"
          placeholder="Your Name"
          value={fromName}
          onChange={(e) => updateConfig({ fromName: e.target.value })}
          disabled={disabled}
        />
        <FormTextField
          label="From Email"
          type="email"
          placeholder="you@example.com"
          value={fromEmail}
          onChange={(e) => updateConfig({ fromEmail: e.target.value })}
          disabled={disabled}
        />
      </div>
      <FormTextField
        label="Reply-To"
        type="email"
        placeholder="reply@example.com (optional)"
        value={replyTo}
        onChange={(e) => updateConfig({ replyTo: e.target.value })}
        disabled={disabled}
        description="Reply-to email address. Leave empty to use the from email."
      />
      <FormTextField
        label="Subject Template"
        placeholder="Email subject line"
        value={subjectTemplate}
        onChange={(e) => updateConfig({ subjectTemplate: e.target.value })}
        disabled={disabled}
        description="Supports merge fields like ~Contact.FirstName~"
      />
      <FormRadioGroup
        label="Content Type"
        value={contentType}
        onValueChange={(value) => updateConfig({ contentType: value })}
        options={[
          { value: 'text/html', label: 'HTML' },
          { value: 'text/plain', label: 'Plain Text' },
        ]}
        disabled={disabled}
      />
      <FormTextArea
        label="Body Template"
        placeholder="Email content..."
        value={bodyTemplate}
        onChange={(e) => updateConfig({ bodyTemplate: e.target.value })}
        disabled={disabled}
        rows={6}
      />
    </div>
  )
}
