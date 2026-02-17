'use client'

// Schema: see schemas.ts > ipNotificationsSchema
import { FieldPicker } from '@/components/field-picker'
import { TagPicker } from '@/components/tag-picker'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { X, Plus } from 'lucide-react'
import type { ConfigFormProps } from './types'

export function IpNotificationsForm({
  config,
  onChange,
  disabled,
  platformId,
  connectionId,
}: ConfigFormProps) {
  const ipAddress = (config.ip_address as string) || ''
  const matchCountries = (config.match_countries as string[]) || []
  const matchRegions = (config.match_regions as string[]) || []
  const applyTag = (config.apply_tag as string) || ''
  const saveLocationTo = (config.save_location_to as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addCountry = () => {
    updateConfig({ match_countries: [...matchCountries, ''] })
  }

  const updateCountry = (index: number, value: string) => {
    const updated = [...matchCountries]
    updated[index] = value.toUpperCase()
    updateConfig({ match_countries: updated })
  }

  const removeCountry = (index: number) => {
    const updated = matchCountries.filter((_, i) => i !== index)
    updateConfig({ match_countries: updated })
  }

  const addRegion = () => {
    updateConfig({ match_regions: [...matchRegions, ''] })
  }

  const updateRegion = (index: number, value: string) => {
    const updated = [...matchRegions]
    updated[index] = value
    updateConfig({ match_regions: updated })
  }

  const removeRegion = (index: number) => {
    const updated = matchRegions.filter((_, i) => i !== index)
    updateConfig({ match_regions: updated })
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
        <label className="text-sm font-medium">Match Countries (optional)</label>
        <div className="space-y-2">
          {matchCountries.map((country, index) => (
            <div key={index} className="flex gap-2">
              <Input
                value={country}
                onChange={(e) => updateCountry(index, e.target.value)}
                placeholder="US, CA, GB..."
                disabled={disabled}
                className="flex-1"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => removeCountry(index)}
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
            onClick={addCountry}
            disabled={disabled}
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Country Code
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">
          Country codes to trigger notification (e.g., US, CA, GB). Leave empty to match all countries.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Match Regions (optional)</label>
        <div className="space-y-2">
          {matchRegions.map((region, index) => (
            <div key={index} className="flex gap-2">
              <Input
                value={region}
                onChange={(e) => updateRegion(index, e.target.value)}
                placeholder="California, Texas..."
                disabled={disabled}
                className="flex-1"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => removeRegion(index)}
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
            onClick={addRegion}
            disabled={disabled}
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Region/State
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">
          Region/state names to trigger notification. Leave empty to match all regions.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Apply Tag (optional)</label>
        <TagPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={applyTag ? [applyTag] : []}
          onChange={(value) => updateConfig({ apply_tag: value[0] || '' })}
          placeholder="Select tag to apply when location matches..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Tag to apply when the IP location matches your criteria.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Save Location To (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={saveLocationTo}
          onChange={(value) => updateConfig({ save_location_to: value })}
          placeholder="Field to save formatted location..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Field to save the formatted location string (e.g., "San Francisco, California, United States").
        </p>
      </div>
    </div>
  )
}
