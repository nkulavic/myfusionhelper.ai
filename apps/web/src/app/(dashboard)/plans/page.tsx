'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { Check, Loader2, Shield, Zap, Clock, ChevronDown, ChevronUp } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useBillingInfo,
  useCreatePortalSession,
  useCreateCheckoutSession,
} from '@/lib/hooks/use-settings'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { toast } from 'sonner'
import {
  PLAN_CONFIGS,
  PAID_PLAN_IDS,
  COMPARISON_ROWS,
  formatLimit,
  getAnnualSavingsPercent,
  isTrialPlan,
  type PlanId,
} from '@/lib/plan-constants'

export default function PlansPage() {
  const searchParams = useSearchParams()
  const { data: billing } = useBillingInfo()
  const createPortal = useCreatePortalSession()
  const createCheckout = useCreateCheckoutSession()
  const [checkoutPlan, setCheckoutPlan] = useState<string | null>(null)
  const [isAnnual, setIsAnnual] = useState(false)
  const [showComparison, setShowComparison] = useState(false)

  useEffect(() => {
    if (searchParams.get('billing') === 'cancelled') {
      toast.info('Checkout cancelled. You can try again anytime.')
      const url = new URL(window.location.href)
      url.searchParams.delete('billing')
      window.history.replaceState({}, '', url.toString())
    }
  }, [searchParams])

  const currentPlan = billing?.plan || 'trial'
  const isOnTrial = isTrialPlan(currentPlan)

  const handleManageSubscription = () => {
    createPortal.mutate(undefined, {
      onSuccess: (res: { data?: { url: string } }) => {
        if (res.data?.url) {
          window.location.href = res.data.url
        }
      },
    })
  }

  const handleSelectPlan = (planId: 'start' | 'grow' | 'deliver') => {
    if (!isOnTrial) {
      handleManageSubscription()
      return
    }
    setCheckoutPlan(planId)
    createCheckout.mutate(
      {
        plan: planId,
        returnUrl: '/plans',
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
      },
    )
  }

  const maxSavings = Math.max(...PAID_PLAN_IDS.map((id) => getAnnualSavingsPercent(id)))

  const plans = PAID_PLAN_IDS.map((id) => {
    const config = PLAN_CONFIGS[id]
    return {
      id,
      name: config.name,
      description: config.description,
      popular: id === 'grow',
      features: [
        `${formatLimit(config.maxHelpers)} active helpers`,
        `${config.maxConnections} CRM connections`,
        `${formatLimit(config.maxExecutions)} monthly executions`,
        `${config.maxApiKeys} API keys`,
        `${config.maxTeamMembers} team members`,
        config.overageRate > 0 ? `$${config.overageRate}/execution overage` : 'Dedicated support',
      ],
    }
  })

  return (
    <div className="mx-auto max-w-5xl space-y-8">
      {/* Header */}
      <div className="text-center">
        <h1 className="text-3xl font-bold tracking-tight">
          Choose the best plan for your business
        </h1>
        <p className="mt-2 text-muted-foreground">
          All plans include a 14-day free trial. No credit card required.
        </p>
      </div>

      {/* Monthly / Annual Toggle */}
      <div className="flex items-center justify-center gap-3">
        <span
          className={cn(
            'text-sm font-medium transition-colors',
            !isAnnual ? 'text-foreground' : 'text-muted-foreground',
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
            isAnnual ? 'bg-primary' : 'bg-muted-foreground/30',
          )}
        >
          <span
            className={cn(
              'pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow-sm ring-0 transition-transform',
              isAnnual ? 'translate-x-5' : 'translate-x-0',
            )}
          />
        </button>
        <span
          className={cn(
            'text-sm font-medium transition-colors',
            isAnnual ? 'text-foreground' : 'text-muted-foreground',
          )}
        >
          Annual
        </span>
        {isAnnual && (
          <Badge variant="secondary" className="text-xs">
            Save up to {maxSavings}%
          </Badge>
        )}
      </div>

      {/* Plan Cards */}
      <div className="grid gap-6 md:grid-cols-3">
        {plans.map((plan) => {
          const config = PLAN_CONFIGS[plan.id]
          const price = isAnnual ? config.annualMonthlyPrice : config.monthlyPrice
          const originalPrice = isAnnual ? config.monthlyPrice : null
          const isCurrentPlan = currentPlan === plan.id
          const isPlanLoading = checkoutPlan === plan.id && createCheckout.isPending

          return (
            <div
              key={plan.id}
              className={cn(
                'relative flex flex-col rounded-xl border p-6 transition-all',
                plan.popular && 'border-primary shadow-md',
                isCurrentPlan && 'bg-primary/5 ring-2 ring-primary',
              )}
            >
              {plan.popular && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                  <Badge className="gap-1">
                    <Zap className="h-3 w-3" />
                    Most Popular
                  </Badge>
                </div>
              )}

              <div className="mb-4">
                <h3 className="text-xl font-bold">{plan.name}</h3>
                <p className="mt-1 text-sm text-muted-foreground">{plan.description}</p>
              </div>

              <div className="mb-6">
                <div className="flex items-baseline gap-1">
                  <span className="text-4xl font-bold">${price}</span>
                  <span className="text-muted-foreground">/month</span>
                </div>
                {isAnnual && originalPrice && (
                  <p className="mt-1 text-sm text-muted-foreground">
                    <span className="line-through">${originalPrice}/mo</span>
                    <span className="ml-1 font-medium text-primary">
                      Save {getAnnualSavingsPercent(plan.id)}%
                    </span>
                  </p>
                )}
              </div>

              <ul className="mb-6 flex-1 space-y-3">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-2 text-sm">
                    <Check className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
                    <span>{feature}</span>
                  </li>
                ))}
              </ul>

              {isCurrentPlan ? (
                <Button variant="outline" disabled className="w-full">
                  Current Plan
                </Button>
              ) : (
                <Button
                  variant={plan.popular ? 'default' : 'outline'}
                  className="w-full"
                  onClick={() => handleSelectPlan(plan.id as 'start' | 'grow' | 'deliver')}
                  disabled={isPlanLoading || createCheckout.isPending}
                >
                  {isPlanLoading && <Loader2 className="h-4 w-4 animate-spin" />}
                  Choose {plan.name}
                </Button>
              )}
            </div>
          )
        })}
      </div>

      {createCheckout.isError && (
        <p className="text-center text-sm text-destructive">
          Failed to create checkout session. Please try again.
        </p>
      )}

      {/* Feature Comparison Table */}
      <div>
        <button
          onClick={() => setShowComparison(!showComparison)}
          className="mx-auto flex items-center gap-2 text-sm font-medium text-muted-foreground hover:text-foreground"
        >
          {showComparison ? 'Hide' : 'Show'} feature comparison
          {showComparison ? (
            <ChevronUp className="h-4 w-4" />
          ) : (
            <ChevronDown className="h-4 w-4" />
          )}
        </button>

        {showComparison && (
          <div className="mt-4 overflow-x-auto rounded-lg border">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="px-4 py-3 text-left font-medium">Feature</th>
                  {PAID_PLAN_IDS.map((id) => (
                    <th key={id} className="px-4 py-3 text-center font-medium">
                      {PLAN_CONFIGS[id].name}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {COMPARISON_ROWS.map((row, i) => (
                  <tr key={row.label} className={cn(i % 2 === 0 && 'bg-muted/25')}>
                    <td className="px-4 py-3 font-medium">{row.label}</td>
                    <td className="px-4 py-3 text-center">{row.start}</td>
                    <td className="px-4 py-3 text-center">{row.grow}</td>
                    <td className="px-4 py-3 text-center">{row.deliver}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* FAQ */}
      <div className="mx-auto max-w-2xl">
        <h2 className="mb-4 text-center text-xl font-bold">Frequently Asked Questions</h2>
        <Accordion type="single" collapsible className="w-full">
          <AccordionItem value="trial">
            <AccordionTrigger>What happens when my free trial ends?</AccordionTrigger>
            <AccordionContent>
              When your 14-day trial ends, you&apos;ll still be able to log in and view your data,
              but you won&apos;t be able to create new helpers or run executions until you choose a
              plan. Your data is never deleted.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="commitment">
            <AccordionTrigger>Are there any long-term commitments?</AccordionTrigger>
            <AccordionContent>
              No. All plans are month-to-month (or annual if you choose). You can cancel, upgrade,
              or downgrade at any time from the billing portal. Annual plans are billed upfront but
              can still be cancelled.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="change">
            <AccordionTrigger>Can I change my plan later?</AccordionTrigger>
            <AccordionContent>
              Absolutely. You can upgrade or downgrade at any time through the Stripe customer
              portal. When you upgrade, you&apos;ll be charged the prorated difference. When you
              downgrade, the credit is applied to your next invoice.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="payment">
            <AccordionTrigger>What payment methods do you accept?</AccordionTrigger>
            <AccordionContent>
              We accept all major credit cards (Visa, Mastercard, American Express) through our
              payment provider Stripe. We also support ACH bank transfers for annual plans.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="overages">
            <AccordionTrigger>What happens if I exceed my execution limit?</AccordionTrigger>
            <AccordionContent>
              On paid plans, you can continue running executions beyond your included limit at the
              overage rate for your plan. On the free trial, executions are hard-capped at the plan
              limit.
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>

      {/* Trust Signals */}
      <div className="flex flex-wrap items-center justify-center gap-6 pb-8 text-sm text-muted-foreground">
        <div className="flex items-center gap-2">
          <Shield className="h-4 w-4" />
          Cancel anytime
        </div>
        <div className="flex items-center gap-2">
          <Clock className="h-4 w-4" />
          No charge until trial ends
        </div>
        <div className="flex items-center gap-2">
          <Shield className="h-4 w-4" />
          Secure checkout via Stripe
        </div>
      </div>
    </div>
  )
}
