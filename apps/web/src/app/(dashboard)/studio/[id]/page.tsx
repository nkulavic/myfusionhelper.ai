'use client'

import { useParams } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Loader2 } from 'lucide-react'
import { useDashboard } from '@/lib/hooks/use-studio'
import { DashboardCanvas } from '@/components/studio/dashboard-canvas'

export default function StudioDashboardPage() {
  const params = useParams()
  const dashboardId = params.id as string

  const { data: dashboard, isLoading } = useDashboard(dashboardId)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!dashboard) {
    return (
      <div className="animate-fade-in-up space-y-6">
        <div className="flex items-center gap-4">
          <Link href="/studio" className="rounded-md p-2 hover:bg-accent">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold">Dashboard not found</h1>
            <p className="text-muted-foreground">
              This dashboard may have been deleted.
            </p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="animate-fade-in-up space-y-4">
      <div className="flex items-center gap-3">
        <Link href="/studio" className="rounded-md p-2 hover:bg-accent">
          <ArrowLeft className="h-5 w-5" />
        </Link>
      </div>
      <DashboardCanvas dashboard={dashboard} />
    </div>
  )
}
