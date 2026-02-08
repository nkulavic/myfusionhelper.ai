'use client'

// Schema: see schemas.ts > actionItSchema
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { DynamicList, AddItemRow } from './form-fields'
import type { ConfigFormProps } from './types'

export function ActionItForm({ config, onChange, disabled }: ConfigFormProps) {
  const automationIds = (config.automationIds as string[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <DynamicList<string>
        label="Automation Sequence"
        description="List of automation IDs to run in order."
        items={automationIds}
        onItemsChange={(items) => updateConfig({ automationIds: items })}
        renderItem={(action, i) => (
          <>
            <span className="font-medium">{i + 1}.</span>
            <span className="flex-1 font-mono">{action}</span>
          </>
        )}
        renderAddForm={(onAdd) => {
          const ActionAddForm = () => {
            const [value, setValue] = useState('')
            const handleAdd = () => {
              if (!value.trim()) return
              onAdd(value.trim())
              setValue('')
            }
            return (
              <AddItemRow onAdd={handleAdd} disabled={disabled} canAdd={!!value.trim()}>
                <Input placeholder="Action set name or ID" value={value} onChange={(e) => setValue(e.target.value)} disabled={disabled} className="flex-1" onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAdd() } }} />
              </AddItemRow>
            )
          }
          return <ActionAddForm />
        }}
        disabled={disabled}
      />
    </div>
  )
}
