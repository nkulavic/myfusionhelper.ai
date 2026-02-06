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
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import { useUIStore } from '@/lib/stores/ui-store'
import { useLogout } from '@/lib/hooks/use-auth'
import { ThemeToggle } from '@/components/theme-toggle'
import { AIChatPanel } from '@/components/ai-chat-panel'
import { Button } from '@/components/ui/button'

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
  const [aiChatOpen, setAiChatOpen] = useState(false)
  const { user } = useAuthStore()
  const { currentAccount, onboardingComplete } = useWorkspaceStore()
  const { sidebarCollapsed, setSidebarCollapsed } = useUIStore()
  const logout = useLogout()

  // Redirect to onboarding if not completed
  useEffect(() => {
    if (!onboardingComplete) {
      router.replace('/onboarding')
    }
  }, [onboardingComplete, router])

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
      onSuccess: () => router.push('/login'),
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
                    'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
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
        <main className="flex-1 p-4 sm:p-6">{children}</main>
      </div>

      {/* AI Chat Floating Button */}
      <Button
        size="icon"
        onClick={() => setAiChatOpen(!aiChatOpen)}
        className="fixed bottom-6 right-6 z-30 h-14 w-14 rounded-full shadow-lg transition-transform hover:scale-105"
        aria-label="Open AI chat"
      >
        <MessageSquare className="h-6 w-6" />
      </Button>

      {/* AI Chat Panel */}
      {aiChatOpen && <AIChatPanel onClose={() => setAiChatOpen(false)} />}
    </div>
  )
}
