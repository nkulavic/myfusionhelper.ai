'use client'

import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

const formatOptions = [
  { value: 'uppercase', label: 'UPPERCASE', description: 'Convert to all uppercase' },
  { value: 'lowercase', label: 'lowercase', description: 'Convert to all lowercase' },
  { value: 'title_case', label: 'Title Case', description: 'Capitalize first letter of each word' },
  { value: 'sentence_case', label: 'Sentence case', description: 'Capitalize first letter only' },
  { value: 'trim', label: 'Trim whitespace', description: 'Remove leading and trailing spaces' },
  { value: 'trim_all', label: 'Trim all spaces', description: 'Collapse multiple spaces to one' },
  { value: 'phone', label: 'Phone format', description: 'Format as (123) 456-7890' },
]

export function FormatItForm({ config, onChange, disabled }: ConfigFormProps) {
  const fields = (config.fields as string[]) || []
  const format = (config.format as string) || 'title_case'
  const [newField, setNewField] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addField = () => {
    const trimmed = newField.trim()
    if (!trimmed || fields.includes(trimmed)) return
    updateConfig({ fields: [...fields, trimmed] })
    setNewField('')
  }

  const removeField = (field: string) => {
    updateConfig({ fields: fields.filter((f) => f !== field) })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="format-type">Format Type</Label>
        <select
          id="format-type"
          value={format}
          onChange={(e) => updateConfig({ format: e.target.value })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          {formatOptions.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        <p className="text-xs text-muted-foreground">
          {formatOptions.find((o) => o.value === format)?.description}
        </p>
      </div>

      <div className="grid gap-2">
        <Label>Fields to Format</Label>
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
        {fields.length > 0 ? (
          <div className="flex flex-wrap gap-1.5 mt-1">
            {fields.map((field) => (
              <span
                key={field}
                className="inline-flex items-center gap-1 rounded-md bg-muted px-2 py-0.5 text-xs font-mono"
              >
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
            Add the contact fields you want to format. e.g. FirstName, LastName, Email
          </p>
        )}
      </div>
    </div>
  )
}
