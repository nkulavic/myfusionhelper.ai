'use client'

import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

export function TagItForm({ config, onChange, disabled }: ConfigFormProps) {
  const action = (config.action as string) || 'apply'
  const tags = (config.tags as string[]) || []
  const [newTag, setNewTag] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addTag = () => {
    const trimmed = newTag.trim()
    if (!trimmed || tags.includes(trimmed)) return
    updateConfig({ tags: [...tags, trimmed] })
    setNewTag('')
  }

  const removeTag = (tag: string) => {
    updateConfig({ tags: tags.filter((t) => t !== tag) })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="tag-action">Action</Label>
        <select
          id="tag-action"
          value={action}
          onChange={(e) => updateConfig({ action: e.target.value })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          <option value="apply">Apply tags</option>
          <option value="remove">Remove tags</option>
          <option value="toggle">Toggle tags</option>
        </select>
        <p className="text-xs text-muted-foreground">
          {action === 'apply'
            ? 'Add these tags to the contact.'
            : action === 'remove'
            ? 'Remove these tags from the contact.'
            : 'If the contact has the tag, remove it. Otherwise, add it.'}
        </p>
      </div>

      <div className="grid gap-2">
        <Label>Tags</Label>
        <div className="flex gap-2">
          <Input
            placeholder="Enter a tag name..."
            value={newTag}
            onChange={(e) => setNewTag(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault()
                addTag()
              }
            }}
            disabled={disabled}
            className="flex-1"
          />
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={addTag}
            disabled={disabled || !newTag.trim()}
            aria-label="Add tag"
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>
        {tags.length > 0 ? (
          <div className="flex flex-wrap gap-1.5 mt-1">
            {tags.map((tag) => (
              <span
                key={tag}
                className="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary"
              >
                {tag}
                {!disabled && (
                  <button
                    type="button"
                    onClick={() => removeTag(tag)}
                    className="ml-0.5 rounded-full p-0.5 hover:bg-primary/20"
                    aria-label={`Remove tag ${tag}`}
                  >
                    <X className="h-3 w-3" />
                  </button>
                )}
              </span>
            ))}
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">No tags added yet. Type a tag name and press Enter.</p>
        )}
      </div>

      <div className="grid gap-2">
        <Label htmlFor="tag-condition">Condition (optional)</Label>
        <select
          id="tag-condition"
          value={(config.condition as string) || 'always'}
          onChange={(e) => updateConfig({ condition: e.target.value })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          <option value="always">Always run</option>
          <option value="if_not_tagged">Only if contact does not have these tags</option>
          <option value="if_tagged">Only if contact already has these tags</option>
        </select>
      </div>
    </div>
  )
}
