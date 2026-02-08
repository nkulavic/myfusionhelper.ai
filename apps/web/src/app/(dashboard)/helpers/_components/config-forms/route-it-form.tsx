'use client'

// Schema: see schemas.ts > routeItSchema
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FormTextField, DynamicList, AddItemRow } from './form-fields'
import type { ConfigFormProps } from './types'

interface Route {
  label: string
  redirectUrl: string
}

export function RouteItForm({ config, onChange, disabled }: ConfigFormProps) {
  const routes = (config.routes as Route[]) || []
  const fallbackUrl = (config.fallbackUrl as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <DynamicList<Route>
        label="Routes"
        description="Define conditions and redirect URLs. Contacts matching a route will be sent to its URL."
        items={routes}
        onItemsChange={(items) => updateConfig({ routes: items })}
        renderItem={(route) => (
          <>
            <span className="font-medium min-w-[80px]">{route.label}</span>
            <span className="text-muted-foreground">→</span>
            <span className="flex-1 font-mono truncate">{route.redirectUrl}</span>
          </>
        )}
        renderAddForm={(onAdd) => {
          const RouteAddForm = () => {
            const [label, setLabel] = useState('')
            const [url, setUrl] = useState('')
            const handleAdd = () => {
              if (!url.trim()) return
              onAdd({ label: label.trim() || `Route ${routes.length + 1}`, redirectUrl: url.trim() })
              setLabel('')
              setUrl('')
            }
            return (
              <AddItemRow onAdd={handleAdd} disabled={disabled} canAdd={!!url.trim()}>
                <Input placeholder="Route label" value={label} onChange={(e) => setLabel(e.target.value)} disabled={disabled} className="w-1/3" />
                <Input placeholder="https://example.com/page" value={url} onChange={(e) => setUrl(e.target.value)} disabled={disabled} className="flex-1" onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAdd() } }} />
              </AddItemRow>
            )
          }
          return <RouteAddForm />
        }}
        disabled={disabled}
      />

      <FormTextField
        label="Fallback URL"
        placeholder="https://example.com/default"
        value={fallbackUrl}
        onChange={(e) => updateConfig({ fallbackUrl: e.target.value })}
        disabled={disabled}
        description="URL to redirect to if no route conditions match."
      />
    </div>
  )
}
