'use client'

// Schema: see schemas.ts > facebookLeadsSchema
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { TagPicker } from '@/components/tag-picker'
import { FormTextField, DynamicList, AddItemRow } from './form-fields'
import type { ConfigFormProps } from './types'

interface FieldMapping {
  fbField: string
  crmField: string
}

export function FacebookLeadsForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const formId = (config.formId as string) || ''
  const pageId = (config.pageId as string) || ''
  const fieldMappings = (config.fieldMappings as FieldMapping[]) || []
  const leadTags = (config.leadTags as string[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3">
        <FormTextField
          label="Page ID"
          placeholder="Facebook Page ID"
          value={pageId}
          onChange={(e) => updateConfig({ pageId: e.target.value })}
          disabled={disabled}
        />
        <FormTextField
          label="Form ID"
          placeholder="Lead Ad Form ID"
          value={formId}
          onChange={(e) => updateConfig({ formId: e.target.value })}
          disabled={disabled}
        />
      </div>

      <DynamicList<FieldMapping>
        label="Field Mappings"
        description="Map Facebook form fields to CRM fields."
        items={fieldMappings}
        onItemsChange={(items) => updateConfig({ fieldMappings: items })}
        renderItem={(m) => (
          <>
            <span className="font-medium">{m.fbField}</span>
            <span className="text-muted-foreground">→</span>
            <span className="flex-1 font-mono">{m.crmField}</span>
          </>
        )}
        renderAddForm={(onAdd) => {
          const MappingAddForm = () => {
            const [fbField, setFbField] = useState('')
            const [crmField, setCrmField] = useState('')
            const handleAdd = () => {
              if (!fbField.trim() || !crmField.trim()) return
              onAdd({ fbField: fbField.trim(), crmField: crmField.trim() })
              setFbField('')
              setCrmField('')
            }
            return (
              <AddItemRow onAdd={handleAdd} disabled={disabled} canAdd={!!fbField.trim() && !!crmField.trim()}>
                <Input placeholder="FB field name" value={fbField} onChange={(e) => setFbField(e.target.value)} disabled={disabled} className="w-1/2" />
                <Input placeholder="CRM field key" value={crmField} onChange={(e) => setCrmField(e.target.value)} disabled={disabled} className="w-1/2" onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAdd() } }} />
              </AddItemRow>
            )
          }
          return <MappingAddForm />
        }}
        disabled={disabled}
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Lead Tags</label>
        <TagPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={leadTags} onChange={(value) => updateConfig({ leadTags: value })} placeholder="Tags for new leads..." disabled={disabled} />
      </div>
    </div>
  )
}
