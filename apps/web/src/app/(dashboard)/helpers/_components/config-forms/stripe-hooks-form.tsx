'use client'

// Schema: see schemas.ts > stripeHooksSchema
import { TagPicker } from '@/components/tag-picker'
import { FormTextField, FormCheckbox } from './form-fields'
import type { ConfigFormProps } from './types'

const eventTypes = [
  { value: 'charge.succeeded', label: 'Charge Succeeded' },
  { value: 'charge.failed', label: 'Charge Failed' },
  { value: 'charge.refunded', label: 'Charge Refunded' },
  { value: 'customer.subscription.created', label: 'Subscription Created' },
  { value: 'customer.subscription.updated', label: 'Subscription Updated' },
  { value: 'customer.subscription.deleted', label: 'Subscription Cancelled' },
  { value: 'invoice.payment_succeeded', label: 'Invoice Paid' },
  { value: 'invoice.payment_failed', label: 'Invoice Payment Failed' },
]

export function StripeHooksForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const selectedEvents = (config.selectedEvents as string[]) || []
  const goalName = (config.goalName as string) || ''
  const eventTags = (config.eventTags as string[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const toggleEvent = (eventValue: string) => {
    const updated = selectedEvents.includes(eventValue)
      ? selectedEvents.filter((e) => e !== eventValue)
      : [...selectedEvents, eventValue]
    updateConfig({ selectedEvents: updated })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Stripe Events</label>
        <p className="text-xs text-muted-foreground">Select which Stripe events to process.</p>
        <div className="space-y-1.5">
          {eventTypes.map((evt) => (
            <FormCheckbox
              key={evt.value}
              label={evt.label}
              checked={selectedEvents.includes(evt.value)}
              onCheckedChange={() => toggleEvent(evt.value)}
              disabled={disabled}
            />
          ))}
        </div>
      </div>

      <FormTextField
        label="API Goal"
        placeholder="Goal on event (optional)"
        value={goalName}
        onChange={(e) => updateConfig({ goalName: e.target.value })}
        disabled={disabled}
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Event Tags</label>
        <TagPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={eventTags} onChange={(value) => updateConfig({ eventTags: value })} placeholder="Tags to apply on events..." disabled={disabled} />
      </div>
    </div>
  )
}
