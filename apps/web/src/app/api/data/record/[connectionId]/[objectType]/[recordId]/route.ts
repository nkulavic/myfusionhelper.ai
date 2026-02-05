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

function findRecord(
  objectType: string,
  recordId: string
): Record<string, unknown> | undefined {
  switch (objectType) {
    case 'contacts': {
      const contact = getMockContacts().find((c) => c.id === recordId)
      if (!contact) return undefined
      const { customFields, ...rest } = contact
      return {
        ...rest,
        leadSource: customFields.leadSource,
        industry: customFields.industry,
        annualRevenue: customFields.annualRevenue,
        preferredContact: customFields.preferredContact,
      } as Record<string, unknown>
    }

    case 'tags': {
      const tag = getMockTags().find((t) => t.id === recordId)
      return tag ? ({ ...tag } as Record<string, unknown>) : undefined
    }

    case 'custom_fields': {
      const field = getMockCustomFields().find((f) => f.id === recordId)
      return field ? ({ ...field } as Record<string, unknown>) : undefined
    }

    case 'deals': {
      const deal = getMockDeals().find((d) => d.id === recordId)
      return deal ? ({ ...deal } as Record<string, unknown>) : undefined
    }

    default: {
      // Fallback: search contacts for any unrecognised type
      const fallback = getMockContacts().find((c) => c.id === recordId)
      if (!fallback) return undefined
      const { customFields, ...rest } = fallback
      return {
        ...rest,
        leadSource: customFields.leadSource,
        industry: customFields.industry,
        annualRevenue: customFields.annualRevenue,
        preferredContact: customFields.preferredContact,
      } as Record<string, unknown>
    }
  }
}

// ---------------------------------------------------------------------------
// GET handler
// ---------------------------------------------------------------------------

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ connectionId: string; objectType: string; recordId: string }> }
) {
  try {
    const { connectionId, objectType, recordId } = await params

    if (!objectType || !recordId) {
      return NextResponse.json(
        { error: 'objectType and recordId are required' },
        { status: 400 }
      )
    }

    const record = findRecord(objectType, recordId)

    if (!record) {
      return NextResponse.json(
        { error: `Record not found: ${objectType}/${recordId}` },
        { status: 404 }
      )
    }

    return NextResponse.json({
      record,
      objectType,
      connectionId,
    })
  } catch (error) {
    console.error('Error fetching record:', error)
    return NextResponse.json(
      { error: 'Failed to fetch record' },
      { status: 500 }
    )
  }
}
