'use client'

import { useState, useEffect } from 'react'
import { Loader2 } from 'lucide-react'
import { useNotificationPreferences, useUpdateNotificationPreferences } from '@/lib/hooks/use-settings'
import { useAuthStore } from '@/lib/stores/auth-store'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { TestEmailLog } from './test-email-log'

const urlRegex = /^https?:\/\/.+/

export function NotificationsTab() {
  const { data: prefs, isLoading } = useNotificationPreferences()
  const updatePrefs = useUpdateNotificationPreferences()
  const user = useAuthStore((s) => s.user)
  const isTestUser = user?.email?.endsWith('@test.myfusionhelper.ai') ?? false
  const [webhookUrl, setWebhookUrl] = useState('')
  const [webhookInitialized, setWebhookInitialized] = useState(false)

  // Initialize webhook URL from server data
  useEffect(() => {
    if (prefs?.webhookUrl && !webhookInitialized) {
      setWebhookUrl(prefs.webhookUrl)
      setWebhookInitialized(true)
    }
  }, [prefs?.webhookUrl, webhookInitialized])

  const handleToggle = (key: string, value: boolean) => {
    updatePrefs.mutate({ [key]: value })
  }

  const webhookValid = !webhookUrl || urlRegex.test(webhookUrl)

  const emailNotifications = [
    {
      key: 'executionFailures',
      label: 'Execution Failures',
      description: 'Get notified when a helper execution fails',
    },
    {
      key: 'connectionIssues',
      label: 'Connection Issues',
      description: 'Alerts when a CRM connection has errors or token expiry',
    },
    {
      key: 'usageAlerts',
      label: 'Usage Alerts',
      description: 'Warnings when approaching plan limits (80% threshold)',
    },
    {
      key: 'weeklySummary',
      label: 'Weekly Summary',
      description: 'Weekly digest of execution stats and insights',
    },
    {
      key: 'newFeatures',
      label: 'New Features',
      description: 'Product updates, new helpers, and platform announcements',
    },
  ]

  const inAppNotifications = [
    {
      key: 'realtimeStatus',
      label: 'Real-time Execution Status',
      description: 'Show running and recently completed executions',
    },
    {
      key: 'aiInsights',
      label: 'AI Insights',
      description: 'Surface AI-powered suggestions and anomaly alerts',
    },
    {
      key: 'systemMaintenance',
      label: 'System Maintenance',
      description: 'Scheduled maintenance and downtime notices',
    },
  ]

  if (isLoading) {
    return (
      <div className="space-y-6">
        {[1, 2, 3].map((i) => (
          <Card key={i}>
            <CardHeader>
              <Skeleton className="h-6 w-48" />
            </CardHeader>
            <CardContent className="space-y-4">
              {[1, 2, 3].map((j) => (
                <Skeleton key={j} className="h-12 w-full" />
              ))}
            </CardContent>
          </Card>
        ))}
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Email Notifications</CardTitle>
          <CardDescription>Choose which notifications you want to receive by email.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {emailNotifications.map((item) => (
            <div key={item.key} className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label className="text-sm">{item.label}</Label>
                <p className="text-xs text-muted-foreground">{item.description}</p>
              </div>
              <Switch
                checked={(prefs?.[item.key as keyof typeof prefs] as boolean) ?? false}
                onCheckedChange={(checked) => handleToggle(item.key, checked)}
              />
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">In-App Notifications</CardTitle>
          <CardDescription>Control the notification bell in the dashboard header.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {inAppNotifications.map((item) => (
            <div key={item.key} className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label className="text-sm">{item.label}</Label>
                <p className="text-xs text-muted-foreground">{item.description}</p>
              </div>
              <Switch
                checked={(prefs?.[item.key as keyof typeof prefs] as boolean) ?? false}
                onCheckedChange={(checked) => handleToggle(item.key, checked)}
              />
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Webhook Notifications</CardTitle>
          <CardDescription>Send notifications to external services via webhooks.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <Label htmlFor="webhook-url">Webhook URL</Label>
            <div className="flex gap-2">
              <Input
                id="webhook-url"
                type="url"
                value={webhookUrl}
                onChange={(e) => setWebhookUrl(e.target.value)}
                placeholder="https://hooks.slack.com/services/..."
                className="flex-1 font-mono"
              />
              <Button
                onClick={() => updatePrefs.mutate({ webhookUrl: webhookUrl || undefined })}
                disabled={updatePrefs.isPending || !webhookValid}
              >
                {updatePrefs.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                Save
              </Button>
            </div>
            {webhookUrl && !webhookValid && (
              <p className="text-xs text-destructive">
                Please enter a valid URL starting with http:// or https://
              </p>
            )}
            <p className="text-xs text-muted-foreground">
              We&apos;ll POST JSON to this URL for critical events (failures, connection issues)
            </p>
          </div>
        </CardContent>
      </Card>

      {isTestUser && <TestEmailLog />}
    </div>
  )
}
