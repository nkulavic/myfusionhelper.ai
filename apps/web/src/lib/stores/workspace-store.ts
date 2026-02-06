import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { Account, PlatformConnection } from '@myfusionhelper/types'

interface WorkspaceState {
  currentAccount: Account | null
  connections: PlatformConnection[]
  activeConnectionId: string | null
  onboardingComplete: boolean

  setAccount: (account: Account) => void
  setConnections: (connections: PlatformConnection[]) => void
  addConnection: (connection: PlatformConnection) => void
  updateConnection: (id: string, updates: Partial<PlatformConnection>) => void
  removeConnection: (id: string) => void
  setActiveConnection: (id: string | null) => void
  getActiveConnection: () => PlatformConnection | undefined
  completeOnboarding: () => void
  reset: () => void
}

export const useWorkspaceStore = create<WorkspaceState>()(
  persist(
    (set, get) => ({
      currentAccount: null,
      connections: [],
      activeConnectionId: null,
      onboardingComplete: false,

      setAccount: (account) => set({ currentAccount: account }),

      setConnections: (connections) => set({ connections }),

      addConnection: (connection) =>
        set((state) => ({ connections: [...state.connections, connection] })),

      updateConnection: (id, updates) =>
        set((state) => ({
          connections: state.connections.map((c) =>
            c.connectionId === id ? { ...c, ...updates } : c
          ),
        })),

      removeConnection: (id) =>
        set((state) => ({
          connections: state.connections.filter((c) => c.connectionId !== id),
          activeConnectionId:
            state.activeConnectionId === id ? null : state.activeConnectionId,
        })),

      setActiveConnection: (id) => set({ activeConnectionId: id }),

      getActiveConnection: () => {
        const state = get()
        return state.connections.find((c) => c.connectionId === state.activeConnectionId)
      },

      completeOnboarding: () => set({ onboardingComplete: true }),

      reset: () =>
        set({ currentAccount: null, connections: [], activeConnectionId: null, onboardingComplete: false }),
    }),
    {
      name: 'mfh-workspace',
      partialize: (state) => ({
        currentAccount: state.currentAccount,
        activeConnectionId: state.activeConnectionId,
        onboardingComplete: state.onboardingComplete,
      }),
    }
  )
)
