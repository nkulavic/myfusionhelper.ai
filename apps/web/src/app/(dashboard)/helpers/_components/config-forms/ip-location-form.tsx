'use client'

// Schema: see schemas.ts > ipLocationSchema
import { FieldPicker } from '@/components/field-picker'
import type { ConfigFormProps } from './types'

export function IpLocationForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const ipField = (config.ipField as string) || ''
  const cityField = (config.cityField as string) || ''
  const stateField = (config.stateField as string) || ''
  const countryField = (config.countryField as string) || ''
  const zipField = (config.zipField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <label className="text-sm font-medium">IP Address Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={ipField}
          onChange={(value) => updateConfig({ ipField: value })}
          placeholder="Field containing IP address..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field that contains the IP address to look up.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">City Field (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={cityField}
          onChange={(value) => updateConfig({ cityField: value })}
          placeholder="Store city..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Field to store the resolved city name.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">State/Region Field (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={stateField}
          onChange={(value) => updateConfig({ stateField: value })}
          placeholder="Store state/region..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Field to store the resolved state or region.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Country Field (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={countryField}
          onChange={(value) => updateConfig({ countryField: value })}
          placeholder="Store country..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Field to store the resolved country name.
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">ZIP/Postal Code Field (optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={zipField}
          onChange={(value) => updateConfig({ zipField: value })}
          placeholder="Store ZIP/postal code..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Field to store the resolved ZIP or postal code.
        </p>
      </div>
    </div>
  )
}
