'use client'

import { Suspense, useCallback } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { AnimatePresence, motion } from 'framer-motion'
import { HelpersCatalog } from './_components/helpers-catalog'
import { HelperDetail } from './_components/helper-detail'
import { HelperBuilder } from './_components/helper-builder'

type HelperView = 'catalog' | 'detail' | 'new'

const slideIn = {
  initial: { opacity: 0, x: 24 },
  animate: { opacity: 1, x: 0 },
  exit: { opacity: 0, x: -24 },
}

const slideBack = {
  initial: { opacity: 0, x: -24 },
  animate: { opacity: 1, x: 0 },
  exit: { opacity: 0, x: 24 },
}

const transition = {
  duration: 0.2,
  ease: [0.25, 0.46, 0.45, 0.94] as [number, number, number, number],
}

function HelpersContent() {
  const searchParams = useSearchParams()
  const router = useRouter()

  const view: HelperView = (searchParams.get('view') as HelperView) || 'catalog'
  const helperId = searchParams.get('id') || ''

  const navigate = useCallback(
    (newView: HelperView, id?: string) => {
      const params = new URLSearchParams()
      if (newView !== 'catalog') params.set('view', newView)
      if (id) params.set('id', id)
      const qs = params.toString()
      router.push(`/helpers${qs ? `?${qs}` : ''}`, { scroll: false })
    },
    [router]
  )

  return (
    <AnimatePresence mode="wait">
      {view === 'catalog' && (
        <motion.div
          key="catalog"
          {...slideBack}
          transition={transition}
        >
          <HelpersCatalog
            onSelectHelper={(id) => navigate('new', id)}
            onNewHelper={() => navigate('new')}
          />
        </motion.div>
      )}

      {view === 'detail' && helperId && (
        <motion.div
          key={`detail-${helperId}`}
          {...slideIn}
          transition={transition}
        >
          <HelperDetail
            helperId={helperId}
            onBack={() => navigate('catalog')}
          />
        </motion.div>
      )}

      {view === 'new' && (
        <motion.div
          key={`new-${helperId}`}
          {...slideIn}
          transition={transition}
        >
          <HelperBuilder
            initialType={helperId || undefined}
            onBack={() => navigate('catalog')}
            onCreated={(id) => navigate('detail', id)}
          />
        </motion.div>
      )}
    </AnimatePresence>
  )
}

function HelpersSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <div className="flex items-center justify-between">
        <div>
          <div className="h-7 w-32 rounded bg-muted" />
          <div className="mt-1 h-4 w-64 rounded bg-muted" />
        </div>
        <div className="h-10 w-32 rounded-md bg-muted" />
      </div>
      <div className="h-10 w-full rounded-md bg-muted" />
      <div className="flex gap-2">
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="h-9 w-24 rounded-full bg-muted" />
        ))}
      </div>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
          <div key={i} className="h-36 rounded-lg bg-muted" />
        ))}
      </div>
    </div>
  )
}

export default function HelpersPage() {
  return (
    <Suspense fallback={<HelpersSkeleton />}>
      <HelpersContent />
    </Suspense>
  )
}
