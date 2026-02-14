'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { X, Plus } from 'lucide-react'
import { FormTextField, InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

export function RouteItByCustomForm({ config, onChange, disabled }: ConfigFormProps) {
  const fieldName = (config.fieldName as string) || ''
  const valueRoutes = (config.valueRoutes as Record<string, string>) || {}
  const fallbackUrl = (config.fallbackUrl as string) || ''
  const saveToField = (config.saveToField as string) || ''
  const applyTag = (config.applyTag as string) || ''

  const [newFieldValue, setNewFieldValue] = useState('')
  const [newUrl, setNewUrl] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addValueRoute = () => {
    if (!newFieldValue.trim() || !newUrl.trim()) return
    updateConfig({
      valueRoutes: { ...valueRoutes, [newFieldValue.trim()]: newUrl.trim() },
    })
    setNewFieldValue('')
    setNewUrl('')
  }

  const removeValueRoute = (value: string) => {
    const updated = { ...valueRoutes }
    delete updated[value]
    updateConfig({ valueRoutes: updated })
  }

  const updateValueRouteUrl = (value: string, url: string) => {
    updateConfig({
      valueRoutes: { ...valueRoutes, [value]: url },
    })
  }

  return (
    <div className="space-y-6">
      <InfoBanner>
        Routes contacts to different URLs based on a custom field value. Perfect for routing by
        status, tier, category, or any custom field.
      </InfoBanner>

      <div className="space-y-4">
        {/* Field name */}
        <FormTextField
          label="Field Name"
          description="CRM field to check for routing (e.g., status, tier, category)"
          value={fieldName}
          onChange={(value) => updateConfig({ fieldName: value })}
          disabled={disabled}
          placeholder="status"
          required
        />

        {/* Value-to-URL mappings */}
        <div>
          <Label className="text-sm font-medium">Value Routing</Label>
          <p className="text-xs text-muted-foreground mb-3">
            Map field values to destination URLs
          </p>

          {/* Existing value routes */}
          {Object.entries(valueRoutes).length > 0 && (
            <div className="space-y-2 mb-3">
              {Object.entries(valueRoutes).map(([value, url]) => (
                <div key={value} className="flex gap-2 items-center">
                  <div className="w-48">
                    <Input value={value} disabled className="bg-muted" />
                  </div>
                  <Input
                    value={url}
                    onChange={(e) => updateValueRouteUrl(value, e.target.value)}
                    disabled={disabled}
                    placeholder="https://example.com/route"
                    type="url"
                    className="flex-1"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => removeValueRoute(value)}
                    disabled={disabled}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          )}

          {/* Add new value route */}
          <div className="flex gap-2 items-end">
            <div className="w-48">
              <Label className="text-xs">Field Value</Label>
              <Input
                value={newFieldValue}
                onChange={(e) => setNewFieldValue(e.target.value)}
                disabled={disabled}
                placeholder="premium"
                className="mt-1"
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    addValueRoute()
                  }
                }}
              />
            </div>
            <div className="flex-1">
              <Label className="text-xs">URL</Label>
              <Input
                type="url"
                value={newUrl}
                onChange={(e) => setNewUrl(e.target.value)}
                disabled={disabled}
                placeholder="https://example.com/premium"
                className="mt-1"
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    addValueRoute()
                  }
                }}
              />
            </div>
            <Button
              type="button"
              variant="outline"
              size="icon"
              onClick={addValueRoute}
              disabled={disabled || !newFieldValue.trim() || !newUrl.trim()}
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Fallback URL */}
        <FormTextField
          label="Fallback URL"
          description="Default URL if the field value doesn't match any route"
          value={fallbackUrl}
          onChange={(value) => updateConfig({ fallbackUrl: value })}
          disabled={disabled}
          placeholder="https://example.com/default"
          type="url"
        />

        {/* Optional: Save to field */}
        <FormTextField
          label="Save to Field (Optional)"
          description="CRM field to save the selected URL to"
          value={saveToField}
          onChange={(value) => updateConfig({ saveToField: value })}
          disabled={disabled}
          placeholder="redirected_url"
        />

        {/* Optional: Apply tag */}
        <FormTextField
          label="Apply Tag (Optional)"
          description="Tag to apply after routing"
          value={applyTag}
          onChange={(value) => updateConfig({ applyTag: value })}
          disabled={disabled}
          placeholder="routed_by_custom"
        />
      </div>
    </div>
  )
}
