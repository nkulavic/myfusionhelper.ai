'use client'

import { useState } from 'react'
import { ArrowLeft, CheckCircle } from 'lucide-react'
import { helpersCatalog, categoryInfo, type HelperDefinition } from '@/lib/helpers-catalog'
import { cn } from '@/lib/utils'

interface PickHelperStepProps {
  onNext: () => void
  onBack: () => void
  onSkip: () => void
}

// Show popular helpers first, limited to the most useful for onboarding
const popularHelpers = helpersCatalog
  .filter((h) => h.popular && h.status === 'available')
  .slice(0, 8)

const categories = categoryInfo.filter((c) => c.id !== 'all')

export function PickHelperStep({ onNext, onBack, onSkip }: PickHelperStepProps) {
  const [selectedHelpers, setSelectedHelpers] = useState<Set<string>>(new Set())
  const [activeCategory, setActiveCategory] = useState<string>('popular')

  const toggleHelper = (id: string) => {
    setSelectedHelpers((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const displayHelpers: HelperDefinition[] =
    activeCategory === 'popular'
      ? popularHelpers
      : helpersCatalog.filter((h) => h.category === activeCategory && h.status === 'available').slice(0, 8)

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h2 className="text-2xl font-bold">Choose Your First Helpers</h2>
        <p className="mt-1 text-muted-foreground">
          Pick the helpers you want to try first. You can always add more later.
        </p>
      </div>

      {/* Category tabs */}
      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => setActiveCategory('popular')}
          className={cn(
            'rounded-full px-3 py-1.5 text-xs font-medium transition-colors',
            activeCategory === 'popular'
              ? 'bg-primary text-primary-foreground'
              : 'bg-muted text-muted-foreground hover:bg-accent hover:text-foreground'
          )}
        >
          Popular
        </button>
        {categories.map((cat) => (
          <button
            key={cat.id}
            onClick={() => setActiveCategory(cat.id)}
            className={cn(
              'rounded-full px-3 py-1.5 text-xs font-medium transition-colors',
              activeCategory === cat.id
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-muted-foreground hover:bg-accent hover:text-foreground'
            )}
          >
            {cat.name}
          </button>
        ))}
      </div>

      {/* Helper grid */}
      <div className="grid gap-3 sm:grid-cols-2">
        {displayHelpers.map((helper) => {
          const isSelected = selectedHelpers.has(helper.id)
          return (
            <button
              key={helper.id}
              onClick={() => toggleHelper(helper.id)}
              className={cn(
                'flex items-start gap-3 rounded-lg border p-3.5 text-left transition-all',
                isSelected
                  ? 'border-primary bg-primary/5 ring-1 ring-primary'
                  : 'bg-card hover:border-primary/50 hover:shadow-sm'
              )}
            >
              <div
                className={cn(
                  'flex h-9 w-9 shrink-0 items-center justify-center rounded-lg transition-colors',
                  isSelected ? 'bg-primary text-primary-foreground' : 'bg-muted'
                )}
              >
                <helper.icon className="h-4 w-4" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <h3 className="text-sm font-medium">{helper.name}</h3>
                  {isSelected && <CheckCircle className="h-4 w-4 text-primary" />}
                </div>
                <p className="mt-0.5 text-xs text-muted-foreground line-clamp-2">
                  {helper.description}
                </p>
              </div>
            </button>
          )
        })}
      </div>

      {selectedHelpers.size > 0 && (
        <p className="text-center text-sm text-muted-foreground">
          {selectedHelpers.size} helper{selectedHelpers.size !== 1 ? 's' : ''} selected
        </p>
      )}

      {/* Navigation */}
      <div className="flex items-center justify-between pt-2">
        <button
          onClick={onBack}
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Back
        </button>
        <div className="flex items-center gap-3">
          <button
            onClick={onSkip}
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            Skip for now
          </button>
          <button
            onClick={onNext}
            className="inline-flex items-center gap-2 rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            {selectedHelpers.size > 0 ? 'Continue' : 'Skip & Continue'}
          </button>
        </div>
      </div>
    </div>
  )
}
