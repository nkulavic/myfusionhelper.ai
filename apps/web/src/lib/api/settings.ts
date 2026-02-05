import { apiClient } from './client'
import type { Account, APIKey, User } from '@myfusionhelper/types'

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

export interface NotificationPreferences {
  executionFailures: boolean
  connectionIssues: boolean
  usageAlerts: boolean
  weeklySummary: boolean
  newFeatures: boolean
  teamActivity: boolean
  realtimeStatus: boolean
  aiInsights: boolean
  systemMaintenance: boolean
  webhookUrl?: string
}

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

  // TODO: Team management endpoints not yet implemented in backend
  // These will need a new handler: cmd/handlers/accounts/clients/team/
  listTeamMembers: (_accountId: string) =>
    Promise.resolve({
      success: true,
      data: [] as { userId: string; role: string; user: User }[],
    }),

  inviteTeamMember: (_accountId: string, _input: InviteTeamMemberInput) =>
    Promise.resolve({ success: true, data: { invitationId: 'pending' } }),

  // TODO: Notification preferences not yet implemented in backend
  getNotificationPreferences: () =>
    Promise.resolve({
      success: true,
      data: {
        executionFailures: true,
        connectionIssues: true,
        usageAlerts: true,
        weeklySummary: false,
        newFeatures: true,
        teamActivity: false,
        realtimeStatus: false,
        aiInsights: true,
        systemMaintenance: true,
      } as NotificationPreferences,
    }),

  updateNotificationPreferences: (_input: Partial<NotificationPreferences>) =>
    Promise.resolve({ success: true, data: undefined }),

  // TODO: Billing endpoints not yet implemented in backend
  // These will need Stripe integration: cmd/handlers/billing/
  getBillingInfo: () =>
    Promise.resolve({
      success: true,
      data: {
        plan: 'pro',
        priceMonthly: 49,
        renewsAt: new Date(Date.now() + 30 * 86400000).toISOString(),
        usage: {
          executions: { used: 12847, limit: 50000 },
          apiCalls: { used: 45892, limit: 100000 },
          helpers: { used: 14, limit: 50 },
          connections: { used: 3, limit: 5 },
        },
      },
    }),

  createPortalSession: () =>
    Promise.resolve({ success: true, data: { url: '#' } }),

  listInvoices: () =>
    Promise.resolve({
      success: true,
      data: [
        { id: 'inv_001', amount: 49, status: 'paid', date: '2026-01-01' },
        { id: 'inv_002', amount: 49, status: 'paid', date: '2025-12-01' },
      ],
    }),
}
