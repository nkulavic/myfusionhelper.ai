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
} from 'lucide-react'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import { useLogout } from '@/lib/hooks/use-auth'
import { ThemeToggle } from '@/components/theme-toggle'
import { AIChatPanel } from '@/components/ai-chat-panel'

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
  const { currentAccount } = useWorkspaceStore()
  const logout = useLogout()

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
      {/* Sidebar */}
      <aside className="fixed inset-y-0 left-0 z-50 w-64 border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
        <div className="flex h-full flex-col">
          {/* Logo */}
          <div className="flex h-14 items-center justify-center border-b border-sidebar-border px-4">
            <Link href="/" className="flex items-center gap-2 font-bold text-sidebar-foreground">
              <Image src="/logo.png" alt="MyFusion Helper" width={160} height={20} className="brightness-0 invert" />
            </Link>
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
          <nav className="flex-1 space-y-1 p-4">
            {navigation.map((item) => {
              const isActive = pathname === item.href || (item.href !== '/' && pathname.startsWith(item.href + '/'))
              return (
                <Link
                  key={item.name}
                  href={item.href}
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
              <button className="relative rounded-md p-2 hover:bg-sidebar-accent" title="Notifications">
                <Bell className="h-4 w-4 text-sidebar-muted-foreground" />
                <span className="absolute right-1.5 top-1.5 flex h-1.5 w-1.5">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-sidebar-primary opacity-75" />
                  <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-sidebar-primary" />
                </span>
              </button>
              <button
                onClick={handleLogout}
                className="rounded-md p-2 hover:bg-sidebar-accent"
                title="Sign out"
              >
                <LogOut className="h-4 w-4 text-sidebar-muted-foreground" />
              </button>
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
      <div className="flex flex-1 flex-col pl-64">
        <main className="flex-1 p-6">{children}</main>
      </div>

      {/* AI Chat Floating Button */}
      <button
        onClick={() => setAiChatOpen(!aiChatOpen)}
        className="fixed bottom-6 right-6 z-50 flex h-14 w-14 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-lg transition-transform hover:scale-105"
      >
        <MessageSquare className="h-6 w-6" />
      </button>

      {/* AI Chat Panel */}
      {aiChatOpen && <AIChatPanel onClose={() => setAiChatOpen(false)} />}
    </div>
  )
}
