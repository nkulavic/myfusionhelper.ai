'use client'

import { useState } from 'react'
import {
  Blocks,
  Link2,
  History,
  Settings,
  Sparkles,
  Database,
  MessageSquare,
  ArrowLeft,
  ArrowRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface QuickTourStepProps {
  onFinish: () => void
  onBack: () => void
}

const tourItems = [
  {
    icon: Blocks,
    title: 'Helpers',
    path: '/helpers',
    description:
      'Your automation toolkit. Browse 62 helpers across 7 categories, configure them through a visual UI, and activate with a single click. Each helper runs automatically when triggered from your CRM.',
  },
  {
    icon: Link2,
    title: 'Connections',
    path: '/connections',
    description:
      'Manage your CRM connections here. Add new platforms, test connections, and re-authorize when needed. Each helper needs a connection to work.',
  },
  {
    icon: History,
    title: 'Executions',
    path: '/executions',
    description:
      'See every helper run in real time. Filter by status, drill into details, and debug any failures. Track how many contacts each helper has processed.',
  },
  {
    icon: Database,
    title: 'Data Explorer',
    path: '/data-explorer',
    description:
      'Browse your CRM data directly. Query contacts, view records, and explore your data with natural language search.',
  },
  {
    icon: Sparkles,
    title: 'AI Insights',
    path: '/insights',
    description:
      'AI-powered suggestions for your workflows. Get recommendations on which helpers to use, spot anomalies, and optimize your automations.',
  },
  {
    icon: MessageSquare,
    title: 'AI Chat',
    path: '#',
    description:
      'Click the chat bubble in the bottom-right corner anytime to ask questions, get help configuring helpers, or analyze your CRM data.',
  },
  {
    icon: Settings,
    title: 'Settings',
    path: '/settings',
    description:
      'Manage your profile, API keys, team members, billing, and notification preferences all in one place.',
  },
]

export function QuickTourStep({ onFinish, onBack }: QuickTourStepProps) {
  const [activeTour, setActiveTour] = useState(0)

  const current = tourItems[activeTour]

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h2 className="text-2xl font-bold">Quick Tour</h2>
        <p className="mt-1 text-muted-foreground">
          Here&apos;s what you can do with MyFusion Helper
        </p>
      </div>

      {/* Tour navigation dots */}
      <div className="flex items-center justify-center gap-3">
        {tourItems.map((item, i) => (
          <button
            key={item.title}
            onClick={() => setActiveTour(i)}
            className={cn(
              'flex h-10 w-10 items-center justify-center rounded-lg transition-all',
              i === activeTour
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'bg-muted text-muted-foreground hover:bg-accent'
            )}
            title={item.title}
          >
            <item.icon className="h-5 w-5" />
          </button>
        ))}
      </div>

      {/* Current tour item */}
      <div className="rounded-lg border bg-card p-6 text-center">
        <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-xl bg-primary/10">
          <current.icon className="h-7 w-7 text-primary" />
        </div>
        <h3 className="text-lg font-semibold">{current.title}</h3>
        {current.path !== '#' && (
          <p className="mt-1 font-mono text-xs text-muted-foreground">{current.path}</p>
        )}
        <p className="mx-auto mt-3 max-w-md text-sm text-muted-foreground">
          {current.description}
        </p>

        {/* Mini pagination */}
        <div className="mt-5 flex items-center justify-center gap-2">
          <button
            onClick={() => setActiveTour(Math.max(0, activeTour - 1))}
            disabled={activeTour === 0}
            className="rounded-md p-1.5 hover:bg-accent disabled:opacity-30"
          >
            <ArrowLeft className="h-4 w-4" />
          </button>
          <span className="text-xs text-muted-foreground">
            {activeTour + 1} / {tourItems.length}
          </span>
          <button
            onClick={() => setActiveTour(Math.min(tourItems.length - 1, activeTour + 1))}
            disabled={activeTour === tourItems.length - 1}
            className="rounded-md p-1.5 hover:bg-accent disabled:opacity-30"
          >
            <ArrowRight className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Navigation */}
      <div className="flex items-center justify-between pt-2">
        <button
          onClick={onBack}
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Back
        </button>
        <button
          onClick={onFinish}
          className="inline-flex items-center gap-2 rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Go to Dashboard
        </button>
      </div>
    </div>
  )
}
