'use client'

import Image from 'next/image'
import { type CRMPlatform } from '@/lib/crm-platforms'
import { type PlatformDefinition } from '@/lib/api/connections'

interface PlatformLogoFromCRM {
  platform: CRMPlatform
  definition?: never
  size?: number
  className?: string
}

interface PlatformLogoFromAPI {
  platform?: never
  definition: PlatformDefinition
  size?: number
  className?: string
}

type PlatformLogoProps = PlatformLogoFromCRM | PlatformLogoFromAPI

export function PlatformLogo({ platform, definition, size = 40, className = '' }: PlatformLogoProps) {
  const logo = platform?.logo || (definition?.logoUrl ? `/images/platforms/${definition.slug}.png` : null)
  const color = platform?.color || definition?.displayConfig?.color || 'hsl(var(--primary))'
  const initial = platform?.initial || definition?.displayConfig?.initial || definition?.name?.charAt(0) || '?'
  const name = platform?.name || definition?.name || ''

  if (logo) {
    return (
      <Image
        src={logo}
        alt={name}
        width={size}
        height={size}
        className={`rounded-md object-contain ${className}`}
      />
    )
  }

  return (
    <div
      className={`flex items-center justify-center rounded-md text-white font-bold ${className}`}
      style={{
        backgroundColor: color,
        width: size,
        height: size,
        fontSize: size * 0.4,
      }}
    >
      {initial}
    </div>
  )
}
