'use client'

// Schema: see schemas.ts > chainItSchema
import { useState } from 'react'
import { Plus, X, GripVertical, ArrowDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { ConfigFormProps } from './types'

export function ChainItForm({ config, onChange, disabled }: ConfigFormProps) {
  const helpers = (config.helpers as string[]) || []
  const [newHelperId, setNewHelperId] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addHelper = () => {
    const trimmed = newHelperId.trim()
    if (!trimmed || helpers.includes(trimmed)) return
    updateConfig({ helpers: [...helpers, trimmed] })
    setNewHelperId('')
  }

  const removeHelper = (id: string) => {
    updateConfig({ helpers: helpers.filter((h) => h !== id) })
  }

  const moveUp = (index: number) => {
    if (index === 0) return
    const updated = [...helpers]
    ;[updated[index - 1], updated[index]] = [updated[index], updated[index - 1]]
    updateConfig({ helpers: updated })
  }

  const moveDown = (index: number) => {
    if (index >= helpers.length - 1) return
    const updated = [...helpers]
    ;[updated[index], updated[index + 1]] = [updated[index + 1], updated[index]]
    updateConfig({ helpers: updated })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">Helper Chain</label>
        <p className="text-xs text-muted-foreground">
          Add helper IDs in the order you want them executed. Each helper runs
          sequentially with the same contact.
        </p>
        <div className="flex gap-2">
          <Input
            placeholder="Enter a helper ID..."
            value={newHelperId}
            onChange={(e) => setNewHelperId(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault()
                addHelper()
              }
            }}
            disabled={disabled}
            className="flex-1"
          />
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={addHelper}
            disabled={disabled || !newHelperId.trim()}
            aria-label="Add helper"
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>

        {helpers.length > 0 ? (
          <div className="space-y-1 mt-1">
            {helpers.map((id, i) => (
              <div key={`${id}-${i}`}>
                <div className="flex items-center gap-2 rounded-md bg-muted px-3 py-2 text-xs">
                  <GripVertical className="h-3 w-3 text-muted-foreground/50 flex-shrink-0" />
                  <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary/10 text-[10px] font-bold text-primary flex-shrink-0">
                    {i + 1}
                  </span>
                  <span className="flex-1 font-mono truncate">{id}</span>
                  {!disabled && (
                    <div className="flex items-center gap-0.5">
                      <button
                        type="button"
                        onClick={() => moveUp(i)}
                        disabled={i === 0}
                        className="rounded p-0.5 hover:bg-accent disabled:opacity-30"
                        aria-label="Move up"
                      >
                        <ArrowDown className="h-3 w-3 rotate-180" />
                      </button>
                      <button
                        type="button"
                        onClick={() => moveDown(i)}
                        disabled={i === helpers.length - 1}
                        className="rounded p-0.5 hover:bg-accent disabled:opacity-30"
                        aria-label="Move down"
                      >
                        <ArrowDown className="h-3 w-3" />
                      </button>
                      <button
                        type="button"
                        onClick={() => removeHelper(id)}
                        className="rounded p-0.5 hover:bg-accent ml-1"
                        aria-label="Remove helper"
                      >
                        <X className="h-3 w-3" />
                      </button>
                    </div>
                  )}
                </div>
                {i < helpers.length - 1 && (
                  <div className="flex justify-center py-0.5">
                    <ArrowDown className="h-3 w-3 text-muted-foreground/30" />
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">
            No helpers in the chain yet. Add helper IDs to build the execution sequence.
          </p>
        )}
      </div>
    </div>
  )
}
