'use client'

import { Check, CheckCircle, ExternalLink, Loader2, Zap } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  PLAN_CONFIGS,
  PAID_PLAN_IDS,
  formatLimit,
  getAnnualSavingsPercent,
  isTrialPlan,
  type PlanId,
} from '@/lib/plan-constants'

interface PlanCardProps {
  planId: PlanId
  billingPeriod: 'monthly' | 'annual'
  currentPlan: string
  isPopular?: boolean
  onSelect: (planId: string) => void
  isLoading?: boolean
  isManaging?: boolean
  variant?: 'compact' | 'full'
}

export function PlanCard({
  planId,
  billingPeriod,
  currentPlan,
  isPopular,
  onSelect,
  isLoading,
  isManaging,
  variant = 'full',
}: PlanCardProps) {
  const config = PLAN_CONFIGS[planId]
  const isAnnual = billingPeriod === 'annual'
  const price = isAnnual ? config.annualMonthlyPrice : config.monthlyPrice
  const originalPrice = isAnnual ? config.monthlyPrice : null
  const isCurrentPlan = currentPlan === planId
  const isOnTrial = isTrialPlan(currentPlan)
  const isCompact = variant === 'compact'

  const features = [
    `${formatLimit(config.maxHelpers)} active helpers`,
    `${config.maxConnections} CRM connections`,
    `${formatLimit(config.maxExecutions)} monthly executions${isCompact ? '' : ' included'}`,
    `${config.maxApiKeys} API keys`,
    `${config.maxTeamMembers} team members`,
    config.overageRate > 0
      ? `${isCompact ? 'Overage: ' : ''}$${config.overageRate}/execution${isCompact ? '' : ' overage'}`
      : 'Dedicated support',
  ]

  const renderButton = () => {
    if (isCurrentPlan) {
      return (
        <Button variant="outline" disabled className="w-full">
          {isCompact ? <CheckCircle className="h-4 w-4" /> : null}
          Current Plan
        </Button>
      )
    }

    if (!isOnTrial) {
      return (
        <Button
          variant={isPopular ? 'default' : 'outline'}
          className="w-full"
          onClick={() => onSelect(planId)}
          disabled={isManaging}
        >
          {isManaging && <Loader2 className="h-4 w-4 animate-spin" />}
          <ExternalLink className="h-4 w-4" />
          {PAID_PLAN_IDS.indexOf(currentPlan as PlanId) < PAID_PLAN_IDS.indexOf(planId)
            ? 'Upgrade'
            : 'Downgrade'}
        </Button>
      )
    }

    return (
      <Button
        variant={isPopular ? 'default' : 'outline'}
        className="w-full"
        onClick={() => onSelect(planId)}
        disabled={isLoading}
      >
        {isLoading && <Loader2 className="h-4 w-4 animate-spin" />}
        {isCompact ? 'Get Started' : `Choose ${config.name}`}
      </Button>
    )
  }

  return (
    <div
      className={cn(
        'relative flex flex-col transition-all',
        isCompact ? 'rounded-lg border p-5' : 'rounded-xl border p-6',
        isPopular && (isCompact ? 'border-primary shadow-sm' : 'border-primary shadow-md'),
        isCurrentPlan &&
          (isCompact ? 'bg-primary/5 ring-1 ring-primary' : 'bg-primary/5 ring-2 ring-primary'),
      )}
    >
      {isPopular && (
        <div className="absolute -top-3 left-1/2 -translate-x-1/2">
          <Badge className="gap-1">
            <Zap className="h-3 w-3" />
            Most Popular
          </Badge>
        </div>
      )}

      <div className="mb-4">
        <h3 className={cn('font-bold', isCompact ? 'text-lg' : 'text-xl')}>{config.name}</h3>
        <p className={cn('mt-1 text-muted-foreground', isCompact ? 'text-xs' : 'text-sm')}>
          {config.description}
        </p>
      </div>

      <div className={isCompact ? 'mb-4' : 'mb-6'}>
        <div className="flex items-baseline gap-1">
          <span className={cn('font-bold', isCompact ? 'text-3xl' : 'text-4xl')}>${price}</span>
          <span className="text-muted-foreground">/month</span>
        </div>
        {isAnnual && originalPrice && (
          <p className="mt-1 text-sm text-muted-foreground">
            <span className="line-through">${originalPrice}/mo</span>
            <span className="ml-1 font-medium text-primary">
              Save {getAnnualSavingsPercent(planId)}%
            </span>
          </p>
        )}
        {isAnnual && isCompact && (
          <span className="text-xs text-muted-foreground">billed yearly</span>
        )}
      </div>

      <ul className={cn('flex-1 space-y-2', isCompact ? 'mb-6' : 'mb-6 space-y-3')}>
        {features.map((feature) => (
          <li key={feature} className="flex items-start gap-2 text-sm">
            <Check className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
            <span>{feature}</span>
          </li>
        ))}
      </ul>

      {renderButton()}
    </div>
  )
}
