import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  settingsApi,
  type UpdateAccountInput,
  type CreateAPIKeyInput,
  type InviteTeamMemberInput,
  type NotificationPreferences,
} from '@/lib/api/settings'

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
    },
  })
}

export function useRevokeAPIKey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => settingsApi.revokeAPIKey(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
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
    },
  })
}
