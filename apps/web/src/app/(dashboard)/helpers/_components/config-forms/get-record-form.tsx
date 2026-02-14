'use client'

// Schema: see schemas.ts > getRecordSchema
import { FieldPicker } from '@/components/field-picker'
import { FormSelect } from './form-fields'
import type { ConfigFormProps } from './types'

const typeOptions = [
  { value: 'invoice', label: 'Invoice' },
  { value: 'job', label: 'Job' },
  { value: 'subscription', label: 'Subscription' },
  { value: 'lead', label: 'Lead' },
  { value: 'creditcard', label: 'Credit Card' },
  { value: 'payment', label: 'Payment' },
]

const fromFieldsByType: Record<string, { value: string; label: string }[]> = {
  invoice: [
    { value: 'Id', label: 'ID' },
    { value: 'InvoiceTotal', label: 'Invoice Total' },
    { value: 'TotalDue', label: 'Total Due' },
    { value: 'TotalPaid', label: 'Total Paid' },
    { value: 'DateCreated', label: 'Date Created' },
    { value: 'PayStatus', label: 'Pay Status' },
    { value: 'Description', label: 'Description' },
    { value: 'ProductSold', label: 'Product Sold' },
    { value: 'PromoCode', label: 'Promo Code' },
    { value: 'RefundStatus', label: 'Refund Status' },
    { value: 'CreditStatus', label: 'Credit Status' },
  ],
  job: [
    { value: 'Id', label: 'ID' },
    { value: 'JobTitle', label: 'Job Title' },
    { value: 'JobStatus', label: 'Status' },
    { value: 'DueDate', label: 'Due Date' },
    { value: 'StartDate', label: 'Start Date' },
    { value: 'DateCreated', label: 'Date Created' },
    { value: 'OrderTotal', label: 'Order Total' },
  ],
  subscription: [
    { value: 'Id', label: 'ID' },
    { value: 'Status', label: 'Status' },
    { value: 'StartDate', label: 'Start Date' },
    { value: 'NextBillDate', label: 'Next Bill Date' },
    { value: 'BillingAmount', label: 'Billing Amount' },
    { value: 'PaidThruDate', label: 'Paid Through Date' },
  ],
  lead: [
    { value: 'Id', label: 'ID' },
    { value: 'OpportunityTitle', label: 'Title' },
    { value: 'StageID', label: 'Stage ID' },
    { value: 'EstimatedCloseDate', label: 'Est. Close Date' },
    { value: 'ProjectedRevenueHigh', label: 'Projected Revenue' },
    { value: 'DateCreated', label: 'Date Created' },
  ],
  creditcard: [
    { value: 'Id', label: 'ID' },
    { value: 'CardType', label: 'Card Type' },
    { value: 'Last4', label: 'Last 4 Digits' },
    { value: 'ExpirationMonth', label: 'Expiration Month' },
    { value: 'ExpirationYear', label: 'Expiration Year' },
    { value: 'Status', label: 'Status' },
  ],
  payment: [
    { value: 'Id', label: 'ID' },
    { value: 'PayAmt', label: 'Payment Amount' },
    { value: 'PayDate', label: 'Payment Date' },
    { value: 'PayType', label: 'Payment Type' },
    { value: 'PayNote', label: 'Payment Note' },
    { value: 'InvoiceId', label: 'Invoice ID' },
  ],
}

export function GetRecordForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const type = (config.type as string) || 'invoice'
  const fromField = (config.fromField as string) || ''
  const toField = (config.toField as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const availableFields = fromFieldsByType[type] || []

  return (
    <div className="space-y-4">
      <FormSelect
        label="Record Type"
        value={type}
        onValueChange={(v) => updateConfig({ type: v, fromField: '' })}
        options={typeOptions}
        disabled={disabled}
      />

      <FormSelect
        label="Field to Retrieve"
        value={fromField}
        onValueChange={(v) => updateConfig({ fromField: v })}
        options={availableFields}
        placeholder="Select a field..."
        disabled={disabled}
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Save To Contact Field</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={toField}
          onChange={(value) => updateConfig({ toField: value })}
          placeholder="Select field to save data..."
          disabled={disabled}
        />
      </div>
    </div>
  )
}
