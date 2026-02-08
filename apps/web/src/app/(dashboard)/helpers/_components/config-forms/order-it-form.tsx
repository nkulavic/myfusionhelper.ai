'use client'

// Schema: see schemas.ts > orderItSchema
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FormTextField, DynamicList, AddItemRow } from './form-fields'
import type { ConfigFormProps } from './types'

interface OrderProduct {
  productId: string
  quantity: number
  price: string
}

export function OrderItForm({ config, onChange, disabled }: ConfigFormProps) {
  const orderTitle = (config.orderTitle as string) || ''
  const products = (config.products as OrderProduct[]) || []
  const paymentPlan = (config.paymentPlan as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Order Title"
        placeholder="Order description"
        value={orderTitle}
        onChange={(e) => updateConfig({ orderTitle: e.target.value })}
        disabled={disabled}
      />

      <DynamicList<OrderProduct>
        label="Products"
        description="Add products to include in the order."
        items={products}
        onItemsChange={(items) => updateConfig({ products: items })}
        renderItem={(p) => (
          <>
            <span className="font-mono">{p.productId}</span>
            <span className="text-muted-foreground">x{p.quantity}</span>
            {p.price && <span className="font-medium">${p.price}</span>}
          </>
        )}
        renderAddForm={(onAdd) => {
          const ProductAddForm = () => {
            const [productId, setProductId] = useState('')
            const [qty, setQty] = useState('1')
            const [price, setPrice] = useState('')
            const handleAdd = () => {
              if (!productId.trim()) return
              onAdd({ productId: productId.trim(), quantity: parseInt(qty, 10) || 1, price: price.trim() })
              setProductId('')
              setQty('1')
              setPrice('')
            }
            return (
              <AddItemRow onAdd={handleAdd} disabled={disabled} canAdd={!!productId.trim()}>
                <Input placeholder="Product ID" value={productId} onChange={(e) => setProductId(e.target.value)} disabled={disabled} className="w-1/3" />
                <Input placeholder="Qty" type="number" min={1} value={qty} onChange={(e) => setQty(e.target.value)} disabled={disabled} className="w-16" />
                <Input placeholder="Price" value={price} onChange={(e) => setPrice(e.target.value)} disabled={disabled} className="w-24" onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAdd() } }} />
              </AddItemRow>
            )
          }
          return <ProductAddForm />
        }}
        disabled={disabled}
      />

      <FormTextField
        label="Payment Plan ID"
        placeholder="Payment plan (optional)"
        value={paymentPlan}
        onChange={(e) => updateConfig({ paymentPlan: e.target.value })}
        disabled={disabled}
      />
    </div>
  )
}
