'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { Check, Loader2, Zap, ArrowRight } from 'lucide-react'
import { useCreateCheckoutSession, useBillingInfo } from '@/lib/hooks/use-settings'
import {
  PLAN_CONFIGS,
  PAID_PLAN_IDS,
  formatLimit,
  getAnnualSavingsPercent,
  type PlanId,
} from '@/lib/plan-constants'
import { cn } from '@/lib/utils'

interface PlanStepProps {
  onNext: () => void
  onBack: () => void
  onSkip: () => void
}

const planCards = PAID_PLAN_IDS.map((id) => {
  const plan = PLAN_CONFIGS[id]
  return {
    id,
    name: plan.name,
    description: plan.description,
    popular: id === 'grow',
    features: [
      `${formatLimit(plan.maxHelpers)} active helpers`,
      `${plan.maxConnections} CRM connections`,
      `${formatLimit(plan.maxExecutions)} executions/mo`,
      `${plan.maxApiKeys} API keys`,
      `${plan.maxTeamMembers} team members`,
      plan.overageRate > 0
        ? `$${plan.overageRate}/execution overage`
        : 'Dedicated support',
    ],
  }
})

export function PlanStep({ onNext, onBack, onSkip }: PlanStepProps) {
  const searchParams = useSearchParams()
  const sessionId = searchParams.get('session_id')
  const createCheckout = useCreateCheckoutSession()
  const { data: billing, refetch: refetchBilling } = useBillingInfo()
  const [checkoutPlan, setCheckoutPlan] = useState<string | null>(null)
  const [activating, setActivating] = useState(false)
  const [isAnnual, setIsAnnual] = useState(false)

  // Handle return from Stripe Checkout — poll until plan is activated
  useEffect(() => {
    if (!sessionId) return

    setActivating(true)
    let attempts = 0
    const maxAttempts = 15

    const interval = setInterval(async () => {
      attempts++
      const result = await refetchBilling()
      const plan = result.data?.plan

      if (plan && plan !== 'free' && plan !== 'trial') {
        clearInterval(interval)
        setActivating(false)
        onNext()
      } else if (attempts >= maxAttempts) {
        clearInterval(interval)
        setActivating(false)
        // Even if polling times out, proceed — webhook may still be processing
        onNext()
      }
    }, 2000)

    return () => clearInterval(interval)
  }, [sessionId, refetchBilling, onNext])

  const handleSelectPlan = (planId: PlanId) => {
    if (planId === 'free') return
    setCheckoutPlan(planId)
    createCheckout.mutate(
      {
        plan: planId as 'start' | 'grow' | 'deliver',
        returnUrl: '/onboarding?step=plan',
        billingPeriod: isAnnual ? 'annual' : 'monthly',
      },
      {
        onSuccess: (res) => {
          if (res.data?.url) {
            window.location.href = res.data.url
          }
          setCheckoutPlan(null)
        },
        onError: () => {
          setCheckoutPlan(null)
        },
      }
    )
  }

  // Show activation spinner when returning from Stripe
  if (activating) {
    return (
      <div className="flex flex-col items-center justify-center space-y-4 py-16">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <div className="text-center">
          <h2 className="text-xl font-semibold">Activating your subscription...</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            This only takes a moment
          </p>
        </div>
      </div>
    )
  }

  const maxSavings = Math.max(
    ...PAID_PLAN_IDS.map((id) => getAnnualSavingsPercent(id))
  )

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h2 className="text-2xl font-bold">Choose Your Plan</h2>
        <p className="mt-1 text-muted-foreground">
          All plans include a 14-day free trial. No charge until the trial ends.
        </p>
      </div>

      {/* Billing toggle */}
      <div className="flex items-center justify-center gap-3">
        <span
          className={cn(
            'text-sm font-medium transition-colors',
            !isAnnual ? 'text-foreground' : 'text-muted-foreground'
          )}
        >
          Monthly
        </span>
        <button
          type="button"
          role="switch"
          aria-checked={isAnnual}
          onClick={() => setIsAnnual(!isAnnual)}
          className={cn(
            'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors',
            isAnnual ? 'bg-primary' : 'bg-muted-foreground/30'
          )}
        >
          <span
            className={cn(
              'pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow-sm ring-0 transition-transform',
              isAnnual ? 'translate-x-5' : 'translate-x-0'
            )}
          />
        </button>
        <span
          className={cn(
            'text-sm font-medium transition-colors',
            isAnnual ? 'text-foreground' : 'text-muted-foreground'
          )}
        >
          Annual
        </span>
        {isAnnual && (
          <span className="rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-semibold text-primary">
            Save up to {maxSavings}%
          </span>
        )}
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {planCards.map((plan) => {
          const config = PLAN_CONFIGS[plan.id]
          const price = isAnnual ? config.annualMonthlyPrice : config.monthlyPrice

          return (
            <div
              key={plan.id}
              className={`relative flex flex-col rounded-lg border p-5 ${
                plan.popular
                  ? 'border-primary ring-2 ring-primary/20'
                  : 'border-border'
              }`}
            >
              {plan.popular && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                  <span className="inline-flex items-center gap-1 rounded-full bg-primary px-3 py-0.5 text-xs font-medium text-primary-foreground">
                    <Zap className="h-3 w-3" /> Most Popular
                  </span>
                </div>
              )}

              <div className="mb-4">
                <h3 className="text-lg font-semibold">{plan.name}</h3>
                <p className="text-sm text-muted-foreground">{plan.description}</p>
                <div className="mt-2">
                  <span className="text-3xl font-bold">${price}</span>
                  <span className="text-muted-foreground">/mo</span>
                  {isAnnual && (
                    <span className="ml-2 text-xs text-muted-foreground">
                      billed yearly
                    </span>
                  )}
                </div>
              </div>

              <ul className="mb-5 flex-1 space-y-2">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-center gap-2 text-sm">
                    <Check className="h-4 w-4 shrink-0 text-primary" />
                    {feature}
                  </li>
                ))}
              </ul>

              <button
                onClick={() => handleSelectPlan(plan.id)}
                disabled={checkoutPlan !== null}
                className={`inline-flex w-full items-center justify-center gap-2 rounded-md px-4 py-2.5 text-sm font-medium transition-colors ${
                  plan.popular
                    ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                    : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
                } disabled:opacity-50`}
              >
                {checkoutPlan === plan.id ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Redirecting...
                  </>
                ) : (
                  <>
                    Start Free Trial
                    <ArrowRight className="h-4 w-4" />
                  </>
                )}
              </button>
            </div>
          )
        })}
      </div>

      <div className="flex items-center justify-between pt-2">
        <button
          onClick={onBack}
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          Back
        </button>
        <button
          onClick={onSkip}
          className="text-sm font-medium text-primary hover:text-primary/80"
        >
          Skip — Start Free Trial
        </button>
      </div>
    </div>
  )
}
