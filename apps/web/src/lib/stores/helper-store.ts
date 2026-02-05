import { create } from 'zustand'
import type { Helper } from '@myfusionhelper/types'

interface HelperState {
  helpers: Helper[]
  selectedHelper: Helper | null
  isLoading: boolean

  // Builder state
  builderConfig: Record<string, unknown>
  builderHelperType: string | null
  builderConnectionId: string | null

  setHelpers: (helpers: Helper[]) => void
  addHelper: (helper: Helper) => void
  updateHelper: (id: string, updates: Partial<Helper>) => void
  removeHelper: (id: string) => void
  setSelectedHelper: (helper: Helper | null) => void
  setLoading: (loading: boolean) => void

  // Builder actions
  setBuilderHelperType: (type: string | null) => void
  setBuilderConnectionId: (id: string | null) => void
  setBuilderConfig: (config: Record<string, unknown>) => void
  updateBuilderConfig: (key: string, value: unknown) => void
  resetBuilder: () => void
}

export const useHelperStore = create<HelperState>()((set) => ({
  helpers: [],
  selectedHelper: null,
  isLoading: false,
  builderConfig: {},
  builderHelperType: null,
  builderConnectionId: null,

  setHelpers: (helpers) => set({ helpers }),

  addHelper: (helper) =>
    set((state) => ({ helpers: [...state.helpers, helper] })),

  updateHelper: (id, updates) =>
    set((state) => ({
      helpers: state.helpers.map((h) => (h.id === id ? { ...h, ...updates } : h)),
      selectedHelper:
        state.selectedHelper?.id === id
          ? { ...state.selectedHelper, ...updates }
          : state.selectedHelper,
    })),

  removeHelper: (id) =>
    set((state) => ({
      helpers: state.helpers.filter((h) => h.id !== id),
      selectedHelper: state.selectedHelper?.id === id ? null : state.selectedHelper,
    })),

  setSelectedHelper: (helper) => set({ selectedHelper: helper }),
  setLoading: (loading) => set({ isLoading: loading }),

  setBuilderHelperType: (type) => set({ builderHelperType: type, builderConfig: {} }),
  setBuilderConnectionId: (id) => set({ builderConnectionId: id }),
  setBuilderConfig: (config) => set({ builderConfig: config }),
  updateBuilderConfig: (key, value) =>
    set((state) => ({
      builderConfig: { ...state.builderConfig, [key]: value },
    })),
  resetBuilder: () =>
    set({ builderHelperType: null, builderConnectionId: null, builderConfig: {} }),
}))
