'use client'

import { Sparkles, TrendingUp, TrendingDown, AlertTriangle, Lightbulb, BarChart3 } from 'lucide-react'

const insights = [
  {
    id: '1',
    type: 'opportunity',
    title: 'Untagged contacts detected',
    description: '47 contacts have no tags assigned. Consider segmenting them for better targeting.',
    action: 'View contacts',
    icon: AlertTriangle,
    color: 'text-yellow-500',
  },
  {
    id: '2',
    type: 'trend',
    title: 'Helper usage up 23%',
    description: 'Your automation usage increased significantly this week compared to last.',
    action: 'View report',
    icon: TrendingUp,
    color: 'text-green-500',
  },
  {
    id: '3',
    type: 'suggestion',
    title: 'Try the Date Calculator helper',
    description: 'Based on your workflow, automating date calculations could save you time.',
    action: 'Learn more',
    icon: Lightbulb,
    color: 'text-blue-500',
  },
  {
    id: '4',
    type: 'alert',
    title: 'API rate limit warning',
    description: "You're approaching your Keap API limit. Consider spacing out automations.",
    action: 'View usage',
    icon: AlertTriangle,
    color: 'text-red-500',
  },
]

const metrics = [
  { label: 'Contacts Processed', value: '12,847', change: '+12%', trend: 'up' },
  { label: 'Automation Runs', value: '3,421', change: '+8%', trend: 'up' },
  { label: 'Time Saved', value: '47 hrs', change: '+15%', trend: 'up' },
  { label: 'Error Rate', value: '1.2%', change: '-0.3%', trend: 'down' },
]

export default function InsightsPage() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold">Insights</h1>
        <p className="text-muted-foreground">AI-powered analytics and recommendations</p>
      </div>

      {/* Metrics */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {metrics.map((metric) => (
          <div key={metric.label} className="rounded-lg border bg-card p-4">
            <p className="text-sm text-muted-foreground">{metric.label}</p>
            <div className="flex items-baseline gap-2">
              <p className="text-2xl font-bold">{metric.value}</p>
              <span
                className={`flex items-center text-sm ${
                  metric.trend === 'up' ? 'text-green-500' : 'text-red-500'
                }`}
              >
                {metric.trend === 'up' ? (
                  <TrendingUp className="mr-1 h-3 w-3" />
                ) : (
                  <TrendingDown className="mr-1 h-3 w-3" />
                )}
                {metric.change}
              </span>
            </div>
          </div>
        ))}
      </div>

      {/* AI Insights */}
      <div>
        <div className="mb-4 flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold">AI Recommendations</h2>
        </div>
        <div className="grid gap-4 lg:grid-cols-2">
          {insights.map((insight) => (
            <div key={insight.id} className="rounded-lg border bg-card p-4">
              <div className="mb-3 flex items-start gap-3">
                <div className={`rounded-lg bg-muted p-2 ${insight.color}`}>
                  <insight.icon className="h-5 w-5" />
                </div>
                <div className="flex-1">
                  <h3 className="font-semibold">{insight.title}</h3>
                  <p className="mt-1 text-sm text-muted-foreground">{insight.description}</p>
                </div>
              </div>
              <button className="text-sm font-medium text-primary hover:underline">
                {insight.action} &rarr;
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Quick Reports */}
      <div>
        <div className="mb-4 flex items-center gap-2">
          <BarChart3 className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold">Quick Reports</h2>
        </div>
        <div className="grid gap-4 sm:grid-cols-3">
          <button className="flex flex-col items-center rounded-lg border bg-card p-6 text-center hover:border-primary hover:shadow-md">
            <span className="text-2xl">📊</span>
            <span className="mt-2 font-medium">Weekly Summary</span>
            <span className="text-sm text-muted-foreground">Last 7 days overview</span>
          </button>
          <button className="flex flex-col items-center rounded-lg border bg-card p-6 text-center hover:border-primary hover:shadow-md">
            <span className="text-2xl">🏷️</span>
            <span className="mt-2 font-medium">Tag Analysis</span>
            <span className="text-sm text-muted-foreground">Tag distribution report</span>
          </button>
          <button className="flex flex-col items-center rounded-lg border bg-card p-6 text-center hover:border-primary hover:shadow-md">
            <span className="text-2xl">⚡</span>
            <span className="mt-2 font-medium">Helper Performance</span>
            <span className="text-sm text-muted-foreground">Execution analytics</span>
          </button>
        </div>
      </div>

      {/* AI Chat Prompt */}
      <div className="rounded-lg border bg-gradient-to-r from-primary/10 to-primary/5 p-6">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-primary-foreground">
            <Sparkles className="h-5 w-5" />
          </div>
          <div className="flex-1">
            <h3 className="font-semibold">Ask AI anything about your data</h3>
            <p className="text-sm text-muted-foreground">
              Try: &quot;Show me contacts who haven&apos;t been emailed in 30 days&quot;
            </p>
          </div>
          <button className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
            Open Chat
          </button>
        </div>
      </div>
    </div>
  )
}
