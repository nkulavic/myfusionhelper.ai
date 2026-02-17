'use client'

import { useState, useEffect, Suspense } from 'react'
import { useSearchParams } from 'next/navigation'
import {
  User,
  Building,
  CreditCard,
  Key,
  Users,
  Bell,
  Loader2,
  Sparkles,
  Shield,
} from 'lucide-react'
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

const tabs = [
  { id: 'profile', name: 'Profile', icon: User },
  { id: 'account', name: 'Account', icon: Building },
  { id: 'security', name: 'Security', icon: Shield },
  { id: 'team', name: 'Team', icon: Users },
  { id: 'api-keys', name: 'API Keys', icon: Key },
  { id: 'ai', name: 'AI Assistant', icon: Sparkles },
  { id: 'billing', name: 'Billing', icon: CreditCard },
  { id: 'notifications', name: 'Notifications', icon: Bell },
]

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
  const tabParam = searchParams.get('tab')
  const validTabs = tabs.map((t) => t.id)
  const [activeTab, setActiveTab] = useState(
    tabParam && validTabs.includes(tabParam) ? tabParam : 'profile'
  )

  useEffect(() => {
    if (tabParam && validTabs.includes(tabParam)) {
      setActiveTab(tabParam)
    }
  }, [tabParam])

  return (
    <div className="animate-fade-in-up space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-muted-foreground">Manage your account and preferences</p>
      </div>

      {/* Mobile: horizontal scrolling tabs */}
      <div className="flex gap-1.5 overflow-x-auto pb-2 md:hidden">
        {tabs.map((tab) => (
          <Button
            key={tab.id}
            variant={activeTab === tab.id ? 'default' : 'ghost'}
            size="sm"
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              'flex-shrink-0',
              activeTab !== tab.id && 'text-muted-foreground'
            )}
          >
            <tab.icon className="h-4 w-4" />
            {tab.name}
          </Button>
        ))}
      </div>

      <div className="flex gap-6">
        {/* Desktop: vertical side nav */}
        <nav className="hidden w-48 space-y-1 md:block">
          {tabs.map((tab) => (
            <Button
              key={tab.id}
              variant={activeTab === tab.id ? 'default' : 'ghost'}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'w-full justify-start',
                activeTab !== tab.id && 'text-muted-foreground'
              )}
            >
              <tab.icon className="h-4 w-4" />
              {tab.name}
            </Button>
          ))}
        </nav>

        <div className="flex-1 min-w-0">
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
    </div>
  )
}
