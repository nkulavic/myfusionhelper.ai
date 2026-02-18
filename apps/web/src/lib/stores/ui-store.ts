import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

interface UIState {
  sidebarCollapsed: boolean
  sidebarMinimized: boolean
  commandPaletteOpen: boolean
  aiChatOpen: boolean

  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void
  toggleSidebarMinimized: () => void
  setCommandPaletteOpen: (open: boolean) => void
  setAIChatOpen: (open: boolean) => void
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      sidebarCollapsed: true,
      sidebarMinimized: false,
      commandPaletteOpen: false,
      aiChatOpen: false,

      toggleSidebar: () =>
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),
      toggleSidebarMinimized: () =>
        set((state) => ({ sidebarMinimized: !state.sidebarMinimized })),
      setCommandPaletteOpen: (open) => set({ commandPaletteOpen: open }),
      setAIChatOpen: (open) => set({ aiChatOpen: open }),
    }),
    {
      name: 'mfh-ui',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ sidebarMinimized: state.sidebarMinimized }),
    }
  )
)
