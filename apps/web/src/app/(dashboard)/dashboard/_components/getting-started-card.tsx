'use client'

import Link from 'next/link'
import { CheckCircle, Link2, Blocks, Activity, X, ArrowRight } from 'lucide-react'
import { useAuthStore } from '@/lib/stores/auth-store'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { PlatformConnection, Helper, HelperExecution } from '@myfusionhelper/types'

interface GettingStartedCardProps {
  connections: PlatformConnection[] | undefined
  helpers: Helper[] | undefined
  executions: HelperExecution[] | undefined
  isLoading: boolean
  onDismiss: () => void
}

const steps = [
  {
    label: 'Connect your CRM',
    description: 'Link your CRM platform so helpers can access your contacts and data.',
    cta: 'Add connection',
    href: '/connections',
    icon: Link2,
  },
  {
    label: 'Create your first helper',
    description: 'Pick from 62 pre-built automations and configure one for your workflow.',
    cta: 'Browse helpers',
    href: '/helpers?view=new',
    icon: Blocks,
  },
  {
    label: 'Watch your first run',
    description: 'Trigger a helper from your CRM and see the execution results here.',
    cta: 'View executions',
    href: '/executions',
    icon: Activity,
  },
]

export function GettingStartedCard({
  connections,
  helpers,
  executions,
  isLoading,
  onDismiss,
}: GettingStartedCardProps) {
  const user = useAuthStore((s) => s.user)
  const firstName = user?.name?.split(' ')[0] || 'there'

  const completedSteps = [
    (connections?.length ?? 0) > 0,
    (helpers?.length ?? 0) > 0,
    (executions?.length ?? 0) > 0,
  ]

  const completedCount = completedSteps.filter(Boolean).length
  const nextStepIndex = completedSteps.indexOf(false)

  if (isLoading) {
    return <GettingStartedSkeleton />
  }

  return (
    <div className="relative rounded-lg border bg-card p-6">
      {/* Dismiss */}
      <button
        onClick={onDismiss}
        className="absolute right-4 top-4 rounded-md p-1 text-muted-foreground hover:bg-accent hover:text-foreground"
        aria-label="Dismiss getting started"
      >
        <X className="h-4 w-4" />
      </button>

      {/* Header */}
      <div className="mb-5">
        <h2 className="text-lg font-bold">Welcome, {firstName}!</h2>
        <p className="mt-0.5 text-sm text-muted-foreground">
          Complete these steps to get your first automation running.
        </p>
        {/* Progress */}
        <div className="mt-3 flex items-center gap-3">
          <div className="h-2 flex-1 rounded-full bg-muted">
            <div
              className="h-2 rounded-full bg-primary transition-all duration-500"
              style={{ width: `${(completedCount / steps.length) * 100}%` }}
            />
          </div>
          <span className="text-xs font-medium text-muted-foreground">
            {completedCount}/{steps.length}
          </span>
        </div>
      </div>

      {/* Steps */}
      <div className="space-y-3">
        {steps.map((step, i) => {
          const isComplete = completedSteps[i]
          const isNext = i === nextStepIndex

          return (
            <div
              key={step.label}
              className={cn(
                'flex items-start gap-4 rounded-lg border p-4 transition-colors',
                isComplete
                  ? 'border-success/20 bg-success/5'
                  : isNext
                    ? 'border-primary/30 bg-primary/5'
                    : 'border-border bg-background'
              )}
            >
              {/* Step indicator */}
              <div
                className={cn(
                  'flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-medium',
                  isComplete
                    ? 'bg-success text-white'
                    : isNext
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted text-muted-foreground'
                )}
              >
                {isComplete ? (
                  <CheckCircle className="h-4 w-4" />
                ) : (
                  i + 1
                )}
              </div>

              {/* Content */}
              <div className="flex-1 min-w-0">
                <p
                  className={cn(
                    'text-sm font-medium',
                    isComplete && 'text-muted-foreground line-through'
                  )}
                >
                  {step.label}
                </p>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  {step.description}
                </p>
              </div>

              {/* CTA */}
              {!isComplete && (
                <Link
                  href={step.href}
                  className={cn(
                    'inline-flex shrink-0 items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                    isNext
                      ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                      : 'text-primary hover:bg-primary/10'
                  )}
                >
                  {step.cta}
                  <ArrowRight className="h-3 w-3" />
                </Link>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

function GettingStartedSkeleton() {
  return (
    <div className="rounded-lg border bg-card p-6">
      <Skeleton className="h-5 w-40" />
      <Skeleton className="mt-2 h-4 w-72" />
      <Skeleton className="mt-3 h-2 w-full rounded-full" />
      <div className="mt-5 space-y-3">
        {[1, 2, 3].map((i) => (
          <div key={i} className="flex items-start gap-4 rounded-lg border p-4">
            <Skeleton className="h-8 w-8 rounded-full" />
            <div className="flex-1">
              <Skeleton className="h-4 w-32" />
              <Skeleton className="mt-1.5 h-3 w-56" />
            </div>
            <Skeleton className="h-8 w-28 rounded-md" />
          </div>
        ))}
      </div>
    </div>
  )
}
