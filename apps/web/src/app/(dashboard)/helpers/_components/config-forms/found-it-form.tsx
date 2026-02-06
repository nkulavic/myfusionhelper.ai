'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

export function FoundItForm({ config, onChange, disabled }: ConfigFormProps) {
  const field = (config.field as string) || ''
  const foundTagId = (config.foundTagId as string) || ''
  const notFoundTagId = (config.notFoundTagId as string) || ''
  const condition = (config.condition as string) || 'not_empty'

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="found-field">Field to Check</Label>
        <Input
          id="found-field"
          placeholder="e.g. Email, Phone1, _CustomField123"
          value={field}
          onChange={(e) => updateConfig({ field: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field to check for a value.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="found-condition">Condition</Label>
        <select
          id="found-condition"
          value={condition}
          onChange={(e) => updateConfig({ condition: e.target.value })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          <option value="not_empty">Field is not empty</option>
          <option value="is_empty">Field is empty</option>
          <option value="contains">Field contains value</option>
          <option value="equals">Field equals value</option>
        </select>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="found-tag">Tag to Apply (when matched)</Label>
        <Input
          id="found-tag"
          placeholder="Tag ID to apply if condition is true"
          value={foundTagId}
          onChange={(e) => updateConfig({ foundTagId: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          This tag will be applied when the condition is met.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="not-found-tag">Tag to Apply (when not matched, optional)</Label>
        <Input
          id="not-found-tag"
          placeholder="Tag ID to apply if condition is false"
          value={notFoundTagId}
          onChange={(e) => updateConfig({ notFoundTagId: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Optional. This tag will be applied when the condition is not met.
        </p>
      </div>
    </div>
  )
}
