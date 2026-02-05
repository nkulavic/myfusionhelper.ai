'use client'

import { useMemo } from 'react'
import Link from 'next/link'
import {
  ArrowLeft,
  Play,
  Settings,
  Trash2,
  ToggleLeft,
  ToggleRight,
  Clock,
  CheckCircle,
  XCircle,
  Copy,
  AlertTriangle,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { helpersCatalog } from '@/lib/helpers-catalog'
import { useHelper, useExecutions, useDeleteHelper, useUpdateHelper } from '@/lib/hooks/use-helpers'
import { Skeleton } from '@/components/ui/skeleton'

interface HelperDetailProps {
  helperId: string
  onBack: () => void
}

function DetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Skeleton className="h-8 w-8" />
        <div>
          <Skeleton className="h-7 w-48" />
          <Skeleton className="mt-1 h-4 w-64" />
        </div>
      </div>
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <div className="grid grid-cols-3 gap-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="rounded-lg border bg-card p-4">
                <Skeleton className="h-3 w-20" />
                <Skeleton className="mt-2 h-6 w-16" />
              </div>
            ))}
          </div>
          <Skeleton className="h-48 rounded-lg" />
        </div>
        <div className="space-y-4">
          <Skeleton className="h-48 rounded-lg" />
          <Skeleton className="h-32 rounded-lg" />
        </div>
      </div>
    </div>
  )
}

