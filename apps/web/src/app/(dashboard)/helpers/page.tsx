'use client'

import { Suspense, useCallback } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { AnimatePresence, motion } from 'framer-motion'
import { Blocks, Grid3X3 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { MyHelpersList } from './_components/my-helpers-list'
import { HelpersCatalog } from './_components/helpers-catalog'
import { HelperBuilder } from './_components/helper-builder'

type HelperView = 'my-helpers' | 'catalog' | 'new'

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

  const view: HelperView =
    (searchParams.get('view') as HelperView) || 'my-helpers'
  const helperTypeId = searchParams.get('type') || ''

  const navigate = useCallback(
    (newView: HelperView, typeId?: string) => {
      const params = new URLSearchParams()
      if (newView !== 'my-helpers') params.set('view', newView)
      if (typeId) params.set('type', typeId)
      const qs = params.toString()
      router.push(`/helpers${qs ? `?${qs}` : ''}`, { scroll: false })
    },
    [router]
  )

  const isListView = view === 'my-helpers' || view === 'catalog'

  return (
    <div>
      {/* Tabs -- only visible in list views */}
      {isListView && (
        <div className="mb-6 flex items-center gap-1 rounded-lg bg-muted p-1 w-fit">
          <button
            onClick={() => navigate('my-helpers')}
            className={cn(
              'inline-flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-colors',
              view === 'my-helpers'
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <Blocks className="h-4 w-4" />
            My Helpers
          </button>
          <button
            onClick={() => navigate('catalog')}
            className={cn(
              'inline-flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-colors',
              view === 'catalog'
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <Grid3X3 className="h-4 w-4" />
            Catalog
          </button>
        </div>
      )}

      <AnimatePresence mode="wait">
        {view === 'my-helpers' && (
          <motion.div key="my-helpers" {...slideBack} transition={transition}>
            <MyHelpersList
              onSelectHelper={(id) => router.push(`/helpers/${id}`)}
              onNewHelper={() => navigate('catalog')}
            />
          </motion.div>
        )}

        {view === 'catalog' && (
          <motion.div key="catalog" {...slideBack} transition={transition}>
            <HelpersCatalog
              onSelectHelper={(typeId) => navigate('new', typeId)}
              onNewHelper={() => navigate('new')}
            />
          </motion.div>
        )}

        {view === 'new' && (
          <motion.div
            key={`new-${helperTypeId}`}
            {...slideIn}
            transition={transition}
          >
            <HelperBuilder
              initialType={helperTypeId || undefined}
              onBack={() => navigate('catalog')}
              onCreated={(id) => router.push(`/helpers/${id}`)}
            />
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

function HelpersSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <div className="h-11 w-64 rounded-lg bg-muted" />
      <div className="flex items-center justify-between">
        <div>
          <div className="h-7 w-32 rounded bg-muted" />
          <div className="mt-1 h-4 w-64 rounded bg-muted" />
        </div>
        <div className="h-10 w-32 rounded-md bg-muted" />
      </div>
      <div className="grid gap-4 sm:grid-cols-3">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-20 rounded-lg bg-muted" />
        ))}
      </div>
      <div className="h-10 w-full rounded-md bg-muted" />
      <div className="space-y-3">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-[72px] rounded-lg bg-muted" />
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
