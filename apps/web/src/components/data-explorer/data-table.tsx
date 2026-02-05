'use client'

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  type ColumnDef,
  type SortingState,
  type PaginationState,
} from '@tanstack/react-table'
import {
  ArrowDown,
  ArrowUp,
  ArrowUpDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Download,
  Filter,
  Inbox,
  MoreHorizontal,
  Plus,
  X,
} from 'lucide-react'

import {
  useDataExplorerStore,
  type FilterCondition,
  type FilterOperator,
} from '@/lib/stores/data-explorer-store'
import type { DataQueryResponse } from '@/lib/api/data-explorer'
import { downloadCsv, downloadJson } from '@/components/data-explorer/export-utils'

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const PAGE_SIZE_OPTIONS = [25, 50, 100, 250] as const

const OPERATOR_LABELS: Record<FilterOperator, string> = {
  eq: 'equals',
  neq: 'not equals',
  contains: 'contains',
  startswith: 'starts with',
  gt: 'greater than',
  gte: 'greater than or equal',
  lt: 'less than',
  lte: 'less than or equal',
  in: 'in',
  between: 'between',
  daterange: 'date range',
}

const FILTER_OPERATORS: { value: FilterOperator; label: string }[] = [
  { value: 'eq', label: 'Equals' },
  { value: 'neq', label: 'Not Equals' },
  { value: 'contains', label: 'Contains' },
  { value: 'startswith', label: 'Starts With' },
  { value: 'gt', label: 'Greater Than' },
  { value: 'lt', label: 'Less Than' },
  { value: 'between', label: 'Between' },
]

// ---------------------------------------------------------------------------
// Formatting helpers
// ---------------------------------------------------------------------------

function capitalize(str: string): string {
  return str
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase())
}

function isDateString(value: unknown): boolean {
  if (typeof value !== 'string') return false
  // ISO date patterns
  if (/^\d{4}-\d{2}-\d{2}(T|\s)/.test(value)) {
    const d = new Date(value)
    return !isNaN(d.getTime())
  }
  return false
}

function formatCellValue(value: unknown): string {
  if (value === null || value === undefined) return '-'

  if (Array.isArray(value)) {
    return value.map((v) => String(v ?? '')).join(', ')
  }

  if (typeof value === 'boolean') {
    return value ? 'Yes' : 'No'
  }

  if (typeof value === 'number') {
    return value.toLocaleString()
  }

  if (typeof value === 'string') {
    if (isDateString(value)) {
      try {
        return new Date(value).toLocaleDateString(undefined, {
          year: 'numeric',
          month: 'short',
          day: 'numeric',
        })
      } catch {
        return value
      }
    }
    if (value.length > 80) {
      return value.slice(0, 80) + '...'
    }
    return value
  }

  if (typeof value === 'object') {
    try {
      const s = JSON.stringify(value)
      return s.length > 80 ? s.slice(0, 80) + '...' : s
    } catch {
      return String(value)
    }
  }

  return String(value)
}

function getRecordDisplayName(
  record: Record<string, unknown>,
  objectType: string | null,
  columns: string[],
): string {
  // For contacts: first_name + last_name
  if (objectType === 'contacts' || objectType === 'contact') {
    const first = record['first_name'] ?? record['firstName'] ?? ''
    const last = record['last_name'] ?? record['lastName'] ?? ''
    const full = `${first} ${last}`.trim()
    if (full) return full
  }

  // For other types: try 'name' field
  if (record['name'] && typeof record['name'] === 'string') {
    return record['name']
  }

  // Fallback: first string column value
  for (const col of columns) {
    if (col === 'id') continue
    const val = record[col]
    if (typeof val === 'string' && val.trim()) {
      return val.length > 60 ? val.slice(0, 60) + '...' : val
    }
  }

  // Last resort: use the id
  const id = record['id'] ?? record['Id'] ?? record['ID']
  return id ? String(id) : 'Record'
}

// ---------------------------------------------------------------------------
// Filter builder form (internal)
// ---------------------------------------------------------------------------

