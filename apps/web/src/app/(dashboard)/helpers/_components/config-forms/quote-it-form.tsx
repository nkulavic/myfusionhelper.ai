'use client'

// Schema: see schemas.ts > quoteItSchema

import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormSelect } from './form-fields'
import type { ConfigFormProps } from './types'

const categoryOptions = [
  { value: 'inspire', label: 'Inspirational' },
  { value: 'management', label: 'Management' },
  { value: 'sports', label: 'Sports' },
  { value: 'life', label: 'Life' },
  { value: 'funny', label: 'Funny' },
  { value: 'love', label: 'Love' },
  { value: 'art', label: 'Art' },
  { value: 'students', label: 'Students' },
]

const formatOptions = [
  { value: 'single_line', label: 'Single line (quote - author)' },
  { value: 'multi_line', label: 'Multi line (quote on one line, author on next)' },
  { value: 'multi_field', label: 'Separate fields (quote and author in different fields)' },
]

export function QuoteItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const category = (config.category as string) || 'inspire'
  const format = (config.format as string) || 'single_line'
  const targetField = (config.targetField as string) || ''
  const quoteField = (config.quoteField as string) || ''
  const authorField = (config.authorField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormSelect
        label="Quote Category"
        value={category}
        onValueChange={(value) => updateConfig({ category: value })}
        options={categoryOptions}
        disabled={disabled}
      />

      <FormSelect
        label="Format"
        value={format}
        onValueChange={(value) => updateConfig({ format: value })}
        options={formatOptions}
        disabled={disabled}
      />

      {format !== 'multi_field' ? (
        <div className="grid gap-2">
          <Label>Save To Field</Label>
          <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={targetField} onChange={(value) => updateConfig({ targetField: value })} placeholder="Select field..." disabled={disabled} />
        </div>
      ) : (
        <>
          <div className="grid gap-2">
            <Label>Quote Field</Label>
            <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={quoteField} onChange={(value) => updateConfig({ quoteField: value })} placeholder="Select field for quote..." disabled={disabled} />
          </div>
          <div className="grid gap-2">
            <Label>Author Field</Label>
            <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={authorField} onChange={(value) => updateConfig({ authorField: value })} placeholder="Select field for author..." disabled={disabled} />
          </div>
        </>
      )}
    </div>
  )
}
