'use client'

// Schema: see schemas.ts > ipRedirectsSchema
import { FieldPicker } from '@/components/field-picker'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { X, Plus } from 'lucide-react'
import type { ConfigFormProps } from './types'

export function IpRedirectsForm({
  config,
  onChange,
  disabled,
  platformId,
  connectionId,
}: ConfigFormProps) {
  const ipAddress = (config.ip_address as string) || ''
  const countryUrls = (config.country_urls as Record<string, string>) || {}
  const defaultUrl = (config.default_url as string) || ''
  const saveRedirectTo = (config.save_redirect_to as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const countryUrlEntries = Object.entries(countryUrls)

  const addCountryUrl = () => {
    updateConfig({
      country_urls: { ...countryUrls, '': '' },
    })
  }

  const updateCountryUrlKey = (oldKey: string, newKey: string) => {
    const updated = { ...countryUrls }
    if (oldKey !== newKey) {
      const value = updated[oldKey]
      delete updated[oldKey]
      updated[newKey.toUpperCase()] = value
    }
    updateConfig({ country_urls: updated })
  }

  const updateCountryUrlValue = (key: string, value: string) => {
    updateConfig({
      country_urls: { ...countryUrls, [key]: value },
    })
  }

  const removeCountryUrl = (key: string) => {
    const updated = { ...countryUrls }
    delete updated[key]
    updateConfig({ country_urls: updated })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">IP Address Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={ipAddress}
          onChange={(value) => updateConfig({ ip_address: value })}
          placeholder="Field containing IP address..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field that contains the IP address to look up.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Country-Specific URLs</label>
        <div className="space-y-2">
          {countryUrlEntries.map(([country, url]) => (
            <div key={country || 'empty'} className="flex gap-2">
              <Input
                value={country}
                onChange={(e) => updateCountryUrlKey(country, e.target.value)}
                placeholder="Country code (US, CA, GB...)"
                disabled={disabled}
                className="w-32"
              />
              <Input
                value={url}
                onChange={(e) => updateCountryUrlValue(country, e.target.value)}
                placeholder="https://example.com/us"
                disabled={disabled}
                className="flex-1"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => removeCountryUrl(country)}
                disabled={disabled}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={addCountryUrl}
            disabled={disabled}
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Country URL
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">
          Map country codes to specific redirect URLs. Visitors from these countries will be redirected to the corresponding URL.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Default URL (optional)</label>
        <Input
          value={defaultUrl}
          onChange={(e) => updateConfig({ default_url: e.target.value })}
          placeholder="https://example.com/default"
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Fallback URL if the visitor's country doesn't match any of the rules above.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Save Redirect URL To (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={saveRedirectTo}
          onChange={(value) => updateConfig({ save_redirect_to: value })}
          placeholder="Field to save redirect URL..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          CRM field to save the determined redirect URL (requires CRM connection).
        </p>
      </div>
    </div>
  )
}
