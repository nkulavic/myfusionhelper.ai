'use client'

import Image from 'next/image'
import Link from 'next/link'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useForgotPassword } from '@/lib/hooks/use-auth'
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
import { AlertCircle, ArrowLeft, CheckCircle, Loader2, Mail } from 'lucide-react'
import { useState, useCallback } from 'react'

const forgotPasswordSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
})

type ForgotPasswordFormValues = z.infer<typeof forgotPasswordSchema>

export default function ForgotPasswordPage() {
  const [error, setError] = useState('')
  const [errorKey, setErrorKey] = useState(0)
  const [submitted, setSubmitted] = useState(false)
  const [submittedEmail, setSubmittedEmail] = useState('')
  const forgotPasswordMutation = useForgotPassword()

  const showError = useCallback((message: string) => {
    setError(message)
    setErrorKey((k) => k + 1)
  }, [])

  const form = useForm<ForgotPasswordFormValues>({
    resolver: zodResolver(forgotPasswordSchema),
    mode: 'onTouched',
    defaultValues: {
      email: '',
    },
  })

  function onSubmit(values: ForgotPasswordFormValues) {
    setError('')
    forgotPasswordMutation.mutate(values, {
      onSuccess: () => {
        setSubmittedEmail(values.email)
        setSubmitted(true)
      },
      onError: (err) => {
        if (err instanceof APIError) {
          switch (err.statusCode) {
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

  if (submitted) {
    return (
      <Card>
        <CardHeader className="text-center">
          <Link href="/" className="mx-auto mb-2 flex items-center gap-2 font-bold">
            <Image
              src="/logo.png"
              alt="MyFusion Helper"
              width={180}
              height={23}
              className="dark:hidden"
            />
            <Image
              src="/logo-full.png"
              alt="MyFusion Helper"
              width={180}
              height={23}
              className="hidden dark:block"
            />
          </Link>
          <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-success/10">
            <CheckCircle className="h-6 w-6 text-success" />
          </div>
          <CardTitle className="text-2xl">Check your email</CardTitle>
          <CardDescription>
            We sent a password reset code to{' '}
            <span className="font-medium text-foreground">{submittedEmail}</span>
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-center text-sm text-muted-foreground">
            Enter the code on the next page to reset your password. The code expires in 10 minutes.
          </p>
          <Button asChild className="w-full">
            <Link href={`/reset-password?email=${encodeURIComponent(submittedEmail)}`}>
              <Mail className="h-4 w-4" />
              Enter Reset Code
            </Link>
          </Button>
          <div className="text-center">
            <button
              onClick={() => {
                setSubmitted(false)
                form.reset()
              }}
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Didn&apos;t receive the email? Try again
            </button>
          </div>
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
            className="dark:hidden"
          />
          <Image
            src="/logo-full.png"
            alt="MyFusion Helper"
            width={180}
            height={23}
            className="hidden dark:block"
          />
        </Link>
        <CardTitle className="text-2xl">Reset your password</CardTitle>
        <CardDescription>
          Enter your email and we&apos;ll send you a code to reset your password
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
                    <Input type="email" placeholder="name@example.com" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button
              type="submit"
              className="w-full"
              disabled={forgotPasswordMutation.isPending}
            >
              {forgotPasswordMutation.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              {forgotPasswordMutation.isPending ? 'Sending...' : 'Send Reset Code'}
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
