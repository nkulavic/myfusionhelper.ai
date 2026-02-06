'use client'

import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

export function HookItForm({ config, onChange, disabled }: ConfigFormProps) {
  const url = (config.url as string) || ''
  const method = (config.method as string) || 'POST'
  const headers = (config.headers as Record<string, string>) || {}
  const includeContactData = (config.includeContactData as boolean) ?? true
  const [newHeaderKey, setNewHeaderKey] = useState('')
  const [newHeaderValue, setNewHeaderValue] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addHeader = () => {
    const key = newHeaderKey.trim()
    const value = newHeaderValue.trim()
    if (!key) return
    updateConfig({ headers: { ...headers, [key]: value } })
    setNewHeaderKey('')
    setNewHeaderValue('')
  }

  const removeHeader = (key: string) => {
    const updated = { ...headers }
    delete updated[key]
    updateConfig({ headers: updated })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="webhook-url">Webhook URL</Label>
        <Input
          id="webhook-url"
          type="url"
          placeholder="https://example.com/webhook"
          value={url}
          onChange={(e) => updateConfig({ url: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The endpoint that will receive the contact data.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="webhook-method">HTTP Method</Label>
        <select
          id="webhook-method"
          value={method}
          onChange={(e) => updateConfig({ method: e.target.value })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="PATCH">PATCH</option>
        </select>
      </div>

      <div className="grid gap-2">
        <Label>Custom Headers</Label>
        <div className="flex gap-2">
          <Input
            placeholder="Header name"
            value={newHeaderKey}
            onChange={(e) => setNewHeaderKey(e.target.value)}
            disabled={disabled}
            className="flex-1"
          />
          <Input
            placeholder="Header value"
            value={newHeaderValue}
            onChange={(e) => setNewHeaderValue(e.target.value)}
            disabled={disabled}
            className="flex-1"
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault()
                addHeader()
              }
            }}
          />
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={addHeader}
            disabled={disabled || !newHeaderKey.trim()}
            aria-label="Add header"
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>
        {Object.keys(headers).length > 0 ? (
          <div className="space-y-1 mt-1">
            {Object.entries(headers).map(([key, value]) => (
              <div
                key={key}
                className="flex items-center gap-2 rounded-md bg-muted px-3 py-1.5 text-xs font-mono"
              >
                <span className="font-medium">{key}:</span>
                <span className="flex-1 truncate text-muted-foreground">{value}</span>
                {!disabled && (
                  <button
                    type="button"
                    onClick={() => removeHeader(key)}
                    className="rounded p-0.5 hover:bg-accent"
                    aria-label={`Remove header ${key}`}
                  >
                    <X className="h-3 w-3" />
                  </button>
                )}
              </div>
            ))}
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">
            Optional. Add headers like Authorization or Content-Type.
          </p>
        )}
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="include-contact"
          checked={includeContactData}
          onChange={(e) => updateConfig({ includeContactData: e.target.checked })}
          disabled={disabled}
          className="h-4 w-4 rounded border-input"
        />
        <Label htmlFor="include-contact" className="text-sm font-normal">
          Include full contact data in request body
        </Label>
      </div>
    </div>
  )
}
