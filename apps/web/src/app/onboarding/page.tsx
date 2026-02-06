'use client'

import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { AnimatePresence, motion } from 'framer-motion'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import { useAuthStore } from '@/lib/stores/auth-store'
import { WelcomeStep } from './_components/welcome-step'
import { ConnectCRMStep } from './_components/connect-crm-step'
import { PickHelperStep } from './_components/pick-helper-step'
import { QuickTourStep } from './_components/quick-tour-step'

const STEPS = ['welcome', 'connect', 'helper', 'tour'] as const
type Step = (typeof STEPS)[number]

const stepLabels: Record<Step, string> = {
  welcome: 'Welcome',
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
  const router = useRouter()
  const { user } = useAuthStore()
  const { completeOnboarding, onboardingComplete } = useWorkspaceStore()
  const [currentStep, setCurrentStep] = useState<Step>('welcome')
  const [direction, setDirection] = useState(0)

  // If onboarding is already complete, redirect to helpers
  if (onboardingComplete) {
    router.replace('/helpers')
    return null
  }

  const currentIndex = STEPS.indexOf(currentStep)

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
    router.push('/helpers')
  }, [completeOnboarding, router])

  const skip = useCallback(() => {
    completeOnboarding()
    router.push('/helpers')
  }, [completeOnboarding, router])

  const firstName = user?.name?.split(' ')[0] || 'there'

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
            {currentStep === 'welcome' && (
              <WelcomeStep firstName={firstName} onNext={next} onSkip={skip} />
            )}
            {currentStep === 'connect' && (
              <ConnectCRMStep onNext={next} onBack={back} onSkip={skip} />
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
