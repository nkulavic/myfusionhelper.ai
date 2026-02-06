'use client'

import Image from 'next/image'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useRegister } from '@/lib/hooks/use-auth'
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
  FormDescription,
} from '@/components/ui/form'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'
import { TermsContent } from '@/components/legal/terms-content'
import { PrivacyContent } from '@/components/legal/privacy-content'
import { EULAContent } from '@/components/legal/eula-content'
import { AlertCircle, Loader2, Check, X } from 'lucide-react'
import { useState, useCallback } from 'react'
import type { Control } from 'react-hook-form'
import PhoneInput, { isValidPhoneNumber } from 'react-phone-number-input'
import 'react-phone-number-input/style.css'

const registerSchema = z
  .object({
    name: z.string().min(2, 'Name must be at least 2 characters'),
    email: z.string().email('Please enter a valid email address'),
    phoneNumber: z
      .string()
      .min(1, 'Phone number is required')
      .refine((val) => isValidPhoneNumber(val), 'Please enter a valid phone number'),
    password: z
      .string()
      .min(8, 'Password must be at least 8 characters')
      .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
      .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
      .regex(/[0-9]/, 'Password must contain at least one number'),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  })

type RegisterFormValues = z.infer<typeof registerSchema>

const passwordRules = [
  { label: 'At least 8 characters', test: (v: string) => v.length >= 8 },
  { label: 'One uppercase letter', test: (v: string) => /[A-Z]/.test(v) },
  { label: 'One lowercase letter', test: (v: string) => /[a-z]/.test(v) },
  { label: 'One number', test: (v: string) => /[0-9]/.test(v) },
]

function PasswordRequirements({ control }: { control: Control<RegisterFormValues> }) {
  const password = useWatch({ control, name: 'password' })
  const confirmPassword = useWatch({ control, name: 'confirmPassword' })

  if (!password && !confirmPassword) return null

  const allChecks = [
    ...passwordRules.map((rule) => ({
      label: rule.label,
      passed: password ? rule.test(password) : false,
    })),
    {
      label: 'Passwords match',
      passed: !!password && !!confirmPassword && password === confirmPassword,
    },
  ]

  return (
    <ul className="mt-2 space-y-1">
      {allChecks.map((check) => (
        <li
          key={check.label}
          className={`flex items-center gap-1.5 text-xs transition-colors ${
            check.passed ? 'text-success' : 'text-muted-foreground'
          }`}
        >
          {check.passed ? (
            <Check className="h-3 w-3" />
          ) : (
            <X className="h-3 w-3" />
          )}
          {check.label}
        </li>
      ))}
    </ul>
  )
}

type LegalSection = 'terms' | 'privacy' | 'eula' | null

const legalTitles: Record<Exclude<LegalSection, null>, string> = {
  terms: 'Terms of Service',
  privacy: 'Privacy Policy',
  eula: 'End User License Agreement',
}

const legalComponents: Record<Exclude<LegalSection, null>, () => JSX.Element> = {
  terms: TermsContent,
  privacy: PrivacyContent,
  eula: EULAContent,
}

function LegalDialog({
  open,
  onClose,
}: {
  open: LegalSection
  onClose: () => void
}) {
  const Content = open ? legalComponents[open] : null
  return (
    <Dialog open={open !== null} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{open ? legalTitles[open] : ''}</DialogTitle>
        </DialogHeader>
        <ScrollArea className="max-h-[60vh]">
          {Content && <Content />}
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}

export default function RegisterPage() {
  const router = useRouter()
  const [error, setError] = useState('')
  const [errorKey, setErrorKey] = useState(0)
  const [legalOpen, setLegalOpen] = useState<LegalSection>(null)
  const registerMutation = useRegister()

  const showError = useCallback((message: string) => {
    setError(message)
    setErrorKey((k) => k + 1)
  }, [])

  const form = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    mode: 'onTouched',
    defaultValues: {
      name: '',
      email: '',
      phoneNumber: '',
      password: '',
      confirmPassword: '',
    },
  })

  function onSubmit({ confirmPassword: _, ...values }: RegisterFormValues) {
    setError('')
    registerMutation.mutate(values, {
      onSuccess: () => {
        router.push('/onboarding')
      },
      onError: (err) => {
        if (err instanceof APIError) {
          switch (err.statusCode) {
            case 409:
              showError('An account with this email already exists. Try signing in instead.')
              break
            case 429:
              showError('Too many attempts. Please try again later.')
              break
            default:
              showError(err.message)
          }
        } else {
          showError('Registration failed. Please try again.')
        }
      },
    })
  }

  return (
    <Card className="animate-scale-in">
      <CardHeader className="text-center">
        <Link href="/" className="mx-auto mb-2 flex items-center gap-2 font-bold">
          <Image src="/logo.png" alt="MyFusion Helper" width={180} height={23} className="dark:brightness-0 dark:invert" />
        </Link>
        <CardTitle className="text-2xl">Create an account</CardTitle>
        <CardDescription>Start your free trial today</CardDescription>
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
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input placeholder="John Doe" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

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
              name="phoneNumber"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Phone Number</FormLabel>
                  <FormControl>
                    <PhoneInput
                      international
                      defaultCountry="US"
                      placeholder="+1 (303) 555-0199"
                      value={field.value}
                      onChange={(value) => field.onChange(value || '')}
                      onBlur={field.onBlur}
                      className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-within:outline-none focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2 [&_.PhoneInputInput]:bg-transparent [&_.PhoneInputInput]:outline-none [&_.PhoneInputInput]:border-none [&_.PhoneInputInput]:text-foreground [&_.PhoneInputInput]:placeholder:text-muted-foreground"
                    />
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
                    <Input type="password" placeholder="Create a password" {...field} />
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
                    <Input type="password" placeholder="Re-enter your password" {...field} />
                  </FormControl>
                  <PasswordRequirements control={form.control} />
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button type="submit" className="w-full" disabled={registerMutation.isPending}>
              {registerMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {registerMutation.isPending ? 'Creating account...' : 'Create account'}
            </Button>
          </form>
        </Form>

        <p className="mt-4 text-center text-xs text-muted-foreground">
          By creating an account, you agree to our{' '}
          <button
            type="button"
            onClick={() => setLegalOpen('terms')}
            className="underline hover:text-foreground"
          >
            Terms
          </button>
          ,{' '}
          <button
            type="button"
            onClick={() => setLegalOpen('privacy')}
            className="underline hover:text-foreground"
          >
            Privacy Policy
          </button>{' '}
          &{' '}
          <button
            type="button"
            onClick={() => setLegalOpen('eula')}
            className="underline hover:text-foreground"
          >
            EULA
          </button>
        </p>

        <LegalDialog open={legalOpen} onClose={() => setLegalOpen(null)} />

        <div className="mt-4 text-center text-sm">
          Already have an account?{' '}
          <Link href="/login" className="font-medium text-primary hover:underline">
            Sign in
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
