'use client'

import Image from 'next/image'
import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useResetPassword } from '@/lib/hooks/use-auth'
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
import { AlertCircle, ArrowLeft, CheckCircle, Loader2 } from 'lucide-react'
import { useState, useCallback, Suspense } from 'react'

const resetPasswordSchema = z
  .object({
    email: z.string().email('Please enter a valid email address'),
    code: z.string().min(1, 'Verification code is required'),
    newPassword: z
      .string()
      .min(8, 'Password must be at least 8 characters')
      .regex(/[A-Z]/, 'Password must contain an uppercase letter')
      .regex(/[a-z]/, 'Password must contain a lowercase letter')
      .regex(/[0-9]/, 'Password must contain a number'),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  })

type ResetPasswordFormValues = z.infer<typeof resetPasswordSchema>

function ResetPasswordForm() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const emailFromQuery = searchParams.get('email') || ''

  const [error, setError] = useState('')
  const [errorKey, setErrorKey] = useState(0)
  const [success, setSuccess] = useState(false)
  const resetPasswordMutation = useResetPassword()

  const showError = useCallback((message: string) => {
    setError(message)
    setErrorKey((k) => k + 1)
  }, [])

  const form = useForm<ResetPasswordFormValues>({
    resolver: zodResolver(resetPasswordSchema),
    mode: 'onTouched',
    defaultValues: {
      email: emailFromQuery,
      code: '',
      newPassword: '',
      confirmPassword: '',
    },
  })

  function onSubmit({ confirmPassword: _, ...values }: ResetPasswordFormValues) {
    setError('')
    resetPasswordMutation.mutate(values, {
      onSuccess: () => {
        setSuccess(true)
      },
      onError: (err) => {
        if (err instanceof APIError) {
          switch (err.statusCode) {
            case 400:
              showError(err.message)
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

  if (success) {
    return (
      <Card>
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
          <CardTitle className="text-2xl">Password reset</CardTitle>
          <CardDescription>
            Your password has been successfully reset. You can now sign in with your new password.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button className="w-full" onClick={() => router.push('/login')}>
            Sign in
          </Button>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
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
        <CardTitle className="text-2xl">Set new password</CardTitle>
        <CardDescription>
          Enter the code from your email and choose a new password
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

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <Input
                      type="email"
                      placeholder="name@example.com"
                      {...field}
                      readOnly={!!emailFromQuery}
                      className={emailFromQuery ? 'bg-muted text-muted-foreground' : ''}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="code"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Verification Code</FormLabel>
                  <FormControl>
                    <Input placeholder="Enter the 6-digit code" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="newPassword"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>New Password</FormLabel>
                  <FormControl>
                    <Input type="password" placeholder="Enter new password" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="confirmPassword"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Confirm Password</FormLabel>
                  <FormControl>
                    <Input type="password" placeholder="Confirm new password" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button
              type="submit"
              className="w-full"
              disabled={resetPasswordMutation.isPending}
            >
              {resetPasswordMutation.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              {resetPasswordMutation.isPending ? 'Resetting...' : 'Reset Password'}
            </Button>
          </form>
        </Form>

        <div className="mt-6 text-center">
          <Link
            href="/login"
            className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="h-3 w-3" />
            Back to sign in
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}

export default function ResetPasswordPage() {
  return (
    <Suspense>
      <ResetPasswordForm />
    </Suspense>
  )
}
