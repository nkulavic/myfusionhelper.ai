'use client'

import { useCallback, useMemo } from 'react'
import { useRouter } from 'next/navigation'
import {
  LayoutDashboard,
  Blocks,
  Link2,
  History,
  Settings,
  Sparkles,
  Database,
  BarChart3,
  Mail,
  Plus,
  Zap,
  MessageSquare,
} from 'lucide-react'
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from '@/components/ui/command'
import { useUIStore } from '@/lib/stores/ui-store'
import { useHelpers } from '@/lib/hooks/use-helpers'
import { useConnections } from '@/lib/hooks/use-connections'
import type { PlatformConnection } from '@myfusionhelper/types'

const navigationItems = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard, keywords: 'home overview' },
  { name: 'Helpers', href: '/helpers', icon: Blocks, keywords: 'automations' },
  { name: 'Connections', href: '/connections', icon: Link2, keywords: 'crm platforms' },
  { name: 'Executions', href: '/executions', icon: History, keywords: 'runs logs' },
  { name: 'Insights', href: '/insights', icon: Sparkles, keywords: 'analytics ai' },
  { name: 'Data Explorer', href: '/data-explorer', icon: Database, keywords: 'query browse contacts' },
  { name: 'Reports', href: '/reports', icon: BarChart3, keywords: 'charts stats' },
  { name: 'Emails', href: '/emails', icon: Mail, keywords: 'templates compose' },
  { name: 'Settings', href: '/settings', icon: Settings, keywords: 'account profile billing' },
]

const actionItems = [
  { name: 'Create New Helper', href: '/helpers?view=new', icon: Plus, keywords: 'add automation' },
  { name: 'Add Connection', href: '/connections', icon: Link2, keywords: 'connect crm' },
  { name: 'Compose Email', href: '/emails', icon: Mail, keywords: 'send write' },
  { name: 'View Execution Trends', href: '/reports/execution-trends', icon: BarChart3, keywords: 'chart analytics' },
  { name: 'Open AI Chat', action: 'toggle-ai-chat', icon: MessageSquare, keywords: 'assistant ask' },
]

export function CommandPalette() {
  const router = useRouter()
  const { commandPaletteOpen, setCommandPaletteOpen, setAIChatOpen, aiChatOpen } = useUIStore()

  const { data: helpers } = useHelpers()
  const { data: connections } = useConnections()

  const recentHelpers = useMemo(() => {
    if (!helpers || !Array.isArray(helpers)) return []
    return helpers
      .slice()
      .sort((a, b) => {
        const aTime = a.lastExecutedAt || a.updatedAt || a.createdAt
        const bTime = b.lastExecutedAt || b.updatedAt || b.createdAt
        return new Date(bTime).getTime() - new Date(aTime).getTime()
      })
      .slice(0, 5)
  }, [helpers])

  const recentConnections = useMemo(() => {
    if (!connections) return []
    return connections.slice(0, 5)
  }, [connections])

  const runCommand = useCallback(
    (command: () => void) => {
      setCommandPaletteOpen(false)
      command()
    },
    [setCommandPaletteOpen]
  )

  return (
    <CommandDialog open={commandPaletteOpen} onOpenChange={setCommandPaletteOpen}>
      <CommandInput placeholder="Type a command or search..." />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>

        {/* Navigation */}
        <CommandGroup heading="Navigation">
          {navigationItems.map((item) => (
            <CommandItem
              key={item.href}
              value={`${item.name} ${item.keywords}`}
              onSelect={() => runCommand(() => router.push(item.href))}
            >
              <item.icon className="mr-2 h-4 w-4" />
              <span>Go to {item.name}</span>
            </CommandItem>
          ))}
        </CommandGroup>

        <CommandSeparator />

        {/* Actions */}
        <CommandGroup heading="Actions">
          {actionItems.map((item) => (
            <CommandItem
              key={item.name}
              value={`${item.name} ${item.keywords}`}
              onSelect={() =>
                runCommand(() => {
                  if (item.action === 'toggle-ai-chat') {
                    setAIChatOpen(!aiChatOpen)
                  } else if (item.href) {
                    router.push(item.href)
                  }
                })
              }
            >
              <item.icon className="mr-2 h-4 w-4" />
              <span>{item.name}</span>
              {item.action === 'toggle-ai-chat' && <CommandShortcut>&#8984;/</CommandShortcut>}
            </CommandItem>
          ))}
        </CommandGroup>

        {/* Recent Helpers */}
        {recentHelpers.length > 0 && (
          <>
            <CommandSeparator />
            <CommandGroup heading="Recent Helpers">
              {recentHelpers.map((helper) => (
                <CommandItem
                  key={helper.helperId}
                  value={`helper ${helper.name} ${helper.helperType} ${helper.category}`}
                  onSelect={() => runCommand(() => router.push(`/helpers?selected=${helper.helperId}`))}
                >
                  <Zap className="mr-2 h-4 w-4" />
                  <span>{helper.name}</span>
                  <span className="ml-auto text-xs text-muted-foreground">{helper.helperType}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          </>
        )}

        {/* Recent Connections */}
        {recentConnections.length > 0 && (
          <>
            <CommandSeparator />
            <CommandGroup heading="Connections">
              {recentConnections.map((conn: PlatformConnection) => (
                <CommandItem
                  key={conn.connectionId}
                  value={`connection ${conn.name} ${conn.platformId}`}
                  onSelect={() => runCommand(() => router.push('/connections'))}
                >
                  <Link2 className="mr-2 h-4 w-4" />
                  <span>{conn.name}</span>
                  <span className={`ml-auto text-xs ${
                    conn.status === 'active' ? 'text-success' :
                    conn.status === 'error' ? 'text-destructive' : 'text-warning'
                  }`}>
                    {conn.status}
                  </span>
                </CommandItem>
              ))}
            </CommandGroup>
          </>
        )}
      </CommandList>
    </CommandDialog>
  )
}
