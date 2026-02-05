'use client'

import { useState, useMemo } from 'react'
import { Search, Plus, Filter, Sparkles } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  helpersCatalog,
  categoryInfo,
  getCategoryCounts,
  type HelperDefinition,
} from '@/lib/helpers-catalog'
import { getCRMPlatform } from '@/lib/crm-platforms'
import { CRMBadges } from '@/components/crm-badges'

interface HelpersCatalogProps {
  onSelectHelper: (id: string) => void
  onNewHelper: () => void
}

export function HelpersCatalog({ onSelectHelper, onNewHelper }: HelpersCatalogProps) {
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [showAvailableOnly, setShowAvailableOnly] = useState(false)

  const categoryCounts = useMemo(() => getCategoryCounts(), [])

  const filteredHelpers = useMemo(() => {
    return helpersCatalog.filter((helper) => {
      const matchesCategory =
        selectedCategory === 'all' || helper.category === selectedCategory
      const matchesSearch =
        searchQuery === '' ||
        helper.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        helper.description.toLowerCase().includes(searchQuery.toLowerCase())
      const matchesAvailability = !showAvailableOnly || helper.status === 'available'
      return matchesCategory && matchesSearch && matchesAvailability
    })
  }, [selectedCategory, searchQuery, showAvailableOnly])

  const availableCount = filteredHelpers.filter((h) => h.status === 'available').length
  const comingSoonCount = filteredHelpers.filter((h) => h.status === 'coming_soon').length

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Helpers</h1>
          <p className="text-muted-foreground">
            {helpersCatalog.length} automation helpers across{' '}
            {categoryInfo.length - 1} categories
          </p>
        </div>
        <button
          onClick={onNewHelper}
          className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" />
          New Helper
        </button>
      </div>

      {/* Search and Filters */}
      <div className="flex gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search helpers by name or description..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-10 w-full rounded-md border border-input bg-background pl-10 pr-4 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
        <button
          onClick={() => setShowAvailableOnly(!showAvailableOnly)}
          className={cn(
            'inline-flex items-center gap-2 rounded-md border px-4 py-2 text-sm font-medium transition-colors',
            showAvailableOnly
              ? 'border-primary bg-primary/10 text-primary'
              : 'border-input bg-background text-muted-foreground hover:bg-accent'
          )}
        >
          <Filter className="h-4 w-4" />
          Available Only
        </button>
      </div>

      {/* Category Tabs */}
      <div className="flex gap-2 overflow-x-auto pb-2">
        {categoryInfo.map((category) => {
          const count = categoryCounts[category.id] || 0
          return (
            <button
              key={category.id}
              onClick={() => setSelectedCategory(category.id)}
              className={cn(
                'inline-flex items-center gap-2 whitespace-nowrap rounded-full px-4 py-2 text-sm font-medium transition-colors',
                selectedCategory === category.id
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-muted text-muted-foreground hover:bg-muted/80'
              )}
            >
              {category.name}
              <span
                className={cn(
                  'rounded-full px-2 py-0.5 text-xs',
                  selectedCategory === category.id
                    ? 'bg-primary-foreground/20'
                    : 'bg-background'
                )}
              >
                {count}
              </span>
            </button>
          )
        })}
      </div>

      {/* Results Summary */}
      <div className="flex items-center gap-4 text-sm text-muted-foreground">
        <span>
          Showing {filteredHelpers.length} helper
          {filteredHelpers.length !== 1 ? 's' : ''}
        </span>
        <span className="text-border">|</span>
        <span className="text-success">{availableCount} available</span>
        {comingSoonCount > 0 && (
          <>
            <span className="text-border">|</span>
            <span className="text-warning">{comingSoonCount} coming soon</span>
          </>
        )}
      </div>

      {/* Helpers Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {filteredHelpers.map((helper) => (
          <HelperCard
            key={helper.id}
            helper={helper}
            onSelect={onSelectHelper}
          />
        ))}
      </div>

      {filteredHelpers.length === 0 && (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Search className="mb-4 h-12 w-12 text-muted-foreground/50" />
          <h3 className="mb-1 font-semibold">No helpers found</h3>
          <p className="text-sm text-muted-foreground">
            Try adjusting your search or filter criteria
          </p>
        </div>
      )}
    </div>
  )
}

function HelperCard({
  helper,
  onSelect,
}: {
  helper: HelperDefinition
  onSelect: (id: string) => void
}) {
  const isAvailable = helper.status === 'available'

  // Determine accent color from first supported CRM, or use primary
  const accentCRM =
    helper.supportedCRMs.length > 0
      ? getCRMPlatform(helper.supportedCRMs[0])
      : null

  return (
    <button
      onClick={() => isAvailable && onSelect(helper.id)}
      className={cn(
        'group relative flex flex-col overflow-hidden rounded-lg border bg-card p-5 text-left transition-all',
        isAvailable
          ? 'hover:-translate-y-0.5 hover:shadow-md cursor-pointer'
          : 'opacity-60 cursor-default'
      )}
    >
      {/* Bottom accent stripe — gradient from brand green to blue */}
      <div
        className="absolute inset-x-0 bottom-0 h-[3px]"
        style={{
          background: accentCRM
            ? `linear-gradient(to right, ${accentCRM.color}, hsl(var(--primary)))`
            : 'linear-gradient(to right, hsl(var(--success)), hsl(var(--primary)))',
          opacity: isAvailable ? 1 : 0.4,
        }}
      />

      {/* Badges */}
      <div className="absolute right-3 top-3 flex gap-1.5">
        {helper.popular && (
          <span className="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
            <Sparkles className="h-3 w-3" />
            Popular
          </span>
        )}
        {helper.status === 'coming_soon' && (
          <span className="rounded-full bg-warning/10 px-2 py-0.5 text-xs font-medium text-warning">
            Soon
          </span>
        )}
        {helper.status === 'beta' && (
          <span className="rounded-full bg-info/10 px-2 py-0.5 text-xs font-medium text-info">
            Beta
          </span>
        )}
      </div>

      {/* Icon */}
      <div
        className="mb-3 flex h-10 w-10 items-center justify-center rounded-lg"
        style={{
          backgroundColor: accentCRM
            ? `${accentCRM.color}18`
            : 'hsl(var(--primary) / 0.1)',
        }}
      >
        <helper.icon
          className="h-5 w-5"
          style={{
            color: accentCRM ? accentCRM.color : 'hsl(var(--primary))',
          }}
        />
      </div>

      {/* Content */}
      <h3 className="mb-1 text-sm font-semibold group-hover:text-primary">
        {helper.name}
      </h3>
      <p className="flex-1 text-xs text-muted-foreground leading-relaxed">
        {helper.description}
      </p>

      {/* CRM Badges */}
      <div className="mt-auto pt-3 border-t border-border/50">
        <CRMBadges crmIds={helper.supportedCRMs} />
      </div>
    </button>
  )
}
