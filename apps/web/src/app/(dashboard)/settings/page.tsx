'use client'

import { Suspense } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { ProfileTab } from './_components/profile-tab'
import { AccountTab } from './_components/account-tab'
import { SecurityTab } from './_components/security-tab'
import { TeamTab } from './_components/team-tab'
import { APIKeysTab } from './_components/api-keys-tab'
import { AITab } from './_components/ai-tab'
import { BillingTab } from './_components/billing-tab'
import { NotificationsTab } from './_components/notifications-tab'
import { settingsTabs } from './_lib/settings-tabs'

export default function SettingsPage() {
  return (
    <Suspense
      fallback={
        <div className="animate-fade-in-up p-6">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      }
    >
      <SettingsContent />
    </Suspense>
  )
}

function SettingsContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const tabParam = searchParams.get('tab')
  const validTabs = settingsTabs.map((t) => t.id)
  const activeTab = tabParam && validTabs.includes(tabParam) ? tabParam : 'profile'

  return (
    <div className="animate-fade-in-up space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-muted-foreground">Manage your account and preferences</p>
      </div>

      {/* Mobile: horizontal scrolling tabs (sidebar is overlay on mobile) */}
      <div className="flex gap-1.5 overflow-x-auto pb-2 lg:hidden">
        {settingsTabs.map((tab) => (
          <Button
            key={tab.id}
            variant={activeTab === tab.id ? 'default' : 'ghost'}
            size="sm"
            onClick={() => router.push(`/settings?tab=${tab.id}`, { scroll: false })}
            className={cn(
              'flex-shrink-0',
              activeTab !== tab.id && 'text-muted-foreground',
            )}
          >
            <tab.icon className="h-4 w-4" />
            {tab.name}
          </Button>
        ))}
      </div>

      {/* Tab content — full width (sidebar IS the nav on desktop) */}
      <div className="min-w-0">
        {activeTab === 'profile' && <ProfileTab />}
        {activeTab === 'account' && <AccountTab />}
        {activeTab === 'security' && <SecurityTab />}
        {activeTab === 'api-keys' && <APIKeysTab />}
        {activeTab === 'ai' && <AITab />}
        {activeTab === 'billing' && <BillingTab />}
        {activeTab === 'team' && <TeamTab />}
        {activeTab === 'notifications' && <NotificationsTab />}
      </div>
    </div>
  )
}
