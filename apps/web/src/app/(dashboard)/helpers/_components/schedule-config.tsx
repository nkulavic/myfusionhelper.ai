'use client'

import { useState } from 'react'
import { Clock, Calendar, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useUpdateHelper } from '@/lib/hooks/use-helpers'
import type { Helper } from '@myfusionhelper/types'

interface ScheduleConfigProps {
  helper: Helper
}

const PRESETS = [
  { label: 'Every 5 minutes', value: 'rate(5 minutes)' },
  { label: 'Every 15 minutes', value: 'rate(15 minutes)' },
  { label: 'Every hour', value: 'rate(1 hour)' },
  { label: 'Every 6 hours', value: 'rate(6 hours)' },
  { label: 'Daily at 9 AM UTC', value: 'cron(0 9 * * ? *)' },
  { label: 'Daily at midnight UTC', value: 'cron(0 0 * * ? *)' },
  { label: 'Weekly (Mon 9 AM UTC)', value: 'cron(0 9 ? * MON *)' },
  { label: 'Monthly (1st at 9 AM)', value: 'cron(0 9 1 * ? *)' },
  { label: 'Custom', value: 'custom' },
] as const

function getPresetLabel(cron: string): string {
  const preset = PRESETS.find((p) => p.value === cron)
  return preset ? preset.label : 'Custom'
}

function getPresetValue(cron: string | undefined): string {
  if (!cron) return ''
  const match = PRESETS.find((p) => p.value === cron)
  return match ? match.value : 'custom'
}

export function ScheduleConfig({ helper }: ScheduleConfigProps) {
  const updateHelper = useUpdateHelper()
  const [enabled, setEnabled] = useState(helper.scheduleEnabled ?? false)
  const [preset, setPreset] = useState(getPresetValue(helper.cronExpression))
  const [customCron, setCustomCron] = useState(
    getPresetValue(helper.cronExpression) === 'custom'
      ? (helper.cronExpression ?? '')
      : ''
  )
  const [saving, setSaving] = useState(false)

  const currentCron =
    preset === 'custom' ? customCron : preset === '' ? '' : preset

  const hasChanges =
    enabled !== (helper.scheduleEnabled ?? false) ||
    (enabled && currentCron !== (helper.cronExpression ?? ''))

  const canSave = hasChanges && (!enabled || currentCron !== '')

  async function handleSave() {
    setSaving(true)
    try {
      await updateHelper.mutateAsync({
        id: helper.helperId,
        input: {
          scheduleEnabled: enabled,
          cronExpression: enabled ? currentCron : undefined,
        },
      })
    } finally {
      setSaving(false)
    }
  }

  function handlePresetChange(value: string) {
    setPreset(value)
    if (value !== 'custom') {
      setCustomCron('')
    }
  }

  function handleToggle(checked: boolean) {
    setEnabled(checked)
  }

  return (
    <div className="rounded-lg border bg-card p-5 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Calendar className="h-4 w-4 text-muted-foreground" />
          <h3 className="font-semibold">Schedule</h3>
        </div>
        <Switch checked={enabled} onCheckedChange={handleToggle} />
      </div>

      {enabled && (
        <div className="space-y-3">
          <div className="space-y-1.5">
            <Label className="text-xs">Frequency</Label>
            <Select value={preset} onValueChange={handlePresetChange}>
              <SelectTrigger className="h-8 text-xs">
                <SelectValue placeholder="Select schedule..." />
              </SelectTrigger>
              <SelectContent>
                {PRESETS.map((p) => (
                  <SelectItem key={p.value} value={p.value} className="text-xs">
                    {p.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {preset === 'custom' && (
            <div className="space-y-1.5">
              <Label className="text-xs">Cron / Rate Expression</Label>
              <Input
                className="h-8 text-xs font-mono"
                placeholder="cron(0 9 * * ? *) or rate(1 hour)"
                value={customCron}
                onChange={(e) => setCustomCron(e.target.value)}
              />
              <p className="text-[10px] text-muted-foreground">
                Uses AWS EventBridge schedule expressions.
              </p>
            </div>
          )}

          {currentCron && (
            <div className="rounded-md bg-muted px-3 py-2">
              <p className="text-[10px] font-mono text-muted-foreground break-all">
                {currentCron}
              </p>
            </div>
          )}

          {helper.lastScheduledAt && (
            <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
              <Clock className="h-3 w-3" />
              Last run: {new Date(helper.lastScheduledAt).toLocaleString()}
            </div>
          )}
        </div>
      )}

      {!enabled && helper.cronExpression && (
        <p className="text-xs text-muted-foreground">
          Schedule paused ({getPresetLabel(helper.cronExpression)})
        </p>
      )}

      {canSave && (
        <Button
          size="sm"
          className="w-full"
          onClick={handleSave}
          disabled={saving}
        >
          {saving ? (
            <>
              <Loader2 className="h-3 w-3 animate-spin" />
              Saving...
            </>
          ) : (
            'Save Schedule'
          )}
        </Button>
      )}
    </div>
  )
}
