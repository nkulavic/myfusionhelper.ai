import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi, type LoginInput, type RegisterInput } from '@/lib/api/auth'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import { cognitoSignIn, cognitoSignUp, cognitoConfirmSignUp, cognitoSignOut } from '@/lib/auth-client'

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
    mutationFn: async (input: LoginInput) => {
      // Step 1: Authenticate with Cognito
      await cognitoSignIn(input.email, input.password)
      // Step 2: Get user context from Go backend
      const res = await authApi.login(input)
      return res
    },
    onSuccess: (res) => {
      if (res.data) {
        setUser(res.data.user)
        setAccount(res.data.account)
      }
    },
  })
}

export function useRegister() {
  return useMutation({
    mutationFn: async (input: RegisterInput) => {
      // Step 1: Create user in Cognito (requires email verification)
      const result = await cognitoSignUp(input.email, input.password, input.name)
      return result
    },
  })
}

export function useConfirmRegistration() {
  const { setUser } = useAuthStore()
  const { setAccount } = useWorkspaceStore()
  return useMutation({
    mutationFn: async (input: { email: string; code: string; name: string; password: string }) => {
      // Step 1: Confirm sign up with verification code
      await cognitoConfirmSignUp(input.email, input.code)
      // Step 2: Sign in with Cognito now that user is confirmed
      await cognitoSignIn(input.email, input.password)
      // Step 3: Create backend records
      const res = await authApi.register({
        email: input.email,
        name: input.name,
        password: input.password,
      })
      return res
    },
    onSuccess: (res) => {
      if (res.data) {
        setUser(res.data.user)
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
    mutationFn: async () => {
      // Step 1: Notify Go backend
      try {
        await authApi.logout()
      } catch {
        // Backend logout failure shouldn't block client cleanup
      }
      // Step 2: Sign out of Cognito
      await cognitoSignOut()
    },
    onSuccess: () => {
      clearAuth()
      reset()
      queryClient.clear()
    },
  })
}
