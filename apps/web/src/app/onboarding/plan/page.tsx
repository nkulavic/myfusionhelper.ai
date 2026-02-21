'use client'

import { Suspense, useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { Check, Loader2, Zap, ArrowRight, Shield } from 'lucide-react'
import { useCreateCheckoutSession, useBillingInfo } from '@/lib/hooks/use-settings'
import {
  PLAN_CONFIGS,
  PAID_PLAN_IDS,
  formatLimit,
  getAnnualSavingsPercent,
  isTrialPlan,
  type PlanId,
} from '@/lib/plan-constants'
import { useAuthStore } from '@/lib/stores/auth-store'
import { cn } from '@/lib/utils'

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

export default function PlanGatePage() {
  return (
    <Suspense fallback={null}>
      <PlanGateContent />
    </Suspense>
  )
}

function PlanGateContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const sessionId = searchParams.get('session_id')
  const { user } = useAuthStore()

  const createCheckout = useCreateCheckoutSession()
  const { data: billing, refetch: refetchBilling, isLoading: billingLoading } = useBillingInfo()
  const [checkoutPlan, setCheckoutPlan] = useState<string | null>(null)
  const [activating, setActivating] = useState(false)
  const [isAnnual, setIsAnnual] = useState(false)

  // If user already has a paid plan (not trial/free), skip to onboarding
  useEffect(() => {
    if (!billingLoading && billing && !isTrialPlan(billing.plan)) {
      router.replace('/onboarding')
    }
  }, [billing, billingLoading, router])

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

      if (plan && !isTrialPlan(plan)) {
        clearInterval(interval)
        setActivating(false)
        router.push('/onboarding')
      } else if (attempts >= maxAttempts) {
        clearInterval(interval)
        setActivating(false)
        // Even if polling times out, proceed — webhook may still be processing
        router.push('/onboarding')
      }
    }, 2000)

    return () => clearInterval(interval)
  }, [sessionId, refetchBilling, router])

  const handleSelectPlan = (planId: PlanId) => {
    if (planId === 'free') return
    setCheckoutPlan(planId)
    createCheckout.mutate(
      {
        plan: planId as 'start' | 'grow' | 'deliver',
        returnUrl: '/onboarding/plan',
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

  const firstName = user?.name?.split(' ')[0] || 'there'

  // Show activation spinner when returning from Stripe
  if (activating) {
    return (
      <div className="flex flex-col items-center justify-center space-y-4 py-24">
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

  // Show loading while checking billing status
  if (billingLoading) {
    return (
      <div className="flex items-center justify-center py-24">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const maxSavings = Math.max(
    ...PAID_PLAN_IDS.map((id) => getAnnualSavingsPercent(id))
  )

  return (
    <div className="space-y-8">
      <div className="text-center">
        <h1 className="text-3xl font-bold">Welcome, {firstName}!</h1>
        <p className="mt-2 text-lg text-muted-foreground">
          Choose a plan to get started. All plans include a 14-day free trial &mdash; no charge until the trial ends.
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

      {/* Trust signals */}
      <div className="flex items-center justify-center gap-6 text-xs text-muted-foreground">
        <span className="flex items-center gap-1.5">
          <Shield className="h-3.5 w-3.5" />
          Cancel anytime
        </span>
        <span className="flex items-center gap-1.5">
          <Shield className="h-3.5 w-3.5" />
          No charge for 14 days
        </span>
        <span className="flex items-center gap-1.5">
          <Shield className="h-3.5 w-3.5" />
          Secure checkout via Stripe
        </span>
      </div>
    </div>
  )
}
