'use client'

import { useState } from 'react'
import { User, Building, CreditCard, Key, Users, Bell, XCircle, Loader2, Copy, Plus, Sparkles, Eye, EyeOff, CheckCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import { useAPIKeys, useCreateAPIKey, useRevokeAPIKey, useUpdateAccount } from '@/lib/hooks/use-settings'
import { Skeleton } from '@/components/ui/skeleton'
import {
  useAISettingsStore,
  providerModels,
  providerLabels,
  modelLabels,
  type AIProvider,
} from '@/lib/stores/ai-settings-store'

const tabs = [
  { id: 'profile', name: 'Profile', icon: User },
  { id: 'account', name: 'Account', icon: Building },
  { id: 'team', name: 'Team', icon: Users },
  { id: 'api-keys', name: 'API Keys', icon: Key },
  { id: 'ai', name: 'AI Assistant', icon: Sparkles },
  { id: 'billing', name: 'Billing', icon: CreditCard },
  { id: 'notifications', name: 'Notifications', icon: Bell },
]

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState('profile')

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-muted-foreground">Manage your account and preferences</p>
      </div>

      <div className="flex gap-6">
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

        <div className="flex-1">
          {activeTab === 'profile' && <ProfileTab />}
          {activeTab === 'account' && <AccountTab />}
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

function ProfileTab() {
  const { user } = useAuthStore()
  const [name, setName] = useState(user?.name || '')
  const [email, setEmail] = useState(user?.email || '')

  const initials = name
    ? name.split(' ').map((n) => n[0]).join('').toUpperCase().slice(0, 2)
    : 'U'

  return (
    <div className="space-y-6">
      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-4 text-lg font-semibold">Profile Information</h2>
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary text-2xl font-bold text-primary-foreground">
              {initials}
            </div>
          </div>
          <div>
            <label className="mb-2 block text-sm font-medium">Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            />
          </div>
          <div>
            <label className="mb-2 block text-sm font-medium">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
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
  )
}

function AccountTab() {
  const { currentAccount } = useWorkspaceStore()
  const updateAccount = useUpdateAccount()
  const [accountName, setAccountName] = useState(currentAccount?.name || '')
  const [company, setCompany] = useState(currentAccount?.company || '')

  const handleSave = () => {
    if (!currentAccount) return
    updateAccount.mutate({
      accountId: currentAccount.accountId,
      input: { name: accountName, company },
    })
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-4 text-lg font-semibold">Workspace Details</h2>
        <div className="space-y-4">
          <div>
            <label className="mb-2 block text-sm font-medium">Workspace Name</label>
            <input
              type="text"
              value={accountName}
              onChange={(e) => setAccountName(e.target.value)}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            />
          </div>
          <div>
            <label className="mb-2 block text-sm font-medium">Company</label>
            <input
              type="text"
              value={company}
              onChange={(e) => setCompany(e.target.value)}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            />
          </div>
          {currentAccount && (
            <div>
              <label className="mb-2 block text-sm font-medium">Workspace ID</label>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  value={currentAccount.accountId}
                  readOnly
                  className="flex h-10 flex-1 rounded-md border border-input bg-muted px-3 py-2 text-sm font-mono text-muted-foreground"
                />
                <button
                  onClick={() => navigator.clipboard.writeText(currentAccount.accountId)}
                  className="rounded-md border border-input bg-background px-3 py-2 text-sm hover:bg-accent"
                >
                  Copy
                </button>
              </div>
            </div>
          )}
        </div>
        <div className="mt-6 flex justify-end">
          <button
            onClick={handleSave}
            disabled={updateAccount.isPending}
            className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            {updateAccount.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
            Save Changes
          </button>
        </div>
      </div>

      <div className="rounded-lg border border-destructive/30 bg-card p-6">
        <h2 className="mb-2 text-lg font-semibold text-destructive">Danger Zone</h2>
        <p className="mb-4 text-sm text-muted-foreground">
          Irreversible actions that affect your entire workspace
        </p>
        <div className="space-y-3">
          <div className="flex items-center justify-between rounded-md border border-destructive/30 p-4">
            <div>
              <p className="text-sm font-medium text-destructive">Delete Workspace</p>
              <p className="text-xs text-muted-foreground">Permanently delete this workspace and all associated data</p>
            </div>
            <button className="rounded-md border border-destructive/30 bg-background px-4 py-2 text-sm font-medium text-destructive hover:bg-destructive/10">
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

function APIKeysTab() {
  const { data: apiKeys, isLoading } = useAPIKeys()
  const createKey = useCreateAPIKey()
  const revokeKey = useRevokeAPIKey()
  const [newKeyName, setNewKeyName] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [newKey, setNewKey] = useState<string | null>(null)

  const handleCreate = () => {
    if (!newKeyName.trim()) return
    createKey.mutate(
      { name: newKeyName, permissions: ['execute_helpers'] },
      {
        onSuccess: (res) => {
          if (res.data && 'key' in res.data) {
            setNewKey(res.data.key)
          }
          setNewKeyName('')
          setShowCreate(false)
        },
      }
    )
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg border bg-card p-6">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">API Keys</h2>
          <button
            onClick={() => setShowCreate(!showCreate)}
            className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            <Plus className="h-4 w-4" />
            Create New Key
          </button>
        </div>
        <p className="mb-4 text-sm text-muted-foreground">
          Use API keys to authenticate helper executions from your CRM automations.
        </p>

        {showCreate && (
          <div className="mb-4 flex items-center gap-2 rounded-md border p-4">
            <input
              type="text"
              value={newKeyName}
              onChange={(e) => setNewKeyName(e.target.value)}
              placeholder="Key name (e.g. Production)"
              className="flex h-10 flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm"
            />
            <button
              onClick={handleCreate}
              disabled={createKey.isPending || !newKeyName.trim()}
              className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              {createKey.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
              Create
            </button>
            <button
              onClick={() => setShowCreate(false)}
              className="rounded-md border border-input px-3 py-2 text-sm hover:bg-accent"
            >
              Cancel
            </button>
          </div>
        )}

        {newKey && (
          <div className="mb-4 rounded-md border border-success/30 bg-success/10 p-4">
            <p className="mb-2 text-sm font-medium text-success">
              API key created! Copy it now — you won&apos;t be able to see it again.
            </p>
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded bg-background p-2 font-mono text-sm">{newKey}</code>
              <button
                onClick={() => { navigator.clipboard.writeText(newKey); setNewKey(null) }}
                className="inline-flex items-center gap-1 rounded-md border px-3 py-2 text-sm hover:bg-accent"
              >
                <Copy className="h-3 w-3" />
                Copy
              </button>
            </div>
          </div>
        )}

        {isLoading ? (
          <div className="space-y-3">
            {[1, 2].map((i) => (
              <div key={i} className="flex items-center justify-between rounded-md border p-4">
                <div>
                  <Skeleton className="h-4 w-32" />
                  <Skeleton className="mt-1 h-3 w-40" />
                </div>
                <Skeleton className="h-8 w-20" />
              </div>
            ))}
          </div>
        ) : apiKeys && apiKeys.length > 0 ? (
          <div className="space-y-3">
            {apiKeys.map((key) => (
              <div key={key.id} className="flex items-center justify-between rounded-md border p-4">
                <div>
                  <p className="font-medium">{key.name}</p>
                  <p className="font-mono text-sm text-muted-foreground">{key.keyPrefix}...</p>
                  <p className="text-xs text-muted-foreground">
                    Created {new Date(key.createdAt).toLocaleDateString()}
                    {key.lastUsedAt && ` · Last used ${new Date(key.lastUsedAt).toLocaleDateString()}`}
                  </p>
                </div>
                <button
                  onClick={() => revokeKey.mutate(key.id)}
                  disabled={revokeKey.isPending}
                  className="rounded-md border border-input bg-background px-3 py-1.5 text-sm text-destructive hover:bg-destructive/10 disabled:opacity-50"
                >
                  Revoke
                </button>
              </div>
            ))}
          </div>
        ) : (
          <div className="py-8 text-center text-sm text-muted-foreground">
            No API keys yet. Create one to start executing helpers via API.
          </div>
        )}
      </div>
    </div>
  )
}

function AITab() {
  const {
    preferredProvider,
    preferredModel,
    anthropicApiKey,
    openaiApiKey,
    setPreferredProvider,
    setPreferredModel,
    setAnthropicKey,
    setOpenAIKey,
  } = useAISettingsStore()

  const [showAnthropicKey, setShowAnthropicKey] = useState(false)
  const [showOpenAIKey, setShowOpenAIKey] = useState(false)
  const [testStatus, setTestStatus] = useState<'idle' | 'testing' | 'success' | 'error'>('idle')
  const [testError, setTestError] = useState('')

  const handleTest = async () => {
    setTestStatus('testing')
    setTestError('')
    try {
      const headers: Record<string, string> = { 'Content-Type': 'application/json' }
      if (preferredProvider === 'anthropic' && anthropicApiKey) {
        headers['x-anthropic-key'] = anthropicApiKey
      }
      if (preferredProvider === 'openai' && openaiApiKey) {
        headers['x-openai-key'] = openaiApiKey
      }

      const res = await fetch('/api/chat', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          messages: [{ role: 'user', content: 'Say "Connection successful!" in 3 words or less.' }],
          provider: preferredProvider,
          model: preferredModel,
        }),
      })

      if (!res.ok) {
        const text = await res.text()
        throw new Error(text || `HTTP ${res.status}`)
      }

      setTestStatus('success')
    } catch (err) {
      setTestStatus('error')
      setTestError(err instanceof Error ? err.message : 'Connection failed')
    }
  }

  const needsKey = preferredProvider === 'anthropic' || preferredProvider === 'openai'
  const currentKey = preferredProvider === 'anthropic' ? anthropicApiKey : openaiApiKey

  return (
    <div className="space-y-6">
      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-2 text-lg font-semibold">AI Assistant Configuration</h2>
        <p className="mb-6 text-sm text-muted-foreground">
          Groq is included free for all users. For Claude or OpenAI, provide your own API key — stored locally in your browser, never sent to our servers.
        </p>

        {/* Provider Selection */}
        <div className="space-y-4">
          <div>
            <label className="mb-2 block text-sm font-medium">AI Provider</label>
            <div className="grid grid-cols-3 gap-3">
              {(Object.keys(providerLabels) as AIProvider[]).map((provider) => (
                <button
                  key={provider}
                  onClick={() => setPreferredProvider(provider)}
                  className={cn(
                    'rounded-lg border p-3 text-left transition-all',
                    preferredProvider === provider
                      ? 'border-primary bg-primary/5 ring-1 ring-primary'
                      : 'hover:border-primary/50'
                  )}
                >
                  <p className="text-sm font-medium">{providerLabels[provider]}</p>
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {provider === 'groq' && 'Fast open-source models'}
                    {provider === 'anthropic' && 'Claude Sonnet & Opus'}
                    {provider === 'openai' && 'GPT-4o & GPT-4o Mini'}
                  </p>
                </button>
              ))}
            </div>
          </div>

          {/* Model Selection */}
          <div>
            <label className="mb-2 block text-sm font-medium">Model</label>
            <select
              value={preferredModel}
              onChange={(e) => setPreferredModel(e.target.value)}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            >
              {providerModels[preferredProvider].map((model) => (
                <option key={model} value={model}>
                  {modelLabels[model] || model}
                </option>
              ))}
            </select>
          </div>

          {/* API Key (for BYOK providers) */}
          {needsKey && (
            <div>
              <label className="mb-2 block text-sm font-medium">
                {preferredProvider === 'anthropic' ? 'Anthropic' : 'OpenAI'} API Key
              </label>
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <input
                    type={
                      (preferredProvider === 'anthropic' ? showAnthropicKey : showOpenAIKey)
                        ? 'text'
                        : 'password'
                    }
                    value={currentKey || ''}
                    onChange={(e) => {
                      const val = e.target.value || null
                      if (preferredProvider === 'anthropic') setAnthropicKey(val)
                      else setOpenAIKey(val)
                    }}
                    placeholder={
                      preferredProvider === 'anthropic'
                        ? 'sk-ant-...'
                        : 'sk-...'
                    }
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 pr-10 py-2 font-mono text-sm"
                  />
                  <button
                    type="button"
                    onClick={() => {
                      if (preferredProvider === 'anthropic') setShowAnthropicKey(!showAnthropicKey)
                      else setShowOpenAIKey(!showOpenAIKey)
                    }}
                    className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 hover:bg-accent"
                  >
                    {(preferredProvider === 'anthropic' ? showAnthropicKey : showOpenAIKey) ? (
                      <EyeOff className="h-4 w-4 text-muted-foreground" />
                    ) : (
                      <Eye className="h-4 w-4 text-muted-foreground" />
                    )}
                  </button>
                </div>
              </div>
              <p className="mt-1.5 text-xs text-muted-foreground">
                Get your API key from{' '}
                {preferredProvider === 'anthropic' ? (
                  <a href="https://console.anthropic.com" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">
                    console.anthropic.com
                  </a>
                ) : (
                  <a href="https://platform.openai.com/api-keys" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">
                    platform.openai.com
                  </a>
                )}
              </p>
            </div>
          )}

          {/* Test Connection */}
          <div className="flex items-center gap-3 pt-2">
            <button
              onClick={handleTest}
              disabled={testStatus === 'testing' || (needsKey && !currentKey)}
              className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              {testStatus === 'testing' && <Loader2 className="h-4 w-4 animate-spin" />}
              Test Connection
            </button>
            {testStatus === 'success' && (
              <span className="inline-flex items-center gap-1.5 text-sm text-success">
                <CheckCircle className="h-4 w-4" />
                Connected successfully
              </span>
            )}
            {testStatus === 'error' && (
              <span className="inline-flex items-center gap-1.5 text-sm text-destructive">
                <XCircle className="h-4 w-4" />
                {testError}
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function BillingTab() {
  // TODO: Connect to real billing API when Stripe integration is built
  return (
    <div className="space-y-6">
      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-4 text-lg font-semibold">Current Plan</h2>
        <div className="flex items-center justify-between rounded-lg bg-primary/10 p-4">
          <div>
            <p className="text-lg font-bold">Pro Plan</p>
            <p className="text-sm text-muted-foreground">$49/month</p>
          </div>
          <button className="rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent">
            Manage Subscription
          </button>
        </div>
      </div>
      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-4 text-lg font-semibold">Usage This Month</h2>
        <div className="space-y-4">
          {[
            { label: 'Helper Executions', used: 12847, limit: 50000 },
            { label: 'API Calls', used: 8234, limit: 100000 },
            { label: 'Active Helpers', used: 14, limit: 50 },
            { label: 'Connections', used: 3, limit: 5 },
          ].map((item) => (
            <div key={item.label}>
              <div className="mb-2 flex justify-between text-sm">
                <span>{item.label}</span>
                <span>{item.used.toLocaleString()} / {item.limit.toLocaleString()}</span>
              </div>
              <div className="h-2 rounded-full bg-muted">
                <div
                  className="h-2 rounded-full bg-primary"
                  style={{ width: `${Math.min((item.used / item.limit) * 100, 100)}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function TeamTab() {
  // TODO: Connect to real team management API when endpoint is built
  const { user } = useAuthStore()
  return (
    <div className="space-y-6">
      <div className="rounded-lg border bg-card p-6">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Team Members</h2>
          <button className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
            Invite Member
          </button>
        </div>
        <div className="space-y-3">
          <div className="flex items-center justify-between rounded-md border p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-sm font-medium text-primary-foreground">
                {user?.name?.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2) || 'U'}
              </div>
              <div>
                <p className="text-sm font-medium">{user?.name || 'You'}</p>
                <p className="text-xs text-muted-foreground">{user?.email || ''}</p>
              </div>
            </div>
            <span className="rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary">
              Owner
            </span>
          </div>
        </div>
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-4 text-lg font-semibold">Roles & Permissions</h2>
        <div className="space-y-3 text-sm">
          {[
            { role: 'Owner', color: 'primary', desc: 'Full access to everything including billing, team management, and workspace deletion' },
            { role: 'Admin', color: 'blue', desc: 'Can manage helpers, connections, team members, and API keys. Cannot access billing or delete workspace.' },
            { role: 'Member', color: 'green', desc: 'Can manage helpers and execute them. Cannot manage connections, team, or API keys.' },
            { role: 'Viewer', color: 'gray', desc: 'Read-only access to helpers, executions, and analytics. Cannot make changes.' },
          ].map((item) => (
            <div key={item.role} className="flex items-start gap-3">
              <span className={cn(
                'mt-0.5 inline-block w-16 rounded-full px-2.5 py-0.5 text-center text-xs font-medium',
                item.color === 'primary' && 'bg-primary/10 text-primary',
                item.color === 'blue' && 'bg-info/10 text-info',
                item.color === 'green' && 'bg-success/10 text-success',
                item.color === 'gray' && 'bg-muted text-muted-foreground',
              )}>
                {item.role}
              </span>
              <p className="flex-1 text-muted-foreground">{item.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function NotificationsTab() {
  // TODO: Connect to real notification preferences API when endpoint is built
  return (
    <div className="space-y-6">
      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-2 text-lg font-semibold">Email Notifications</h2>
        <p className="mb-4 text-sm text-muted-foreground">
          Choose which notifications you want to receive by email.
        </p>
        <div className="space-y-4">
          <NotificationToggle label="Execution Failures" description="Get notified when a helper execution fails" defaultChecked={true} />
          <NotificationToggle label="Connection Issues" description="Alerts when a CRM connection has errors or token expiry" defaultChecked={true} />
          <NotificationToggle label="Usage Alerts" description="Warnings when approaching plan limits (80% threshold)" defaultChecked={true} />
          <NotificationToggle label="Weekly Summary" description="Weekly digest of execution stats and insights" defaultChecked={false} />
          <NotificationToggle label="New Features" description="Product updates, new helpers, and platform announcements" defaultChecked={false} />
        </div>
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-2 text-lg font-semibold">In-App Notifications</h2>
        <p className="mb-4 text-sm text-muted-foreground">
          Control the notification bell in the dashboard header.
        </p>
        <div className="space-y-4">
          <NotificationToggle label="Real-time Execution Status" description="Show running and recently completed executions" defaultChecked={true} />
          <NotificationToggle label="AI Insights" description="Surface AI-powered suggestions and anomaly alerts" defaultChecked={true} />
          <NotificationToggle label="System Maintenance" description="Scheduled maintenance and downtime notices" defaultChecked={true} />
        </div>
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-2 text-lg font-semibold">Webhook Notifications</h2>
        <p className="mb-4 text-sm text-muted-foreground">
          Send notifications to external services via webhooks.
        </p>
        <div>
          <label className="mb-2 block text-sm font-medium">Webhook URL</label>
          <div className="flex gap-2">
            <input
              type="url"
              placeholder="https://hooks.slack.com/services/..."
              className="flex h-10 flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
            <button className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
              Save
            </button>
          </div>
          <p className="mt-1.5 text-xs text-muted-foreground">
            We&apos;ll POST JSON to this URL for critical events (failures, connection issues)
          </p>
        </div>
      </div>
    </div>
  )
}

function NotificationToggle({
  label,
  description,
  defaultChecked,
}: {
  label: string
  description: string
  defaultChecked: boolean
}) {
  const [checked, setChecked] = useState(defaultChecked)

  return (
    <div className="flex items-center justify-between">
      <div>
        <p className="text-sm font-medium">{label}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
      <button
        onClick={() => setChecked(!checked)}
        className={cn(
          'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
          checked ? 'bg-primary' : 'bg-muted'
        )}
      >
        <span
          className={cn(
            'inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
            checked ? 'translate-x-6' : 'translate-x-1'
          )}
        />
      </button>
    </div>
  )
}
