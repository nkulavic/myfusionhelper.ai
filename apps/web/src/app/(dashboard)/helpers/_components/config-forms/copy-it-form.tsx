'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

export function CopyItForm({ config, onChange, disabled }: ConfigFormProps) {
  const sourceField = (config.sourceField as string) || ''
  const targetField = (config.targetField as string) || ''
  const overwrite = (config.overwrite as boolean) ?? true

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="source-field">Source Field</Label>
        <Input
          id="source-field"
          placeholder="e.g. Email, Phone1, _CustomField123"
          value={sourceField}
          onChange={(e) => updateConfig({ sourceField: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field to copy the value from.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="target-field">Target Field</Label>
        <Input
          id="target-field"
          placeholder="e.g. Phone2, _CustomField456"
          value={targetField}
          onChange={(e) => updateConfig({ targetField: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field to copy the value to.
        </p>
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="overwrite"
          checked={overwrite}
          onChange={(e) => updateConfig({ overwrite: e.target.checked })}
          disabled={disabled}
          className="h-4 w-4 rounded border-input"
        />
        <Label htmlFor="overwrite" className="text-sm font-normal">
          Overwrite existing value in target field
        </Label>
      </div>
      <p className="text-xs text-muted-foreground -mt-2">
        When unchecked, the value will only be copied if the target field is empty.
      </p>
    </div>
  )
}
