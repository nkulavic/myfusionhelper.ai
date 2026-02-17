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
  XCircle,
  Loader2,
  Copy,
  Plus,
  Sparkles,
  Eye,
  EyeOff,
  Check,
  CheckCircle,
  ExternalLink,
  Receipt,
  Shield,
  Smartphone,
  Lock,
  Trash2,
  Zap,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import {
  useUpdateProfile,
  useUpdatePassword,
  useAPIKeys,
  useCreateAPIKey,
  useRevokeAPIKey,
  useUpdateAccount,
  useBillingInfo,
  useInvoices,
  useCreatePortalSession,
  useCreateCheckoutSession,
  useTeamMembers,
  useInviteTeamMember,
  useUpdateTeamMember,
  useRemoveTeamMember,
  useNotificationPreferences,
  useUpdateNotificationPreferences,
} from '@/lib/hooks/use-settings'
import {
  useAISettingsStore,
  providerModels,
  providerLabels,
  modelLabels,
  type AIProvider,
} from '@/lib/stores/ai-settings-store'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Skeleton } from '@/components/ui/skeleton'
import { Progress } from '@/components/ui/progress'
import { toast } from 'sonner'
import {
  PLAN_CONFIGS,
  PAID_PLAN_IDS,
  formatLimit,
  getAnnualSavingsPercent,
  getPlanLabel,
  type PlanId,
} from '@/lib/plan-constants'
import { usePlanLimits } from '@/lib/hooks/use-plan-limits'

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
    <Suspense fallback={<div className="animate-fade-in-up p-6"><Loader2 className="h-6 w-6 animate-spin" /></div>}>
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