export function HelperDetail({ helperId, onBack }: HelperDetailProps) {
  const { data: helper, isLoading, error } = useHelper(helperId)
  const { data: executions } = useExecutions({ helperId, limit: 5 })
  const deleteHelper = useDeleteHelper()
  const updateHelper = useUpdateHelper()

  const helperTemplate = useMemo(
    () => helpersCatalog.find((h) => h.id === helper?.type),
    [helper?.type]
  )

  if (isLoading) return <DetailSkeleton />

  if (error || !helper) {
    return (
      <div>
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <AlertTriangle className="mb-4 h-12 w-12 text-muted-foreground/50" />
          <h2 className="text-lg font-semibold">Helper not found</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            {error instanceof Error ? error.message : 'The helper could not be loaded.'}
          </p>
          <button
            onClick={onBack}
            className="mt-4 inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to Helpers
          </button>
        </div>
      </div>
    )
  }

  const isEnabled = helper.status === 'active'

  const handleToggle = () => {
    updateHelper.mutate({
      id: helperId,
      input: { enabled: !isEnabled },
    })
  }

  const handleDelete = () => {
    deleteHelper.mutate(helperId, {
      onSuccess: () => onBack(),
    })
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <button onClick={onBack} className="rounded-md p-2 hover:bg-accent">
            <ArrowLeft className="h-4 w-4" />
          </button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold">{helper.name}</h1>
              <span
                className={cn(
                  'rounded-full px-2.5 py-0.5 text-xs font-medium',
                  isEnabled
                    ? 'bg-success/10 text-success'
                    : 'bg-muted text-muted-foreground'
                )}
              >
                {isEnabled ? 'Active' : 'Inactive'}
              </span>
            </div>
            <p className="text-sm text-muted-foreground">
              {helper.description || helperTemplate?.description || 'Custom automation helper'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button className="inline-flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm font-medium hover:bg-accent">
            <Play className="h-4 w-4" />
            Test Run
          </button>
          <button className="inline-flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
            <Settings className="h-4 w-4" />
            Edit Config
          </button>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Stats */}
          <div className="grid grid-cols-3 gap-4">
            <div className="rounded-lg border bg-card p-4">
              <p className="text-xs text-muted-foreground">Type</p>
              <p className="mt-1 text-lg font-bold">{helper.type}</p>
            </div>
            <div className="rounded-lg border bg-card p-4">
              <p className="text-xs text-muted-foreground">Category</p>
              <p className="mt-1 text-lg font-bold capitalize">{helper.category}</p>
            </div>
            <div className="rounded-lg border bg-card p-4">
              <p className="text-xs text-muted-foreground">Connection</p>
              <p className="mt-1 text-lg font-bold truncate">{helper.connectionId || 'None'}</p>
            </div>
          </div>

          {/* Configuration */}
          <div className="rounded-lg border bg-card">
            <div className="border-b px-5 py-4">
              <h2 className="font-semibold">Configuration</h2>
            </div>
            <div className="p-5">
              <pre className="rounded-md bg-muted p-4 text-sm font-mono overflow-x-auto">
                {JSON.stringify(helper.config, null, 2)}
              </pre>
            </div>
          </div>

          {/* Recent Executions */}
          <div className="rounded-lg border bg-card">
            <div className="flex items-center justify-between border-b px-5 py-4">
              <h2 className="font-semibold">Recent Executions</h2>
              <Link
                href={`/executions?helper=${helperId}`}
                className="text-xs text-primary hover:underline"
              >
                View all
              </Link>
            </div>
            {executions && executions.length > 0 ? (
              <div className="divide-y">
                {executions.map((exec) => (
                  <Link
                    key={exec.id}
                    href={`/executions/${exec.id}`}
                    className="flex items-center gap-4 px-5 py-3 hover:bg-accent/50"
                  >
                    <div className="flex-shrink-0">
                      {exec.status === 'completed' ? (
                        <CheckCircle className="h-4 w-4 text-success" />
                      ) : exec.status === 'failed' ? (
                        <XCircle className="h-4 w-4 text-destructive" />
                      ) : (
                        <Clock className="h-4 w-4 animate-spin text-info" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-mono truncate">
                        {exec.contactId || 'No contact'}
                      </p>
                    </div>
                    <p className="text-xs font-mono text-muted-foreground">
                      {exec.durationMs ? `${exec.durationMs}ms` : '-'}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(exec.startedAt).toLocaleTimeString()}
                    </p>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="py-8 text-center text-sm text-muted-foreground">
                No executions yet
              </div>
            )}
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-4">
          {/* Details */}
          <div className="rounded-lg border bg-card p-5 space-y-4">
            <h3 className="font-semibold">Details</h3>
            <div className="space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">ID</span>
                <span className="font-mono text-xs">{helper.id}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Created</span>
                <span className="font-medium">
                  {new Date(helper.createdAt).toLocaleDateString()}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Updated</span>
                <span className="font-medium">
                  {new Date(helper.updatedAt).toLocaleDateString()}
                </span>
              </div>
            </div>
          </div>

          {/* API Endpoint */}
          <div className="rounded-lg border bg-card p-5 space-y-3">
            <h3 className="font-semibold">API Endpoint</h3>
            <div className="rounded-md bg-muted p-3">
              <p className="text-[11px] font-mono text-muted-foreground break-all">
                POST /helpers/{helper.id}/execute
              </p>
            </div>
            <button
              onClick={() => navigator.clipboard.writeText(`POST /helpers/${helper.id}/execute`)}
              className="inline-flex items-center gap-1.5 text-xs text-primary hover:underline"
            >
              <Copy className="h-3 w-3" />
              Copy endpoint
            </button>
          </div>

          {/* Actions */}
          <div className="rounded-lg border bg-card p-5 space-y-3">
            <h3 className="font-semibold">Actions</h3>
            <div className="space-y-2">
              <button
                onClick={handleToggle}
                disabled={updateHelper.isPending}
                className="flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-accent disabled:opacity-50"
              >
                {isEnabled ? (
                  <ToggleRight className="h-4 w-4 text-success" />
                ) : (
                  <ToggleLeft className="h-4 w-4 text-muted-foreground" />
                )}
                {isEnabled ? 'Disable Helper' : 'Enable Helper'}
              </button>
              <button
                onClick={handleDelete}
                disabled={deleteHelper.isPending}
                className="flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm text-destructive hover:bg-destructive/10 disabled:opacity-50"
              >
                <Trash2 className="h-4 w-4" />
                {deleteHelper.isPending ? 'Deleting...' : 'Delete Helper'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
