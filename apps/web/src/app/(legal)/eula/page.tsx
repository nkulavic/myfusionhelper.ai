import type { Metadata } from 'next'
import { EULAContent } from '@/components/legal/eula-content'

export const metadata: Metadata = {
  title: 'End User License Agreement | MyFusion Helper',
  description: 'MyFusion Helper End User License Agreement (EULA)',
}

export default function EULAPage() {
  return (
    <article className="rounded-lg border bg-card p-8 shadow-sm">
      <h1 className="mb-6 text-3xl font-bold text-foreground">
        End User License Agreement
      </h1>
      <EULAContent />
    </article>
  )
}
