'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { Check, Loader2, Zap, ArrowRight } from 'lucide-react'
import { useCreateCheckoutSession, useBillingInfo } from '@/lib/hooks/use-settings'

interface PlanStepProps {
  onNext: () => void
  onBack: () => void
  onSkip: () => void
}

const plans = [
  {
    id: 'start' as const,
    name: 'Start',
    price: 39,
    description: 'For solopreneurs getting started',
    features: [
      '10 active helpers',
      '2 CRM connections',
      '5,000 executions/mo',
      '2 API keys',
      'Email support',
    ],
  },
  {
    id: 'grow' as const,
    name: 'Grow',
    price: 59,
    popular: true,
    description: 'For growing businesses',
    features: [
      '50 active helpers',
      '5 CRM connections',
      '25,000 executions/mo',
      '10 API keys',
      '5 team members',
      'Priority support',
    ],
  },
  {
    id: 'deliver' as const,
    name: 'Deliver',
    price: 79,
    description: 'For teams that need it all',
    features: [
      'Unlimited helpers',
      '20 CRM connections',
      '100,000 executions/mo',
      '100 API keys',
      '100 team members',
      'Dedicated support',
    ],
  },
]

export function PlanStep({ onNext, onBack, onSkip }: PlanStepProps) {
  const searchParams = useSearchParams()
  const sessionId = searchParams.get('session_id')
  const createCheckout = useCreateCheckoutSession()
  const { data: billing, refetch: refetchBilling } = useBillingInfo()
  const [checkoutPlan, setCheckoutPlan] = useState<string | null>(null)
  const [activating, setActivating] = useState(false)

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

      if (plan && plan !== 'free') {
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

  const handleSelectPlan = (planId: 'start' | 'grow' | 'deliver') => {
    setCheckoutPlan(planId)
    createCheckout.mutate(
      { plan: planId, returnUrl: '/onboarding?step=plan' },
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

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h2 className="text-2xl font-bold">Choose Your Plan</h2>
        <p className="mt-1 text-muted-foreground">
          All plans include a 14-day free trial. No charge until the trial ends.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {plans.map((plan) => (
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
                <span className="text-3xl font-bold">${plan.price}</span>
                <span className="text-muted-foreground">/mo</span>
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
        ))}
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
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          Continue with free sandbox
        </button>
      </div>
    </div>
  )
}
