'use client'

import { Suspense, useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { Shield, Clock, ChevronDown, ChevronUp } from 'lucide-react'
import { cn } from '@/lib/utils'

import {
  useBillingInfo,
  useCreatePortalSession,
  useCreateCheckoutSession,
} from '@/lib/hooks/use-settings'
import { Button } from '@/components/ui/button'
import { PlanCard } from '@/components/plan-card'
import { BillingToggle } from '@/components/billing-toggle'
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
  getAnnualSavingsPercent,
  isTrialPlan,
} from '@/lib/plan-constants'

export default function PlansPage() {
  return (
    <Suspense fallback={null}>
      <PlansContent />
    </Suspense>
  )
}

function PlansContent() {
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
      <div className="flex justify-center">
        <BillingToggle
          billingPeriod={isAnnual ? 'annual' : 'monthly'}
          onChange={(p) => setIsAnnual(p === 'annual')}
          maxSavings={maxSavings}
        />
      </div>

      {/* Plan Cards */}
      <div className="grid gap-6 md:grid-cols-3">
        {PAID_PLAN_IDS.map((planId) => (
          <PlanCard
            key={planId}
            planId={planId}
            billingPeriod={isAnnual ? 'annual' : 'monthly'}
            currentPlan={currentPlan}
            isPopular={planId === 'grow'}
            onSelect={(id) => handleSelectPlan(id as 'start' | 'grow' | 'deliver')}
            isLoading={checkoutPlan === planId && createCheckout.isPending}
            isManaging={createPortal.isPending}
            variant="full"
          />
        ))}
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
