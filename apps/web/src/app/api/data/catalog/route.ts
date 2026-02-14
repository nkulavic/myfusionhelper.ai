import { NextResponse } from 'next/server'
import { getMockCatalog } from '@/lib/mock-data/crm-mock-data'
import type { DataCatalogResponse } from '@/lib/api/data-explorer'

export async function GET() {
  try {
    const sources = getMockCatalog()
    const response: DataCatalogResponse = { sources }
    return NextResponse.json(response)
  } catch (error) {
    console.error('Error fetching catalog:', error)
    return NextResponse.json(
      { error: 'Failed to fetch data catalog' },
      { status: 500 }
    )
  }
}
