'use client'

import Image from 'next/image'
import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import {
  LayoutDashboard,
  Blocks,
  Link2,
  History,
  Settings,
  Sparkles,
  LogOut,
  ChevronDown,
  Bell,
  BarChart3,
  Mail,
  MessageSquare,
  Database,
  Menu,
  X,
  Search,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import { useUIStore } from '@/lib/stores/ui-store'
import { useLogout } from '@/lib/hooks/use-auth'
import { useGlobalKeyboardShortcuts } from '@/lib/hooks/use-keyboard-shortcuts'
import { ThemeToggle } from '@/components/theme-toggle'
import { AIChatPanel } from '@/components/ai-chat-panel'
import { CommandPalette } from '@/components/command-palette'
import { Button } from '@/components/ui/button'
import { useBillingInfo } from '@/lib/hooks/use-settings'

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Helpers', href: '/helpers', icon: Blocks },
  { name: 'Connections', href: '/connections', icon: Link2 },
  { name: 'Executions', href: '/executions', icon: History },
  { name: 'Insights', href: '/insights', icon: Sparkles },
  { name: 'Data Explorer', href: '/data-explorer', icon: Database },
  { name: 'Reports', href: '/reports', icon: BarChart3 },
  { name: 'Emails', href: '/emails', icon: Mail },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()
  const { user } = useAuthStore()
  const { currentAccount, onboardingComplete, _hasHydrated } = useWorkspaceStore()
  const {
    sidebarCollapsed,
    setSidebarCollapsed,
    aiChatOpen,
    setAIChatOpen,
    setCommandPaletteOpen,
  } = useUIStore()
  const logout = useLogout()

  // Register global keyboard shortcuts
  useGlobalKeyboardShortcuts()

  // Redirect to onboarding if not completed (wait for hydration first)
  useEffect(() => {
    if (_hasHydrated && !onboardingComplete) {
      router.replace('/onboarding')
    }
  }, [_hasHydrated, onboardingComplete, router])

  const userInitials = user?.name
    ? user.name
        .split(' ')
        .map((n) => n[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)
    : 'U'

  const handleLogout = () => {
    logout.mutate(undefined, {
      onSuccess: () => router.push('/'),
    })
  }

  return (
    <div className="flex min-h-screen">
      {/* Mobile Header */}
      <div className="fixed inset-x-0 top-0 z-40 flex h-14 items-center gap-3 border-b bg-background px-4 lg:hidden">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
          aria-label="Toggle sidebar"
        >
          <Menu className="h-5 w-5" />
        </Button>
        <Link href="/" className="flex items-center">
          <Image src="/logo.png" alt="MyFusion Helper" width={140} height={18} className="dark:brightness-0 dark:invert" />
        </Link>
      </div>

      {/* Mobile Overlay */}
      {!sidebarCollapsed && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={() => setSidebarCollapsed(true)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-50 w-64 border-r border-sidebar-border bg-sidebar text-sidebar-foreground transition-transform duration-200 ease-in-out',
          'lg:translate-x-0',
          sidebarCollapsed ? '-translate-x-full' : 'translate-x-0'
        )}
      >
        <div className="flex h-full flex-col">
          {/* Logo */}
          <div className="flex h-14 items-center justify-between border-b border-sidebar-border px-4">
            <Link href="/" className="flex items-center gap-2 font-bold text-sidebar-foreground">
              <Image src="/logo.png" alt="MyFusion Helper" width={160} height={20} className="brightness-0 invert" />
            </Link>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 text-sidebar-muted-foreground hover:bg-sidebar-accent lg:hidden"
              onClick={() => setSidebarCollapsed(true)}
              aria-label="Close sidebar"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>

          {/* Account Switcher */}
          <div className="border-b border-sidebar-border p-4">
            <button className="flex w-full items-center justify-between rounded-md border border-sidebar-border bg-sidebar-accent px-3 py-2 text-sm hover:bg-sidebar-accent/80">
              <div className="flex flex-col items-start">
                <span className="font-medium text-sidebar-foreground">
                  {currentAccount?.name || 'My Business'}
                </span>
                <span className="text-xs text-sidebar-muted-foreground capitalize">
                  {currentAccount?.plan || 'Free'} Plan
                </span>
              </div>
              <ChevronDown className="h-4 w-4 text-sidebar-muted-foreground" />
            </button>
          </div>

          {/* Search / Command Palette Trigger */}
          <div className="px-4 pt-4">
            <button
              onClick={() => setCommandPaletteOpen(true)}
              className="flex w-full items-center gap-2 rounded-md border border-sidebar-border bg-sidebar-accent/50 px-3 py-2 text-sm text-sidebar-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
            >
              <Search className="h-3.5 w-3.5" />
              <span className="flex-1 text-left">Search...</span>
              <kbd className="hidden rounded border border-sidebar-border bg-sidebar px-1.5 py-0.5 font-mono text-[10px] text-sidebar-muted-foreground sm:inline-block">
                &#8984;K
              </kbd>
            </button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 space-y-1 overflow-y-auto p-4">
            {navigation.map((item) => {
              const isActive = pathname === item.href || (item.href !== '/' && pathname.startsWith(item.href + '/'))
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  onClick={() => setSidebarCollapsed(true)}
                  className={cn(
                    'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-all active:scale-[0.97]',
                    isActive
                      ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                      : 'text-sidebar-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground'
                  )}
                >
                  <item.icon className="h-4 w-4" />
                  {item.name}
                </Link>
              )
            })}
          </nav>

          {/* User Menu */}
          <div className="border-t border-sidebar-border p-4 space-y-3">
            <div className="flex items-center justify-center gap-1">
              <ThemeToggle variant="sidebar" />
              <Button
                variant="ghost"
                size="icon"
                className="relative h-9 w-9 hover:bg-sidebar-accent"
                title="Notifications"
              >
                <Bell className="h-4 w-4 text-sidebar-muted-foreground" />
                <span className="absolute right-1.5 top-1.5 flex h-1.5 w-1.5">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-sidebar-primary opacity-75" />
                  <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-sidebar-primary" />
                </span>
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="h-9 w-9 hover:bg-sidebar-accent"
                onClick={handleLogout}
                title="Sign out"
              >
                <LogOut className="h-4 w-4 text-sidebar-muted-foreground" />
              </Button>
            </div>
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-full bg-sidebar-primary text-sm font-medium text-sidebar-primary-foreground">
                {userInitials}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate text-sidebar-foreground">
                  {user?.name || 'User'}
                </p>
                <p className="text-xs text-sidebar-muted-foreground truncate">
                  {user?.email || ''}
                </p>
              </div>
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <div className="flex flex-1 flex-col pt-14 lg:pt-0 lg:pl-64">
        <UpgradeBanner />
        <main className="flex-1 p-4 sm:p-6">
          {children}
        </main>
      </div>

      {/* AI Chat Floating Button */}
      <Button
        size="icon"
        onClick={() => setAIChatOpen(!aiChatOpen)}
        className="fixed bottom-6 right-6 z-30 h-14 w-14 rounded-full shadow-lg transition-transform hover:scale-105"
        aria-label="Open AI chat"
        title="AI Chat (&#8984;/)"
      >
        <MessageSquare className="h-6 w-6" />
      </Button>

      {/* AI Chat Panel */}
      {aiChatOpen && <AIChatPanel onClose={() => setAIChatOpen(false)} />}

      {/* Command Palette */}
      <CommandPalette />
    </div>
  )
}

