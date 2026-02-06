import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  settingsApi,
  type UpdateProfileInput,
  type UpdatePasswordInput,
  type UpdateAccountInput,
  type CreateAPIKeyInput,
  type InviteTeamMemberInput,
  type UpdateTeamMemberInput,
  type NotificationPreferences,
} from '@/lib/api/settings'
import { toast } from 'sonner'

export function useUpdateProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: UpdateProfileInput) => settingsApi.updateProfile(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auth-status'] })
      toast.success('Profile updated')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update profile')
    },
  })
}

export function useUpdatePassword() {
  return useMutation({
    mutationFn: (input: UpdatePasswordInput) => settingsApi.updatePassword(input),
    onSuccess: () => {
      toast.success('Password updated')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update password')
    },
  })
}

export function useAccount(accountId: string) {
  return useQuery({
    queryKey: ['account', accountId],
    queryFn: async () => {
      const res = await settingsApi.getAccount(accountId)
      return res.data
    },
    enabled: !!accountId,
  })
}

export function useUpdateAccount() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      accountId,
      input,
    }: {
      accountId: string
      input: UpdateAccountInput
    }) => settingsApi.updateAccount(accountId, input),
    onSuccess: (_, { accountId }) => {
      queryClient.invalidateQueries({ queryKey: ['account', accountId] })
    },
  })
}

export function useAPIKeys() {
  return useQuery({
    queryKey: ['api-keys'],
    queryFn: async () => {
      const res = await settingsApi.listAPIKeys()
      return res.data ?? []
    },
  })
}

export function useCreateAPIKey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateAPIKeyInput) => settingsApi.createAPIKey(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
      toast.success('API key created')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create API key')
    },
  })
}

export function useRevokeAPIKey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => settingsApi.revokeAPIKey(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
      toast.success('API key revoked')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to revoke API key')
    },
  })
}

export function useBillingInfo() {
  return useQuery({
    queryKey: ['billing'],
    queryFn: async () => {
      const res = await settingsApi.getBillingInfo()
      return res.data
    },
  })
}

export function useInvoices() {
  return useQuery({
    queryKey: ['invoices'],
    queryFn: async () => {
      const res = await settingsApi.listInvoices()
      return res.data ?? []
    },
  })
}

export function useCreatePortalSession() {
  return useMutation({
    mutationFn: () => settingsApi.createPortalSession(),
  })
}

export function useCreateCheckoutSession() {
  return useMutation({
    mutationFn: (plan: 'start' | 'grow' | 'deliver') =>
      settingsApi.createCheckoutSession(plan),
  })
}

export function useTeamMembers(accountId: string) {
  return useQuery({
    queryKey: ['team-members', accountId],
    queryFn: async () => {
      const res = await settingsApi.listTeamMembers(accountId)
      return res.data?.members ?? []
    },
    enabled: !!accountId,
  })
}

export function useInviteTeamMember() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      accountId,
      input,
    }: {
      accountId: string
      input: InviteTeamMemberInput
    }) => settingsApi.inviteTeamMember(accountId, input),
    onSuccess: (_, { accountId }) => {
      queryClient.invalidateQueries({ queryKey: ['team-members', accountId] })
      toast.success('Team member invited')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to invite team member')
    },
  })
}

export function useUpdateTeamMember() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      accountId,
      userId,
      input,
    }: {
      accountId: string
      userId: string
      input: UpdateTeamMemberInput
    }) => settingsApi.updateTeamMember(accountId, userId, input),
    onSuccess: (_, { accountId }) => {
      queryClient.invalidateQueries({ queryKey: ['team-members', accountId] })
      toast.success('Team member updated')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update team member')
    },
  })
}

export function useRemoveTeamMember() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      accountId,
      userId,
    }: {
      accountId: string
      userId: string
    }) => settingsApi.removeTeamMember(accountId, userId),
    onSuccess: (_, { accountId }) => {
      queryClient.invalidateQueries({ queryKey: ['team-members', accountId] })
      toast.success('Team member removed')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to remove team member')
    },
  })
}

export function useNotificationPreferences() {
  return useQuery({
    queryKey: ['notification-preferences'],
    queryFn: async () => {
      const res = await settingsApi.getNotificationPreferences()
      return res.data
    },
  })
}

export function useUpdateNotificationPreferences() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<NotificationPreferences>) =>
      settingsApi.updateNotificationPreferences(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notification-preferences'] })
      toast.success('Notification preferences saved')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to save preferences')
    },
  })
}
