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

export const authApi = {
  login: (input: LoginInput) =>
    apiClient.post<AuthResponse>('/auth/login', input),

  register: (input: RegisterInput) =>
    apiClient.post<AuthResponse>('/auth/register', input),

  status: () =>
    apiClient.get<AuthStatusResponse>('/auth/status'),

  logout: () => apiClient.post<void>('/auth/logout'),

  refresh: () =>
    apiClient.post<{ token: string }>('/auth/refresh'),

  forgotPassword: (input: ForgotPasswordInput) =>
    apiClient.post<void>('/auth/forgot-password', input),

  resetPassword: (input: ResetPasswordInput) =>
    apiClient.post<void>('/auth/reset-password', input),
}
