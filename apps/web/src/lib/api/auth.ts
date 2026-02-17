import { apiClient } from './client'
import type { User, Account } from '@myfusionhelper/types'

export interface LoginInput {
  email: string
  password: string
}

export interface RegisterInput {
  email: string
  password: string
  name: string
  phoneNumber: string
}

export interface ForgotPasswordInput {
  email: string
}

export interface ResetPasswordInput {
  email: string
  code: string
  newPassword: string
}

export interface AuthStatusResponse {
  user: User
  account: Account
}

interface AuthResponse {
  token: string
  refreshToken: string
  user: User
  account: Account
}

export interface MfaChallengeResponse {
  mfaRequired: true
  challengeName: 'SMS_MFA' | 'SOFTWARE_TOKEN_MFA'
  session: string
}

export interface MfaStatusResponse {
  enabled: boolean
  method: 'totp' | 'sms' | null
  phoneVerified: boolean
}

export interface TotpSetupResponse {
  secret: string
  qrCodeUri: string
}

export type LoginResponse = AuthResponse | MfaChallengeResponse

export const authApi = {
  login: (input: LoginInput) =>
    apiClient.post<LoginResponse>('/auth/login', input),

  register: (input: RegisterInput) =>
    apiClient.post<AuthResponse>('/auth/register', input),

  status: () =>
    apiClient.get<AuthStatusResponse>('/auth/status'),

  logout: () => apiClient.post<void>('/auth/logout'),

  refresh: () =>
    apiClient.post<{ token: string }>('/auth/refresh'),

  completeOnboarding: () =>
    apiClient.patch<{ onboardingComplete: boolean }>('/auth/onboarding-complete'),

  forgotPassword: (input: ForgotPasswordInput) =>
    apiClient.post<void>('/auth/forgot-password', input),

  resetPassword: (input: ResetPasswordInput) =>
    apiClient.post<void>('/auth/reset-password', input),

  // MFA
  submitMfaChallenge: (input: { session: string; code: string; challengeName: string }) =>
    apiClient.post<AuthResponse>('/auth/mfa-challenge', input),

  getMfaStatus: () =>
    apiClient.get<MfaStatusResponse>('/auth/mfa/status'),

  setupTotp: () =>
    apiClient.post<TotpSetupResponse>('/auth/mfa/setup-totp'),

  verifyTotp: (input: { code: string }) =>
    apiClient.post<void>('/auth/mfa/verify-totp', input),

  enableSmsMfa: () =>
    apiClient.post<void>('/auth/mfa/enable-sms'),

  disableMfa: () =>
    apiClient.post<void>('/auth/mfa/disable'),
}