interface FilterFormProps {
  columns: string[]
  onAdd: (filter: FilterCondition) => void
  onClose: () => void
}

function FilterForm({ columns, onAdd, onClose }: FilterFormProps) {
  const [column, setColumn] = useState<string>(columns[0] ?? '')
  const [operator, setOperator] = useState<FilterOperator>('contains')
  const [value, setValue] = useState('')
  const [value2, setValue2] = useState('')

  const handleSubmit = () => {
    if (!column || !value) return
    const filter: FilterCondition = { column, operator, value }
    if (operator === 'between' && value2) {
      filter.value2 = value2
    }
    onAdd(filter)
    setValue('')
    setValue2('')
    onClose()
  }

  return (
    <div className="space-y-3">
      <div className="text-sm font-medium">Add Filter</div>

      <Select value={column} onValueChange={(v) => setColumn(v)}>
        <SelectTrigger className="h-9 text-xs">
          <SelectValue placeholder="Column" />
        </SelectTrigger>
        <SelectContent>
          {columns.map((col) => (
            <SelectItem key={col} value={col}>
              {capitalize(col)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select value={operator} onValueChange={(v) => setOperator(v as FilterOperator)}>
        <SelectTrigger className="h-9 text-xs">
          <SelectValue placeholder="Operator" />
        </SelectTrigger>
        <SelectContent>
          {FILTER_OPERATORS.map((op) => (
            <SelectItem key={op.value} value={op.value}>
              {op.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Input
        className="h-9 text-xs"
        placeholder="Value"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') handleSubmit()
        }}
      />

      {operator === 'between' && (
        <Input
          className="h-9 text-xs"
          placeholder="Value 2"
          value={value2}
          onChange={(e) => setValue2(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleSubmit()
          }}
        />
      )}

      <div className="flex justify-end gap-2">
        <Button variant="ghost" size="sm" onClick={onClose}>
          Cancel
        </Button>
        <Button size="sm" onClick={handleSubmit} disabled={!column || !value}>
          <Plus className="mr-1 h-3 w-3" />
          Add
        </Button>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Pagination helpers
// ---------------------------------------------------------------------------

function getPageNumbers(currentPage: number, totalPages: number): (number | 'ellipsis')[] {
  if (totalPages <= 5) {
    return Array.from({ length: totalPages }, (_, i) => i + 1)
  }

  const pages: (number | 'ellipsis')[] = []

  if (currentPage <= 3) {
    pages.push(1, 2, 3, 4, 'ellipsis', totalPages)
  } else if (currentPage >= totalPages - 2) {
    pages.push(1, 'ellipsis', totalPages - 3, totalPages - 2, totalPages - 1, totalPages)
  } else {
    pages.push(1, 'ellipsis', currentPage - 1, currentPage, currentPage + 1, 'ellipsis', totalPages)
  }

  return pages
}

// ---------------------------------------------------------------------------
// DataTable component
// ---------------------------------------------------------------------------

export function DataTable() {
  const {
    selection,
    pagination,
    sorting,
    filterConditions,
    nlQuery,
    generatedDescription,
    setPage,
    setPageSize,
    setSorting,
    setTotal,
    selectRecord,
    addFilter,
    removeFilter,
    clearFilters,
  } = useDataExplorerStore()

  const { connectionId, objectType, objectTypeLabel } = selection

  // Local state
  const [data, setData] = useState<DataQueryResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [filterOpen, setFilterOpen] = useState(false)
  const [search, setSearch] = useState('')
  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [debouncedSearch, setDebouncedSearch] = useState('')

  // Debounce search input
  useEffect(() => {
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current)
    }
    searchTimeoutRef.current = setTimeout(() => {
      setDebouncedSearch(search)
    }, 300)
    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current)
      }
    }
  }, [search])

  // -------------------------------------------------------------------------
  // Data fetching
  // -------------------------------------------------------------------------

  const fetchData = useCallback(async () => {
    if (!connectionId || !objectType) return

    setLoading(true)
    setError(null)

    try {
      const response = await fetch('/api/data/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          connectionId,
          objectType,
          page: pagination.page,
          pageSize: pagination.pageSize,
          sortBy: sorting.sortBy,
          sortOrder: sorting.sortOrder,
          filterConditions,
          nlQuery,
          search: debouncedSearch || undefined,
        }),
      })

      if (!response.ok) {
        throw new Error(`Query failed: ${response.status} ${response.statusText}`)
      }

      const result: DataQueryResponse = await response.json()
      setData(result)
      setTotal(result.totalRecords)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred while fetching data')
      setData(null)
    } finally {
      setLoading(false)
    }
  }, [
    connectionId,
    objectType,
    pagination.page,
    pagination.pageSize,
    sorting.sortBy,
    sorting.sortOrder,
    filterConditions,
    nlQuery,
    debouncedSearch,
    setTotal,
  ])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  // -------------------------------------------------------------------------
  // Dynamic column definitions
  // -------------------------------------------------------------------------

  const columns = useMemo<ColumnDef<Record<string, unknown>>[]>(() => {
    if (!data?.columns?.length) return []

    return data.columns.map((col): ColumnDef<Record<string, unknown>> => ({
      id: col,
      accessorFn: (row) => row[col],
      header: () => capitalize(col),
      cell: ({ getValue }) => {
        const val = getValue()
        return (
          <span className="block max-w-[300px] truncate" title={String(val ?? '')}>
            {formatCellValue(val)}
          </span>
        )
      },
      enableSorting: true,
    }))
  }, [data?.columns])

  // -------------------------------------------------------------------------
  // TanStack table instance
  // -------------------------------------------------------------------------

  const tableSorting = useMemo<SortingState>(
    () => (sorting.sortBy ? [{ id: sorting.sortBy, desc: sorting.sortOrder === 'desc' }] : []),
    [sorting.sortBy, sorting.sortOrder],
  )

  const tablePagination = useMemo<PaginationState>(
    () => ({ pageIndex: pagination.page - 1, pageSize: pagination.pageSize }),
    [pagination.page, pagination.pageSize],
  )

  const table = useReactTable({
    data: data?.records ?? [],
    columns,
    pageCount: data?.totalPages ?? -1,
    state: {
      sorting: tableSorting,
      pagination: tablePagination,
    },
    onSortingChange: (updaterOrValue) => {
      const next =
        typeof updaterOrValue === 'function'
          ? updaterOrValue(tableSorting)
          : updaterOrValue
      if (next.length === 0) {
        setSorting(null, 'asc')
      } else {
        setSorting(next[0].id, next[0].desc ? 'desc' : 'asc')
      }
    },
    onPaginationChange: (updaterOrValue) => {
      const next =
        typeof updaterOrValue === 'function'
          ? updaterOrValue(tablePagination)
          : updaterOrValue
      setPage(next.pageIndex + 1)
      if (next.pageSize !== pagination.pageSize) {
        setPageSize(next.pageSize)
      }
    },
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    manualSorting: true,
  })

  // -------------------------------------------------------------------------
  // Export handlers
  // -------------------------------------------------------------------------

  const handleExportCsv = useCallback(() => {
    if (!data?.records?.length) return
    const filename = `${objectTypeLabel ?? objectType ?? 'export'}.csv`
    downloadCsv(data.records, filename)
  }, [data, objectType, objectTypeLabel])

  const handleExportJson = useCallback(() => {
    if (!data?.records?.length) return
    const filename = `${objectTypeLabel ?? objectType ?? 'export'}.json`
    downloadJson(data.records, filename)
  }, [data, objectType, objectTypeLabel])

  // -------------------------------------------------------------------------
  // Row click handler
  // -------------------------------------------------------------------------

  const handleRowClick = useCallback(
    (record: Record<string, unknown>) => {
      const recordId = String(record['id'] ?? record['Id'] ?? record['ID'] ?? '')
      if (!recordId) return
      const displayName = getRecordDisplayName(record, objectType, data?.columns ?? [])
      selectRecord(recordId, displayName)
    },
    [objectType, data?.columns, selectRecord],
  )

  // -------------------------------------------------------------------------
  // Early returns
  // -------------------------------------------------------------------------

  if (!connectionId || !objectType) {
    return (
      <div className="flex h-64 flex-col items-center justify-center gap-3 text-muted-foreground">
        <Inbox className="h-10 w-10" />
        <p className="text-sm">Select an object type to explore data</p>
      </div>
    )
  }

  // -------------------------------------------------------------------------
  // Computed values
  // -------------------------------------------------------------------------

  const totalPages = data?.totalPages ?? 0
  const currentPage = pagination.page
  const startRecord = data?.records?.length
    ? (currentPage - 1) * pagination.pageSize + 1
    : 0
  const endRecord = data?.records?.length
    ? startRecord + data.records.length - 1
    : 0
  const totalRecords = pagination.total
  const pageNumbers = getPageNumbers(currentPage, totalPages)

  // -------------------------------------------------------------------------
  // Render
  // -------------------------------------------------------------------------

  return (
    <div className="flex flex-col gap-0">
      {/* Toolbar */}
      <div className="flex flex-wrap items-center gap-2 border-b px-4 py-3">
        {/* Filter builder */}
        <Popover open={filterOpen} onOpenChange={setFilterOpen}>
          <PopoverTrigger asChild>
            <Button variant="outline" size="sm">
              <Filter className="mr-1 h-3.5 w-3.5" />
              Filter
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-72" align="start">
            <FilterForm
              columns={data?.columns ?? []}
              onAdd={(filter) => {
                addFilter(filter)
                setFilterOpen(false)
              }}
              onClose={() => setFilterOpen(false)}
            />
          </PopoverContent>
        </Popover>

        {/* Export dropdown */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="sm" disabled={!data?.records?.length}>
              <Download className="mr-1 h-3.5 w-3.5" />
              Export
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start">
            <DropdownMenuItem onClick={handleExportCsv}>
              Export CSV
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleExportJson}>
              Export JSON
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Page size selector */}
        <Select
          value={String(pagination.pageSize)}
          onValueChange={(v) => setPageSize(Number(v))}
        >
          <SelectTrigger className="h-9 w-[100px] text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {PAGE_SIZE_OPTIONS.map((size) => (
              <SelectItem key={size} value={String(size)}>
                {size} rows
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Search input */}
        <div className="ml-auto">
          <Input
            className="h-9 w-[200px] text-xs"
            placeholder="Search records..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>

      {/* Filter pills */}
      {filterConditions.length > 0 && (
        <div className="flex flex-wrap items-center gap-1.5 border-b px-4 py-2">
          {filterConditions.map((fc, idx) => (
            <Badge key={idx} variant="secondary" className="gap-1 pr-1 text-xs font-normal">
              <span>
                {capitalize(fc.column)}{' '}
                {OPERATOR_LABELS[fc.operator] ?? fc.operator}{' '}
                &quot;{String(fc.value)}&quot;
                {fc.value2 != null && ` and "${String(fc.value2)}"`}
              </span>
              <button
                className="ml-0.5 rounded-full p-0.5 hover:bg-muted"
                onClick={() => removeFilter(idx)}
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          ))}
          {filterConditions.length > 1 && (
            <Button
              variant="ghost"
              size="sm"
              className="h-6 text-xs text-muted-foreground"
              onClick={clearFilters}
            >
              Clear all
            </Button>
          )}
        </div>
      )}

      {/* NL description */}
      {generatedDescription && (
        <div className="border-b bg-muted/30 px-4 py-2 text-xs text-muted-foreground">
          {generatedDescription}
        </div>
      )}

      {/* Error state */}
      {error && (
        <div className="border-b bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Table */}
      <div className="relative overflow-auto">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  const sorted = header.column.getIsSorted()
                  return (
                    <TableHead
                      key={header.id}
                      className="whitespace-nowrap"
                    >
                      {header.isPlaceholder ? null : (
                        <button
                          className="inline-flex items-center gap-1 hover:text-foreground"
                          onClick={header.column.getToggleSortingHandler()}
                        >
                          {flexRender(header.column.columnDef.header, header.getContext())}
                          {sorted === 'asc' ? (
                            <ArrowUp className="h-3.5 w-3.5" />
                          ) : sorted === 'desc' ? (
                            <ArrowDown className="h-3.5 w-3.5" />
                          ) : (
                            <ArrowUpDown className="h-3.5 w-3.5 opacity-30" />
                          )}
                        </button>
                      )}
                    </TableHead>
                  )
                })}
              </TableRow>
            ))}
          </TableHeader>

          <TableBody>
            {loading ? (
              // Skeleton loading rows
              Array.from({ length: pagination.pageSize > 10 ? 10 : pagination.pageSize }).map(
                (_, i) => (
                  <TableRow key={`skeleton-${i}`}>
                    {(data?.columns ?? ['a', 'b', 'c', 'd', 'e']).map((col, j) => (
                      <TableCell key={`skeleton-${i}-${j}`}>
                        <Skeleton className="h-4 w-full max-w-[180px]" />
                      </TableCell>
                    ))}
                  </TableRow>
                ),
              )
            ) : table.getRowModel().rows.length === 0 ? (
              // Empty state
              <TableRow>
                <TableCell
                  colSpan={columns.length || 1}
                  className="h-48 text-center"
                >
                  <div className="flex flex-col items-center justify-center gap-2 text-muted-foreground">
                    <Inbox className="h-8 w-8" />
                    <p className="text-sm">No records found</p>
                    {filterConditions.length > 0 && (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-xs"
                        onClick={clearFilters}
                      >
                        Clear filters
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              // Data rows
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  className="cursor-pointer hover:bg-muted/50"
                  onClick={() => handleRowClick(row.original)}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id} className="py-2.5">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination footer */}
      {data && (
        <div className="flex items-center justify-between border-t px-4 py-3">
          {/* Record count */}
          <div className="text-xs text-muted-foreground">
            {totalRecords > 0
              ? `Showing ${startRecord}-${endRecord} of ${totalRecords.toLocaleString()} records`
              : 'No records'}
          </div>

          {/* Page buttons */}
          <div className="flex items-center gap-1">
            <Button
              variant="outline"
              size="sm"
              className="h-8 w-8 p-0"
              disabled={currentPage === 1}
              onClick={() => setPage(1)}
            >
              <ChevronsLeft className="h-3.5 w-3.5" />
            </Button>

            <Button
              variant="outline"
              size="sm"
              className="h-8 w-8 p-0"
              disabled={!data.hasPrevPage}
              onClick={() => setPage(currentPage - 1)}
            >
              <ChevronLeft className="h-3.5 w-3.5" />
            </Button>

            {pageNumbers.map((pageNum, idx) =>
              pageNum === 'ellipsis' ? (
                <span
                  key={`ellipsis-${idx}`}
                  className="flex h-8 w-8 items-center justify-center text-xs text-muted-foreground"
                >
                  <MoreHorizontal className="h-3.5 w-3.5" />
                </span>
              ) : (
                <Button
                  key={pageNum}
                  variant={pageNum === currentPage ? 'default' : 'outline'}
                  size="sm"
                  className="h-8 w-8 p-0 text-xs"
                  onClick={() => setPage(pageNum)}
                >
                  {pageNum}
                </Button>
              ),
            )}

            <Button
              variant="outline"
              size="sm"
              className="h-8 w-8 p-0"
              disabled={!data.hasNextPage}
              onClick={() => setPage(currentPage + 1)}
            >
              <ChevronRight className="h-3.5 w-3.5" />
            </Button>

            <Button
              variant="outline"
              size="sm"
              className="h-8 w-8 p-0"
              disabled={currentPage === totalPages}
              onClick={() => setPage(totalPages)}
            >
              <ChevronsRight className="h-3.5 w-3.5" />
            </Button>
          </div>

          {/* Query time */}
          <div className="text-xs text-muted-foreground">
            {data.queryTimeMs != null && `${data.queryTimeMs}ms`}
          </div>
        </div>
      )}
    </div>
  )
}
