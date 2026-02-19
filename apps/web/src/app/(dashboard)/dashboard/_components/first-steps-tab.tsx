'use client'

import Link from 'next/link'
import { CheckCircle, Link2, Blocks, Activity, ArrowRight, Clock } from 'lucide-react'
import { useAuthStore } from '@/lib/stores/auth-store'
import { usePlanLimits } from '@/lib/hooks/use-plan-limits'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { PlatformConnection, Helper, HelperExecution } from '@myfusionhelper/types'

interface FirstStepsTabProps {
  connections: PlatformConnection[] | undefined
  helpers: Helper[] | undefined
  executions: HelperExecution[] | undefined
  isLoading: boolean
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

export function FirstStepsTab({
  connections,
  helpers,
  executions,
  isLoading,
}: FirstStepsTabProps) {
  const user = useAuthStore((s) => s.user)
  const { isTrialing, daysRemaining, isTrialExpired } = usePlanLimits()
  const firstName = user?.name?.split(' ')[0] || 'there'

  const completedSteps = [
    (connections?.length ?? 0) > 0,
    (helpers?.length ?? 0) > 0,
    (executions?.length ?? 0) > 0,
  ]

  const completedCount = completedSteps.filter(Boolean).length
  const nextStepIndex = completedSteps.indexOf(false)
  const totalDays = 14
  const daysPassed = Math.max(0, totalDays - daysRemaining)
  const progressPercent = Math.min(100, Math.round((daysPassed / totalDays) * 100))

  if (isLoading) {
    return <FirstStepsSkeleton />
  }

  return (
    <div className="space-y-6">
      {/* Trial Progress Widget */}
      {(isTrialing || isTrialExpired) && (
        <div className="rounded-lg border bg-card p-6">
          <div className="flex items-start justify-between">
            <div>
              <div className="flex items-center gap-2">
                <Clock className="h-5 w-5 text-primary" />
                <h3 className="text-lg font-bold">14-Day Free Trial</h3>
              </div>
              <p className="mt-1 text-sm text-muted-foreground">
                {isTrialExpired
                  ? 'Your trial has expired. Choose a plan to continue using all features.'
                  : `You have ${daysRemaining} day${daysRemaining !== 1 ? 's' : ''} remaining in your free trial.`}
              </p>
            </div>
            <Button asChild size="sm">
              <Link href="/plans">
                {isTrialExpired ? 'Choose Plan' : 'Upgrade'}
              </Link>
            </Button>
          </div>
          <div className="mt-4">
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Day {daysPassed}</span>
              <span>Day {totalDays}</span>
            </div>
            <div className="mt-1 h-2 rounded-full bg-muted">
              <div
                className={cn(
                  'h-2 rounded-full transition-all duration-500',
                  isTrialExpired
                    ? 'bg-destructive'
                    : daysRemaining <= 2
                      ? 'bg-destructive'
                      : daysRemaining <= 7
                        ? 'bg-amber-500'
                        : 'bg-primary',
                )}
                style={{ width: `${progressPercent}%` }}
              />
            </div>
          </div>
        </div>
      )}

      {/* Getting Started Steps */}
      <div className="rounded-lg border bg-card p-6">
        <div className="mb-5">
          <h2 className="text-lg font-bold">Welcome, {firstName}!</h2>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Complete these steps to get your first automation running.
          </p>
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
                      : 'border-border bg-background',
                )}
              >
                <div
                  className={cn(
                    'flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-medium',
                    isComplete
                      ? 'bg-success text-white'
                      : isNext
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-muted text-muted-foreground',
                  )}
                >
                  {isComplete ? <CheckCircle className="h-4 w-4" /> : i + 1}
                </div>

                <div className="min-w-0 flex-1">
                  <p
                    className={cn(
                      'text-sm font-medium',
                      isComplete && 'text-muted-foreground line-through',
                    )}
                  >
                    {step.label}
                  </p>
                  <p className="mt-0.5 text-xs text-muted-foreground">{step.description}</p>
                </div>

                {!isComplete && (
                  <Link
                    href={step.href}
                    className={cn(
                      'inline-flex shrink-0 items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                      isNext
                        ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                        : 'text-primary hover:bg-primary/10',
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

      {/* Choose Your Plan CTA */}
      {(isTrialing || isTrialExpired) && (
        <div className="rounded-lg border border-primary/20 bg-primary/5 p-6 text-center">
          <h3 className="text-lg font-bold">Ready to pick a plan?</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Compare features and pricing to find the best fit for your business.
          </p>
          <Button asChild className="mt-4">
            <Link href="/plans">View Plans</Link>
          </Button>
        </div>
      )}
    </div>
  )
}

function FirstStepsSkeleton() {
  return (
    <div className="space-y-6">
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-start justify-between">
          <div>
            <Skeleton className="h-5 w-40" />
            <Skeleton className="mt-2 h-4 w-64" />
          </div>
          <Skeleton className="h-9 w-24 rounded-md" />
        </div>
        <Skeleton className="mt-4 h-2 w-full rounded-full" />
      </div>
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
    </div>
  )
}
