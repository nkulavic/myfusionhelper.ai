'use client'

// Go keys: lcv_total_orders, lcv_total_spend, lcv_average_order, lcv_total_due, include_zero

import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { FormSelect, InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

const includeZeroOptions = [
  { value: 'true', label: 'Yes' },
  { value: 'false', label: 'No' },
]

export function ClvForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const lcvTotalOrders = (config.lcvTotalOrders as string) || ''
  const lcvTotalSpend = (config.lcvTotalSpend as string) || ''
  const lcvAverageOrder = (config.lcvAverageOrder as string) || ''
  const lcvTotalDue = (config.lcvTotalDue as string) || ''
  const includeZero = config.includeZero === true || config.includeZero === 'true'

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <InfoBanner>
        Calculate customer lifetime value metrics from purchase history. Select the CRM fields where each value will be saved.
      </InfoBanner>

      <div className="grid gap-2">
        <Label>Total Orders Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={lcvTotalOrders}
          onChange={(value) => updateConfig({ lcvTotalOrders: value })}
          placeholder="Field to store total order count..."
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <Label>Total Spend Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={lcvTotalSpend}
          onChange={(value) => updateConfig({ lcvTotalSpend: value })}
          placeholder="Field to store total spend..."
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <Label>Average Order Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={lcvAverageOrder}
          onChange={(value) => updateConfig({ lcvAverageOrder: value })}
          placeholder="Field to store average order value..."
          disabled={disabled}
        />
      </div>

      <div className="grid gap-2">
        <Label>Total Due Field</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={lcvTotalDue}
          onChange={(value) => updateConfig({ lcvTotalDue: value })}
          placeholder="Field to store total amount due..."
          disabled={disabled}
        />
      </div>

      <FormSelect
        label="Include Zero-Value Orders"
        value={includeZero ? 'true' : 'false'}
        onValueChange={(value) => updateConfig({ includeZero: value === 'true' })}
        options={includeZeroOptions}
        disabled={disabled}
        description="Whether to include orders with a zero dollar value in calculations."
      />
    </div>
  )
}