// ---------------------------------------------------------------------------
// Profile Tab
// ---------------------------------------------------------------------------
function ProfileTab() {
  const { user, updateUserData } = useAuthStore()
  const [name, setName] = useState(user?.name || '')
  const [email, setEmail] = useState(user?.email || '')
  const [phone, setPhone] = useState(user?.phoneNumber || '')
  const updateProfile = useUpdateProfile()

  const handleSaveProfile = () => {
    updateProfile.mutate(
      { name, email },
      {
        onSuccess: (res) => {
          if (res.data) {
            updateUserData({ name: res.data.name, email: res.data.email })
          }
        },
      }
    )
  }

  const initials = name
    ? name
        .split(' ')
        .map((n) => n[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)
    : 'U'

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Profile Information</CardTitle>
          <CardDescription>Update your personal details</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <Avatar className="h-16 w-16">
              <AvatarFallback className="bg-primary text-2xl font-bold text-primary-foreground">
                {initials}
              </AvatarFallback>
            </Avatar>
            <div>
              <p className="text-sm font-medium">{name || 'Your Name'}</p>
              <p className="text-xs text-muted-foreground">{email}</p>
            </div>
          </div>
          <Separator />
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="profile-name">Full Name</Label>
              <Input
                id="profile-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Your full name"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="profile-email">Email Address</Label>
              <Input
                id="profile-email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="profile-phone">Phone Number</Label>
              <Input
                id="profile-phone"
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                placeholder="+1 (555) 000-0000"
              />
            </div>
          </div>
        </CardContent>
        <CardFooter className="justify-end">
          <Button onClick={handleSaveProfile} disabled={updateProfile.isPending}>
            {updateProfile.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
            Save Changes
          </Button>
        </CardFooter>
      </Card>

      <ChangePasswordCard />
    </div>
  )
}

// ---------------------------------------------------------------------------
// Change Password Card (shared between Profile & Security tabs)
// ---------------------------------------------------------------------------
function ChangePasswordCard() {
  const updatePassword = useUpdatePassword()
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showCurrent, setShowCurrent] = useState(false)
  const [showNew, setShowNew] = useState(false)
  const [success, setSuccess] = useState(false)

  const passwordErrors: string[] = []
  if (newPassword && newPassword.length < 8) passwordErrors.push('At least 8 characters')
  if (newPassword && !/[A-Z]/.test(newPassword)) passwordErrors.push('One uppercase letter')
  if (newPassword && !/[a-z]/.test(newPassword)) passwordErrors.push('One lowercase letter')
  if (newPassword && !/[0-9]/.test(newPassword)) passwordErrors.push('One number')
  const mismatch = confirmPassword !== '' && confirmPassword !== newPassword
  const canSubmit =
    currentPassword.length > 0 &&
    newPassword.length >= 8 &&
    passwordErrors.length === 0 &&
    confirmPassword === newPassword &&
    !updatePassword.isPending

  const handleSubmit = () => {
    setSuccess(false)
    updatePassword.mutate(
      { currentPassword, newPassword },
      {
        onSuccess: () => {
          setCurrentPassword('')
          setNewPassword('')
          setConfirmPassword('')
          setSuccess(true)
        },
      }
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Change Password</CardTitle>
        <CardDescription>Update your account password</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {success && (
          <div className="flex items-center gap-2 rounded-md border border-success/30 bg-success/10 p-3 text-sm text-success">
            <CheckCircle className="h-4 w-4 shrink-0" />
            Password updated successfully.
          </div>
        )}
        {updatePassword.isError && (
          <div className="flex items-center gap-2 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
            <XCircle className="h-4 w-4 shrink-0" />
            {(updatePassword.error as Error)?.message || 'Failed to update password'}
          </div>
        )}
        <div className="space-y-2">
          <Label htmlFor="pwd-current">Current Password</Label>
          <div className="relative">
            <Input
              id="pwd-current"
              type={showCurrent ? 'text' : 'password'}
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              placeholder="Enter current password"
              className="pr-10"
            />
            <button
              type="button"
              onClick={() => setShowCurrent(!showCurrent)}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 hover:bg-accent"
            >
              {showCurrent ? (
                <EyeOff className="h-4 w-4 text-muted-foreground" />
              ) : (
                <Eye className="h-4 w-4 text-muted-foreground" />
              )}
            </button>
          </div>
        </div>
        <div className="space-y-2">
          <Label htmlFor="pwd-new">New Password</Label>
          <div className="relative">
            <Input
              id="pwd-new"
              type={showNew ? 'text' : 'password'}
              value={newPassword}
              onChange={(e) => {
                setNewPassword(e.target.value)
                setSuccess(false)
              }}
              placeholder="Enter new password"
              className="pr-10"
            />
            <button
              type="button"
              onClick={() => setShowNew(!showNew)}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 hover:bg-accent"
            >
              {showNew ? (
                <EyeOff className="h-4 w-4 text-muted-foreground" />
              ) : (
                <Eye className="h-4 w-4 text-muted-foreground" />
              )}
            </button>
          </div>
          {newPassword && passwordErrors.length > 0 && (
            <ul className="space-y-1 text-xs text-destructive">
              {passwordErrors.map((err) => (
                <li key={err} className="flex items-center gap-1">
                  <XCircle className="h-3 w-3 shrink-0" />
                  {err}
                </li>
              ))}
            </ul>
          )}
          {newPassword && passwordErrors.length === 0 && (
            <p className="flex items-center gap-1 text-xs text-success">
              <Check className="h-3 w-3" />
              Password meets requirements
            </p>
          )}
        </div>
        <div className="space-y-2">
          <Label htmlFor="pwd-confirm">Confirm New Password</Label>
          <Input
            id="pwd-confirm"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="Confirm new password"
          />
          {mismatch && (
            <p className="flex items-center gap-1 text-xs text-destructive">
              <XCircle className="h-3 w-3" />
              Passwords do not match
            </p>
          )}
        </div>
      </CardContent>
      <CardFooter>
        <Button onClick={handleSubmit} disabled={!canSubmit}>
          {updatePassword.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
          Update Password
        </Button>
      </CardFooter>
    </Card>
  )
}

// ---------------------------------------------------------------------------
// Account Tab
// ---------------------------------------------------------------------------
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
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Workspace Details</CardTitle>
          <CardDescription>Manage your workspace name and company info</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="workspace-name">Workspace Name</Label>
            <Input
              id="workspace-name"
              value={accountName}
              onChange={(e) => setAccountName(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="workspace-company">Company</Label>
            <Input
              id="workspace-company"
              value={company}
              onChange={(e) => setCompany(e.target.value)}
            />
          </div>
          {currentAccount && (
            <div className="space-y-2">
              <Label>Workspace ID</Label>
              <div className="flex items-center gap-2">
                <Input
                  value={currentAccount.accountId}
                  readOnly
                  className="bg-muted font-mono text-muted-foreground"
                />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => navigator.clipboard.writeText(currentAccount.accountId)}
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
        <CardFooter className="justify-end">
          <Button onClick={handleSave} disabled={updateAccount.isPending}>
            {updateAccount.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
            Save Changes
          </Button>
        </CardFooter>
      </Card>

      <Card className="border-destructive/30">
        <CardHeader>
          <CardTitle className="text-lg text-destructive">Danger Zone</CardTitle>
          <CardDescription>Irreversible actions that affect your entire workspace</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between rounded-md border border-destructive/30 p-4">
            <div>
              <p className="text-sm font-medium text-destructive">Delete Workspace</p>
              <p className="text-xs text-muted-foreground">
                Permanently delete this workspace and all associated data
              </p>
            </div>
            <Button variant="destructive" size="sm">
              Delete
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

// ---------------------------------------------------------------------------
// API Keys Tab
// ---------------------------------------------------------------------------
function APIKeysTab() {
  const { data: apiKeys, isLoading } = useAPIKeys()
  const createKey = useCreateAPIKey()
  const revokeKey = useRevokeAPIKey()
  const { canCreate: canCreateResource, getUsage, getLimit } = usePlanLimits()
  const [newKeyName, setNewKeyName] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [newKey, setNewKey] = useState<string | null>(null)
  const atApiKeyLimit = !canCreateResource('apiKeys')

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
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-lg">API Keys</CardTitle>
              <CardDescription>
                Use API keys to authenticate helper executions from your CRM automations.
              </CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xs text-muted-foreground">
                {getUsage('apiKeys')} / {getLimit('apiKeys')} keys
              </span>
              <Button onClick={() => setShowCreate(!showCreate)} disabled={atApiKeyLimit}>
                <Plus className="h-4 w-4" />
                Create New Key
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {showCreate && (
            <div className="flex items-center gap-2 rounded-md border p-4">
              <Input
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                placeholder="Key name (e.g. Production)"
                className="flex-1"
              />
              <Button onClick={handleCreate} disabled={createKey.isPending || !newKeyName.trim()}>
                {createKey.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                Create
              </Button>
              <Button variant="outline" onClick={() => setShowCreate(false)}>
                Cancel
              </Button>
            </div>
          )}

          {newKey && (
            <div className="rounded-md border border-success/30 bg-success/10 p-4">
              <p className="mb-2 text-sm font-medium text-success">
                API key created! Copy it now -- you won&apos;t be able to see it again.
              </p>
              <div className="flex items-center gap-2">
                <code className="flex-1 rounded bg-background p-2 font-mono text-sm">{newKey}</code>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    navigator.clipboard.writeText(newKey)
                    setNewKey(null)
                  }}
                >
                  <Copy className="h-3 w-3" />
                  Copy
                </Button>
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
                <div key={key.keyId} className="flex items-center justify-between rounded-md border p-4">
                  <div>
                    <p className="font-medium">{key.name}</p>
                    <p className="font-mono text-sm text-muted-foreground">{key.keyPrefix}...</p>
                    <p className="text-xs text-muted-foreground">
                      Created {new Date(key.createdAt).toLocaleDateString()}
                      {key.lastUsedAt &&
                        ` -- Last used ${new Date(key.lastUsedAt).toLocaleDateString()}`}
                    </p>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-destructive hover:bg-destructive/10"
                    onClick={() => revokeKey.mutate(key.keyId)}
                    disabled={revokeKey.isPending}
                  >
                    Revoke
                  </Button>
                </div>
              ))}
            </div>
          ) : (
            <div className="py-8 text-center text-sm text-muted-foreground">
              No API keys yet. Create one to start executing helpers via API.
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

// ---------------------------------------------------------------------------
// AI Tab
// ---------------------------------------------------------------------------
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
          messages: [
            { role: 'user', content: 'Say "Connection successful!" in 3 words or less.' },
          ],
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
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">AI Assistant Configuration</CardTitle>
          <CardDescription>
            Groq is included free for all users. For Claude or OpenAI, provide your own API key --
            stored locally in your browser, never sent to our servers.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Provider Selection */}
          <div className="space-y-2">
            <Label>AI Provider</Label>
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
          <div className="space-y-2">
            <Label htmlFor="ai-model">Model</Label>
            <select
              id="ai-model"
              value={preferredModel}
              onChange={(e) => setPreferredModel(e.target.value)}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
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
            <div className="space-y-2">
              <Label htmlFor="ai-api-key">
                {preferredProvider === 'anthropic' ? 'Anthropic' : 'OpenAI'} API Key
              </Label>
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <Input
                    id="ai-api-key"
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
                    placeholder={preferredProvider === 'anthropic' ? 'sk-ant-...' : 'sk-...'}
                    className="pr-10 font-mono"
                  />
                  <button
                    type="button"
                    onClick={() => {
                      if (preferredProvider === 'anthropic')
                        setShowAnthropicKey(!showAnthropicKey)
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
              <p className="text-xs text-muted-foreground">
                Get your API key from{' '}
                {preferredProvider === 'anthropic' ? (
                  <a
                    href="https://console.anthropic.com"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline"
                  >
                    console.anthropic.com
                  </a>
                ) : (
                  <a
                    href="https://platform.openai.com/api-keys"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline"
                  >
                    platform.openai.com
                  </a>
                )}
              </p>
            </div>
          )}

          {/* Test Connection */}
          <div className="flex items-center gap-3 pt-2">
            <Button
              onClick={handleTest}
              disabled={testStatus === 'testing' || (needsKey && !currentKey)}
            >
              {testStatus === 'testing' && <Loader2 className="h-4 w-4 animate-spin" />}
              Test Connection
            </Button>
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
        </CardContent>
      </Card>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Billing Tab
// ---------------------------------------------------------------------------
function BillingTab() {
  const searchParams = useSearchParams()
  const { data: billing, isLoading: billingLoading } = useBillingInfo()
  const { data: invoices, isLoading: invoicesLoading } = useInvoices()
  const createPortal = useCreatePortalSession()
  const createCheckout = useCreateCheckoutSession()
  const [checkoutPlan, setCheckoutPlan] = useState<string | null>(null)
  const [isAnnual, setIsAnnual] = useState(false)

  // Handle checkout cancellation toast
  useEffect(() => {
    if (searchParams.get('billing') === 'cancelled') {
      toast.info('Checkout cancelled. You can try again anytime.')
      // Clear the param from URL
      const url = new URL(window.location.href)
      url.searchParams.delete('billing')
      window.history.replaceState({}, '', url.toString())
    }
  }, [searchParams])

  const plans = PAID_PLAN_IDS.map((id) => {
    const config = PLAN_CONFIGS[id]
    return {
      id,
      name: config.name,
      description: config.description,
      popular: id === 'grow',
      features: [
        `${formatLimit(config.maxHelpers)} active helpers`,
        `${config.maxConnections} CRM connections`,
        `${formatLimit(config.maxExecutions)} monthly executions included`,
        `${config.maxApiKeys} API keys`,
        `${config.maxTeamMembers} team members`,
        config.overageRate > 0
          ? `Overage: $${config.overageRate}/execution`
          : 'Dedicated support',
      ],
    }
  })

  const handleManageSubscription = () => {
    createPortal.mutate(undefined, {
      onSuccess: (res: { data?: { url: string } }) => {
        if (res.data?.url) {
          window.location.href = res.data.url
        }
      },
    })
  }

  const handleSelectPlan = (planId: 'start' | 'grow' | 'deliver') => {
    // If user already has a paid subscription, route through Stripe Portal
    // This preserves trial periods and handles proration correctly
    if (currentPlan !== 'free') {
      handleManageSubscription()
      return
    }
    // Free users get a new checkout session
    setCheckoutPlan(planId)
    createCheckout.mutate(
      {
        plan: planId,
        returnUrl: '/settings?tab=billing',
        billingPeriod: isAnnual ? 'annual' : 'monthly',
      },
      {
        onSuccess: (res) => {
          if (res.data?.url) {
            window.location.href = res.data.url
          }
          setCheckoutPlan(null)
        },
        onError: () => {
          setCheckoutPlan(null)
        },
      }
    )
  }

  const currentPlan = billing?.plan || 'free'
  const maxSavings = Math.max(
    ...PAID_PLAN_IDS.map((id) => getAnnualSavingsPercent(id))
  )

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Current Plan</CardTitle>
          <CardDescription>Manage your subscription and billing</CardDescription>
        </CardHeader>
        <CardContent>
          {billingLoading ? (
            <Skeleton className="h-20 w-full" />
          ) : billing ? (
            <div className="flex items-center justify-between rounded-lg bg-primary/10 p-4">
              <div>
                <p className="text-lg font-bold">
                  {getPlanLabel(billing.plan)} Plan
                </p>
                {billing.priceMonthly > 0 && (
                  <p className="text-sm text-muted-foreground">
                    {billing.billingPeriod === 'annual'
                      ? `$${billing.priceAnnually}/year ($${Math.round(billing.priceAnnually / 12)}/mo)`
                      : `$${billing.priceMonthly}/month`}
                  </p>
                )}
                {billing.billingPeriod === 'annual' && billing.plan !== 'free' && (
                  <Badge variant="secondary" className="mt-1 text-xs">
                    Annual billing
                  </Badge>
                )}
                {billing.renewsAt && (
                  <p className="mt-1 text-xs text-muted-foreground">
                    Renews {new Date(billing.renewsAt * 1000).toLocaleDateString()}
                  </p>
                )}
                {billing.trialEndsAt && (
                  <p className="mt-1 text-xs text-muted-foreground">
                    Trial ends {new Date(billing.trialEndsAt * 1000).toLocaleDateString()}
                  </p>
                )}
                {billing.cancelAt && (
                  <p className="mt-1 text-xs text-destructive">
                    Cancels {new Date(billing.cancelAt * 1000).toLocaleDateString()}
                  </p>
                )}
              </div>
              {billing.plan !== 'free' && (
                <Button
                  variant="outline"
                  onClick={handleManageSubscription}
                  disabled={createPortal.isPending}
                >
                  {createPortal.isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <ExternalLink className="h-4 w-4" />
                  )}
                  Manage Subscription
                </Button>
              )}
            </div>
          ) : null}
        </CardContent>
      </Card>

      {/* Plan Tiers */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-lg">
                {currentPlan === 'free' ? 'Choose a Plan' : 'Change Plan'}
              </CardTitle>
              <CardDescription>
                {currentPlan === 'free'
                  ? 'Select the plan that best fits your needs'
                  : 'Upgrade or change your current plan'}
              </CardDescription>
            </div>
            {/* Billing toggle */}
            <div className="flex items-center gap-3">
              <span
                className={cn(
                  'text-sm font-medium transition-colors',
                  !isAnnual ? 'text-foreground' : 'text-muted-foreground'
                )}
              >
                Monthly
              </span>
              <button
                type="button"
                role="switch"
                aria-checked={isAnnual}
                onClick={() => setIsAnnual(!isAnnual)}
                className={cn(
                  'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors',
                  isAnnual ? 'bg-primary' : 'bg-muted-foreground/30'
                )}
              >
                <span
                  className={cn(
                    'pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow-sm ring-0 transition-transform',
                    isAnnual ? 'translate-x-5' : 'translate-x-0'
                  )}
                />
              </button>
              <span
                className={cn(
                  'text-sm font-medium transition-colors',
                  isAnnual ? 'text-foreground' : 'text-muted-foreground'
                )}
              >
                Annual
              </span>
              {isAnnual && (
                <Badge variant="secondary" className="text-xs">
                  Save up to {maxSavings}%
                </Badge>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            {plans.map((plan) => {
              const config = PLAN_CONFIGS[plan.id]
              const price = isAnnual ? config.annualMonthlyPrice : config.monthlyPrice
              const isCurrentPlan = currentPlan === plan.id
              const isPlanLoading = checkoutPlan === plan.id && createCheckout.isPending
              return (
                <div
                  key={plan.id}
                  className={cn(
                    'relative flex flex-col rounded-lg border p-5 transition-all',
                    plan.popular && 'border-primary shadow-sm',
                    isCurrentPlan && 'bg-primary/5 ring-1 ring-primary'
                  )}
                >
                  {plan.popular && (
                    <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                      <Badge className="gap-1">
                        <Zap className="h-3 w-3" />
                        Most Popular
                      </Badge>
                    </div>
                  )}
                  <div className="mb-4">
                    <h3 className="text-lg font-bold">{plan.name}</h3>
                    <p className="mt-1 text-xs text-muted-foreground">{plan.description}</p>
                  </div>
                  <div className="mb-4">
                    <span className="text-3xl font-bold">${price}</span>
                    <span className="text-sm text-muted-foreground">/month</span>
                    {isAnnual && (
                      <span className="ml-2 text-xs text-muted-foreground">
                        billed yearly
                      </span>
                    )}
                  </div>
                  <ul className="mb-6 flex-1 space-y-2">
                    {plan.features.map((feature) => (
                      <li key={feature} className="flex items-start gap-2 text-sm">
                        <Check className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
                        <span>{feature}</span>
                      </li>
                    ))}
                  </ul>
                  {isCurrentPlan ? (
                    <Button variant="outline" disabled className="w-full">
                      <CheckCircle className="h-4 w-4" />
                      Current Plan
                    </Button>
                  ) : currentPlan !== 'free' ? (
                    <Button
                      variant={plan.popular ? 'default' : 'outline'}
                      className="w-full"
                      onClick={handleManageSubscription}
                      disabled={createPortal.isPending}
                    >
                      {createPortal.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                      <ExternalLink className="h-4 w-4" />
                      {PAID_PLAN_IDS.indexOf(currentPlan as PlanId) <
                        PAID_PLAN_IDS.indexOf(plan.id)
                        ? 'Upgrade'
                        : 'Downgrade'}
                    </Button>
                  ) : (
                    <Button
                      variant={plan.popular ? 'default' : 'outline'}
                      className="w-full"
                      onClick={() => handleSelectPlan(plan.id as 'start' | 'grow' | 'deliver')}
                      disabled={isPlanLoading || createCheckout.isPending}
                    >
                      {isPlanLoading && <Loader2 className="h-4 w-4 animate-spin" />}
                      Get Started
                    </Button>
                  )}
                </div>
              )
            })}
          </div>
          {createCheckout.isError && (
            <p className="mt-4 text-center text-sm text-destructive">
              Failed to create checkout session. Please try again.
            </p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Usage This Month</CardTitle>
          <CardDescription>Track your resource consumption against plan limits</CardDescription>
        </CardHeader>
        <CardContent>
          {billingLoading ? (
            <div className="space-y-6">
              {[1, 2, 3, 4].map((i) => (
                <Skeleton key={i} className="h-8 w-full" />
              ))}
            </div>
          ) : billing?.usage && billing?.limits ? (
            <div className="space-y-5">
              {[
                {
                  label: 'Helper Executions',
                  used: billing.usage.monthlyExecutions,
                  limit: billing.limits.maxExecutions,
                },
                {
                  label: 'Active Helpers',
                  used: billing.usage.helpers,
                  limit: billing.limits.maxHelpers,
                },
                {
                  label: 'Connections',
                  used: billing.usage.connections,
                  limit: billing.limits.maxConnections,
                },
                {
                  label: 'API Keys',
                  used: billing.usage.apiKeys,
                  limit: billing.limits.maxApiKeys,
                },
              ].map((item) => {
                const pct = item.limit > 0 ? Math.min((item.used / item.limit) * 100, 100) : 0
                return (
                  <div key={item.label}>
                    <div className="mb-2 flex justify-between text-sm">
                      <span className="font-medium">{item.label}</span>
                      <span className="text-muted-foreground">
                        {item.used.toLocaleString()} / {item.limit.toLocaleString()}
                      </span>
                    </div>
                    <Progress value={pct} className="h-2" />
                  </div>
                )
              })}
            </div>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Invoice History</CardTitle>
          <CardDescription>View and download past invoices</CardDescription>
        </CardHeader>
        <CardContent>
          {invoicesLoading ? (
            <div className="space-y-3">
              {[1, 2].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : invoices && invoices.length > 0 ? (
            <div className="space-y-2">
              {invoices.map((inv) => (
                <div
                  key={inv.id}
                  className="flex items-center justify-between rounded-md border p-3"
                >
                  <div className="flex items-center gap-3">
                    <Receipt className="h-4 w-4 text-muted-foreground" />
                    <div>
                      <p className="text-sm font-medium">${inv.amount.toFixed(2)}</p>
                      <p className="text-xs text-muted-foreground">
                        {new Date(inv.date * 1000).toLocaleDateString('en-US', {
                          year: 'numeric',
                          month: 'long',
                        })}
                      </p>
                    </div>
                  </div>
                  <Badge variant={inv.status === 'paid' ? 'success' : 'warning'}>
                    {inv.status}
                  </Badge>
                </div>
              ))}
            </div>
          ) : (
            <p className="py-4 text-center text-sm text-muted-foreground">No invoices yet.</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Team Tab
// ---------------------------------------------------------------------------
function TeamTab() {
  const { user } = useAuthStore()
  const { currentAccount } = useWorkspaceStore()
  const { data: members, isLoading, isError } = useTeamMembers(currentAccount?.accountId || '')
  const inviteMember = useInviteTeamMember()
  const updateMember = useUpdateTeamMember()
  const removeMember = useRemoveTeamMember()
  const [showInvite, setShowInvite] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState<'admin' | 'member' | 'viewer'>('member')
  const [confirmRemove, setConfirmRemove] = useState<string | null>(null)

  const handleInvite = () => {
    if (!currentAccount || !inviteEmail.trim()) return
    inviteMember.mutate(
      { accountId: currentAccount.accountId, input: { email: inviteEmail, role: inviteRole } },
      {
        onSuccess: () => {
          setInviteEmail('')
          setShowInvite(false)
        },
      }
    )
  }

  const handleRoleChange = (userId: string, role: 'admin' | 'member' | 'viewer') => {
    if (!currentAccount) return
    updateMember.mutate({
      accountId: currentAccount.accountId,
      userId,
      input: { role },
    })
  }

  const handleRemove = (userId: string) => {
    if (!currentAccount) return
    removeMember.mutate(
      { accountId: currentAccount.accountId, userId },
      { onSuccess: () => setConfirmRemove(null) }
    )
  }

  const userInitials = user?.name
    ?.split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2) || 'U'

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-lg">Team Members</CardTitle>
              <CardDescription>Manage who has access to this workspace</CardDescription>
            </div>
            <Button onClick={() => setShowInvite(!showInvite)}>
              <Plus className="h-4 w-4" />
              Invite Member
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {inviteMember.isError && (
            <div className="flex items-center gap-2 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
              <XCircle className="h-4 w-4 shrink-0" />
              {(inviteMember.error as Error)?.message || 'Failed to send invite'}
            </div>
          )}

          {showInvite && (
            <div className="flex items-center gap-2 rounded-md border p-4">
              <Input
                type="email"
                value={inviteEmail}
                onChange={(e) => setInviteEmail(e.target.value)}
                placeholder="colleague@company.com"
                className="flex-1"
              />
              <select
                value={inviteRole}
                onChange={(e) => setInviteRole(e.target.value as 'admin' | 'member' | 'viewer')}
                className="flex h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="admin">Admin</option>
                <option value="member">Member</option>
                <option value="viewer">Viewer</option>
              </select>
              <Button
                onClick={handleInvite}
                disabled={inviteMember.isPending || !inviteEmail.trim()}
              >
                {inviteMember.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                Send Invite
              </Button>
              <Button variant="outline" onClick={() => setShowInvite(false)}>
                Cancel
              </Button>
            </div>
          )}

          {/* Current user always shows */}
          <div className="flex items-center justify-between rounded-md border p-4">
            <div className="flex items-center gap-3">
              <Avatar>
                <AvatarFallback className="bg-primary text-sm font-medium text-primary-foreground">
                  {userInitials}
                </AvatarFallback>
              </Avatar>
              <div>
                <p className="text-sm font-medium">{user?.name || 'You'}</p>
                <p className="text-xs text-muted-foreground">{user?.email || ''}</p>
              </div>
            </div>
            <Badge>Owner</Badge>
          </div>

          {isLoading ? (
            <div className="space-y-3">
              {[1, 2].map((i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : isError ? (
            <div className="py-4 text-center text-sm text-muted-foreground">
              Failed to load team members. Please try again.
            </div>
          ) : members && members.length > 0 ? (
            members.map((member: { userId: string; name: string; email: string; role: string; status: string }) => (
              <div
                key={member.userId}
                className="flex items-center justify-between rounded-md border p-4"
              >
                <div className="flex items-center gap-3">
                  <Avatar>
                    <AvatarFallback>
                      {member.name
                        ?.split(' ')
                        .map((n: string) => n[0])
                        .join('')
                        .toUpperCase()
                        .slice(0, 2) || '??'}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <p className="text-sm font-medium">
                      {member.name}
                      {member.status === 'Pending' && (
                        <span className="ml-2 text-xs text-muted-foreground">(pending)</span>
                      )}
                    </p>
                    <p className="text-xs text-muted-foreground">{member.email}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {member.role !== 'Owner' ? (
                    <>
                      <select
                        value={member.role.toLowerCase()}
                        onChange={(e) =>
                          handleRoleChange(
                            member.userId,
                            e.target.value as 'admin' | 'member' | 'viewer'
                          )
                        }
                        disabled={updateMember.isPending}
                        className="flex h-8 rounded-md border border-input bg-background px-2 py-1 text-xs"
                      >
                        <option value="admin">Admin</option>
                        <option value="member">Member</option>
                        <option value="viewer">Viewer</option>
                      </select>
                      {confirmRemove === member.userId ? (
                        <div className="flex items-center gap-1">
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => handleRemove(member.userId)}
                            disabled={removeMember.isPending}
                          >
                            {removeMember.isPending ? (
                              <Loader2 className="h-3 w-3 animate-spin" />
                            ) : (
                              'Confirm'
                            )}
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setConfirmRemove(null)}
                          >
                            Cancel
                          </Button>
                        </div>
                      ) : (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 text-muted-foreground hover:text-destructive"
                          onClick={() => setConfirmRemove(member.userId)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                    </>
                  ) : (
                    <Badge>Owner</Badge>
                  )}
                </div>
              </div>
            ))
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Roles & Permissions</CardTitle>
          <CardDescription>Understanding access levels in your workspace</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {[
            {
              role: 'Owner',
              variant: 'default' as const,
              icon: Shield,
              desc: 'Full access to everything including billing, team management, and workspace deletion',
            },
            {
              role: 'Admin',
              variant: 'info' as const,
              icon: Key,
              desc: 'Can manage helpers, connections, team members, and API keys. Cannot access billing or delete workspace.',
            },
            {
              role: 'Member',
              variant: 'success' as const,
              icon: Users,
              desc: 'Can manage helpers and execute them. Cannot manage connections, team, or API keys.',
            },
            {
              role: 'Viewer',
              variant: 'secondary' as const,
              icon: Eye,
              desc: 'Read-only access to helpers, executions, and analytics. Cannot make changes.',
            },
          ].map((item) => (
            <div key={item.role} className="flex items-start gap-3">
              <Badge variant={item.variant} className="mt-0.5 w-16 justify-center">
                {item.role}
              </Badge>
              <p className="flex-1 text-sm text-muted-foreground">{item.desc}</p>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Security Tab (2FA/MFA)
// ---------------------------------------------------------------------------
function SecurityTab() {
  const [mfaEnabled, setMfaEnabled] = useState(false)
  const [mfaMethod, setMfaMethod] = useState<'totp' | 'sms'>('totp')

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Shield className="h-5 w-5" />
            Two-Factor Authentication (2FA)
          </CardTitle>
          <CardDescription>
            Add an extra layer of security to your account by requiring a second form of verification when signing in.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="flex items-center gap-3">
              <div className={cn(
                'flex h-10 w-10 items-center justify-center rounded-full',
                mfaEnabled ? 'bg-success/10' : 'bg-muted'
              )}>
                <Shield className={cn('h-5 w-5', mfaEnabled ? 'text-success' : 'text-muted-foreground')} />
              </div>
              <div>
                <p className="text-sm font-medium">
                  {mfaEnabled ? '2FA is enabled' : '2FA is not enabled'}
                </p>
                <p className="text-xs text-muted-foreground">
                  {mfaEnabled
                    ? 'Your account is protected with two-factor authentication'
                    : 'Enable 2FA to add an extra layer of security'}
                </p>
              </div>
            </div>
            <Badge variant={mfaEnabled ? 'default' : 'secondary'}>
              {mfaEnabled ? 'Active' : 'Inactive'}
            </Badge>
          </div>

          <Separator />

          <div className="space-y-4">
            <h4 className="text-sm font-medium">Authentication Methods</h4>

            <div
              className={cn(
                'flex items-start gap-4 rounded-lg border p-4 cursor-pointer transition-colors',
                mfaMethod === 'totp' && 'border-primary bg-primary/5'
              )}
              onClick={() => setMfaMethod('totp')}
            >
              <div className={cn(
                'mt-0.5 flex h-4 w-4 items-center justify-center rounded-full border-2',
                mfaMethod === 'totp' ? 'border-primary bg-primary' : 'border-muted-foreground'
              )}>
                {mfaMethod === 'totp' && <div className="h-1.5 w-1.5 rounded-full bg-white" />}
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <Smartphone className="h-4 w-4" />
                  <span className="text-sm font-medium">Authenticator App</span>
                  <Badge variant="outline" className="text-xs">Recommended</Badge>
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  Use an authenticator app like Google Authenticator, Authy, or 1Password to generate verification codes.
                </p>
              </div>
            </div>

            <div
              className={cn(
                'flex items-start gap-4 rounded-lg border p-4 cursor-pointer transition-colors',
                mfaMethod === 'sms' && 'border-primary bg-primary/5'
              )}
              onClick={() => setMfaMethod('sms')}
            >
              <div className={cn(
                'mt-0.5 flex h-4 w-4 items-center justify-center rounded-full border-2',
                mfaMethod === 'sms' ? 'border-primary bg-primary' : 'border-muted-foreground'
              )}>
                {mfaMethod === 'sms' && <div className="h-1.5 w-1.5 rounded-full bg-white" />}
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <Lock className="h-4 w-4" />
                  <span className="text-sm font-medium">SMS Verification</span>
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  Receive a verification code via text message to your registered phone number.
                </p>
              </div>
            </div>
          </div>

          <Button
            onClick={() => setMfaEnabled(!mfaEnabled)}
            variant={mfaEnabled ? 'outline' : 'default'}
            className="w-full"
          >
            {mfaEnabled ? 'Disable 2FA' : 'Enable 2FA'}
          </Button>
        </CardContent>
      </Card>

      <ChangePasswordCard />

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Active Sessions</CardTitle>
          <CardDescription>View and manage your active login sessions.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4 rounded-lg border p-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-success/10">
              <CheckCircle className="h-5 w-5 text-success" />
            </div>
            <div className="flex-1">
              <p className="text-sm font-medium">Current Session</p>
              <p className="text-xs text-muted-foreground">This device - Active now</p>
            </div>
            <Badge variant="outline">Current</Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Notifications Tab
// ---------------------------------------------------------------------------
function NotificationsTab() {
  const { data: prefs, isLoading } = useNotificationPreferences()
  const updatePrefs = useUpdateNotificationPreferences()
  const [webhookUrl, setWebhookUrl] = useState('')

  const handleToggle = (key: string, value: boolean) => {
    updatePrefs.mutate({ [key]: value })
  }

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
                checked={prefs?.[item.key as keyof typeof prefs] as boolean ?? false}
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
                checked={prefs?.[item.key as keyof typeof prefs] as boolean ?? false}
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
                value={webhookUrl || prefs?.webhookUrl || ''}
                onChange={(e) => setWebhookUrl(e.target.value)}
                placeholder="https://hooks.slack.com/services/..."
                className="flex-1 font-mono"
              />
              <Button
                onClick={() => updatePrefs.mutate({ webhookUrl })}
                disabled={updatePrefs.isPending}
              >
                {updatePrefs.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                Save
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              We&apos;ll POST JSON to this URL for critical events (failures, connection issues)
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
