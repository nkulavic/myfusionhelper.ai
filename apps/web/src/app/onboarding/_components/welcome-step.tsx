'use client'

import { Sparkles, Blocks, Link2, BarChart3 } from 'lucide-react'

interface WelcomeStepProps {
  firstName: string
  onNext: () => void
  onSkip: () => void
}

const highlights = [
  {
    icon: Link2,
    title: 'Connect Your CRM',
    description: 'Link Keap, GoHighLevel, ActiveCampaign, Ontraport, or HubSpot in under 60 seconds',
  },
  {
    icon: Blocks,
    title: '62 Pre-Built Helpers',
    description: 'Tag contacts, format data, score leads, sync to Google Sheets, and much more',
  },
  {
    icon: Sparkles,
    title: 'AI-Powered Insights',
    description: 'Get intelligent recommendations to optimize your automations and spot issues early',
  },
  {
    icon: BarChart3,
    title: 'Track Every Execution',
    description: 'Monitor success rates, debug failures, and measure the impact of every helper run',
  },
]

export function WelcomeStep({ firstName, onNext, onSkip }: WelcomeStepProps) {
  return (
    <div className="space-y-8">
      <div className="text-center">
        <h1 className="text-3xl font-bold">Welcome, {firstName}!</h1>
        <p className="mt-2 text-lg text-muted-foreground">
          Let&apos;s get you set up in just a few steps. You&apos;ll have your first automation
          running in minutes.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        {highlights.map((item) => (
          <div
            key={item.title}
            className="flex items-start gap-3 rounded-lg border bg-card p-4"
          >
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
              <item.icon className="h-5 w-5 text-primary" />
            </div>
            <div>
              <h3 className="font-medium">{item.title}</h3>
              <p className="mt-0.5 text-sm text-muted-foreground">{item.description}</p>
            </div>
          </div>
        ))}
      </div>

      <div className="flex items-center justify-between pt-4">
        <button
          onClick={onSkip}
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          I&apos;ll explore on my own
        </button>
        <button
          onClick={onNext}
          className="inline-flex items-center gap-2 rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Get Started
        </button>
      </div>
    </div>
  )
}
