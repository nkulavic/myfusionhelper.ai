'use client'

import { use } from 'react'
import { HelperDetail } from '../_components/helper-detail'

export default function HelperDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)
  return <HelperDetail helperId={id} />
}
