'use client'

import { useState } from 'react'
import { User, Building, CreditCard, Key, Users, Bell } from 'lucide-react'
import { cn } from '@/lib/utils'

const tabs = [
  { id: 'profile', name: 'Profile', icon: User },
  { id: 'account', name: 'Account', icon: Building },
  { id: 'team', name: 'Team', icon: Users },
  { id: 'api-keys', name: 'API Keys', icon: Key },
  { id: 'billing', name: 'Billing', icon: CreditCard },
  { id: 'notifications', name: 'Notifications', icon: Bell },
]

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState('profile')

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-muted-foreground">Manage your account and preferences</p>
      </div>

      <div className="flex gap-6">
        {/* Sidebar */}
        <nav className="w-48 space-y-1">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                activeTab === tab.id
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-accent hover:text-foreground'
              )}
            >
              <tab.icon className="h-4 w-4" />
              {tab.name}
            </button>
          ))}
        </nav>

        {/* Content */}
        <div className="flex-1">
          {activeTab === 'profile' && (
            <div className="space-y-6">
              <div className="rounded-lg border bg-card p-6">
                <h2 className="mb-4 text-lg font-semibold">Profile Information</h2>
                <div className="space-y-4">
                  <div className="flex items-center gap-4">
                    <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary text-2xl font-bold text-primary-foreground">
                      JD
                    </div>
                    <button className="rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent">
                      Change Avatar
                    </button>
                  </div>
                  <div className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <label className="mb-2 block text-sm font-medium">First Name</label>
                      <input
                        type="text"
                        defaultValue="John"
                        className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      />
                    </div>
                    <div>
                      <label className="mb-2 block text-sm font-medium">Last Name</label>
                      <input
                        type="text"
                        defaultValue="Doe"
                        className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      />
                    </div>
                  </div>
                  <div>
                    <label className="mb-2 block text-sm font-medium">Email</label>
                    <input
                      type="email"
                      defaultValue="john@example.com"
                      className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    />
                  </div>
                </div>
                <div className="mt-6 flex justify-end">
                  <button className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
                    Save Changes
                  </button>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'api-keys' && (
            <div className="space-y-6">
              <div className="rounded-lg border bg-card p-6">
                <div className="mb-4 flex items-center justify-between">
                  <h2 className="text-lg font-semibold">API Keys</h2>
                  <button className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
                    Create New Key
                  </button>
                </div>
                <p className="mb-4 text-sm text-muted-foreground">
                  Use API keys to authenticate helper executions from your CRM automations.
                </p>
                <div className="space-y-3">
                  <div className="flex items-center justify-between rounded-md border p-4">
                    <div>
                      <p className="font-medium">Production Key</p>
                      <p className="font-mono text-sm text-muted-foreground">mfh_live_abc...xyz</p>
                    </div>
                    <div className="flex gap-2">
                      <button className="rounded-md border border-input bg-background px-3 py-1.5 text-sm hover:bg-accent">
                        Copy
                      </button>
                      <button className="rounded-md border border-input bg-background px-3 py-1.5 text-sm text-red-500 hover:bg-red-50">
                        Revoke
                      </button>
                    </div>
                  </div>
                  <div className="flex items-center justify-between rounded-md border p-4">
                    <div>
                      <p className="font-medium">Test Key</p>
                      <p className="font-mono text-sm text-muted-foreground">mfh_test_def...uvw</p>
                    </div>
                    <div className="flex gap-2">
                      <button className="rounded-md border border-input bg-background px-3 py-1.5 text-sm hover:bg-accent">
                        Copy
                      </button>
                      <button className="rounded-md border border-input bg-background px-3 py-1.5 text-sm text-red-500 hover:bg-red-50">
                        Revoke
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'billing' && (
            <div className="space-y-6">
              <div className="rounded-lg border bg-card p-6">
                <h2 className="mb-4 text-lg font-semibold">Current Plan</h2>
                <div className="flex items-center justify-between rounded-lg bg-primary/10 p-4">
                  <div>
                    <p className="text-lg font-bold">Pro Plan</p>
                    <p className="text-sm text-muted-foreground">$49/month • Renews Feb 15, 2026</p>
                  </div>
                  <button className="rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent">
                    Manage Subscription
                  </button>
                </div>
              </div>
              <div className="rounded-lg border bg-card p-6">
                <h2 className="mb-4 text-lg font-semibold">Usage This Month</h2>
                <div className="space-y-4">
                  <div>
                    <div className="mb-2 flex justify-between text-sm">
                      <span>Helper Executions</span>
                      <span>12,847 / 50,000</span>
                    </div>
                    <div className="h-2 rounded-full bg-muted">
                      <div className="h-2 w-1/4 rounded-full bg-primary" />
                    </div>
                  </div>
                  <div>
                    <div className="mb-2 flex justify-between text-sm">
                      <span>API Calls</span>
                      <span>8,234 / 100,000</span>
                    </div>
                    <div className="h-2 rounded-full bg-muted">
                      <div className="h-2 w-[8%] rounded-full bg-primary" />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab !== 'profile' && activeTab !== 'api-keys' && activeTab !== 'billing' && (
            <div className="flex flex-col items-center justify-center rounded-lg border bg-card p-12 text-center">
              <p className="text-lg font-medium">Coming Soon</p>
              <p className="text-sm text-muted-foreground">
                This section is under development
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
