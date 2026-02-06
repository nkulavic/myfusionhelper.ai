'use client'

import Link from 'next/link'
import { useParams } from 'next/navigation'
import {
  ArrowLeft,
  CheckCircle,
  XCircle,
  Clock,
  RefreshCw,
  Copy,
  RotateCcw,
  Blocks,
  Timer,
  Calendar,
  AlertTriangle,
  ChevronRight,
  Loader2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useExecution } from '@/lib/hooks/use-helpers'
import { Skeleton } from '@/components/ui/skeleton'

export default function ExecutionDetailPage() {
  const params = useParams()
  const executionId = params.id as string
  const { data: execution, isLoading, error } = useExecution(executionId)

  const getStatusBadge = (status: string) => {
    const styles: Record<string, string> = {
      completed: 'bg-success/10 text-success',
      failed: 'bg-destructive/10 text-destructive',
      running: 'bg-info/10 text-info',
      pending: 'bg-warning/10 text-warning',
    }
    return (
      <span className={cn('rounded-full px-2.5 py-0.5 text-xs font-medium capitalize', styles[status] || '')}>
        {status}
      </span>
    )
  }

  if (isLoading) {
    return (
      <div className="mx-auto max-w-4xl space-y-6">
        <div className="flex items-center gap-4">
          <Link href="/executions" className="rounded-md p-2 hover:bg-accent">
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div className="space-y-2">
            <Skeleton className="h-7 w-48" />
            <Skeleton className="h-4 w-32" />
          </div>
        </div>
        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2 space-y-6">
            <div className="rounded-lg border bg-card p-5">
              <Skeleton className="h-5 w-40 mb-4" />
              <div className="space-y-3">
                <Skeleton className="h-12 w-full" />
                <Skeleton className="h-12 w-full" />
                <Skeleton className="h-12 w-full" />
              </div>
            </div>
          </div>
          <div>
            <div className="rounded-lg border bg-card p-5">
              <Skeleton className="h-5 w-20 mb-4" />
              <div className="space-y-3">
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-full" />
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (error || !execution) {
    return (
      <div className="mx-auto max-w-4xl space-y-6">
        <div className="flex items-center gap-4">
          <Link href="/executions" className="rounded-md p-2 hover:bg-accent">
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold">Execution Not Found</h1>
            <p className="text-sm text-muted-foreground">
              {error ? 'Unable to load execution details.' : `Execution ${executionId} was not found.`}
            </p>
          </div>
        </div>
        <Link
          href="/executions"
          className="inline-flex items-center gap-2 rounded-md border px-4 py-2 text-sm font-medium hover:bg-accent"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Executions
        </Link>
      </div>
    )
  }

  const durationStr = execution.durationMs
    ? execution.durationMs > 1000
      ? `${(execution.durationMs / 1000).toFixed(1)}s`
      : `${execution.durationMs}ms`
    : '-'

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <Link href="/executions" className="rounded-md p-2 hover:bg-accent">
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold font-mono">{execution.executionId}</h1>
              {getStatusBadge(execution.status)}
            </div>
            <p className="text-sm text-muted-foreground">
              Helper: {execution.helperId}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => navigator.clipboard.writeText(execution.executionId)}
            className="inline-flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm font-medium hover:bg-accent"
          >
            <Copy className="h-4 w-4" />
            Copy ID
          </button>
        </div>
      </div>

      {/* Error Banner */}
      {execution.errorMessage && (
        <div className="flex items-start gap-3 rounded-lg border border-destructive/30 bg-destructive/10 p-4">
          <AlertTriangle className="mt-0.5 h-5 w-5 flex-shrink-0 text-destructive" />
          <div>
            <p className="text-sm font-medium text-destructive">Execution Failed</p>
            <p className="mt-0.5 text-sm text-destructive/80">{execution.errorMessage}</p>
          </div>
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Timeline */}
          <div className="rounded-lg border bg-card">
            <div className="border-b px-5 py-4">
              <h2 className="font-semibold">Execution Timeline</h2>
            </div>
            <div className="p-5">
              <div className="space-y-4">
                <TimelineItem
                  time={execution.startedAt}
                  label="Started"
                  description="Execution started processing"
                  status="completed"
                />
                {execution.completedAt && (
                  <TimelineItem
                    time={execution.completedAt}
                    label={execution.status === 'failed' ? 'Failed' : 'Completed'}
                    description={
                      execution.errorMessage || `Execution completed in ${durationStr}`
                    }
                    status={execution.status === 'failed' ? 'failed' : 'completed'}
                    isLast
                  />
                )}
                {!execution.completedAt && execution.status === 'running' && (
                  <TimelineItem
                    time={new Date().toISOString()}
                    label="Running"
                    description="Execution is in progress..."
                    status="running"
                    isLast
                  />
                )}
              </div>
            </div>
          </div>

          {/* Input Configuration */}
          {execution.input && Object.keys(execution.input).length > 0 && (
            <div className="rounded-lg border bg-card">
              <div className="border-b px-5 py-4">
                <h2 className="font-semibold">Input Configuration</h2>
              </div>
              <div className="p-5">
                <pre className="rounded-md bg-muted p-4 text-xs font-mono overflow-x-auto">
                  {JSON.stringify(execution.input, null, 2)}
                </pre>
              </div>
            </div>
          )}

          {/* Output */}
          {execution.output && Object.keys(execution.output).length > 0 && (
            <div className="rounded-lg border bg-card">
              <div className="border-b px-5 py-4">
                <h2 className="font-semibold">Output</h2>
              </div>
              <div className="p-5">
                <pre className="rounded-md bg-muted p-4 text-xs font-mono overflow-x-auto">
                  {JSON.stringify(execution.output, null, 2)}
                </pre>
              </div>
            </div>
          )}
        </div>

        {/* Sidebar */}
        <div className="space-y-4">
          {/* Details */}
          <div className="rounded-lg border bg-card p-5 space-y-4">
            <h3 className="font-semibold">Details</h3>
            <div className="space-y-3 text-sm">
              <DetailRow icon={Blocks} label="Helper" value={execution.helperId} />
              {execution.contactId && (
                <DetailRow icon={Calendar} label="Contact" value={execution.contactId} />
              )}
              <DetailRow icon={Timer} label="Duration" value={durationStr} />
              <DetailRow
                icon={Calendar}
                label="Started"
                value={new Date(execution.startedAt).toLocaleString()}
              />
              {execution.completedAt && (
                <DetailRow
                  icon={Calendar}
                  label="Completed"
                  value={new Date(execution.completedAt).toLocaleString()}
                />
              )}
            </div>
          </div>

          {/* Related Links */}
          <div className="rounded-lg border bg-card p-5 space-y-3">
            <h3 className="font-semibold">Related</h3>
            <div className="space-y-2">
              <Link
                href={`/helpers/${execution.helperId}`}
                className="flex items-center justify-between rounded-md px-3 py-2 text-sm hover:bg-accent"
              >
                <span>View Helper</span>
                <ChevronRight className="h-4 w-4 text-muted-foreground" />
              </Link>
              <Link
                href="/executions"
                className="flex items-center justify-between rounded-md px-3 py-2 text-sm hover:bg-accent"
              >
                <span>All Executions</span>
                <ChevronRight className="h-4 w-4 text-muted-foreground" />
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function TimelineItem({
  time,
  label,
  description,
  status,
  isLast = false,
}: {
  time: string
  label: string
  description: string
  status: 'completed' | 'failed' | 'running'
  isLast?: boolean
}) {
  return (
    <div className="flex gap-3">
      <div className="flex flex-col items-center">
        <div
          className={cn(
            'flex h-6 w-6 items-center justify-center rounded-full',
            status === 'completed' && 'bg-success/10',
            status === 'failed' && 'bg-destructive/10',
            status === 'running' && 'bg-info/10'
          )}
        >
          {status === 'completed' && <CheckCircle className="h-3.5 w-3.5 text-success" />}
          {status === 'failed' && <XCircle className="h-3.5 w-3.5 text-destructive" />}
          {status === 'running' && <RefreshCw className="h-3.5 w-3.5 animate-spin text-info" />}
        </div>
        {!isLast && <div className="mt-1 h-full w-px bg-border" />}
      </div>
      <div className="flex-1 pb-4">
        <div className="flex items-center justify-between">
          <p className="text-sm font-medium">{label}</p>
          <p className="text-xs font-mono text-muted-foreground">
            {new Date(time).toLocaleTimeString()}
          </p>
        </div>
        <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>
      </div>
    </div>
  )
}

function DetailRow({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ComponentType<{ className?: string }>
  label: string
  value: string
}) {
  return (
    <div className="flex items-center gap-2">
      <Icon className="h-4 w-4 text-muted-foreground" />
      <span className="text-muted-foreground">{label}</span>
      <span className="ml-auto font-medium text-right truncate max-w-[150px]">{value}</span>
    </div>
  )
}
