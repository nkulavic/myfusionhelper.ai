'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { ExternalLink, Loader2, Receipt } from 'lucide-react'
import {
  useBillingInfo,
  useInvoices,
  useCreatePortalSession,
  useCreateCheckoutSession,
} from '@/lib/hooks/use-settings'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Progress } from '@/components/ui/progress'
import { PlanCard } from '@/components/plan-card'
import { BillingToggle } from '@/components/billing-toggle'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { toast } from 'sonner'
import {
  PAID_PLAN_IDS,
  getAnnualSavingsPercent,
  getPlanLabel,
  isTrialPlan,
  type PlanId,
} from '@/lib/plan-constants'

export function BillingTab() {
  const searchParams = useSearchParams()
  const { data: billing, isLoading: billingLoading } = useBillingInfo()
  const { data: invoices, isLoading: invoicesLoading } = useInvoices()
  const createPortal = useCreatePortalSession()
  const createCheckout = useCreateCheckoutSession()
  const [checkoutPlan, setCheckoutPlan] = useState<string | null>(null)
  const [isAnnual, setIsAnnual] = useState(false)

  // Handle checkout cancellation toast
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
        returnUrl: '/settings?tab=billing',
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

  const maxSavings = Math.max(...PAID_PLAN_IDS.map((id) => getAnnualSavingsPercent(id)))

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Current Plan</CardTitle>
          <CardDescription>Manage your subscription and billing</CardDescription>
        </CardHeader>
        <CardContent>
          {billingLoading ? (
            <Skeleton className="h-20 w-full" />
          ) : billing ? (
            <div className="flex items-center justify-between rounded-lg bg-primary/10 p-4">
              <div>
                <p className="text-lg font-bold">{getPlanLabel(billing.plan)} Plan</p>
                {billing.priceMonthly > 0 && (
                  <p className="text-sm text-muted-foreground">
                    {billing.billingPeriod === 'annual'
                      ? `$${billing.priceAnnually}/year ($${Math.round(billing.priceAnnually / 12)}/mo)`
                      : `$${billing.priceMonthly}/month`}
                  </p>
                )}
                {billing.billingPeriod === 'annual' && !isTrialPlan(billing.plan) && (
                  <Badge variant="secondary" className="mt-1 text-xs">
                    Annual billing
                  </Badge>
                )}
                {billing.renewsAt && (
                  <p className="mt-1 text-xs text-muted-foreground">
                    Renews {new Date(billing.renewsAt * 1000).toLocaleDateString()}
                  </p>
                )}
                {billing.trialEndsAt && (
                  <p className="mt-1 text-xs text-muted-foreground">
                    Trial ends {new Date(billing.trialEndsAt * 1000).toLocaleDateString()}
                  </p>
                )}
                {billing.cancelAt && (
                  <p className="mt-1 text-xs text-destructive">
                    Cancels {new Date(billing.cancelAt * 1000).toLocaleDateString()}
                  </p>
                )}
              </div>
              {!isTrialPlan(billing.plan) && (
                <Button
                  variant="outline"
                  onClick={handleManageSubscription}
                  disabled={createPortal.isPending}
                >
                  {createPortal.isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <ExternalLink className="h-4 w-4" />
                  )}
                  Manage Subscription
                </Button>
              )}
            </div>
          ) : null}
        </CardContent>
      </Card>

      {/* Plan Tiers */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-lg">
                {isOnTrial ? 'Choose a Plan' : 'Change Plan'}
              </CardTitle>
              <CardDescription>
                {isOnTrial
                  ? 'Select the plan that best fits your needs'
                  : 'Upgrade or change your current plan'}
              </CardDescription>
            </div>
            <BillingToggle
              billingPeriod={isAnnual ? 'annual' : 'monthly'}
              onChange={(p) => setIsAnnual(p === 'annual')}
              maxSavings={maxSavings}
            />
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
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
                variant="compact"
              />
            ))}
          </div>
          {createCheckout.isError && (
            <p className="mt-4 text-center text-sm text-destructive">
              Failed to create checkout session. Please try again.
            </p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Usage This Month</CardTitle>
          <CardDescription>Track your resource consumption against plan limits</CardDescription>
        </CardHeader>
        <CardContent>
          {billingLoading ? (
            <div className="space-y-6">
              {[1, 2, 3, 4].map((i) => (
                <Skeleton key={i} className="h-8 w-full" />
              ))}
            </div>
          ) : billing?.usage && billing?.limits ? (
            <div className="space-y-5">
              {[
                {
                  label: 'Helper Executions',
                  used: billing.usage.monthlyExecutions,
                  limit: billing.limits.maxExecutions,
                },
                {
                  label: 'Active Helpers',
                  used: billing.usage.helpers,
                  limit: billing.limits.maxHelpers,
                },
                {
                  label: 'Connections',
                  used: billing.usage.connections,
                  limit: billing.limits.maxConnections,
                },
                {
                  label: 'API Keys',
                  used: billing.usage.apiKeys,
                  limit: billing.limits.maxApiKeys,
                },
              ].map((item) => {
                const pct = item.limit > 0 ? Math.min((item.used / item.limit) * 100, 100) : 0
                return (
                  <div key={item.label}>
                    <div className="mb-2 flex justify-between text-sm">
                      <span className="font-medium">{item.label}</span>
                      <span className="text-muted-foreground">
                        {item.used.toLocaleString()} / {item.limit.toLocaleString()}
                      </span>
                    </div>
                    <Progress value={pct} className="h-2" />
                  </div>
                )
              })}
            </div>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Invoice History</CardTitle>
          <CardDescription>View and download past invoices</CardDescription>
        </CardHeader>
        <CardContent>
          {invoicesLoading ? (
            <div className="space-y-3">
              {[1, 2].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : invoices && invoices.length > 0 ? (
            <div className="space-y-2">
              {invoices.map((inv) => (
                <div
                  key={inv.id}
                  className="flex items-center justify-between rounded-md border p-3"
                >
                  <div className="flex items-center gap-3">
                    <Receipt className="h-4 w-4 text-muted-foreground" />
                    <div>
                      <p className="text-sm font-medium">${inv.amount.toFixed(2)}</p>
                      <p className="text-xs text-muted-foreground">
                        {new Date(inv.date * 1000).toLocaleDateString('en-US', {
                          year: 'numeric',
                          month: 'long',
                        })}
                      </p>
                    </div>
                  </div>
                  <Badge variant={inv.status === 'paid' ? 'success' : 'warning'}>
                    {inv.status}
                  </Badge>
                </div>
              ))}
            </div>
          ) : (
            <p className="py-4 text-center text-sm text-muted-foreground">No invoices yet.</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
