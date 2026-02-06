'use client'

import { useState } from 'react'
import { Plus, X, GripVertical } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

interface ScoringRule {
  condition: string
  tagId: string
  points: number
}

export function ScoreItForm({ config, onChange, disabled }: ConfigFormProps) {
  const targetField = (config.targetField as string) || ''
  const rules = (config.rules as ScoringRule[]) || []
  const resetBefore = (config.resetBefore as boolean) ?? false

  const [newCondition, setNewCondition] = useState('has_tag')
  const [newTagId, setNewTagId] = useState('')
  const [newPoints, setNewPoints] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addRule = () => {
    const tagId = newTagId.trim()
    const points = parseInt(newPoints, 10)
    if (!tagId || isNaN(points)) return
    updateConfig({
      rules: [...rules, { condition: newCondition, tagId, points }],
    })
    setNewTagId('')
    setNewPoints('')
  }

  const removeRule = (index: number) => {
    updateConfig({ rules: rules.filter((_, i) => i !== index) })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="score-target">Target Field</Label>
        <Input
          id="score-target"
          placeholder="e.g. _LeadScore, custom_score_1"
          value={targetField}
          onChange={(e) => updateConfig({ targetField: e.target.value })}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          The contact field where the calculated score will be stored.
        </p>
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="reset-before"
          checked={resetBefore}
          onChange={(e) => updateConfig({ resetBefore: e.target.checked })}
          disabled={disabled}
          className="h-4 w-4 rounded border-input"
        />
        <Label htmlFor="reset-before" className="text-sm font-normal">
          Reset score to 0 before calculating
        </Label>
      </div>

      <div className="grid gap-2">
        <Label>Scoring Rules</Label>
        <div className="rounded-md border p-3 space-y-3">
          <div className="flex gap-2">
            <select
              value={newCondition}
              onChange={(e) => setNewCondition(e.target.value)}
              disabled={disabled}
              className="h-9 w-[140px] rounded-md border border-input bg-background px-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
            >
              <option value="has_tag">Has tag</option>
              <option value="missing_tag">Missing tag</option>
              <option value="tag_count_gt">Tag count &gt;</option>
              <option value="tag_count_lt">Tag count &lt;</option>
            </select>
            <Input
              placeholder="Tag ID"
              value={newTagId}
              onChange={(e) => setNewTagId(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  addRule()
                }
              }}
              disabled={disabled}
              className="h-9 flex-1"
            />
            <Input
              type="number"
              placeholder="Pts"
              value={newPoints}
              onChange={(e) => setNewPoints(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  addRule()
                }
              }}
              disabled={disabled}
              className="h-9 w-20"
            />
            <Button
              type="button"
              variant="outline"
              size="icon"
              className="h-9 w-9"
              onClick={addRule}
              disabled={disabled || !newTagId.trim() || !newPoints}
              aria-label="Add rule"
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>

          {rules.length > 0 ? (
            <div className="space-y-1.5">
              {rules.map((rule, i) => (
                <div
                  key={i}
                  className="flex items-center gap-2 rounded-md bg-muted px-3 py-1.5 text-xs"
                >
                  <GripVertical className="h-3 w-3 text-muted-foreground/50 flex-shrink-0" />
                  <span className="font-medium capitalize">
                    {rule.condition.replace(/_/g, ' ')}
                  </span>
                  <span className="font-mono text-muted-foreground">{rule.tagId}</span>
                  <span className="ml-auto font-semibold">
                    {rule.points > 0 ? '+' : ''}
                    {rule.points} pts
                  </span>
                  {!disabled && (
                    <button
                      type="button"
                      onClick={() => removeRule(i)}
                      className="rounded p-0.5 hover:bg-accent"
                      aria-label="Remove rule"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <p className="text-xs text-muted-foreground">
              Add rules to calculate the score. Each matching rule adds (or subtracts) points.
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
