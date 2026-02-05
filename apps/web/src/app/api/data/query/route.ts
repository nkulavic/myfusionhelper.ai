import { NextRequest, NextResponse } from 'next/server'
import {
  getMockContacts,
  getMockTags,
  getMockCustomFields,
  getMockDeals,
} from '@/lib/mock-data/crm-mock-data'
import type {
  DataQueryRequest,
  DataQueryResponse,
  FilterCondition,
  FieldSchema,
} from '@/lib/api/data-explorer'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function getDataset(objectType: string): Record<string, unknown>[] {
  switch (objectType) {
    case 'contacts':
      return getMockContacts().map(flattenContact)
    case 'tags':
      return getMockTags().map((r) => ({ ...r }) as Record<string, unknown>)
    case 'custom_fields':
      return getMockCustomFields().map((r) => ({ ...r }) as Record<string, unknown>)
    case 'deals':
      return getMockDeals().map((r) => ({ ...r }) as Record<string, unknown>)
    default:
      // Fallback: return contacts for any unrecognised object type (all platforms)
      return getMockContacts().map(flattenContact)
  }
}

function flattenContact(
  contact: ReturnType<typeof getMockContacts>[number]
): Record<string, unknown> {
  const { customFields, ...rest } = contact
  return {
    ...rest,
    leadSource: customFields.leadSource,
    industry: customFields.industry,
    annualRevenue: customFields.annualRevenue,
    preferredContact: customFields.preferredContact,
  } as Record<string, unknown>
}

// ---------------------------------------------------------------------------
// Filtering
// ---------------------------------------------------------------------------

function matchesFilter(
  record: Record<string, unknown>,
  filter: FilterCondition
): boolean {
  const val = record[filter.column]
  const filterValue = filter.value

  switch (filter.operator) {
    case 'eq':
      return String(val) === String(filterValue)

    case 'neq':
      return String(val) !== String(filterValue)

    case 'gt':
      return Number(val) > Number(filterValue)

    case 'gte':
      return Number(val) >= Number(filterValue)

    case 'lt':
      return Number(val) < Number(filterValue)

    case 'lte':
      return Number(val) <= Number(filterValue)

    case 'contains':
      return String(val).toLowerCase().includes(String(filterValue).toLowerCase())

    case 'startswith':
      return String(val).toLowerCase().startsWith(String(filterValue).toLowerCase())

    case 'in':
      if (Array.isArray(filterValue)) {
        return filterValue.map(String).includes(String(val))
      }
      return false

    case 'between':
      return (
        Number(val) >= Number(filterValue) &&
        Number(val) <= Number(filter.value2)
      )

    case 'daterange': {
      const dateVal = new Date(String(val)).getTime()
      const dateFrom = new Date(String(filterValue)).getTime()
      const dateTo = new Date(String(filter.value2)).getTime()
      return dateVal >= dateFrom && dateVal <= dateTo
    }

    default:
      return true
  }
}

function applyFilters(
  records: Record<string, unknown>[],
  filters: FilterCondition[]
): Record<string, unknown>[] {
  return records.filter((record) =>
    filters.every((filter) => matchesFilter(record, filter))
  )
}

// ---------------------------------------------------------------------------
// Search (global text search across all string values)
// ---------------------------------------------------------------------------

function applySearch(
  records: Record<string, unknown>[],
  search: string
): Record<string, unknown>[] {
  const term = search.toLowerCase()
  return records.filter((record) =>
    Object.values(record).some((value) => {
      if (value == null) return false
      if (Array.isArray(value)) {
        return value.some(
          (v) => typeof v === 'string' && v.toLowerCase().includes(term)
        )
      }
      if (typeof value === 'string') {
        return value.toLowerCase().includes(term)
      }
      return String(value).toLowerCase().includes(term)
    })
  )
}

// ---------------------------------------------------------------------------
// Sorting
// ---------------------------------------------------------------------------

function applySorting(
  records: Record<string, unknown>[],
  sortBy: string,
  sortOrder: 'asc' | 'desc'
): Record<string, unknown>[] {
  return [...records].sort((a, b) => {
    const aVal = a[sortBy]
    const bVal = b[sortBy]

    if (aVal == null && bVal == null) return 0
    if (aVal == null) return sortOrder === 'asc' ? -1 : 1
    if (bVal == null) return sortOrder === 'asc' ? 1 : -1

    // Attempt numeric comparison
    const aNum = Number(aVal)
    const bNum = Number(bVal)
    if (!isNaN(aNum) && !isNaN(bNum)) {
      return sortOrder === 'asc' ? aNum - bNum : bNum - aNum
    }

    // Fall back to string comparison
    const cmp = String(aVal).localeCompare(String(bVal))
    return sortOrder === 'asc' ? cmp : -cmp
  })
}

// ---------------------------------------------------------------------------
// Schema inference
// ---------------------------------------------------------------------------

function inferSchema(record: Record<string, unknown>): FieldSchema[] {
  return Object.entries(record).map(([key, value]) => {
    let type = 'string'
    if (typeof value === 'number') type = 'number'
    else if (typeof value === 'boolean') type = 'boolean'
    else if (Array.isArray(value)) type = 'json'
    else if (typeof value === 'string') {
      // Check if it looks like a date
      if (/^\d{4}-\d{2}-\d{2}T/.test(value)) {
        type = 'date'
      }
    }

    return {
      name: key,
      type,
      displayName: key
        .replace(/([A-Z])/g, ' $1')
        .replace(/^./, (s) => s.toUpperCase())
        .trim(),
    }
  })
}

// ---------------------------------------------------------------------------
// POST handler
// ---------------------------------------------------------------------------

export async function POST(request: NextRequest) {
  const startTime = performance.now()

  try {
    const body = (await request.json()) as DataQueryRequest
    const {
      objectType,
      page = 1,
      pageSize = 50,
      sortBy,
      sortOrder = 'asc',
      filterConditions,
      search,
    } = body

    if (!objectType) {
      return NextResponse.json(
        { error: 'objectType is required' },
        { status: 400 }
      )
    }

    // 1. Get the raw dataset
    let records = getDataset(objectType)

    // 2. Apply filter conditions
    if (filterConditions && filterConditions.length > 0) {
      records = applyFilters(records, filterConditions)
    }

    // 3. Apply global text search
    if (search && search.trim().length > 0) {
      records = applySearch(records, search.trim())
    }

    // 4. Apply sorting
    if (sortBy) {
      records = applySorting(records, sortBy, sortOrder)
    }

    // 5. Paginate
    const totalRecords = records.length
    const totalPages = Math.max(1, Math.ceil(totalRecords / pageSize))
    const safePage = Math.max(1, Math.min(page, totalPages))
    const startIdx = (safePage - 1) * pageSize
    const pagedRecords = records.slice(startIdx, startIdx + pageSize)

    // 6. Derive columns and schema from first record
    const columns =
      pagedRecords.length > 0 ? Object.keys(pagedRecords[0]) : []
    const schema =
      pagedRecords.length > 0 ? inferSchema(pagedRecords[0]) : []

    const queryTimeMs = Math.round((performance.now() - startTime) * 100) / 100

    const response: DataQueryResponse = {
      records: pagedRecords,
      totalRecords,
      page: safePage,
      pageSize,
      totalPages,
      hasNextPage: safePage < totalPages,
      hasPrevPage: safePage > 1,
      columns,
      schema,
      queryTimeMs,
    }

    return NextResponse.json(response)
  } catch (error) {
    console.error('Error processing data query:', error)
    return NextResponse.json(
      { error: 'Failed to process data query' },
      { status: 500 }
    )
  }
}
