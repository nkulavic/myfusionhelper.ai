'use client'

// Schema: see schemas.ts > hubspotDealStagerSchema
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { FormTextField, DynamicList, AddItemRow } from './form-fields'
import type { ConfigFormProps } from './types'

interface StageTrigger {
  stageName: string
  action: string
}

export function HubspotDealStagerForm({ config, onChange, disabled }: ConfigFormProps) {
  const pipelineId = (config.pipelineId as string) || ''
  const stageTriggers = (config.stageTriggers as StageTrigger[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Pipeline ID"
        placeholder="HubSpot pipeline ID"
        value={pipelineId}
        onChange={(e) => updateConfig({ pipelineId: e.target.value })}
        disabled={disabled}
      />

      <DynamicList<StageTrigger>
        label="Stage Triggers"
        description="Define what happens when a deal enters each stage."
        items={stageTriggers}
        onItemsChange={(items) => updateConfig({ stageTriggers: items })}
        renderItem={(t) => (
          <>
            <span className="font-medium min-w-[100px]">{t.stageName}</span>
            <span className="text-muted-foreground">→</span>
            <span className="flex-1 font-mono">{t.action}</span>
          </>
        )}
        renderAddForm={(onAdd) => {
          const TriggerAddForm = () => {
            const [stage, setStage] = useState('')
            const [action, setAction] = useState('')
            const handleAdd = () => {
              if (!stage.trim() || !action.trim()) return
              onAdd({ stageName: stage.trim(), action: action.trim() })
              setStage('')
              setAction('')
            }
            return (
              <AddItemRow onAdd={handleAdd} disabled={disabled} canAdd={!!stage.trim() && !!action.trim()}>
                <Input placeholder="Stage name" value={stage} onChange={(e) => setStage(e.target.value)} disabled={disabled} className="w-1/2" />
                <Input placeholder="Action/goal" value={action} onChange={(e) => setAction(e.target.value)} disabled={disabled} className="w-1/2" onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAdd() } }} />
              </AddItemRow>
            )
          }
          return <TriggerAddForm />
        }}
        disabled={disabled}
      />
    </div>
  )
}
