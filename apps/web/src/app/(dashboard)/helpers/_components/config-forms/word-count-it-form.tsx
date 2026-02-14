'use client'

// Schema: see schemas.ts > wordCountItSchema
import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormRadioGroup } from './form-fields'
import type { ConfigFormProps } from './types'

const countTypeOptions = [
  { value: 'words', label: 'Word Count' },
  { value: 'characters', label: 'Character Count' },
]

export function WordCountItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const sourceField = (config.sourceField as string) || ''
  const targetField = (config.targetField as string) || ''
  const countType = (config.countType as string) || 'words'

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label>From Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={sourceField}
          onChange={(value) => updateConfig({ sourceField: value })}
          placeholder="Select field to count..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field whose content will be counted.
        </p>
      </div>

      <div className="grid gap-2">
        <Label>Save Count To</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={targetField}
          onChange={(value) => updateConfig({ targetField: value })}
          placeholder="Select field to save count..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field where the word or character count will be stored.
        </p>
      </div>

      <FormRadioGroup
        label="Count Type"
        value={countType}
        onValueChange={(v) => updateConfig({ countType: v })}
        options={countTypeOptions}
        disabled={disabled}
      />
    </div>
  )
}
