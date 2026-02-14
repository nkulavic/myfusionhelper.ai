import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type FilterOperator = 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte'
  | 'contains' | 'startswith' | 'in' | 'between' | 'daterange'

export interface FilterCondition {
  column: string
  operator: FilterOperator
  value: string | number | string[]
  value2?: string | number
}

export type SelectionLevel = 'none' | 'platform' | 'connection' | 'objectType' | 'record'

export interface Selection {
  level: SelectionLevel
  platformId: string | null
  platformName: string | null
  connectionId: string | null
  connectionName: string | null
  objectType: string | null
  objectTypeLabel: string | null
  recordId: string | null
  recordSummary: string | null
}

const defaultSelection: Selection = {
  level: 'none',
  platformId: null,
  platformName: null,
  connectionId: null,
  connectionName: null,
  objectType: null,
  objectTypeLabel: null,
  recordId: null,
  recordSummary: null,
}

const defaultPagination = { page: 1, pageSize: 50, total: 0 }
const defaultSorting = { sortBy: null as string | null, sortOrder: 'asc' as const }

interface DataExplorerState {
  // Selection
  selection: Selection

  // Navigation
  expandedNodes: string[]

  // UI
  sidebarOpen: boolean
  sidebarWidth: number
  searchQuery: string

  // Query state
  pagination: { page: number; pageSize: number; total: number }
  sorting: { sortBy: string | null; sortOrder: 'asc' | 'desc' }
  filterConditions: FilterCondition[]
  nlQuery: string | null
  generatedDescription: string | null

  // Selection actions
  selectPlatform: (id: string, name: string) => void
  selectConnection: (platformId: string, platformName: string, connectionId: string, connectionName: string) => void
  selectObjectType: (platformId: string, platformName: string, connectionId: string, connectionName: string, objectType: string, label: string) => void
  selectRecord: (recordId: string, summary: string) => void
  clearSelection: () => void

  // Navigation actions
  toggleNodeExpansion: (nodeId: string) => void
  expandNode: (nodeId: string) => void
  collapseAll: () => void

  // UI actions
  setSidebarOpen: (open: boolean) => void
  setSidebarWidth: (width: number) => void
  setSearchQuery: (query: string) => void

  // Query actions
  setPage: (page: number) => void
  setPageSize: (pageSize: number) => void
  setSorting: (sortBy: string | null, sortOrder: 'asc' | 'desc') => void
  addFilter: (filter: FilterCondition) => void
  removeFilter: (index: number) => void
  clearFilters: () => void
  setNLQuery: (query: string | null) => void
  setGeneratedDescription: (desc: string | null) => void
  setTotal: (total: number) => void

  // Reset
  reset: () => void
}

export const useDataExplorerStore = create<DataExplorerState>()(
  persist(
    (set, get) => ({
      // Selection
      selection: { ...defaultSelection },

      // Navigation
      expandedNodes: [],

      // UI
      sidebarOpen: true,
      sidebarWidth: 300,
      searchQuery: '',

      // Query state
      pagination: { ...defaultPagination },
      sorting: { ...defaultSorting },
      filterConditions: [],
      nlQuery: null,
      generatedDescription: null,

      // Selection actions
      selectPlatform: (id, name) =>
        set({
          selection: {
            level: 'platform',
            platformId: id,
            platformName: name,
            connectionId: null,
            connectionName: null,
            objectType: null,
            objectTypeLabel: null,
            recordId: null,
            recordSummary: null,
          },
        }),

      selectConnection: (platformId, platformName, connectionId, connectionName) =>
        set({
          selection: {
            level: 'connection',
            platformId,
            platformName,
            connectionId,
            connectionName,
            objectType: null,
            objectTypeLabel: null,
            recordId: null,
            recordSummary: null,
          },
        }),

      selectObjectType: (platformId, platformName, connectionId, connectionName, objectType, label) =>
        set({
          selection: {
            level: 'objectType',
            platformId,
            platformName,
            connectionId,
            connectionName,
            objectType,
            objectTypeLabel: label,
            recordId: null,
            recordSummary: null,
          },
          pagination: { ...get().pagination, page: 1 },
          filterConditions: [],
          nlQuery: null,
          generatedDescription: null,
        }),

      selectRecord: (recordId, summary) =>
        set((state) => ({
          selection: {
            ...state.selection,
            level: 'record',
            recordId,
            recordSummary: summary,
          },
        })),

      clearSelection: () =>
        set({
          selection: { ...defaultSelection },
        }),

      // Navigation actions
      toggleNodeExpansion: (nodeId) =>
        set((state) => ({
          expandedNodes: state.expandedNodes.includes(nodeId)
            ? state.expandedNodes.filter((id) => id !== nodeId)
            : [...state.expandedNodes, nodeId],
        })),

      expandNode: (nodeId) =>
        set((state) => ({
          expandedNodes: state.expandedNodes.includes(nodeId)
            ? state.expandedNodes
            : [...state.expandedNodes, nodeId],
        })),

      collapseAll: () => set({ expandedNodes: [] }),

      // UI actions
      setSidebarOpen: (open) => set({ sidebarOpen: open }),
      setSidebarWidth: (width) => set({ sidebarWidth: width }),
      setSearchQuery: (query) => set({ searchQuery: query }),

      // Query actions
      setPage: (page) =>
        set((state) => ({
          pagination: { ...state.pagination, page },
        })),

      setPageSize: (pageSize) =>
        set((state) => ({
          pagination: { ...state.pagination, pageSize, page: 1 },
        })),

      setSorting: (sortBy, sortOrder) =>
        set({ sorting: { sortBy, sortOrder } }),

      addFilter: (filter) =>
        set((state) => ({
          filterConditions: [...state.filterConditions, filter],
          pagination: { ...state.pagination, page: 1 },
        })),

      removeFilter: (index) =>
        set((state) => ({
          filterConditions: state.filterConditions.filter((_, i) => i !== index),
          pagination: { ...state.pagination, page: 1 },
        })),

      clearFilters: () =>
        set((state) => ({
          filterConditions: [],
          pagination: { ...state.pagination, page: 1 },
        })),

      setNLQuery: (query) => set({ nlQuery: query }),
      setGeneratedDescription: (desc) => set({ generatedDescription: desc }),

      setTotal: (total) =>
        set((state) => ({
          pagination: { ...state.pagination, total },
        })),

      // Reset
      reset: () =>
        set({
          selection: { ...defaultSelection },
          expandedNodes: [],
          searchQuery: '',
          pagination: { ...defaultPagination },
          sorting: { ...defaultSorting },
          filterConditions: [],
          nlQuery: null,
          generatedDescription: null,
        }),
    }),
    {
      name: 'mfh-data-explorer',
      partialize: (state) => ({
        sidebarOpen: state.sidebarOpen,
        sidebarWidth: state.sidebarWidth,
        expandedNodes: state.expandedNodes,
        pagination: { pageSize: state.pagination.pageSize },
      }),
    }
  )
)
