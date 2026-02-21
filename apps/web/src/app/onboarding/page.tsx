'use client'

import { Suspense, useState, useCallback, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { AnimatePresence, motion } from 'framer-motion'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useCompleteOnboarding } from '@/lib/hooks/use-auth'
import { useBillingInfo } from '@/lib/hooks/use-settings'
import { isTrialPlan } from '@/lib/plan-constants'
import { ConnectCRMStep } from './_components/connect-crm-step'
import { PickHelperStep } from './_components/pick-helper-step'
import { QuickTourStep } from './_components/quick-tour-step'

const STEPS = ['connect', 'helper', 'tour'] as const
type Step = (typeof STEPS)[number]

const stepLabels: Record<Step, string> = {
  connect: 'Connect CRM',
  helper: 'First Helper',
  tour: 'Quick Tour',
}

const slideVariants = {
  enter: (direction: number) => ({
    x: direction > 0 ? 80 : -80,
    opacity: 0,
  }),
  center: {
    x: 0,
    opacity: 1,
  },
  exit: (direction: number) => ({
    x: direction > 0 ? -80 : 80,
    opacity: 0,
  }),
}

const transition = {
  duration: 0.25,
  ease: [0.25, 0.46, 0.45, 0.94] as [number, number, number, number],
}

export default function OnboardingPage() {
  return (
    <Suspense fallback={null}>
      <OnboardingContent />
    </Suspense>
  )
}

function OnboardingContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { user } = useAuthStore()
  const { completeOnboarding, onboardingComplete, _hasHydrated } = useWorkspaceStore()
  const completeOnboardingMutation = useCompleteOnboarding()
  const { data: billing, isLoading: billingLoading } = useBillingInfo()

  // Resume at correct step when returning
  const stepParam = searchParams.get('step') as Step | null
  const initialStep: Step =
    stepParam && STEPS.includes(stepParam as Step) ? (stepParam as Step) : 'connect'

  const [currentStep, setCurrentStep] = useState<Step>(initialStep)
  const [direction, setDirection] = useState(0)
  const [mounted, setMounted] = useState(false)

  const currentIndex = STEPS.indexOf(currentStep)

  useEffect(() => { setMounted(true) }, [])

  // If onboarding is already complete, redirect to dashboard
  useEffect(() => {
    if (mounted && _hasHydrated && (onboardingComplete || user?.onboardingComplete)) {
      router.replace('/')
    }
  }, [mounted, _hasHydrated, onboardingComplete, user?.onboardingComplete, router])

  // If user hasn't selected a paid plan yet, redirect to plan selection
  useEffect(() => {
    if (mounted && !billingLoading && billing && isTrialPlan(billing.plan)) {
      router.replace('/onboarding/plan')
    }
  }, [mounted, billingLoading, billing, router])

  const goToStep = useCallback(
    (step: Step) => {
      const newIndex = STEPS.indexOf(step)
      setDirection(newIndex > currentIndex ? 1 : -1)
      setCurrentStep(step)
    },
    [currentIndex]
  )

  const next = useCallback(() => {
    if (currentIndex < STEPS.length - 1) {
      goToStep(STEPS[currentIndex + 1])
    }
  }, [currentIndex, goToStep])

  const back = useCallback(() => {
    if (currentIndex > 0) {
      goToStep(STEPS[currentIndex - 1])
    }
  }, [currentIndex, goToStep])

  const finish = useCallback(() => {
    completeOnboarding()
    completeOnboardingMutation.mutate()
    router.push('/')
  }, [completeOnboarding, completeOnboardingMutation, router])

  const skip = useCallback(() => {
    completeOnboarding()
    completeOnboardingMutation.mutate()
    router.push('/')
  }, [completeOnboarding, completeOnboardingMutation, router])

  if (!mounted) {
    return null
  }

  return (
    <div className="space-y-8">
      {/* Progress indicator */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {STEPS.map((step, i) => (
            <div key={step} className="flex items-center gap-2">
              <button
                onClick={() => i <= currentIndex && goToStep(step)}
                disabled={i > currentIndex}
                className="flex items-center gap-2"
              >
                <div
                  className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-medium transition-colors ${
                    i < currentIndex
                      ? 'bg-primary text-primary-foreground'
                      : i === currentIndex
                        ? 'bg-primary text-primary-foreground ring-2 ring-primary ring-offset-2 ring-offset-background'
                        : 'bg-muted text-muted-foreground'
                  }`}
                >
                  {i < currentIndex ? (
                    <svg
                      className="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={3}
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                  ) : (
                    i + 1
                  )}
                </div>
                <span
                  className={`hidden text-sm font-medium sm:inline ${
                    i <= currentIndex ? 'text-foreground' : 'text-muted-foreground'
                  }`}
                >
                  {stepLabels[step]}
                </span>
              </button>
              {i < STEPS.length - 1 && (
                <div
                  className={`h-px w-8 transition-colors ${
                    i < currentIndex ? 'bg-primary' : 'bg-border'
                  }`}
                />
              )}
            </div>
          ))}
        </div>
        <button
          onClick={skip}
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          Skip setup
        </button>
      </div>

      {/* Step content */}
      <div className="relative min-h-[420px]">
        <AnimatePresence mode="wait" custom={direction}>
          <motion.div
            key={currentStep}
            custom={direction}
            variants={slideVariants}
            initial="enter"
            animate="center"
            exit="exit"
            transition={transition}
          >
            {currentStep === 'connect' && (
              <ConnectCRMStep onNext={next} onBack={() => router.push('/onboarding/plan')} onSkip={skip} />
            )}
            {currentStep === 'helper' && (
              <PickHelperStep onNext={next} onBack={back} onSkip={skip} />
            )}
            {currentStep === 'tour' && (
              <QuickTourStep onFinish={finish} onBack={back} />
            )}
          </motion.div>
        </AnimatePresence>
      </div>
    </div>
  )
}
