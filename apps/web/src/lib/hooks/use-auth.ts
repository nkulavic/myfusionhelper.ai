import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi, type LoginInput, type RegisterInput } from '@/lib/api/auth'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'

export function useAuthStatus() {
  const { isAuthenticated } = useAuthStore()
  return useQuery({
    queryKey: ['auth-status'],
    queryFn: async () => {
      const res = await authApi.status()
      return res.data
    },
    enabled: isAuthenticated,
  })
}

export function useLogin() {
  const { setUser } = useAuthStore()
  const { setAccount } = useWorkspaceStore()
  return useMutation({
    mutationFn: (input: LoginInput) => authApi.login(input),
    onSuccess: (res) => {
      if (res.data) {
        setUser(res.data.user, res.data.token, res.data.refreshToken)
        setAccount(res.data.account)
      }
    },
  })
}

export function useRegister() {
  const { setUser } = useAuthStore()
  const { setAccount } = useWorkspaceStore()
  return useMutation({
    mutationFn: (input: RegisterInput) => authApi.register(input),
    onSuccess: (res) => {
      if (res.data) {
        setUser(res.data.user, res.data.token, res.data.refreshToken)
        setAccount(res.data.account)
      }
    },
  })
}

export function useLogout() {
  const { clearAuth } = useAuthStore()
  const { reset } = useWorkspaceStore()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => authApi.logout(),
    onSuccess: () => {
      clearAuth()
      reset()
      queryClient.clear()
    },
  })
}
