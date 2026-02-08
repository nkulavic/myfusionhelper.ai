'use client'

// Schema: see schemas.ts > textItSchema
import { FieldPicker } from '@/components/field-picker'
import { FormSelect, FormTextField } from './form-fields'
import type { ConfigFormProps } from './types'

const operations = [
  { value: 'prepend', label: 'Prepend Text', description: 'Add text before the field value' },
  { value: 'append', label: 'Append Text', description: 'Add text after the field value' },
  { value: 'replace', label: 'Find & Replace', description: 'Replace occurrences of a pattern' },
  { value: 'remove', label: 'Remove Text', description: 'Remove occurrences of a pattern' },
  { value: 'truncate', label: 'Truncate', description: 'Limit the field to a max length' },
  { value: 'extract_email_domain', label: 'Extract Email Domain', description: 'Extract the domain from an email address' },
  { value: 'extract_numbers', label: 'Extract Numbers', description: 'Extract numeric characters from the field' },
  { value: 'slug', label: 'Slugify', description: 'Convert to URL-friendly slug format' },
  { value: 'reverse', label: 'Reverse', description: 'Reverse the text in the field' },
]

export function TextItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const field = (config.field as string) || ''
  const operation = (config.operation as string) || 'prepend'
  const value = (config.value as string) || ''
  const replaceWith = (config.replaceWith as string) || ''
  const maxLength = (config.maxLength as number) ?? 100
  const targetField = (config.targetField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const needsValue = ['prepend', 'append', 'replace', 'remove'].includes(operation)
  const needsReplaceWith = operation === 'replace'
  const needsMaxLength = operation === 'truncate'

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={field}
          onChange={(value) => updateConfig({ field: value })}
          placeholder="Select field to manipulate..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field to manipulate.
        </p>
      </div>

      <FormSelect
        label="Operation"
        description={operations.find((o) => o.value === operation)?.description}
        value={operation}
        onValueChange={(v) => updateConfig({ operation: v })}
        options={operations}
        disabled={disabled}
      />

      {needsMaxLength && (
        <FormTextField
          label="Max Length"
          type="number"
          placeholder="100"
          value={maxLength}
          onChange={(e) => updateConfig({ maxLength: parseInt(e.target.value, 10) || 0 })}
          disabled={disabled}
        />
      )}

      {needsValue && (
        <FormTextField
          label={operation === 'replace' ? 'Search Text' : operation === 'remove' ? 'Text to Remove' : 'Text'}
          placeholder={
            operation === 'replace'
              ? 'Text to find...'
              : operation === 'remove'
              ? 'Text to remove...'
              : 'Text to prepend or append...'
          }
          value={value}
          onChange={(e) => updateConfig({ value: e.target.value })}
          disabled={disabled}
        />
      )}

      {needsReplaceWith && (
        <FormTextField
          label="Replace With"
          placeholder="Replace with..."
          value={replaceWith}
          onChange={(e) => updateConfig({ replaceWith: e.target.value })}
          disabled={disabled}
        />
      )}

      <div className="grid gap-2">
        <label className="text-sm font-medium">Target Field (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={targetField}
          onChange={(value) => updateConfig({ targetField: value })}
          placeholder="Leave empty to update the source field"
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Store the result in a different field instead of overwriting the source.
        </p>
      </div>
    </div>
  )
}