function UpgradeBanner() {
  const { data: billing } = useBillingInfo()
  const [dismissed, setDismissed] = useState(false)

  // Check sessionStorage on mount
  useEffect(() => {
    if (typeof window !== 'undefined' && sessionStorage.getItem('mfh_banner_dismissed') === '1') {
      setDismissed(true)
    }
  }, [])

  if (dismissed || !billing) return null

  const handleDismiss = () => {
    setDismissed(true)
    sessionStorage.setItem('mfh_banner_dismissed', '1')
  }

  // Sandbox upgrade banner
  if (billing.plan === 'free') {
    return (
      <div className="flex items-center justify-between gap-4 border-b bg-primary/5 px-4 py-2 sm:px-6">
        <p className="text-sm text-foreground">
          You&apos;re on the free sandbox (1 helper, 100 executions).{' '}
          <Link href="/settings?tab=billing" className="font-medium text-primary hover:underline">
            Choose a plan
          </Link>{' '}
          to unlock full features.
        </p>
        <button
          onClick={handleDismiss}
          className="shrink-0 text-muted-foreground hover:text-foreground"
          aria-label="Dismiss"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    )
  }

  // Trial countdown (< 7 days remaining)
  if (billing.trialEndsAt) {
    const now = Date.now() / 1000
    const daysLeft = Math.ceil((billing.trialEndsAt - now) / 86400)
    if (daysLeft > 0 && daysLeft <= 7) {
      return (
        <div className="flex items-center justify-between gap-4 border-b bg-amber-50 px-4 py-2 dark:bg-amber-950/30 sm:px-6">
          <p className="text-sm text-foreground">
            Your trial ends in {daysLeft} day{daysLeft !== 1 ? 's' : ''}.{' '}
            <Link href="/settings?tab=billing" className="font-medium text-primary hover:underline">
              Manage subscription
            </Link>
          </p>
          <button
            onClick={handleDismiss}
            className="shrink-0 text-muted-foreground hover:text-foreground"
            aria-label="Dismiss"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      )
    }
  }

  return null
}
