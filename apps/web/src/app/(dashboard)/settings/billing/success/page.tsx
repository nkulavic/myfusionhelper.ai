'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { CheckCircle2, ArrowRight, Settings } from 'lucide-react'
import confetti from 'canvas-confetti'

export default function BillingSuccessPage() {
  const router = useRouter()
  const queryClient = useQueryClient()

  useEffect(() => {
    confetti({
      particleCount: 100,
      spread: 70,
      origin: { y: 0.6 },
    })

    queryClient.invalidateQueries({ queryKey: ['billing'] })
  }, [queryClient])

  return (
    <div className="flex min-h-[60vh] items-center justify-center">
      <Card className="mx-4 w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mb-4 flex justify-center">
            <CheckCircle2 className="h-16 w-16 text-success" />
          </div>
          <CardTitle className="text-2xl">Subscription Activated!</CardTitle>
          <CardDescription>
            Your plan has been upgraded successfully
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
              onClick={() => router.push('/settings')}
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
