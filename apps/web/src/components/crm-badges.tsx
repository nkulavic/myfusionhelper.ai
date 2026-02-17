'use client'

import { usePlatforms } from '@/lib/hooks/use-connections'
import type { PlatformDefinition } from '@/lib/api/connections'

interface CRMBadgesProps {
  crmIds: string[]
  max?: number
}

export function CRMBadges({ crmIds, max = 3 }: CRMBadgesProps) {
  const { data: platforms } = usePlatforms()

  const findPlatform = (id: string) =>
    platforms?.find((p: PlatformDefinition) => p.slug === id || p.platformId === id)

  if (crmIds.length === 0) {
    return (
      <div className="flex items-center gap-1.5">
        <div className="flex -space-x-1">
          {platforms?.map((p: PlatformDefinition) => (
            <div
              key={p.platformId}
              className="flex h-5 w-5 items-center justify-center rounded-full border-2 border-card text-[9px] font-bold text-white"
              style={{ backgroundColor: p.displayConfig?.color || 'hsl(var(--primary))' }}
              title={p.name}
            >
              {p.displayConfig?.initial || p.name.charAt(0)}
            </div>
          ))}
        </div>
        <span className="text-[10px] text-muted-foreground">All Platforms</span>
      </div>
    )
  }

  const visible = crmIds.slice(0, max)
  const remaining = crmIds.length - max

  return (
    <div className="flex items-center gap-1.5">
      <div className="flex -space-x-1">
        {visible.map((id) => {
          const platform = findPlatform(id)
          if (!platform) return null
          return (
            <div
              key={id}
              className="flex h-5 w-5 items-center justify-center rounded-full border-2 border-card text-[9px] font-bold text-white"
              style={{ backgroundColor: platform.displayConfig?.color || 'hsl(var(--primary))' }}
              title={platform.name}
            >
              {platform.displayConfig?.initial || platform.name.charAt(0)}
            </div>
          )
        })}
      </div>
      {remaining > 0 && (
        <span className="text-[10px] text-muted-foreground">+{remaining}</span>
      )}
      {visible.length === 1 && (
        <span className="text-[10px] text-muted-foreground">
          {findPlatform(visible[0])?.name} only
        </span>
      )}
    </div>
  )
}
