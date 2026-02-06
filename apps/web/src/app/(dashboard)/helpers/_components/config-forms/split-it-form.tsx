'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ConfigFormProps } from './types'

const ratioOptions = [
  { value: 50, label: '50/50', description: 'Even split between groups' },
  { value: 60, label: '60/40', description: '60% Group A, 40% Group B' },
  { value: 70, label: '70/30', description: '70% Group A, 30% Group B' },
  { value: 80, label: '80/20', description: '80% Group A, 20% Group B' },
  { value: 90, label: '90/10', description: '90% Group A, 10% Group B' },
]

export function SplitItForm({ config, onChange, disabled }: ConfigFormProps) {
  const groupATag = (config.groupATag as string) || ''
  const groupBTag = (config.groupBTag as string) || ''
  const ratio = (config.ratio as number) ?? 50

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label htmlFor="split-ratio">Split Ratio</Label>
        <select
          id="split-ratio"
          value={ratio}
          onChange={(e) => updateConfig({ ratio: parseInt(e.target.value, 10) })}
          disabled={disabled}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        >
          {ratioOptions.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        <p className="text-xs text-muted-foreground">
          {ratioOptions.find((o) => o.value === ratio)?.description}
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="grid gap-2">
          <Label htmlFor="split-group-a">
            Group A Tag ({ratio}%)
          </Label>
          <Input
            id="split-group-a"
            placeholder="Tag ID for Group A"
            value={groupATag}
            onChange={(e) => updateConfig({ groupATag: e.target.value })}
            disabled={disabled}
          />
        </div>
        <div className="grid gap-2">
          <Label htmlFor="split-group-b">
            Group B Tag ({100 - ratio}%)
          </Label>
          <Input
            id="split-group-b"
            placeholder="Tag ID for Group B"
            value={groupBTag}
            onChange={(e) => updateConfig({ groupBTag: e.target.value })}
            disabled={disabled}
          />
        </div>
      </div>

      <div className="rounded-md border border-dashed bg-muted/50 p-3">
        <p className="text-xs text-muted-foreground">
          Each contact processed by this helper will be randomly assigned to Group A
          or Group B based on the split ratio. The corresponding tag will be applied.
          Useful for A/B testing campaigns or splitting audiences.
        </p>
      </div>
    </div>
  )
}
