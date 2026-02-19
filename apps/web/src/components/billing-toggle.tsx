'use client'

import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'

interface BillingToggleProps {
  billingPeriod: 'monthly' | 'annual'
  onChange: (period: 'monthly' | 'annual') => void
  maxSavings: number
}

export function BillingToggle({ billingPeriod, onChange, maxSavings }: BillingToggleProps) {
  const isAnnual = billingPeriod === 'annual'

  return (
    <div className="flex items-center gap-3">
      <span
        className={cn(
          'text-sm font-medium transition-colors',
          !isAnnual ? 'text-foreground' : 'text-muted-foreground',
        )}
      >
        Monthly
      </span>
      <button
        type="button"
        role="switch"
        aria-checked={isAnnual}
        onClick={() => onChange(isAnnual ? 'monthly' : 'annual')}
        className={cn(
          'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors',
          isAnnual ? 'bg-primary' : 'bg-muted-foreground/30',
        )}
      >
        <span
          className={cn(
            'pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow-sm ring-0 transition-transform',
            isAnnual ? 'translate-x-5' : 'translate-x-0',
          )}
        />
      </button>
      <span
        className={cn(
          'text-sm font-medium transition-colors',
          isAnnual ? 'text-foreground' : 'text-muted-foreground',
        )}
      >
        Annual
      </span>
      {isAnnual && (
        <Badge variant="secondary" className="text-xs">
          Save up to {maxSavings}%
        </Badge>
      )}
    </div>
  )
}
