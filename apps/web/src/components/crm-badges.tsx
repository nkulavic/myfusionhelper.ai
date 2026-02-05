import { crmPlatforms, getCRMPlatform } from '@/lib/crm-platforms'

interface CRMBadgesProps {
  crmIds: string[]
  max?: number
}

export function CRMBadges({ crmIds, max = 3 }: CRMBadgesProps) {
  if (crmIds.length === 0) {
    return (
      <div className="flex items-center gap-1.5">
        <div className="flex -space-x-1">
          {crmPlatforms.map((p) => (
            <div
              key={p.id}
              className="flex h-5 w-5 items-center justify-center rounded-full border-2 border-card text-[9px] font-bold text-white"
              style={{ backgroundColor: p.color }}
              title={p.name}
            >
              {p.initial}
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
          const platform = getCRMPlatform(id)
          if (!platform) return null
          return (
            <div
              key={id}
              className="flex h-5 w-5 items-center justify-center rounded-full border-2 border-card text-[9px] font-bold text-white"
              style={{ backgroundColor: platform.color }}
              title={platform.name}
            >
              {platform.initial}
            </div>
          )
        })}
      </div>
      {remaining > 0 && (
        <span className="text-[10px] text-muted-foreground">+{remaining}</span>
      )}
      {visible.length === 1 && (
        <span className="text-[10px] text-muted-foreground">
          {getCRMPlatform(visible[0])?.name} only
        </span>
      )}
    </div>
  )
}
