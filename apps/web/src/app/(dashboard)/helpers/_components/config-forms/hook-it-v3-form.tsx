'use client'

// Hook It V3 (Field Mapper)
// Go keys: field_mappings (object), nested_separator (string), skip_null_values (boolean)

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FormTextField, FormSwitch, DynamicList, AddItemRow, InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

interface FieldMapping {
  webhookField: string
  crmField: string
}

export function HookItV3Form({ config, onChange, disabled }: ConfigFormProps) {
  const nestedSeparator = (config.nestedSeparator as string) || '.'
  const skipNullValues = config.skipNullValues !== undefined ? (config.skipNullValues as boolean) : true

  // Convert object to array for DynamicList
  const fieldMappings = (config.fieldMappings as Record<string, string>) || {}
  const mappings: FieldMapping[] = Object.entries(fieldMappings).map(([webhookField, crmField]) => ({
    webhookField,
    crmField,
  }))

  const [newWebhookField, setNewWebhookField] = useState('')
  const [newCrmField, setNewCrmField] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const handleMappingsChange = (items: FieldMapping[]) => {
    const newMap: Record<string, string> = {}
    for (const item of items) {
      newMap[item.webhookField] = item.crmField
    }
    updateConfig({ fieldMappings: newMap })
  }

  return (
    <div className="space-y-4">
      <InfoBanner>
        Extracts webhook payload fields and saves them to CRM contact custom fields. For example: map "total_amount" → "OrderTotal", "order_id" → "LastOrderID".
      </InfoBanner>

      <DynamicList<FieldMapping>
        label="Webhook Field → CRM Field Mappings"
        description="Map webhook payload field names to CRM custom field names."
        items={mappings}
        onItemsChange={handleMappingsChange}
        disabled={disabled}
        renderAddForm={(onAdd) => (
          <AddItemRow
            onAdd={() => {
              const webhookField = newWebhookField.trim()
              const crmField = newCrmField.trim()
              if (!webhookField || !crmField) return
              onAdd({ webhookField, crmField })
              setNewWebhookField('')
              setNewCrmField('')
            }}
            disabled={disabled}
            canAdd={!!newWebhookField.trim() && !!newCrmField.trim()}
          >
            <Input
              placeholder="Webhook field (e.g., order.total)"
              value={newWebhookField}
              onChange={(e) => setNewWebhookField(e.target.value)}
              disabled={disabled}
              className="flex-1"
            />
            <Input
              placeholder="CRM field name"
              value={newCrmField}
              onChange={(e) => setNewCrmField(e.target.value)}
              disabled={disabled}
              className="flex-1"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  const webhookField = newWebhookField.trim()
                  const crmField = newCrmField.trim()
                  if (!webhookField || !crmField) return
                  onAdd({ webhookField, crmField })
                  setNewWebhookField('')
                  setNewCrmField('')
                }
              }}
            />
          </AddItemRow>
        )}
        renderItem={(entry) => (
          <span className="font-mono text-xs">
            <span className="font-medium">{entry.webhookField}</span>{' '}
            <span className="text-muted-foreground">→ {entry.crmField}</span>
          </span>
        )}
      />

      <FormTextField
        label="Nested Field Separator"
        placeholder="."
        value={nestedSeparator}
        onChange={(e) => updateConfig({ nestedSeparator: e.target.value })}
        disabled={disabled}
        description="Separator for nested field access (e.g., '.' for 'order.id'). Defaults to '.'."
      />

      <FormSwitch
        label="Skip Null/Empty Values"
        description="Skip setting fields when webhook value is null or empty."
        checked={skipNullValues}
        onCheckedChange={(checked) => updateConfig({ skipNullValues: checked })}
        disabled={disabled}
      />
    </div>
  )
}
