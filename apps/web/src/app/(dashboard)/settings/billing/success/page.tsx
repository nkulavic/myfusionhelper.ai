'use client'

import { useEffect, useState, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { CheckCircle2, ArrowRight, Settings, Loader2 } from 'lucide-react'
import confetti from 'canvas-confetti'
import { useBillingInfo } from '@/lib/hooks/use-settings'
import { getPlanLabel } from '@/lib/plan-constants'

export default function BillingSuccessPage() {
  const router = useRouter()
  const { data: billing, refetch } = useBillingInfo()
  const [activated, setActivated] = useState(false)
  const confettiFired = useRef(false)

  // Poll until plan activates (webhook processes)
  useEffect(() => {
    if (activated) return

    let attempts = 0
    const maxAttempts = 15

    const interval = setInterval(async () => {
      attempts++
      const result = await refetch()
      const plan = result.data?.plan

      if (plan && plan !== 'free') {
        clearInterval(interval)
        setActivated(true)
      } else if (attempts >= maxAttempts) {
        clearInterval(interval)
        // Timeout — show success anyway (webhook may still be processing)
        setActivated(true)
      }
    }, 2000)

    return () => clearInterval(interval)
  }, [activated, refetch])

  // Fire confetti once activated
  useEffect(() => {
    if (activated && !confettiFired.current) {
      confettiFired.current = true
      confetti({
        particleCount: 100,
        spread: 70,
        origin: { y: 0.6 },
      })
    }
  }, [activated])

  // Show activation spinner while waiting
  if (!activated) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center space-y-4">
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

  const planName = billing?.plan ? getPlanLabel(billing.plan) : 'your new'
  const trialEndsAt = billing?.trialEndsAt
    ? new Date(billing.trialEndsAt * 1000).toLocaleDateString()
    : null

  return (
    <div className="flex min-h-[60vh] items-center justify-center">
      <Card className="mx-4 w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mb-4 flex justify-center">
            <CheckCircle2 className="h-16 w-16 text-success" />
          </div>
          <CardTitle className="text-2xl">{planName} Plan Activated!</CardTitle>
          <CardDescription>
            Your subscription is now active
            {trialEndsAt && (
              <span className="block mt-1">
                Your 14-day free trial ends on {trialEndsAt}
              </span>
            )}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="rounded-lg bg-muted p-4">
            <h3 className="mb-2 font-semibold">What&apos;s next?</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li>- Connect your CRM platforms</li>
              <li>- Set up your first helpers</li>
              <li>- Invite team members</li>
              <li>- Explore AI-powered insights</li>
            </ul>
          </div>

          <div className="space-y-2">
            <Button
              className="w-full"
              onClick={() => router.push('/helpers')}
            >
              Get Started with Helpers
              <ArrowRight className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              className="w-full"
              onClick={() => router.push('/settings?tab=billing')}
            >
              <Settings className="h-4 w-4" />
              Back to Settings
            </Button>
          </div>

          <p className="text-center text-xs text-muted-foreground">
            You&apos;ll receive a confirmation email shortly with your subscription details
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
