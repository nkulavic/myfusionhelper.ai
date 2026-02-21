'use client'

import Image from 'next/image'
import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { useVerifyEmail, useResendVerification } from '@/lib/hooks/use-auth'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import { APIError } from '@/lib/api/client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertCircle, CheckCircle, Loader2, Mail } from 'lucide-react'
import { useState, useCallback, useEffect, Suspense } from 'react'

function VerifyEmailForm() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const emailFromQuery = searchParams.get('email') || ''
  const codeFromQuery = searchParams.get('code') || ''
  const { user } = useAuthStore()
  const { onboardingComplete } = useWorkspaceStore()

  const email = emailFromQuery || user?.email || ''

  const [code, setCode] = useState(codeFromQuery)
  const [error, setError] = useState('')
  const [errorKey, setErrorKey] = useState(0)
  const [success, setSuccess] = useState(false)
  const [resendCooldown, setResendCooldown] = useState(0)
  const [autoSubmitted, setAutoSubmitted] = useState(false)

  const verifyMutation = useVerifyEmail()
  const resendMutation = useResendVerification()

  const showError = useCallback((message: string) => {
    setError(message)
    setErrorKey((k) => k + 1)
  }, [])

  // Auto-submit if both email and code are provided via query params
  useEffect(() => {
    if (autoSubmitted) return
    if (email && codeFromQuery && codeFromQuery.length === 6) {
      setAutoSubmitted(true)
      verifyMutation.mutate(
        { email, code: codeFromQuery },
        {
          onSuccess: () => {
            setSuccess(true)
            setTimeout(() => {
              router.push(onboardingComplete ? '/' : '/onboarding/plan')
            }, 1500)
          },
          onError: (err) => {
            if (err instanceof APIError) {
              showError(err.message)
            } else {
              showError('Verification failed. Please try again.')
            }
          },
        },
      )
    }
  }, [email, codeFromQuery, autoSubmitted, verifyMutation, router, onboardingComplete, showError])

  // Cooldown timer for resend
  useEffect(() => {
    if (resendCooldown <= 0) return
    const timer = setTimeout(() => setResendCooldown((c) => c - 1), 1000)
    return () => clearTimeout(timer)
  }, [resendCooldown])

  function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!code.trim() || !email) return

    setError('')
    verifyMutation.mutate(
      { email, code: code.trim() },
      {
        onSuccess: () => {
          setSuccess(true)
          setTimeout(() => {
            router.push(onboardingComplete ? '/' : '/onboarding/plan')
          }, 1500)
        },
        onError: (err) => {
          if (err instanceof APIError) {
            showError(err.message)
          } else {
            showError('Verification failed. Please try again.')
          }
        },
      },
    )
  }

  function handleResend() {
    if (resendCooldown > 0 || !email) return

    resendMutation.mutate(
      { email },
      {
        onSuccess: () => {
          setResendCooldown(60)
        },
        onError: () => {
          showError('Failed to resend code. Please try again.')
        },
      },
    )
  }

  if (success) {
    return (
      <Card className="animate-scale-in">
        <CardHeader className="text-center">
          <Link href="/" className="mx-auto mb-2 flex items-center gap-2 font-bold">
            <Image
              src="/logo.png"
              alt="MyFusion Helper"
              width={180}
              height={23}
              className="dark:brightness-0 dark:invert"
            />
          </Link>
          <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-success/10">
            <CheckCircle className="h-6 w-6 text-success" />
          </div>
          <CardTitle className="text-2xl">Email verified</CardTitle>
          <CardDescription>
            Your email has been verified. Redirecting you now...
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex justify-center">
            <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="animate-scale-in">
      <CardHeader className="text-center">
        <Link href="/" className="mx-auto mb-2 flex items-center gap-2 font-bold">
          <Image
            src="/logo.png"
            alt="MyFusion Helper"
            width={180}
            height={23}
            className="dark:brightness-0 dark:invert"
          />
        </Link>
        <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
          <Mail className="h-6 w-6 text-primary" />
        </div>
        <CardTitle className="text-2xl">Verify your email</CardTitle>
        <CardDescription>
          We sent a 6-digit code to <span className="font-medium text-foreground">{email}</span>.
          Enter it below to verify your email address.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {error && (
          <Alert
            key={errorKey}
            variant="destructive"
            className="mb-4 animate-shake border-destructive/30 bg-destructive/10"
          >
            <AlertCircle className="h-4 w-4 text-destructive" />
            <AlertDescription className="animate-fade-in font-medium text-destructive">
              {error}
            </AlertDescription>
          </Alert>
        )}

        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="verify-code" className="text-sm font-medium">
              Verification code
            </label>
            <Input
              id="verify-code"
              type="text"
              inputMode="numeric"
              autoComplete="one-time-code"
              placeholder="000000"
              maxLength={6}
              value={code}
              onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
              className="text-center text-lg tracking-widest"
              autoFocus
            />
          </div>

          <Button
            type="submit"
            className="w-full"
            disabled={verifyMutation.isPending || code.length < 6}
          >
            {verifyMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {verifyMutation.isPending ? 'Verifying...' : 'Verify Email'}
          </Button>
        </form>

        <div className="mt-4 text-center">
          <button
            type="button"
            onClick={handleResend}
            disabled={resendCooldown > 0 || resendMutation.isPending}
            className="text-sm text-muted-foreground hover:text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {resendMutation.isPending
              ? 'Sending...'
              : resendCooldown > 0
                ? `Resend code in ${resendCooldown}s`
                : "Didn't receive a code? Resend"}
          </button>
        </div>

        <div className="mt-4 text-center">
          <Link
            href="/login"
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            Back to sign in
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}

export default function VerifyEmailPage() {
  return (
    <Suspense>
      <VerifyEmailForm />
    </Suspense>
  )
}
