'use client'

// Go keys: steps ([]string), state_field
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FieldPicker } from '@/components/field-picker'
import { DynamicList, AddItemRow } from './form-fields'
import type { ConfigFormProps } from './types'

export function DripItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const steps = (config.steps as string[]) || []
  const stateField = (config.stateField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <DynamicList<string>
        label="Steps"
        description="Add automation or helper IDs to execute in sequence."
        items={steps}
        onItemsChange={(items) => updateConfig({ steps: items })}
        disabled={disabled}
        renderAddForm={(onAdd) => {
          const StepAddForm = () => {
            const [value, setValue] = useState('')
            const handleAdd = () => {
              if (!value.trim()) return
              onAdd(value.trim())
              setValue('')
            }
            return (
              <AddItemRow onAdd={handleAdd} disabled={disabled} canAdd={!!value.trim()}>
                <Input
                  placeholder="Automation or helper ID"
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  disabled={disabled}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault()
                      handleAdd()
                    }
                  }}
                />
              </AddItemRow>
            )
          }
          return <StepAddForm />
        }}
        renderItem={(step) => (
          <span className="font-mono">{step}</span>
        )}
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">State Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={stateField}
          onChange={(value) => updateConfig({ stateField: value })}
          placeholder="Select field to track drip state..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          A contact field used to track which step the contact is currently on.
        </p>
      </div>
    </div>
  )
}
