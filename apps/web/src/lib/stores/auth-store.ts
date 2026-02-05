import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@myfusionhelper/types'
import { setTokens, clearTokens } from '@/lib/auth-client'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean

  setUser: (user: User, token: string, refreshToken?: string) => void
  updateToken: (token: string) => void
  clearAuth: () => void
  setLoading: (loading: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      isLoading: true,

      setUser: (user, token, refreshToken) => {
        setTokens(token, refreshToken)
        set({ user, isAuthenticated: true, isLoading: false })
      },

      updateToken: (token) => {
        setTokens(token)
      },

      clearAuth: () => {
        clearTokens()
        set({ user: null, isAuthenticated: false, isLoading: false })
      },

      setLoading: (loading) => set({ isLoading: loading }),
    }),
    {
      name: 'mfh-auth',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)
