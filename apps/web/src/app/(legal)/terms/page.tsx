import type { Metadata } from 'next'
import { TermsContent } from '@/components/legal/terms-content'

export const metadata: Metadata = {
  title: 'Terms of Service | MyFusion Helper',
  description: 'MyFusion Helper Terms of Service',
}

export default function TermsPage() {
  return (
    <article className="rounded-lg border bg-card p-8 shadow-sm">
      <h1 className="mb-6 text-3xl font-bold text-foreground">Terms of Service</h1>
      <TermsContent />
    </article>
  )
}
