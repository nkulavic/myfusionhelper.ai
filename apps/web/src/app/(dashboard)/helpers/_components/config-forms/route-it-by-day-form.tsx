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
import { FormTextField, InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

const DAYS_OF_WEEK = [
  { value: 'Monday', label: 'Monday' },
  { value: 'Tuesday', label: 'Tuesday' },
  { value: 'Wednesday', label: 'Wednesday' },
  { value: 'Thursday', label: 'Thursday' },
  { value: 'Friday', label: 'Friday' },
  { value: 'Saturday', label: 'Saturday' },
  { value: 'Sunday', label: 'Sunday' },
]

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

export function RouteItByDayForm({ config, onChange, disabled }: ConfigFormProps) {
  const dayRoutes = (config.dayRoutes as Record<string, string>) || {}
  const fallbackUrl = (config.fallbackUrl as string) || ''
  const timezone = (config.timezone as string) || 'UTC'
  const saveToField = (config.saveToField as string) || ''
  const applyTag = (config.applyTag as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const updateDayRoute = (day: string, url: string) => {
    const updated = { ...dayRoutes }
    if (!url.trim()) {
      delete updated[day]
    } else {
      updated[day] = url.trim()
    }
    updateConfig({ dayRoutes: updated })
  }

  return (
    <div className="space-y-6">
      <InfoBanner>
        Routes contacts to different URLs based on the day of the week. Timezone-aware for accurate
        day detection.
      </InfoBanner>

      <div className="space-y-4">
        {/* Day-to-URL mappings */}
        <div>
          <Label className="text-sm font-medium">Day Routing</Label>
          <p className="text-xs text-muted-foreground mb-3">
            Map each day of the week to a destination URL
          </p>

          <div className="space-y-2">
            {DAYS_OF_WEEK.map(({ value, label }) => (
              <div key={value} className="flex gap-2 items-center">
                <div className="w-32">
                  <Label className="text-sm">{label}</Label>
                </div>
                <Input
                  value={dayRoutes[value] || ''}
                  onChange={(e) => updateDayRoute(value, e.target.value)}
                  disabled={disabled}
                  placeholder="https://example.com/monday"
                  type="url"
                  className="flex-1"
                />
              </div>
            ))}
          </div>
        </div>

        {/* Timezone selector */}
        <div>
          <Label htmlFor="timezone" className="text-sm font-medium">
            Timezone
          </Label>
          <p className="text-xs text-muted-foreground mb-2">
            Timezone for day-of-week calculation
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
          description="Default URL if no day route is configured"
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
          placeholder="routed_by_day"
        />
      </div>
    </div>
  )
}
