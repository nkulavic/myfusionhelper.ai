'use client'

// Schema: see schemas.ts > noteItSchema
import { FormSelect, FormTextField, FormTextArea } from './form-fields'
import type { ConfigFormProps } from './types'

const noteTypes = [
  { value: 'general', label: 'General' },
  { value: 'call', label: 'Call' },
  { value: 'email', label: 'Email' },
  { value: 'fax', label: 'Fax' },
  { value: 'letter', label: 'Letter' },
  { value: 'other', label: 'Other' },
]

export function NoteItForm({ config, onChange, disabled }: ConfigFormProps) {
  const noteType = (config.noteType as string) || 'general'
  const subject = (config.subject as string) || ''
  const body = (config.body as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Note Type"
        value={noteType}
        onValueChange={(v) => updateConfig({ noteType: v })}
        options={noteTypes}
        disabled={disabled}
      />

      <FormTextField
        label="Subject"
        placeholder="Note subject (supports {{field_name}} tokens)"
        value={subject}
        onChange={(e) => updateConfig({ subject: e.target.value })}
        disabled={disabled}
        description={'Use {{field_name}} tokens to include contact data.'}
      />

      <FormTextArea
        label="Note Body"
        rows={5}
        placeholder="Note body (supports {{field_name}} tokens)"
        value={body}
        onChange={(e) => updateConfig({ body: e.target.value })}
        disabled={disabled}
        description={'Use {{field_name}} tokens to include dynamic contact data in the body.'}
      />
    </div>
  )
}
