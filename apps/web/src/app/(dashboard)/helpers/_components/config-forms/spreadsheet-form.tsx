'use client'

// Schema: see schemas.ts > spreadsheetSchema
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FieldPicker } from '@/components/field-picker'
import { FormTextField, FormSelect, DynamicList, AddItemRow } from './form-fields'
import type { ConfigFormProps } from './types'

interface ColumnMapping {
  column: string
  field: string
}

const actions = [
  { value: 'append', label: 'Append Row' },
  { value: 'update', label: 'Update Row' },
  { value: 'read', label: 'Read Row' },
]

export function SpreadsheetForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const sheetUrl = (config.sheetUrl as string) || ''
  const sheetName = (config.sheetName as string) || ''
  const action = (config.action as string) || 'append'
  const columnMappings = (config.columnMappings as ColumnMapping[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Spreadsheet URL"
        placeholder="https://docs.google.com/spreadsheets/d/..."
        value={sheetUrl}
        onChange={(e) => updateConfig({ sheetUrl: e.target.value })}
        disabled={disabled}
      />

      <div className="grid grid-cols-2 gap-3">
        <FormTextField
          label="Sheet Name"
          placeholder="Sheet1"
          value={sheetName}
          onChange={(e) => updateConfig({ sheetName: e.target.value })}
          disabled={disabled}
        />
        <FormSelect
          label="Action"
          value={action}
          onValueChange={(value) => updateConfig({ action: value })}
          options={actions}
          disabled={disabled}
        />
      </div>

      <DynamicList<ColumnMapping>
        label="Column → Field Mappings"
        description="Map spreadsheet columns to CRM fields."
        items={columnMappings}
        onItemsChange={(items) => updateConfig({ columnMappings: items })}
        renderItem={(m) => (
          <>
            <span className="font-medium min-w-[80px]">{m.column}</span>
            <span className="text-muted-foreground">→</span>
            <span className="flex-1 font-mono">{m.field}</span>
          </>
        )}
        renderAddForm={(onAdd) => {
          const MappingAddForm = () => {
            const [col, setCol] = useState('')
            const [field, setField] = useState('')
            const handleAdd = () => {
              if (!col.trim() || !field.trim()) return
              onAdd({ column: col.trim(), field: field.trim() })
              setCol('')
              setField('')
            }
            return (
              <AddItemRow onAdd={handleAdd} disabled={disabled} canAdd={!!col.trim() && !!field.trim()}>
                <Input placeholder="Column header" value={col} onChange={(e) => setCol(e.target.value)} disabled={disabled} className="w-1/3" />
                <div className="flex-1">
                  <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={field} onChange={(value) => setField(value as string)} placeholder="Select field..." disabled={disabled} />
                </div>
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
