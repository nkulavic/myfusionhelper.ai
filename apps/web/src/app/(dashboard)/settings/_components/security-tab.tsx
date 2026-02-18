'use client'

import { useState } from 'react'
import {
  Shield,
  Smartphone,
  Loader2,
  CheckCircle,
  XCircle,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useMfaStatus, useSetupTotp, useVerifyTotp, useDisableMfa } from '@/lib/hooks/use-mfa'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { ChangePasswordCard } from './change-password-card'

type SetupStep = 'idle' | 'qr' | 'verify' | 'success'

export function SecurityTab() {
  const { data: mfaStatus, isLoading: statusLoading } = useMfaStatus()
  const setupTotp = useSetupTotp()
  const verifyTotp = useVerifyTotp()
  const disableMfa = useDisableMfa()

  const [totpStep, setTotpStep] = useState<SetupStep>('idle')
  const [totpSecret, setTotpSecret] = useState('')
  const [totpUri, setTotpUri] = useState('')
  const [verifyCode, setVerifyCode] = useState('')
  const [showDisableConfirm, setShowDisableConfirm] = useState(false)

  const mfaEnabled = mfaStatus?.enabled ?? false
  const mfaMethod = mfaStatus?.method ?? null

  const handleStartTotp = () => {
    setupTotp.mutate(undefined, {
      onSuccess: (res) => {
        if (res.data) {
          setTotpSecret(res.data.secret)
          setTotpUri(res.data.qrCodeUri)
          setTotpStep('qr')
        }
      },
    })
  }

  const handleVerifyTotp = () => {
    if (verifyCode.length !== 6) return
    verifyTotp.mutate(
      { code: verifyCode },
      {
        onSuccess: () => {
          setTotpStep('success')
          setVerifyCode('')
        },
      }
    )
  }

  const handleDisable = () => {
    disableMfa.mutate(undefined, {
      onSuccess: () => {
        setShowDisableConfirm(false)
        setTotpStep('idle')
        setTotpSecret('')
        setTotpUri('')
      },
    })
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Shield className="h-5 w-5" />
            Two-Factor Authentication (2FA)
          </CardTitle>
          <CardDescription>
            Add an extra layer of security to your account by requiring a second form of
            verification when signing in.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Status Badge */}
          {statusLoading ? (
            <Skeleton className="h-16 w-full" />
          ) : (
            <div className="flex items-center justify-between rounded-lg border p-4">
              <div className="flex items-center gap-3">
                <div
                  className={cn(
                    'flex h-10 w-10 items-center justify-center rounded-full',
                    mfaEnabled ? 'bg-success/10' : 'bg-muted'
                  )}
                >
                  <Shield
                    className={cn(
                      'h-5 w-5',
                      mfaEnabled ? 'text-success' : 'text-muted-foreground'
                    )}
                  />
                </div>
                <div>
                  <p className="text-sm font-medium">
                    {mfaEnabled ? '2FA is enabled' : '2FA is not enabled'}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {mfaEnabled
                      ? 'Protected with authenticator app'
                      : 'Enable 2FA to add an extra layer of security'}
                  </p>
                </div>
              </div>
              <Badge variant={mfaEnabled ? 'default' : 'secondary'}>
                {mfaEnabled ? 'Active' : 'Inactive'}
              </Badge>
            </div>
          )}

          {/* Disable Flow */}
          {mfaEnabled && !statusLoading && (
            <>
              <Separator />
              {showDisableConfirm ? (
                <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-4 space-y-3">
                  <p className="text-sm font-medium text-destructive">
                    Are you sure you want to disable 2FA?
                  </p>
                  <p className="text-xs text-muted-foreground">
                    This will remove the extra security layer from your account.
                  </p>
                  <div className="flex gap-2">
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={handleDisable}
                      disabled={disableMfa.isPending}
                    >
                      {disableMfa.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                      Yes, Disable 2FA
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setShowDisableConfirm(false)}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              ) : (
                <Button
                  variant="outline"
                  className="w-full"
                  onClick={() => setShowDisableConfirm(true)}
                >
                  Disable 2FA
                </Button>
              )}
            </>
          )}

          {/* Setup Flow (only when not enabled) */}
          {!mfaEnabled && !statusLoading && (
            <>
              <Separator />

              <div className="space-y-4">
                {/* TOTP Setup */}
                <div className="rounded-lg border p-4 space-y-4">
                  <div className="flex items-center gap-2">
                    <Smartphone className="h-4 w-4" />
                    <span className="text-sm font-medium">Authenticator App</span>
                    <Badge variant="outline" className="text-xs">
                      Recommended
                    </Badge>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Use an authenticator app like Google Authenticator, Authy, or 1Password to
                    generate verification codes.
                  </p>

                  {totpStep === 'idle' && (
                    <Button onClick={handleStartTotp} disabled={setupTotp.isPending}>
                      {setupTotp.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                      Set Up Authenticator
                    </Button>
                  )}

                  {totpStep === 'qr' && (
                    <div className="space-y-4 rounded-md border bg-muted/30 p-4">
                      <p className="text-sm font-medium">
                        Scan this QR code with your authenticator app:
                      </p>
                      {/* QR Code rendered as an image via Google Charts API */}
                      <div className="flex justify-center">
                        <img
                          src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(totpUri)}`}
                          alt="TOTP QR Code"
                          className="h-48 w-48 rounded-md border bg-white p-2"
                        />
                      </div>
                      <div className="space-y-1">
                        <p className="text-xs text-muted-foreground">
                          Or enter this secret manually:
                        </p>
                        <code className="block rounded bg-background p-2 text-center font-mono text-sm select-all">
                          {totpSecret}
                        </code>
                      </div>
                      <Separator />
                      <div className="space-y-2">
                        <Label htmlFor="totp-verify">Enter the 6-digit code from your app:</Label>
                        <div className="flex gap-2">
                          <Input
                            id="totp-verify"
                            value={verifyCode}
                            onChange={(e) =>
                              setVerifyCode(e.target.value.replace(/\D/g, '').slice(0, 6))
                            }
                            placeholder="000000"
                            className="w-32 text-center font-mono text-lg tracking-widest"
                            maxLength={6}
                          />
                          <Button
                            onClick={handleVerifyTotp}
                            disabled={verifyCode.length !== 6 || verifyTotp.isPending}
                          >
                            {verifyTotp.isPending && (
                              <Loader2 className="h-4 w-4 animate-spin" />
                            )}
                            Verify
                          </Button>
                        </div>
                        {verifyTotp.isError && (
                          <p className="flex items-center gap-1 text-xs text-destructive">
                            <XCircle className="h-3 w-3" />
                            {(verifyTotp.error as Error)?.message || 'Invalid code. Try again.'}
                          </p>
                        )}
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          setTotpStep('idle')
                          setTotpSecret('')
                          setTotpUri('')
                          setVerifyCode('')
                        }}
                      >
                        Cancel
                      </Button>
                    </div>
                  )}

                  {totpStep === 'success' && (
                    <div className="flex items-center gap-2 rounded-md border border-success/30 bg-success/10 p-3 text-sm text-success">
                      <CheckCircle className="h-4 w-4 shrink-0" />
                      Two-factor authentication enabled successfully!
                    </div>
                  )}

                  {setupTotp.isError && totpStep === 'idle' && (
                    <p className="flex items-center gap-1 text-xs text-destructive">
                      <XCircle className="h-3 w-3" />
                      {(setupTotp.error as Error)?.message || 'Failed to start setup'}
                    </p>
                  )}
                </div>

              </div>
            </>
          )}
        </CardContent>
      </Card>

      <ChangePasswordCard />

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Active Sessions</CardTitle>
          <CardDescription>View and manage your active login sessions.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4 rounded-lg border p-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-success/10">
              <CheckCircle className="h-5 w-5 text-success" />
            </div>
            <div className="flex-1">
              <p className="text-sm font-medium">Current Session</p>
              <p className="text-xs text-muted-foreground">This device - Active now</p>
            </div>
            <Badge variant="outline">Current</Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
