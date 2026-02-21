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
  LayoutPanelTop,
  Menu,
  X,
  Search,
  PanelLeftClose,
  PanelLeftOpen,
  CreditCard,
  Clock,
  ArrowLeft,
} from 'lucide-react'
import { useEffect } from 'react'
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
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { useBillingInfo } from '@/lib/hooks/use-settings'
import { usePlanLimits } from '@/lib/hooks/use-plan-limits'
import { SettingsSidebarNav } from './settings/_components/settings-sidebar-nav'

const baseNavigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Helpers', href: '/helpers', icon: Blocks },
  { name: 'Connections', href: '/connections', icon: Link2 },
  { name: 'Executions', href: '/executions', icon: History },
  { name: 'Insights', href: '/insights', icon: Sparkles },
  { name: 'Data Explorer', href: '/data-explorer', icon: Database },
  { name: 'Reports', href: '/reports', icon: BarChart3 },
  { name: 'Studio', href: '/studio', icon: LayoutPanelTop },
  { name: 'Emails', href: '/emails', icon: Mail },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()
  const isSettingsPage = pathname.startsWith('/settings')
  const { user } = useAuthStore()
  const { currentAccount, onboardingComplete, _hasHydrated } = useWorkspaceStore()
  const {
    sidebarCollapsed,
    setSidebarCollapsed,
    sidebarMinimized,
    toggleSidebarMinimized,
    aiChatOpen,
    setAIChatOpen,
    setCommandPaletteOpen,
  } = useUIStore()
  const logout = useLogout()
  const { isTrialing, isTrialExpired } = usePlanLimits()

  // Build navigation with conditional Plans item
  const navigation = [
    ...baseNavigation.slice(0, 1), // Dashboard
    ...(isTrialing || isTrialExpired ? [{ name: 'Plans', href: '/plans', icon: CreditCard }] : []),
    ...baseNavigation.slice(1), // Rest of nav
  ]

  // Register global keyboard shortcuts
  useGlobalKeyboardShortcuts()

  // Redirect unverified users to verify-email page
  useEffect(() => {
    if (_hasHydrated && user && user.emailVerified === false) {
      router.replace(`/verify-email?email=${encodeURIComponent(user.email)}`)
    }
  }, [_hasHydrated, user, router])

  // Redirect to onboarding if not completed
  useEffect(() => {
    if (_hasHydrated && !onboardingComplete) {
      router.replace('/onboarding/plan')
    }
  }, [_hasHydrated, onboardingComplete, router])

  // Soft lock: redirect expired trial users to /plans (allow /plans, /settings, /dashboard)
  const allowedWhileExpired = ['/', '/plans', '/settings', '/dashboard']
  useEffect(() => {
    if (!isTrialExpired) return
    const isAllowed = allowedWhileExpired.some(
      (route) => pathname === route || pathname.startsWith(route + '/'),
    )
    if (!isAllowed) {
      router.replace('/plans')
    }
  }, [isTrialExpired, pathname, router])

  const userInitials = user?.name
    ? user.name
        .split(' ')
        .map((n) => n[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)
    : 'U'

  const handleLogout = () => {
    logout.mutate()
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
          'fixed inset-y-0 left-0 z-50 border-r border-sidebar-border bg-sidebar text-sidebar-foreground transition-all duration-200 ease-in-out',
          'lg:translate-x-0',
          sidebarCollapsed ? '-translate-x-full' : 'translate-x-0',
          sidebarMinimized ? 'lg:w-16' : 'lg:w-64',
          'w-64'
        )}
      >
        <div className="flex h-full flex-col">
          {/* Logo / Settings header */}
          <div className="flex h-14 items-center justify-between border-b border-sidebar-border px-4">
            {isSettingsPage ? (
              <>
                {/* Settings: Desktop minimized — back arrow icon */}
                {sidebarMinimized && (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Link
                        href="/"
                        className="mx-auto hidden h-8 w-8 items-center justify-center rounded-md text-sidebar-muted-foreground hover:bg-sidebar-accent lg:flex"
                      >
                        <ArrowLeft className="h-4 w-4" />
                      </Link>
                    </TooltipTrigger>
                    <TooltipContent side="right">Back to app</TooltipContent>
                  </Tooltip>
                )}
                {/* Settings: Desktop expanded — back link + "Settings" label */}
                {!sidebarMinimized && (
                  <>
                    <Link
                      href="/"
                      className="hidden items-center gap-2 text-sm font-medium text-sidebar-muted-foreground hover:text-sidebar-foreground lg:flex"
                    >
                      <ArrowLeft className="h-4 w-4" />
                      Back
                    </Link>
                    <span className="hidden text-sm font-semibold text-sidebar-foreground lg:block">
                      Settings
                    </span>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="hidden h-8 w-8 text-sidebar-muted-foreground hover:bg-sidebar-accent lg:flex"
                      onClick={toggleSidebarMinimized}
                      aria-label="Collapse sidebar"
                    >
                      <PanelLeftClose className="h-4 w-4" />
                    </Button>
                  </>
                )}
                {/* Settings: Mobile — "Settings" text + close button */}
                <span className="text-sm font-semibold text-sidebar-foreground lg:hidden">
                  Settings
                </span>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 text-sidebar-muted-foreground hover:bg-sidebar-accent lg:hidden"
                  onClick={() => setSidebarCollapsed(true)}
                  aria-label="Close sidebar"
                >
                  <X className="h-4 w-4" />
                </Button>
              </>
            ) : (
              <>
                {/* Default: Desktop minimized — expand button */}
                {sidebarMinimized && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="mx-auto hidden h-8 w-8 text-sidebar-muted-foreground hover:bg-sidebar-accent lg:flex"
                    onClick={toggleSidebarMinimized}
                    aria-label="Expand sidebar"
                  >
                    <PanelLeftOpen className="h-4 w-4" />
                  </Button>
                )}
                {/* Default: Desktop expanded — logo + collapse button */}
                {!sidebarMinimized && (
                  <>
                    <Link href="/" className="hidden items-center gap-2 font-bold text-sidebar-foreground lg:flex">
                      <Image src="/logo.png" alt="MyFusion Helper" width={160} height={20} className="brightness-0 invert" />
                    </Link>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="hidden h-8 w-8 text-sidebar-muted-foreground hover:bg-sidebar-accent lg:flex"
                      onClick={toggleSidebarMinimized}
                      aria-label="Collapse sidebar"
                    >
                      <PanelLeftClose className="h-4 w-4" />
                    </Button>
                  </>
                )}
                {/* Default: Mobile — logo + close button */}
                <Link href="/" className="flex items-center gap-2 font-bold text-sidebar-foreground lg:hidden">
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
              </>
            )}
          </div>

          {/* Account Switcher - minimized (desktop only) */}
          {sidebarMinimized && (
            <div className="hidden border-b border-sidebar-border p-2 lg:block">
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className="mx-auto flex h-8 w-8 items-center justify-center rounded-md bg-sidebar-accent text-xs font-bold text-sidebar-foreground">
                    {(currentAccount?.name || 'M')[0].toUpperCase()}
                  </div>
                </TooltipTrigger>
                <TooltipContent side="right">
                  <p className="font-medium">{currentAccount?.name || 'My Business'}</p>
                  <p className="text-xs text-muted-foreground capitalize">{currentAccount?.plan || 'Free'} Plan</p>
                </TooltipContent>
              </Tooltip>
            </div>
          )}
          {/* Account Switcher - expanded (always on mobile, desktop when not minimized) */}
          <div className={cn('border-b border-sidebar-border p-4', sidebarMinimized && 'lg:hidden')}>
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

          {/* Search - minimized (desktop only) */}
          {sidebarMinimized && (
            <div className="hidden px-2 pt-3 lg:block">
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    onClick={() => setCommandPaletteOpen(true)}
                    className="mx-auto flex h-9 w-9 items-center justify-center rounded-md border border-sidebar-border bg-sidebar-accent/50 text-sidebar-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
                  >
                    <Search className="h-4 w-4" />
                  </button>
                </TooltipTrigger>
                <TooltipContent side="right">Search (&#8984;K)</TooltipContent>
              </Tooltip>
            </div>
          )}
          {/* Search - expanded (always on mobile, desktop when not minimized) */}
          <div className={cn('px-4 pt-4', sidebarMinimized && 'lg:hidden')}>
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
          <nav className={cn(
            'flex-1 space-y-1 overflow-y-auto p-4',
            sidebarMinimized && 'lg:p-2'
          )}>
            {isSettingsPage ? (
              <SettingsSidebarNav
                sidebarMinimized={sidebarMinimized}
                onNavClick={() => setSidebarCollapsed(true)}
              />
            ) : (
              navigation.map((item) => {
                const isActive = pathname === item.href || (item.href !== '/' && pathname.startsWith(item.href + '/'))
                const linkContent = (
                  <Link
                    key={item.name}
                    href={item.href}
                    onClick={() => setSidebarCollapsed(true)}
                    className={cn(
                      'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-all active:scale-[0.97]',
                      sidebarMinimized && 'lg:justify-center lg:gap-0 lg:px-0',
                      isActive
                        ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                        : 'text-sidebar-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground'
                    )}
                  >
                    <item.icon className="h-4 w-4 shrink-0" />
                    <span className={cn(sidebarMinimized && 'lg:hidden')}>{item.name}</span>
                  </Link>
                )

                if (sidebarMinimized) {
                  return (
                    <Tooltip key={item.name}>
                      <TooltipTrigger asChild>{linkContent}</TooltipTrigger>
                      <TooltipContent side="right" className="hidden lg:block">{item.name}</TooltipContent>
                    </Tooltip>
                  )
                }

                return linkContent
              })
            )}
          </nav>

          {/* User Menu */}
          <div className={cn(
            'border-t border-sidebar-border space-y-3 p-4',
            sidebarMinimized && 'lg:p-2'
          )}>
            {/* Minimized user menu (desktop only) */}
            {sidebarMinimized && (
              <div className="hidden lg:block space-y-2">
                <div className="flex flex-col items-center gap-1">
                  <ThemeToggle variant="sidebar" />
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="relative h-9 w-9 hover:bg-sidebar-accent"
                      >
                        <Bell className="h-4 w-4 text-sidebar-muted-foreground" />
                        <span className="absolute right-1.5 top-1.5 flex h-1.5 w-1.5">
                          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-sidebar-primary opacity-75" />
                          <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-sidebar-primary" />
                        </span>
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent side="right">Notifications</TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-9 w-9 hover:bg-sidebar-accent"
                        onClick={handleLogout}
                      >
                        <LogOut className="h-4 w-4 text-sidebar-muted-foreground" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent side="right">Sign out</TooltipContent>
                  </Tooltip>
                </div>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div className="mx-auto flex h-9 w-9 items-center justify-center rounded-full bg-sidebar-primary text-sm font-medium text-sidebar-primary-foreground">
                      {userInitials}
                    </div>
                  </TooltipTrigger>
                  <TooltipContent side="right">
                    <p className="font-medium">{user?.name || 'User'}</p>
                    <p className="text-xs text-muted-foreground">{user?.email || ''}</p>
                  </TooltipContent>
                </Tooltip>
              </div>
            )}
            {/* Full user menu (always on mobile, desktop when not minimized) */}
            <div className={cn(sidebarMinimized && 'lg:hidden', 'space-y-3')}>
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
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-sidebar-primary text-sm font-medium text-sidebar-primary-foreground">
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
        </div>
      </aside>

      {/* Main Content */}
      <div className={cn(
        'flex flex-1 flex-col pt-14 transition-all duration-200 lg:pt-0',
        sidebarMinimized ? 'lg:pl-16' : 'lg:pl-64'
      )}>
        <TrialBanner />
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

function TrialBanner() {
  const { data: billing } = useBillingInfo()

  if (!billing) return null

  const { isTrialing, daysRemaining, totalTrialDays, trialExpired } = billing

  // Hide for paid subscribers
  if (!isTrialing && !trialExpired) return null

  // Expired state
  if (trialExpired) {
    return (
      <div className="border-b bg-red-100 px-4 py-3 dark:bg-red-950/50 sm:px-6">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <Clock className="h-4 w-4 text-red-600 dark:text-red-400" />
            <p className="text-sm font-medium text-red-800 dark:text-red-200">
              Your free trial has expired. Choose a plan to continue using MyFusion Helper.
            </p>
          </div>
          <Link href="/plans">
            <Button size="sm" variant="default">
              Choose Plan
            </Button>
          </Link>
        </div>
      </div>
    )
  }

  // Active trial
  const dayNumber = totalTrialDays - daysRemaining + 1
  const progress = Math.min(100, Math.round(((dayNumber - 1) / totalTrialDays) * 100))

  // Color by urgency
  let bgClass = 'bg-blue-50 dark:bg-blue-950/30'
  let textClass = 'text-blue-800 dark:text-blue-200'
  let iconClass = 'text-blue-600 dark:text-blue-400'
  let progressClass = 'bg-blue-500'

  if (daysRemaining <= 2) {
    bgClass = 'bg-red-50 dark:bg-red-950/30'
    textClass = 'text-red-800 dark:text-red-200'
    iconClass = 'text-red-600 dark:text-red-400'
    progressClass = 'bg-red-500'
  } else if (daysRemaining <= 7) {
    bgClass = 'bg-amber-50 dark:bg-amber-950/30'
    textClass = 'text-amber-800 dark:text-amber-200'
    iconClass = 'text-amber-600 dark:text-amber-400'
    progressClass = 'bg-amber-500'
  }

  return (
    <div className={cn('border-b px-4 py-3 sm:px-6', bgClass)}>
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3 min-w-0">
          <Clock className={cn('h-4 w-4 shrink-0', iconClass)} />
          <div className="flex items-center gap-3 min-w-0">
            <p className={cn('text-sm font-medium whitespace-nowrap', textClass)}>
              {daysRemaining} day{daysRemaining !== 1 ? 's' : ''} left in your free trial
            </p>
            <div className="hidden sm:flex items-center gap-2 min-w-0">
              <div className="h-1.5 w-24 rounded-full bg-black/10 dark:bg-white/10">
                <div
                  className={cn('h-full rounded-full transition-all', progressClass)}
                  style={{ width: `${progress}%` }}
                />
              </div>
              <span className={cn('text-xs whitespace-nowrap', textClass)}>
                Day {dayNumber} / {totalTrialDays}
              </span>
            </div>
          </div>
        </div>
        <Link href="/plans">
          <Button size="sm" variant="default">
            Choose Plan
          </Button>
        </Link>
      </div>
    </div>
  )
}
