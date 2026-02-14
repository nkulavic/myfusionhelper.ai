import { NextRequest, NextResponse } from 'next/server'
import {
  getMockContacts,
  getMockTags,
  getMockCustomFields,
  getMockDeals,
} from '@/lib/mock-data/crm-mock-data'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function getExportRecords(objectType: string): Record<string, unknown>[] {
  switch (objectType) {
    case 'contacts':
      return getMockContacts().map((c) => {
        const { customFields, ...rest } = c
        return {
          ...rest,
          leadSource: customFields.leadSource,
          industry: customFields.industry,
          annualRevenue: customFields.annualRevenue,
          preferredContact: customFields.preferredContact,
        } as Record<string, unknown>
      })

    case 'tags':
      return getMockTags().map((r) => ({ ...r }) as Record<string, unknown>)

    case 'custom_fields':
      return getMockCustomFields().map((r) => ({ ...r }) as Record<string, unknown>)

    case 'deals':
      return getMockDeals().map((r) => ({ ...r }) as Record<string, unknown>)

    default:
      // Fallback: return contacts for any unrecognised type
      return getMockContacts().map((c) => {
        const { customFields, ...rest } = c
        return {
          ...rest,
          leadSource: customFields.leadSource,
          industry: customFields.industry,
          annualRevenue: customFields.annualRevenue,
          preferredContact: customFields.preferredContact,
        } as Record<string, unknown>
      })
  }
}

function escapeCsvValue(value: unknown): string {
  if (value == null) return ''
  const str = Array.isArray(value) ? value.join('; ') : String(value)
  // Wrap in quotes if the value contains commas, quotes, or newlines
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return `"${str.replace(/"/g, '""')}"`
  }
  return str
}

function recordsToCsv(records: Record<string, unknown>[]): string {
  if (records.length === 0) return ''

  const columns = Object.keys(records[0])
  const headerRow = columns.map(escapeCsvValue).join(',')
  const dataRows = records.map((record) =>
    columns.map((col) => escapeCsvValue(record[col])).join(',')
  )

  return [headerRow, ...dataRows].join('\n')
}

// ---------------------------------------------------------------------------
// POST handler
// ---------------------------------------------------------------------------

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { objectType, format } = body as {
      connectionId?: string
      objectType?: string
      format?: 'json' | 'csv'
    }

    if (!objectType) {
      return NextResponse.json(
        { error: 'objectType is required' },
        { status: 400 }
      )
    }

    if (!format || (format !== 'json' && format !== 'csv')) {
      return NextResponse.json(
        { error: 'format must be "json" or "csv"' },
        { status: 400 }
      )
    }

    const records = getExportRecords(objectType)

    if (format === 'csv') {
      const csv = recordsToCsv(records)
      return new NextResponse(csv, {
        status: 200,
        headers: {
          'Content-Type': 'text/csv; charset=utf-8',
          'Content-Disposition': `attachment; filename="${objectType}-export.csv"`,
        },
      })
    }

    // JSON format
    return new NextResponse(JSON.stringify(records, null, 2), {
      status: 200,
      headers: {
        'Content-Type': 'application/json; charset=utf-8',
        'Content-Disposition': `attachment; filename="${objectType}-export.json"`,
      },
    })
  } catch (error) {
    console.error('Error exporting data:', error)
    return NextResponse.json(
      { error: 'Failed to export data' },
      { status: 500 }
    )
  }
}
