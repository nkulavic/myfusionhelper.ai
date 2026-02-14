'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { X, Plus } from 'lucide-react'
import { FormTextField, InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

const COMMON_TIMEZONES = [
  'UTC',
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'America/Phoenix',
  'Europe/London',
  'Europe/Paris',
  'Asia/Tokyo',
  'Australia/Sydney',
]

interface TimeRoute {
  startTime: string
  endTime: string
  url: string
  label?: string
}

export function RouteItByTimeForm({ config, onChange, disabled }: ConfigFormProps) {
  const timeRoutes = (config.timeRoutes as TimeRoute[]) || []
  const fallbackUrl = (config.fallbackUrl as string) || ''
  const timezone = (config.timezone as string) || 'UTC'
  const saveToField = (config.saveToField as string) || ''
  const applyTag = (config.applyTag as string) || ''

  const [newStartTime, setNewStartTime] = useState('09:00')
  const [newEndTime, setNewEndTime] = useState('17:00')
  const [newUrl, setNewUrl] = useState('')
  const [newLabel, setNewLabel] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addTimeRoute = () => {
    if (!newUrl.trim() || !newStartTime || !newEndTime) return

    const newRoute: TimeRoute = {
      startTime: newStartTime,
      endTime: newEndTime,
      url: newUrl.trim(),
      label: newLabel.trim() || undefined,
    }

    updateConfig({ timeRoutes: [...timeRoutes, newRoute] })
    setNewStartTime('09:00')
    setNewEndTime('17:00')
    setNewUrl('')
    setNewLabel('')
  }

  const removeTimeRoute = (index: number) => {
    updateConfig({
      timeRoutes: timeRoutes.filter((_, i) => i !== index),
    })
  }

  const updateTimeRoute = (index: number, updates: Partial<TimeRoute>) => {
    const updated = [...timeRoutes]
    updated[index] = { ...updated[index], ...updates }
    updateConfig({ timeRoutes: updated })
  }

  return (
    <div className="space-y-6">
      <InfoBanner>
        Routes contacts to different URLs based on the current time of day. Supports overnight
        ranges (e.g., 22:00-06:00). First matching time range wins.
      </InfoBanner>

      <div className="space-y-4">
        {/* Existing time routes */}
        {timeRoutes.length > 0 && (
          <div>
            <Label className="text-sm font-medium">Time Routes</Label>
            <p className="text-xs text-muted-foreground mb-3">Evaluated in order (first match wins)</p>

            <div className="space-y-3">
              {timeRoutes.map((route, index) => (
                <div key={index} className="p-3 border rounded-md space-y-2 bg-muted/50">
                  <div className="flex gap-2 items-start">
                    <div className="flex-1 space-y-2">
                      <div className="flex gap-2">
                        <div className="flex-1">
                          <Label className="text-xs">Start Time</Label>
                          <Input
                            type="time"
                            value={route.startTime}
                            onChange={(e) => updateTimeRoute(index, { startTime: e.target.value })}
                            disabled={disabled}
                            className="mt-1"
                          />
                        </div>
                        <div className="flex-1">
                          <Label className="text-xs">End Time</Label>
                          <Input
                            type="time"
                            value={route.endTime}
                            onChange={(e) => updateTimeRoute(index, { endTime: e.target.value })}
                            disabled={disabled}
                            className="mt-1"
                          />
                        </div>
                      </div>
                      <div>
                        <Label className="text-xs">URL</Label>
                        <Input
                          type="url"
                          value={route.url}
                          onChange={(e) => updateTimeRoute(index, { url: e.target.value })}
                          disabled={disabled}
                          placeholder="https://example.com/morning"
                          className="mt-1"
                        />
                      </div>
                      <div>
                        <Label className="text-xs">Label (Optional)</Label>
                        <Input
                          value={route.label || ''}
                          onChange={(e) => updateTimeRoute(index, { label: e.target.value })}
                          disabled={disabled}
                          placeholder="Business hours"
                          className="mt-1"
                        />
                      </div>
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={() => removeTimeRoute(index)}
                      disabled={disabled}
                      className="mt-5"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Add new time route */}
        <div>
          <Label className="text-sm font-medium">Add Time Route</Label>
          <div className="mt-2 p-3 border rounded-md space-y-2">
            <div className="flex gap-2">
              <div className="flex-1">
                <Label className="text-xs">Start Time</Label>
                <Input
                  type="time"
                  value={newStartTime}
                  onChange={(e) => setNewStartTime(e.target.value)}
                  disabled={disabled}
                  className="mt-1"
                />
              </div>
              <div className="flex-1">
                <Label className="text-xs">End Time</Label>
                <Input
                  type="time"
                  value={newEndTime}
                  onChange={(e) => setNewEndTime(e.target.value)}
                  disabled={disabled}
                  className="mt-1"
                />
              </div>
            </div>
            <div>
              <Label className="text-xs">URL</Label>
              <Input
                type="url"
                value={newUrl}
                onChange={(e) => setNewUrl(e.target.value)}
                disabled={disabled}
                placeholder="https://example.com/route"
                className="mt-1"
              />
            </div>
            <div>
              <Label className="text-xs">Label (Optional)</Label>
              <Input
                value={newLabel}
                onChange={(e) => setNewLabel(e.target.value)}
                disabled={disabled}
                placeholder="Business hours"
                className="mt-1"
              />
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={addTimeRoute}
              disabled={disabled || !newUrl.trim() || !newStartTime || !newEndTime}
              className="w-full"
            >
              <Plus className="h-4 w-4 mr-2" />
              Add Time Route
            </Button>
          </div>
        </div>

        {/* Timezone selector */}
        <div>
          <Label htmlFor="timezone" className="text-sm font-medium">
            Timezone
          </Label>
          <p className="text-xs text-muted-foreground mb-2">
            Timezone for time-of-day calculation
          </p>
          <Select
            value={timezone}
            onValueChange={(value) => updateConfig({ timezone: value })}
            disabled={disabled}
          >
            <SelectTrigger id="timezone">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {COMMON_TIMEZONES.map((tz) => (
                <SelectItem key={tz} value={tz}>
                  {tz}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Fallback URL */}
        <FormTextField
          label="Fallback URL"
          description="Default URL if no time route matches"
          value={fallbackUrl}
          onChange={(value) => updateConfig({ fallbackUrl: value })}
          disabled={disabled}
          placeholder="https://example.com/default"
          type="url"
        />

        {/* Optional: Save to field */}
        <FormTextField
          label="Save to Field (Optional)"
          description="CRM field to save the selected URL to"
          value={saveToField}
          onChange={(value) => updateConfig({ saveToField: value })}
          disabled={disabled}
          placeholder="redirected_url"
        />

        {/* Optional: Apply tag */}
        <FormTextField
          label="Apply Tag (Optional)"
          description="Tag to apply after routing"
          value={applyTag}
          onChange={(value) => updateConfig({ applyTag: value })}
          disabled={disabled}
          placeholder="routed_by_time"
        />
      </div>
    </div>
  )
}
