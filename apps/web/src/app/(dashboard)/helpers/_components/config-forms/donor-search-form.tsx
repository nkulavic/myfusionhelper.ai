'use client'

// Schema: see schemas.ts > donorSearchSchema
import { FieldPicker } from '@/components/field-picker'
import { TagPicker } from '@/components/tag-picker'
import { InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

export function DonorSearchForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const resultField = (config.resultField as string) || ''
  const capacityField = (config.capacityField as string) || ''
  const foundTags = (config.foundTags as string[]) || []
  const notFoundTags = (config.notFoundTags as string[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <InfoBanner>
        Look up donor profiles via Donor Search API using the contact&apos;s name and email.
      </InfoBanner>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Result Field</label>
        <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={resultField} onChange={(value) => updateConfig({ resultField: value })} placeholder="Store search result..." disabled={disabled} />
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Giving Capacity Field</label>
        <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={capacityField} onChange={(value) => updateConfig({ capacityField: value })} placeholder="Store giving capacity..." disabled={disabled} />
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Found Tags</label>
        <TagPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={foundTags} onChange={(value) => updateConfig({ foundTags: value })} placeholder="Tags when donor found..." disabled={disabled} />
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Not Found Tags</label>
        <TagPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={notFoundTags} onChange={(value) => updateConfig({ notFoundTags: value })} placeholder="Tags when not found..." disabled={disabled} />
      </div>
    </div>
  )
}
