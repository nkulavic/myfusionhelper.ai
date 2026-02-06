'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

const operations = [
  { value: 'prepend', label: 'Prepend Text', description: 'Add text before the field value' },
  { value: 'append', label: 'Append Text', description: 'Add text after the field value' },
  { value: 'replace', label: 'Find & Replace', description: 'Replace occurrences of a pattern' },
  { value: 'extract', label: 'Extract (Regex)', description: 'Extract text matching a pattern' },
  { value: 'truncate', label: 'Truncate', description: 'Limit the field to a max length' },
]

export function TextItForm({ config, onChange, disabled }: ConfigFormProps) {
  const field = (config.field as string) || ''
  const operation = (config.operation as string) || 'prepend'
  const value = (config.value as string) || ''
  const replacement = (config.replacement as string) || ''
  const maxLength = (config.maxLength as number) ?? 100
  const targetField = (config.targetField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="text-field">Field</Label>
        <Input
          id="text-field"
          placeholder="e.g. FirstName, _CustomNote"
          value={field}
          onChange={(e) => updateConfig({ field: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field to manipulate.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="text-operation">Operation</Label>
        <select
          id="text-operation"
          value={operation}
          onChange={(e) => updateConfig({ operation: e.target.value })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          {operations.map((op) => (
            <option key={op.value} value={op.value}>
              {op.label}
            </option>
          ))}
        </select>
        <p className="text-xs text-muted-foreground">
          {operations.find((o) => o.value === operation)?.description}
        </p>
      </div>

      {operation === 'truncate' ? (
        <div className="grid gap-2">
          <Label htmlFor="text-maxlength">Max Length</Label>
          <Input
            id="text-maxlength"
            type="number"
            placeholder="100"
            value={maxLength}
            onChange={(e) => updateConfig({ maxLength: parseInt(e.target.value, 10) || 0 })}
            disabled={disabled}
          />
        </div>
      ) : (
        <div className="grid gap-2">
          <Label htmlFor="text-value">
            {operation === 'replace' ? 'Search Pattern' : operation === 'extract' ? 'Regex Pattern' : 'Text'}
          </Label>
          <Input
            id="text-value"
            placeholder={
              operation === 'replace'
                ? 'Text to find...'
                : operation === 'extract'
                ? 'e.g. \\d{3}-\\d{4}'
                : 'Text to prepend or append...'
            }
            value={value}
            onChange={(e) => updateConfig({ value: e.target.value })}
            disabled={disabled}
          />
        </div>
      )}

      {operation === 'replace' && (
        <div className="grid gap-2">
          <Label htmlFor="text-replacement">Replacement Text</Label>
          <Input
            id="text-replacement"
            placeholder="Replace with..."
            value={replacement}
            onChange={(e) => updateConfig({ replacement: e.target.value })}
            disabled={disabled}
          />
        </div>
      )}

      <div className="grid gap-2">
        <Label htmlFor="text-target">Target Field (optional)</Label>
        <Input
          id="text-target"
          placeholder="Leave empty to update the source field"
          value={targetField}
          onChange={(e) => updateConfig({ targetField: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Store the result in a different field instead of overwriting the source.
        </p>
      </div>
    </div>
  )
}
