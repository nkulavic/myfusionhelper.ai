'use client'

// Schema: see schemas.ts > hubspotListSyncSchema
import { FormTextField, FormSelect } from './form-fields'
import type { ConfigFormProps } from './types'

const syncDirections = [
  { value: 'to_list', label: 'Add to List' },
  { value: 'from_list', label: 'Read from List' },
  { value: 'sync', label: 'Two-way Sync' },
]

export function HubspotListSyncForm({ config, onChange, disabled }: ConfigFormProps) {
  const listId = (config.listId as string) || ''
  const syncDirection = (config.syncDirection as string) || 'to_list'
  const filterCriteria = (config.filterCriteria as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="List ID"
        description="The ID of the HubSpot list to sync contacts with."
        placeholder="HubSpot list ID"
        value={listId}
        onChange={(e) => updateConfig({ listId: e.target.value })}
        disabled={disabled}
      />
      <FormSelect
        label="Sync Direction"
        value={syncDirection}
        onValueChange={(v) => updateConfig({ syncDirection: v })}
        options={syncDirections}
        disabled={disabled}
      />
      <FormTextField
        label="Filter Criteria"
        description="Optional criteria to filter which contacts are synced."
        placeholder="Optional filter expression"
        value={filterCriteria}
        onChange={(e) => updateConfig({ filterCriteria: e.target.value })}
        disabled={disabled}
      />
    </div>
  )
}
