'use client'

// Go keys: event_type_uri, api_token, email_field, name_field, result_field
import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

export function CalendlyForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const eventTypeUri = (config.eventTypeUri as string) || ''
  const apiToken = (config.apiToken as string) || ''
  const emailField = (config.emailField as string) || ''
  const nameField = (config.nameField as string) || ''
  const resultField = (config.resultField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Event Type URI"
        placeholder="https://api.calendly.com/event_types/..."
        value={eventTypeUri}
        onChange={(e) => updateConfig({ eventTypeUri: e.target.value })}
        disabled={disabled}
        description="The Calendly event type URI to schedule."
      />

      <FormTextField
        label="API Token"
        placeholder="Calendly personal access token"
        value={apiToken}
        onChange={(e) => updateConfig({ apiToken: e.target.value })}
        disabled={disabled}
        type="password"
        description="Your Calendly API token for authentication."
      />

      <div className="grid gap-2">
        <Label>Email Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={emailField}
          onChange={(value) => updateConfig({ emailField: value })}
          placeholder="Select email field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field containing the email address for scheduling.
        </p>
      </div>

      <div className="grid gap-2">
        <Label>Name Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={nameField}
          onChange={(value) => updateConfig({ nameField: value })}
          placeholder="Select name field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field containing the name for the invitee.
        </p>
      </div>

      <div className="grid gap-2">
        <Label>Result Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={resultField}
          onChange={(value) => updateConfig({ resultField: value })}
          placeholder="Select result field..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field where the scheduling result will be stored.
        </p>
      </div>
    </div>
  )
}
