'use client'

// Schema: see schemas.ts > emailAttachItSchema
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FormTextField, DynamicList, AddItemRow } from './form-fields'
import type { ConfigFormProps } from './types'

export function EmailAttachItForm({ config, onChange, disabled }: ConfigFormProps) {
  const authorizedSenders = (config.authorizedSenders as string[]) || []
  const goalName = (config.goalName as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <DynamicList<string>
        label="Authorized Senders"
        description="Only emails from these addresses will be attached to contact records."
        items={authorizedSenders}
        onItemsChange={(items) => updateConfig({ authorizedSenders: items })}
        renderItem={(sender) => (
          <span className="font-mono">{sender}</span>
        )}
        renderAddForm={(onAdd) => {
          const SenderAddForm = () => {
            const [value, setValue] = useState('')
            const handleAdd = () => {
              if (!value.trim()) return
              onAdd(value.trim())
              setValue('')
            }
            return (
              <AddItemRow onAdd={handleAdd} disabled={disabled} canAdd={!!value.trim()}>
                <Input placeholder="sender@example.com" value={value} onChange={(e) => setValue(e.target.value)} disabled={disabled} className="flex-1" onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAdd() } }} />
              </AddItemRow>
            )
          }
          return <SenderAddForm />
        }}
        disabled={disabled}
      />

      <FormTextField
        label="API Goal on Attachment"
        placeholder="Goal name (optional)"
        value={goalName}
        onChange={(e) => updateConfig({ goalName: e.target.value })}
        disabled={disabled}
      />
    </div>
  )
}
