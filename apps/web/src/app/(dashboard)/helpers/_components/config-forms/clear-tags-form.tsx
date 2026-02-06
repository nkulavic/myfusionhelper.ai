'use client'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

export function ClearTagsForm({ config, onChange, disabled }: ConfigFormProps) {
  const mode = (config.mode as string) || 'all'
  const prefix = (config.prefix as string) || ''
  const category = (config.category as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="clear-mode">Clear Mode</Label>
        <select
          id="clear-mode"
          value={mode}
          onChange={(e) => updateConfig({ mode: e.target.value })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          <option value="all">Remove all tags</option>
          <option value="prefix">Remove tags matching a prefix</option>
          <option value="category">Remove tags in a category</option>
        </select>
        <p className="text-xs text-muted-foreground">
          {mode === 'all'
            ? 'Remove every tag from the contact.'
            : mode === 'prefix'
            ? 'Remove tags whose name starts with a specific prefix.'
            : 'Remove tags belonging to a specific category.'}
        </p>
      </div>

      {mode === 'prefix' && (
        <div className="grid gap-2">
          <Label htmlFor="clear-prefix">Tag Prefix</Label>
          <Input
            id="clear-prefix"
            placeholder="e.g. campaign_"
            value={prefix}
            onChange={(e) => updateConfig({ prefix: e.target.value })}
            disabled={disabled}
          />
          <p className="text-xs text-muted-foreground">
            All tags starting with this prefix will be removed.
          </p>
        </div>
      )}

      {mode === 'category' && (
        <div className="grid gap-2">
          <Label htmlFor="clear-category">Tag Category</Label>
          <Input
            id="clear-category"
            placeholder="e.g. marketing"
            value={category}
            onChange={(e) => updateConfig({ category: e.target.value })}
            disabled={disabled}
          />
          <p className="text-xs text-muted-foreground">
            All tags in this category will be removed.
          </p>
        </div>
      )}

      {mode === 'all' && (
        <div className="rounded-md border border-warning/30 bg-warning/5 p-3">
          <p className="text-xs text-warning">
            Warning: This will remove ALL tags from the contact. Use with caution.
          </p>
        </div>
      )}
    </div>
  )
}
