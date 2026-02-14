'use client'

// Go keys: rules (array of {tag_id, has_tag, points}), target_field
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { FieldPicker } from '@/components/field-picker'
import { DynamicList } from './form-fields'
import type { ConfigFormProps } from './types'

interface ScoringRule {
  tagId: string
  hasTag: boolean
  points: number
}

export function ScoreItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const targetField = (config.targetField as string) || ''
  const rules = (config.rules as ScoringRule[]) || []

  const [newTagId, setNewTagId] = useState('')
  const [newHasTag, setNewHasTag] = useState(true)
  const [newPoints, setNewPoints] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label>Save Score To</Label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={targetField}
          onChange={(value) => updateConfig({ targetField: value })}
          placeholder="Select field to save score..."
          filterType="number"
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field where the calculated score will be stored.
        </p>
      </div>

      <DynamicList<ScoringRule>
        label="Scoring Rules"
        description="Add rules to calculate the score. Each matching rule adds (or subtracts) points."
        items={rules}
        onItemsChange={(items) => updateConfig({ rules: items })}
        disabled={disabled}
        renderAddForm={(onAdd) => (
          <div className="rounded-md border p-3 space-y-3">
            <div className="grid gap-2">
              <Label className="text-xs">Tag ID</Label>
              <Input
                placeholder="Enter tag ID"
                value={newTagId}
                onChange={(e) => setNewTagId(e.target.value)}
                disabled={disabled}
                className="h-9"
              />
            </div>
            <div className="flex items-center justify-between gap-4 rounded-lg border p-3">
              <div className="space-y-0.5">
                <Label className="text-sm font-medium">Has Tag</Label>
                <p className="text-xs text-muted-foreground">
                  Award points when the contact has this tag (on) or is missing it (off).
                </p>
              </div>
              <Switch
                checked={newHasTag}
                onCheckedChange={setNewHasTag}
                disabled={disabled}
              />
            </div>
            <div className="grid gap-2">
              <Label className="text-xs">Points</Label>
              <Input
                type="number"
                placeholder="Points (e.g. 10 or -5)"
                value={newPoints}
                onChange={(e) => setNewPoints(e.target.value)}
                disabled={disabled}
                className="h-9 w-32"
              />
            </div>
            <button
              type="button"
              className="text-xs text-primary hover:underline"
              disabled={disabled || !newTagId.trim() || !newPoints}
              onClick={() => {
                const points = parseInt(newPoints, 10)
                if (!newTagId.trim() || isNaN(points)) return
                onAdd({ tagId: newTagId.trim(), hasTag: newHasTag, points })
                setNewTagId('')
                setNewHasTag(true)
                setNewPoints('')
              }}
            >
              + Add rule
            </button>
          </div>
        )}
        renderItem={(rule) => (
          <div className="flex items-center gap-2">
            <span className="font-medium">
              {rule.hasTag ? 'Has' : 'Missing'} tag
            </span>
            <span className="font-mono text-muted-foreground">{rule.tagId}</span>
            <span className="ml-auto font-semibold">
              {rule.points > 0 ? '+' : ''}
              {rule.points} pts
            </span>
          </div>
        )}
      />
    </div>
  )
}
