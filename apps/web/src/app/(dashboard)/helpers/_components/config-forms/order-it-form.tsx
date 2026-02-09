'use client'

// Schema: see schemas.ts > orderItSchema

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FormTextField, DynamicList, AddItemRow } from './form-fields'
import type { ConfigFormProps } from './types'

export function OrderItForm({ config, onChange, disabled }: ConfigFormProps) {
  const productId = (config.product_id as number) || 0
  const quantity = (config.quantity as number) || 1
  const promoCodes = (config.promo_codes as string[]) || []
  const applyTag = (config.apply_tag as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Product ID"
        type="number"
        placeholder="Enter Keap product ID"
        value={productId.toString()}
        onChange={(e) => updateConfig({ product_id: parseInt(e.target.value, 10) || 0 })}
        disabled={disabled}
        description="The Keap product ID to add to the order."
        required
      />

      <FormTextField
        label="Quantity"
        type="number"
        placeholder="1"
        value={quantity.toString()}
        onChange={(e) => updateConfig({ quantity: parseInt(e.target.value, 10) || 1 })}
        disabled={disabled}
        description="Number of items to order (default: 1)."
      />

      <DynamicList<string>
        label="Promo Codes (optional)"
        description="Add promotional codes to apply to the order."
        items={promoCodes}
        onItemsChange={(items) => updateConfig({ promo_codes: items })}
        renderItem={(code) => <span className="font-mono">{code}</span>}
        renderAddForm={(onAdd) => {
          const PromoCodeAddForm = () => {
            const [code, setCode] = useState('')
            const handleAdd = () => {
              if (!code.trim()) return
              onAdd(code.trim())
              setCode('')
            }
            return (
              <AddItemRow onAdd={handleAdd} disabled={disabled} canAdd={!!code.trim()}>
                <Input
                  placeholder="Enter promo code"
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  disabled={disabled}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault()
                      handleAdd()
                    }
                  }}
                />
              </AddItemRow>
            )
          }
          return <PromoCodeAddForm />
        }}
        disabled={disabled}
      />

      <FormTextField
        label="Apply Tag (optional)"
        placeholder="Tag ID to apply after order creation"
        value={applyTag}
        onChange={(e) => updateConfig({ apply_tag: e.target.value })}
        disabled={disabled}
        description="Tag ID to apply to the contact after the order is created."
      />
    </div>
  )
}
