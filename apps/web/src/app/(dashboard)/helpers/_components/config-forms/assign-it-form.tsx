'use client'

import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

export function AssignItForm({ config, onChange, disabled }: ConfigFormProps) {
  const mode = (config.mode as string) || 'specific'
  const ownerIds = (config.ownerIds as string[]) || []
  const [newOwnerId, setNewOwnerId] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addOwner = () => {
    const trimmed = newOwnerId.trim()
    if (!trimmed || ownerIds.includes(trimmed)) return
    updateConfig({ ownerIds: [...ownerIds, trimmed] })
    setNewOwnerId('')
  }

  const removeOwner = (id: string) => {
    updateConfig({ ownerIds: ownerIds.filter((o) => o !== id) })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="assign-mode">Assignment Mode</Label>
        <select
          id="assign-mode"
          value={mode}
          onChange={(e) => updateConfig({ mode: e.target.value })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          <option value="specific">Specific Owner</option>
          <option value="round_robin">Round Robin</option>
        </select>
        <p className="text-xs text-muted-foreground">
          {mode === 'specific'
            ? 'Always assign to the same owner.'
            : 'Distribute contacts evenly across multiple owners.'}
        </p>
      </div>

      <div className="grid gap-2">
        <Label>
          {mode === 'specific' ? 'Owner ID' : 'Owner IDs (rotation order)'}
        </Label>
        <div className="flex gap-2">
          <Input
            placeholder="Enter an owner/user ID..."
            value={newOwnerId}
            onChange={(e) => setNewOwnerId(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault()
                addOwner()
              }
            }}
            disabled={disabled}
            className="flex-1"
          />
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={addOwner}
            disabled={disabled || !newOwnerId.trim()}
            aria-label="Add owner"
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>
        {ownerIds.length > 0 ? (
          <div className="flex flex-wrap gap-1.5 mt-1">
            {ownerIds.map((id, i) => (
              <span
                key={id}
                className="inline-flex items-center gap-1 rounded-md bg-muted px-2 py-0.5 text-xs font-mono"
              >
                {mode === 'round_robin' && (
                  <span className="text-muted-foreground/60 mr-0.5">
                    {i + 1}.
                  </span>
                )}
                {id}
                {!disabled && (
                  <button
                    type="button"
                    onClick={() => removeOwner(id)}
                    className="ml-0.5 rounded p-0.5 hover:bg-accent"
                    aria-label={`Remove owner ${id}`}
                  >
                    <X className="h-3 w-3" />
                  </button>
                )}
              </span>
            ))}
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">
            {mode === 'specific'
              ? 'Add the user ID of the contact owner.'
              : 'Add multiple owner IDs. Contacts will be distributed in order.'}
          </p>
        )}
      </div>

      {mode === 'round_robin' && ownerIds.length > 1 && (
        <div className="rounded-md border border-dashed bg-muted/50 p-3">
          <p className="text-xs text-muted-foreground">
            Contacts will be distributed evenly across {ownerIds.length} owners
            in a round-robin fashion.
          </p>
        </div>
      )}
    </div>
  )
}
