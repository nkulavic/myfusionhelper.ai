'use client'

import Image from 'next/image'
import { type CRMPlatform } from '@/lib/crm-platforms'

interface PlatformLogoProps {
  platform: CRMPlatform
  size?: number
  className?: string
}

export function PlatformLogo({ platform, size = 40, className = '' }: PlatformLogoProps) {
  if (platform.logo) {
    return (
      <Image
        src={platform.logo}
        alt={platform.name}
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
        backgroundColor: platform.color,
        width: size,
        height: size,
        fontSize: size * 0.4,
      }}
    >
      {platform.initial}
    </div>
  )
}
