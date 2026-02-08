'use client'

// Go keys: options { recency_calculation { thresholds }, frequency_calculation { thresholds },
//   monetary_calculation { thresholds }, save_data { 12 field mappings } }

import { useState } from 'react'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { FormTextField, InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

interface Thresholds {
  '12Threshold'?: number
  '23Threshold'?: number
  '34Threshold'?: number
  '45Threshold'?: number
}

interface CalculationConfig {
  thresholds?: Thresholds
}

interface SaveDataConfig {
  recencyScore?: string
  frequencyScore?: string
  monetaryScore?: string
  rfmCompositeScore?: string
  totalOrderValue?: string
  averageOrderValue?: string
  totalOrderCount?: string
  firstOrderDate?: string
  firstOrderValue?: string
  lastOrderDate?: string
  lastOrderValue?: string
  daysSinceLastOrder?: string
}

interface OptionsConfig {
  recencyCalculation?: CalculationConfig
  frequencyCalculation?: CalculationConfig
  monetaryCalculation?: CalculationConfig
  saveData?: SaveDataConfig
}

const thresholdLabels = [
  { key: '12Threshold', label: '1-2 Threshold', description: 'Boundary between score 1 and 2.' },
  { key: '23Threshold', label: '2-3 Threshold', description: 'Boundary between score 2 and 3.' },
  { key: '34Threshold', label: '3-4 Threshold', description: 'Boundary between score 3 and 4.' },
  { key: '45Threshold', label: '4-5 Threshold', description: 'Boundary between score 4 and 5.' },
]

const saveDataFields = [
  { key: 'recencyScore', label: 'Recency Score Field' },
  { key: 'frequencyScore', label: 'Frequency Score Field' },
  { key: 'monetaryScore', label: 'Monetary Score Field' },
  { key: 'rfmCompositeScore', label: 'RFM Composite Score Field' },
  { key: 'totalOrderValue', label: 'Total Order Value Field' },
  { key: 'averageOrderValue', label: 'Average Order Value Field' },
  { key: 'totalOrderCount', label: 'Total Order Count Field' },
  { key: 'firstOrderDate', label: 'First Order Date Field' },
  { key: 'firstOrderValue', label: 'First Order Value Field' },
  { key: 'lastOrderDate', label: 'Last Order Date Field' },
  { key: 'lastOrderValue', label: 'Last Order Value Field' },
  { key: 'daysSinceLastOrder', label: 'Days Since Last Order Field' },
]

function CollapsibleSection({
  title,
  description,
  defaultOpen = false,
  children,
}: {
  title: string
  description?: string
  defaultOpen?: boolean
  children: React.ReactNode
}) {
  const [open, setOpen] = useState(defaultOpen)
  return (
    <Collapsible open={open} onOpenChange={setOpen} className="rounded-lg border">
      <CollapsibleTrigger className="flex w-full items-center justify-between p-4 text-left hover:bg-muted/50 transition-colors">
        <div>
          <p className="text-sm font-semibold">{title}</p>
          {description && <p className="text-xs text-muted-foreground mt-0.5">{description}</p>}
        </div>
        <ChevronDown className={cn('h-4 w-4 text-muted-foreground transition-transform', open && 'rotate-180')} />
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="space-y-3 px-4 pb-4">{children}</div>
      </CollapsibleContent>
    </Collapsible>
  )
}

export function RfmForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const options = ((config.options as OptionsConfig) || {}) as OptionsConfig

  const updateOptions = (key: string, value: unknown) => {
    const currentOptions = (config.options as Record<string, unknown>) || {}
    onChange({ ...config, options: { ...currentOptions, [key]: value } })
  }

  const getThresholds = (calculationType: 'recencyCalculation' | 'frequencyCalculation' | 'monetaryCalculation'): Thresholds => {
    return options[calculationType]?.thresholds || {}
  }

  const updateThreshold = (
    calculationType: 'recencyCalculation' | 'frequencyCalculation' | 'monetaryCalculation',
    thresholdKey: string,
    value: number,
  ) => {
    const current = options[calculationType] || {}
    const currentThresholds = current.thresholds || {}
    updateOptions(calculationType, {
      ...current,
      thresholds: { ...currentThresholds, [thresholdKey]: value },
    })
  }

  const saveData = options.saveData || {}

  const updateSaveData = (key: string, value: string) => {
    const currentSaveData = options.saveData || {}
    updateOptions('saveData', { ...currentSaveData, [key]: value })
  }

  return (
    <div className="space-y-4">
      <InfoBanner>
        RFM (Recency, Frequency, Monetary) analysis calculates customer value scores based on purchase behavior.
        Configure the score thresholds for each dimension and select where to save the results.
      </InfoBanner>

      {/* Recency Calculation */}
      <CollapsibleSection
        title="Recency Calculation"
        description="Thresholds for scoring how recently a customer purchased."
      >
        {thresholdLabels.map((t) => (
          <FormTextField
            key={`recency-${t.key}`}
            label={t.label}
            type="number"
            min={0}
            value={String(getThresholds('recencyCalculation')[t.key as keyof Thresholds] ?? '')}
            onChange={(e) =>
              updateThreshold('recencyCalculation', t.key, Number(e.target.value) || 0)
            }
            disabled={disabled}
            description={t.description}
          />
        ))}
      </CollapsibleSection>

      {/* Frequency Calculation */}
      <CollapsibleSection
        title="Frequency Calculation"
        description="Thresholds for scoring how often a customer purchases."
      >
        {thresholdLabels.map((t) => (
          <FormTextField
            key={`frequency-${t.key}`}
            label={t.label}
            type="number"
            min={0}
            value={String(getThresholds('frequencyCalculation')[t.key as keyof Thresholds] ?? '')}
            onChange={(e) =>
              updateThreshold('frequencyCalculation', t.key, Number(e.target.value) || 0)
            }
            disabled={disabled}
            description={t.description}
          />
        ))}
      </CollapsibleSection>

      {/* Monetary Calculation */}
      <CollapsibleSection
        title="Monetary Calculation"
        description="Thresholds for scoring customer spend amounts."
      >
        {thresholdLabels.map((t) => (
          <FormTextField
            key={`monetary-${t.key}`}
            label={t.label}
            type="number"
            min={0}
            value={String(getThresholds('monetaryCalculation')[t.key as keyof Thresholds] ?? '')}
            onChange={(e) =>
              updateThreshold('monetaryCalculation', t.key, Number(e.target.value) || 0)
            }
            disabled={disabled}
            description={t.description}
          />
        ))}
      </CollapsibleSection>

      {/* Save Data Field Mappings */}
      <CollapsibleSection
        title="Save Data Fields"
        description="CRM fields where calculated RFM data will be stored."
        defaultOpen
      >
        {saveDataFields.map((f) => (
          <div key={f.key} className="grid gap-2">
            <Label>{f.label}</Label>
            <FieldPicker
              platformId={platformId ?? ''}
              connectionId={connectionId ?? ''}
              value={(saveData as Record<string, string>)[f.key] ?? ''}
              onChange={(value) => updateSaveData(f.key, value)}
              placeholder={`Select ${f.label.toLowerCase()}...`}
              disabled={disabled}
            />
          </div>
        ))}
      </CollapsibleSection>
    </div>
  )
}
