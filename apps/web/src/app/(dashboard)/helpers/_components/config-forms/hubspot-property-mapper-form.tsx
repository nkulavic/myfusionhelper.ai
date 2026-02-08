'use client'

// Schema: see schemas.ts > hubspotPropertyMapperSchema
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FormSelect, DynamicList, AddItemRow } from './form-fields'
import type { ConfigFormProps } from './types'

interface PropertyMapping {
  source: string
  target: string
}

const objectTypes = [
  { value: 'contacts', label: 'Contacts' },
  { value: 'companies', label: 'Companies' },
  { value: 'deals', label: 'Deals' },
  { value: 'tickets', label: 'Tickets' },
]

export function HubspotPropertyMapperForm({ config, onChange, disabled }: ConfigFormProps) {
  const objectType = (config.objectType as string) || 'contacts'
  const mappings = (config.mappings as PropertyMapping[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Object Type"
        value={objectType}
        onValueChange={(value) => updateConfig({ objectType: value })}
        options={objectTypes}
        disabled={disabled}
      />

      <DynamicList<PropertyMapping>
        label="Property Mappings"
        description="Map source properties to target properties."
        items={mappings}
        onItemsChange={(items) => updateConfig({ mappings: items })}
        renderItem={(m) => (
          <>
            <span className="font-medium min-w-[80px]">{m.source}</span>
            <span className="text-muted-foreground">→</span>
            <span className="flex-1 font-mono">{m.target}</span>
          </>
        )}
        renderAddForm={(onAdd) => {
          const MappingAddForm = () => {
            const [source, setSource] = useState('')
            const [target, setTarget] = useState('')
            const handleAdd = () => {
              if (!source.trim() || !target.trim()) return
              onAdd({ source: source.trim(), target: target.trim() })
              setSource('')
              setTarget('')
            }
            return (
              <AddItemRow onAdd={handleAdd} disabled={disabled} canAdd={!!source.trim() && !!target.trim()}>
                <Input placeholder="Source property" value={source} onChange={(e) => setSource(e.target.value)} disabled={disabled} className="w-1/2" />
                <Input placeholder="Target property" value={target} onChange={(e) => setTarget(e.target.value)} disabled={disabled} className="w-1/2" onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAdd() } }} />
              </AddItemRow>
            )
          }
          return <MappingAddForm />
        }}
        disabled={disabled}
      />
    </div>
  )
}
