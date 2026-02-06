import type { Metadata } from 'next'
import { PrivacyContent } from '@/components/legal/privacy-content'

export const metadata: Metadata = {
  title: 'Privacy Policy | MyFusion Helper',
  description: 'MyFusion Helper Privacy Policy',
}

export default function PrivacyPage() {
  return (
    <article className="rounded-lg border bg-card p-8 shadow-sm">
      <h1 className="mb-6 text-3xl font-bold text-foreground">Privacy Policy</h1>
      <PrivacyContent />
    </article>
  )
}
