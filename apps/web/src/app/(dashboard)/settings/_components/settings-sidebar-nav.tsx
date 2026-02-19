'use client'

import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import { cn } from '@/lib/utils'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { settingsTabs } from '../_lib/settings-tabs'

interface SettingsSidebarNavProps {
  sidebarMinimized: boolean
  onNavClick: () => void
}

export function SettingsSidebarNav({ sidebarMinimized, onNavClick }: SettingsSidebarNavProps) {
  const searchParams = useSearchParams()
  const activeTab = searchParams.get('tab') || 'profile'

  return (
    <>
      {settingsTabs.map((tab) => {
        const isActive = activeTab === tab.id
        const linkContent = (
          <Link
            key={tab.id}
            href={`/settings?tab=${tab.id}`}
            onClick={onNavClick}
            scroll={false}
            className={cn(
              'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-all active:scale-[0.97]',
              sidebarMinimized && 'lg:justify-center lg:gap-0 lg:px-0',
              isActive
                ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                : 'text-sidebar-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground',
            )}
          >
            <tab.icon className="h-4 w-4 shrink-0" />
            <span className={cn(sidebarMinimized && 'lg:hidden')}>{tab.name}</span>
          </Link>
        )

        if (sidebarMinimized) {
          return (
            <Tooltip key={tab.id}>
              <TooltipTrigger asChild>{linkContent}</TooltipTrigger>
              <TooltipContent side="right" className="hidden lg:block">
                {tab.name}
              </TooltipContent>
            </Tooltip>
          )
        }

        return linkContent
      })}
    </>
  )
}
