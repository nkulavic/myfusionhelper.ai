'use client'

import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

export function MergeItForm({ config, onChange, disabled }: ConfigFormProps) {
  const sourceFields = (config.sourceFields as string[]) || []
  const targetField = (config.targetField as string) || ''
  const separator = (config.separator as string) ?? ' '
  const [newField, setNewField] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addField = () => {
    const trimmed = newField.trim()
    if (!trimmed || sourceFields.includes(trimmed)) return
    updateConfig({ sourceFields: [...sourceFields, trimmed] })
    setNewField('')
  }

  const removeField = (field: string) => {
    updateConfig({ sourceFields: sourceFields.filter((f) => f !== field) })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label>Source Fields</Label>
        <div className="flex gap-2">
          <Input
            placeholder="Enter a field name..."
            value={newField}
            onChange={(e) => setNewField(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault()
                addField()
              }
            }}
            disabled={disabled}
            className="flex-1"
          />
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={addField}
            disabled={disabled || !newField.trim()}
            aria-label="Add field"
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>
        {sourceFields.length > 0 ? (
          <div className="flex flex-wrap gap-1.5 mt-1">
            {sourceFields.map((field, i) => (
              <span
                key={field}
                className="inline-flex items-center gap-1 rounded-md bg-muted px-2 py-0.5 text-xs font-mono"
              >
                {i > 0 && (
                  <span className="text-muted-foreground/60 mr-0.5">
                    {separator || ' '}
                  </span>
                )}
                {field}
                {!disabled && (
                  <button
                    type="button"
                    onClick={() => removeField(field)}
                    className="ml-0.5 rounded p-0.5 hover:bg-accent"
                    aria-label={`Remove field ${field}`}
                  >
                    <X className="h-3 w-3" />
                  </button>
                )}
              </span>
            ))}
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">
            Add the fields you want to merge together. e.g. FirstName, LastName
          </p>
        )}
      </div>

      <div className="grid gap-2">
        <Label htmlFor="merge-separator">Separator</Label>
        <Input
          id="merge-separator"
          placeholder="e.g. a space, comma, or dash"
          value={separator}
          onChange={(e) => updateConfig({ separator: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The character(s) placed between each field value. Default is a space.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="merge-target">Target Field</Label>
        <Input
          id="merge-target"
          placeholder="e.g. FullName, _DisplayName"
          value={targetField}
          onChange={(e) => updateConfig({ targetField: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The field where the merged value will be stored.
        </p>
      </div>
    </div>
  )
}
