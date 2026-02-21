'use client'

import Image from 'next/image'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useLogin } from '@/lib/hooks/use-auth'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import { useAuthStore } from '@/lib/stores/auth-store'
import { authApi, type MfaChallengeResponse } from '@/lib/api/auth'
import { APIError } from '@/lib/api/client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { AlertCircle, Loader2, ShieldCheck } from 'lucide-react'
import { useState, useCallback } from 'react'

const loginSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
  password: z.string().min(1, 'Password is required'),
})

type LoginFormValues = z.infer<typeof loginSchema>

export default function LoginPage() {
  const router = useRouter()
  const { onboardingComplete } = useWorkspaceStore()
  const { setUser } = useAuthStore()
  const { setAccount } = useWorkspaceStore()
  const [error, setError] = useState('')
  const [errorKey, setErrorKey] = useState(0)
  const loginMutation = useLogin()

  // MFA challenge state
  const [mfaChallenge, setMfaChallenge] = useState<MfaChallengeResponse | null>(null)
  const [mfaCode, setMfaCode] = useState('')
  const [mfaSubmitting, setMfaSubmitting] = useState(false)

  const showError = useCallback((message: string) => {
    setError(message)
    setErrorKey((k) => k + 1)
  }, [])

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    mode: 'onTouched',
    defaultValues: {
      email: '',
      password: '',
    },
  })

  function onSubmit(values: LoginFormValues) {
    setError('')
    loginMutation.mutate(values, {
      onSuccess: (res) => {
        if (res.data && 'mfaRequired' in res.data && res.data.mfaRequired) {
          setMfaChallenge(res.data as MfaChallengeResponse)
          return
        }
        if (res.data && 'user' in res.data && res.data.user?.emailVerified === false) {
          router.push(`/verify-email?email=${encodeURIComponent(values.email)}`)
          return
        }
        router.push(onboardingComplete ? '/' : '/onboarding/plan')
      },
      onError: (err) => {
        if (err instanceof APIError) {
          switch (err.statusCode) {
            case 401:
              showError('Invalid email or password')
              break
            case 429:
              showError('Too many attempts. Please try again later.')
              break
            default:
              showError(err.message)
          }
        } else {
          showError('An error occurred. Please try again.')
        }
      },
    })
  }

  async function onMfaSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!mfaChallenge || !mfaCode.trim()) return

    setError('')
    setMfaSubmitting(true)
    try {
      const res = await authApi.submitMfaChallenge({
        session: mfaChallenge.session,
        code: mfaCode.trim(),
        challengeName: mfaChallenge.challengeName,
        username: mfaChallenge.username,
      })
      if (res.data) {
        setUser(res.data.user, res.data.token, res.data.refreshToken)
        setAccount(res.data.account)
        if (res.data.user?.emailVerified === false) {
          router.push(`/verify-email?email=${encodeURIComponent(res.data.user.email)}`)
        } else {
          router.push(onboardingComplete ? '/' : '/onboarding/plan')
        }
      }
    } catch (err) {
      if (err instanceof APIError) {
        if (err.statusCode === 401) {
          showError(err.message || 'Invalid verification code')
        } else {
          showError(err.message)
        }
      } else {
        showError('Verification failed. Please try again.')
      }
    } finally {
      setMfaSubmitting(false)
    }
  }

  // MFA code input screen
  if (mfaChallenge) {
    const isSms = mfaChallenge.challengeName === 'SMS_MFA'
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
            <ShieldCheck className="h-6 w-6 text-primary" />
          </div>
          <CardTitle className="text-2xl">Two-factor authentication</CardTitle>
          <CardDescription>
            {isSms
              ? 'Enter the code sent to your phone'
              : 'Enter the code from your authenticator app'}
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

          <form onSubmit={onMfaSubmit} className="space-y-4">
            <div className="space-y-2">
              <label htmlFor="mfa-code" className="text-sm font-medium">
                Verification code
              </label>
              <Input
                id="mfa-code"
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                placeholder="000000"
                maxLength={6}
                value={mfaCode}
                onChange={(e) => setMfaCode(e.target.value.replace(/\D/g, ''))}
                className="text-center text-lg tracking-widest"
                autoFocus
              />
            </div>

            <Button
              type="submit"
              className="w-full"
              disabled={mfaSubmitting || mfaCode.length < 6}
            >
              {mfaSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {mfaSubmitting ? 'Verifying...' : 'Verify'}
            </Button>
          </form>

          <div className="mt-4 text-center">
            <button
              type="button"
              onClick={() => {
                setMfaChallenge(null)
                setMfaCode('')
                setError('')
              }}
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Back to sign in
            </button>
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
        <CardTitle className="text-2xl">Welcome back</CardTitle>
        <CardDescription>Enter your credentials to sign in</CardDescription>
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

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <Input type="email" placeholder="name@example.com" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Password</FormLabel>
                  <FormControl>
                    <Input type="password" {...field} />
                  </FormControl>
                  <FormMessage />
                  <div className="flex justify-end">
                    <Link
                      href="/forgot-password"
                      className="text-sm text-muted-foreground hover:text-foreground"
                    >
                      Forgot password?
                    </Link>
                  </div>
                </FormItem>
              )}
            />

            <Button type="submit" className="w-full" disabled={loginMutation.isPending}>
              {loginMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {loginMutation.isPending ? 'Signing in...' : 'Sign in'}
            </Button>
          </form>
        </Form>

        <div className="mt-6 text-center text-sm">
          Don&apos;t have an account?{' '}
          <Link href="/register" className="font-medium text-primary hover:underline">
            Sign up
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
