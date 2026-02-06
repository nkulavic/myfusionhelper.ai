import { apiClient } from './client'
import type { Account, APIKey, User, NotificationPreferences } from '@myfusionhelper/types'

export interface UpdateProfileInput {
  name?: string
  email?: string
}

export interface UpdateAccountInput {
  name?: string
  company?: string
  timezone?: string
}

export interface CreateAPIKeyInput {
  name: string
  permissions: string[]
  expiresAt?: string
}

export interface InviteTeamMemberInput {
  email: string
  role: 'admin' | 'member' | 'viewer'
}

export interface UpdateTeamMemberInput {
  role: 'admin' | 'member' | 'viewer'
}

export type { NotificationPreferences }

export const settingsApi = {
  // Profile - uses auth status endpoint
  getProfile: () =>
    apiClient.get<{ user: User; account: Account }>('/auth/status'),

  // Account - requires account ID
  getAccount: (accountId: string) =>
    apiClient.get<Account>(`/accounts/${accountId}`),

  updateAccount: (accountId: string, input: UpdateAccountInput) =>
    apiClient.put<Account>(`/accounts/${accountId}`, input),

  listAccounts: () =>
    apiClient.get<Account[]>('/accounts'),

  switchAccount: (accountId: string) =>
    apiClient.post<{ account: Account }>('/accounts/switch', { accountId }),

  // API Keys
  listAPIKeys: () => apiClient.get<APIKey[]>('/api-keys'),

  createAPIKey: (input: CreateAPIKeyInput) =>
    apiClient.post<APIKey & { key: string }>('/api-keys', input),

  revokeAPIKey: (id: string) =>
    apiClient.delete<void>(`/api-keys/${id}`),

  // Profile update
  updateProfile: (input: UpdateProfileInput) =>
    apiClient.put<User>('/auth/profile', input),

  // Team management
  listTeamMembers: (accountId: string) =>
    apiClient.get<{ members: { userId: string; role: string; email: string; name: string; status: string; linkedAt: string }[]; totalCount: number }>(
      `/accounts/${accountId}/team`
    ),

  inviteTeamMember: (accountId: string, input: InviteTeamMemberInput) =>
    apiClient.post<{ userId: string; email: string; role: string; status: string }>(
      `/accounts/${accountId}/team`,
      input
    ),

  updateTeamMember: (accountId: string, userId: string, input: UpdateTeamMemberInput) =>
    apiClient.put<{ userId: string; role: string }>(
      `/accounts/${accountId}/team/${userId}`,
      input
    ),

  removeTeamMember: (accountId: string, userId: string) =>
    apiClient.delete<{ userId: string }>(
      `/accounts/${accountId}/team/${userId}`
    ),

  // Notification preferences
  getNotificationPreferences: () =>
    apiClient.get<NotificationPreferences>('/accounts/preferences'),

  updateNotificationPreferences: (input: Partial<NotificationPreferences>) =>
    apiClient.put<NotificationPreferences>('/accounts/preferences', input),

  // Billing
  getBillingInfo: () =>
    apiClient.get<BillingInfo>('/billing'),

  createPortalSession: () =>
    apiClient.post<{ url: string }>('/billing/portal-session'),

  listInvoices: () =>
    apiClient.get<Invoice[]>('/billing/invoices'),

  createCheckoutSession: (plan: 'start' | 'grow' | 'deliver') =>
    apiClient.post<{ url: string; sessionId: string }>('/billing/checkout/sessions', { plan }),
}

export interface BillingInfo {
  plan: string
  status: string
  priceMonthly: number
  renewsAt?: number
  trialEndsAt?: number
  cancelAt?: number
  stripeCustomerId?: string
  usage: {
    helpers: number
    connections: number
    apiKeys: number
    teamMembers: number
    monthlyExecutions: number
    monthlyApiRequests: number
  }
  limits: {
    maxHelpers: number
    maxConnections: number
    maxApiKeys: number
    maxTeamMembers: number
    maxExecutions: number
    webhooksEnabled: boolean
  }
}

export interface Invoice {
  id: string
  amount: number
  currency: string
  status: string
  date: number
  pdfUrl?: string
  hostedUrl?: string
}
