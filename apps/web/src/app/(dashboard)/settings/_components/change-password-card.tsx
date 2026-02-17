'use client'

import { useState } from 'react'
import { Eye, EyeOff, XCircle, Check, CheckCircle, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { useUpdatePassword } from '@/lib/hooks/use-settings'

export function ChangePasswordCard() {
  const updatePassword = useUpdatePassword()
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showCurrent, setShowCurrent] = useState(false)
  const [showNew, setShowNew] = useState(false)
  const [success, setSuccess] = useState(false)

  const passwordErrors: string[] = []
  if (newPassword && newPassword.length < 8) passwordErrors.push('At least 8 characters')
  if (newPassword && !/[A-Z]/.test(newPassword)) passwordErrors.push('One uppercase letter')
  if (newPassword && !/[a-z]/.test(newPassword)) passwordErrors.push('One lowercase letter')
  if (newPassword && !/[0-9]/.test(newPassword)) passwordErrors.push('One number')
  const mismatch = confirmPassword !== '' && confirmPassword !== newPassword
  const canSubmit =
    currentPassword.length > 0 &&
    newPassword.length >= 8 &&
    passwordErrors.length === 0 &&
    confirmPassword === newPassword &&
    !updatePassword.isPending

  const handleSubmit = () => {
    setSuccess(false)
    updatePassword.mutate(
      { currentPassword, newPassword },
      {
        onSuccess: () => {
          setCurrentPassword('')
          setNewPassword('')
          setConfirmPassword('')
          setSuccess(true)
        },
      }
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Change Password</CardTitle>
        <CardDescription>Update your account password</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {success && (
          <div className="flex items-center gap-2 rounded-md border border-success/30 bg-success/10 p-3 text-sm text-success">
            <CheckCircle className="h-4 w-4 shrink-0" />
            Password updated successfully.
          </div>
        )}
        {updatePassword.isError && (
          <div className="flex items-center gap-2 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
            <XCircle className="h-4 w-4 shrink-0" />
            {(updatePassword.error as Error)?.message || 'Failed to update password'}
          </div>
        )}
        <div className="space-y-2">
          <Label htmlFor="pwd-current">Current Password</Label>
          <div className="relative">
            <Input
              id="pwd-current"
              type={showCurrent ? 'text' : 'password'}
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              placeholder="Enter current password"
              className="pr-10"
            />
            <button
              type="button"
              onClick={() => setShowCurrent(!showCurrent)}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 hover:bg-accent"
            >
              {showCurrent ? (
                <EyeOff className="h-4 w-4 text-muted-foreground" />
              ) : (
                <Eye className="h-4 w-4 text-muted-foreground" />
              )}
            </button>
          </div>
        </div>
        <div className="space-y-2">
          <Label htmlFor="pwd-new">New Password</Label>
          <div className="relative">
            <Input
              id="pwd-new"
              type={showNew ? 'text' : 'password'}
              value={newPassword}
              onChange={(e) => {
                setNewPassword(e.target.value)
                setSuccess(false)
              }}
              placeholder="Enter new password"
              className="pr-10"
            />
            <button
              type="button"
              onClick={() => setShowNew(!showNew)}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 hover:bg-accent"
            >
              {showNew ? (
                <EyeOff className="h-4 w-4 text-muted-foreground" />
              ) : (
                <Eye className="h-4 w-4 text-muted-foreground" />
              )}
            </button>
          </div>
          {newPassword && passwordErrors.length > 0 && (
            <ul className="space-y-1 text-xs text-destructive">
              {passwordErrors.map((err) => (
                <li key={err} className="flex items-center gap-1">
                  <XCircle className="h-3 w-3 shrink-0" />
                  {err}
                </li>
              ))}
            </ul>
          )}
          {newPassword && passwordErrors.length === 0 && (
            <p className="flex items-center gap-1 text-xs text-success">
              <Check className="h-3 w-3" />
              Password meets requirements
            </p>
          )}
        </div>
        <div className="space-y-2">
          <Label htmlFor="pwd-confirm">Confirm New Password</Label>
          <Input
            id="pwd-confirm"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="Confirm new password"
          />
          {mismatch && (
            <p className="flex items-center gap-1 text-xs text-destructive">
              <XCircle className="h-3 w-3" />
              Passwords do not match
            </p>
          )}
        </div>
      </CardContent>
      <CardFooter>
        <Button onClick={handleSubmit} disabled={!canSubmit}>
          {updatePassword.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
          Update Password
        </Button>
      </CardFooter>
    </Card>
  )
}
