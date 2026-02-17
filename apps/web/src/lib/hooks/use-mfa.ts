import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi } from '@/lib/api/auth'
import { toast } from 'sonner'

export function useMfaStatus() {
  return useQuery({
    queryKey: ['mfa-status'],
    queryFn: async () => {
      const res = await authApi.getMfaStatus()
      return res.data
    },
  })
}

export function useSetupTotp() {
  return useMutation({
    mutationFn: () => authApi.setupTotp(),
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to start TOTP setup')
    },
  })
}

export function useVerifyTotp() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: { code: string }) => authApi.verifyTotp(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mfa-status'] })
      toast.success('Two-factor authentication enabled')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Invalid verification code')
    },
  })
}

export function useEnableSmsMfa() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => authApi.enableSmsMfa(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mfa-status'] })
      toast.success('SMS two-factor authentication enabled')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to enable SMS 2FA')
    },
  })
}

export function useDisableMfa() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => authApi.disableMfa(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mfa-status'] })
      toast.success('Two-factor authentication disabled')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to disable 2FA')
    },
  })
}
