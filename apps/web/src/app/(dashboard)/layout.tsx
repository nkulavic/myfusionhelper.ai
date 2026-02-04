'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  Zap,
  LayoutDashboard,
  Blocks,
  Link2,
  History,
  Settings,
  Sparkles,
  LogOut,
  ChevronDown,
  Bell,
} from 'lucide-react'
import { cn } from '@/lib/utils'

const navigation = [
  { name: 'Dashboard', href: '/helpers', icon: LayoutDashboard },
  { name: 'Helpers', href: '/helpers', icon: Blocks },
  { name: 'Connections', href: '/connections', icon: Link2 },
  { name: 'Executions', href: '/executions', icon: History },
  { name: 'Insights', href: '/insights', icon: Sparkles },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="fixed inset-y-0 left-0 z-50 w-64 border-r bg-card">
        <div className="flex h-full flex-col">
          {/* Logo */}
          <div className="flex h-14 items-center border-b px-4">
            <Link href="/helpers" className="flex items-center gap-2 font-bold">
              <Zap className="h-6 w-6" />
              <span>MyFusion Helper</span>
            </Link>
          </div>

          {/* Account Switcher */}
          <div className="border-b p-4">
            <button className="flex w-full items-center justify-between rounded-md border bg-background px-3 py-2 text-sm hover:bg-accent">
              <div className="flex flex-col items-start">
                <span className="font-medium">My Business</span>
                <span className="text-xs text-muted-foreground">Pro Plan</span>
              </div>
              <ChevronDown className="h-4 w-4 text-muted-foreground" />
            </button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 space-y-1 p-4">
            {navigation.map((item) => {
              const isActive = pathname === item.href || pathname.startsWith(item.href + '/')
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  className={cn(
                    'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                    isActive
                      ? 'bg-primary text-primary-foreground'
                      : 'text-muted-foreground hover:bg-accent hover:text-foreground'
                  )}
                >
                  <item.icon className="h-4 w-4" />
                  {item.name}
                </Link>
              )
            })}
          </nav>

          {/* User Menu */}
          <div className="border-t p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-full bg-primary text-sm font-medium text-primary-foreground">
                JD
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium">John Doe</p>
                <p className="text-xs text-muted-foreground">john@example.com</p>
              </div>
              <button className="rounded-md p-2 hover:bg-accent">
                <LogOut className="h-4 w-4 text-muted-foreground" />
              </button>
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <div className="flex flex-1 flex-col pl-64">
        {/* Top Bar */}
        <header className="sticky top-0 z-40 flex h-14 items-center justify-between border-b bg-background px-6">
          <div />
          <div className="flex items-center gap-4">
            <button className="relative rounded-md p-2 hover:bg-accent">
              <Bell className="h-5 w-5 text-muted-foreground" />
              <span className="absolute right-1 top-1 flex h-2 w-2">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary opacity-75" />
                <span className="relative inline-flex h-2 w-2 rounded-full bg-primary" />
              </span>
            </button>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  )
}
