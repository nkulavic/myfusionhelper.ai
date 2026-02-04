'use client'

import { useState } from 'react'
import Link from 'next/link'
import {
  Search,
  Plus,
  Tag,
  Copy,
  Calculator,
  Calendar,
  Bell,
  FileSpreadsheet,
  Webhook,
  Users,
  Filter,
} from 'lucide-react'
import { cn } from '@/lib/utils'

const categories = [
  { id: 'all', name: 'All Helpers', count: 92 },
  { id: 'contact', name: 'Contact', count: 24 },
  { id: 'data', name: 'Data', count: 18 },
  { id: 'tagging', name: 'Tagging', count: 12 },
  { id: 'automation', name: 'Automation', count: 15 },
  { id: 'integration', name: 'Integration', count: 14 },
  { id: 'notification', name: 'Notification', count: 5 },
  { id: 'analytics', name: 'Analytics', count: 4 },
]

const helpers = [
  {
    id: 'tag-it',
    name: 'Tag It',
    description: 'Apply or remove tags based on conditions',
    category: 'tagging',
    icon: Tag,
    popular: true,
  },
  {
    id: 'copy-it',
    name: 'Copy It',
    description: 'Copy field values between contacts or to other fields',
    category: 'contact',
    icon: Copy,
    popular: true,
  },
  {
    id: 'date-calc',
    name: 'Date Calculator',
    description: 'Calculate dates, add/subtract days, find differences',
    category: 'data',
    icon: Calendar,
    popular: true,
  },
  {
    id: 'math',
    name: 'Math Helper',
    description: 'Perform calculations on field values',
    category: 'data',
    icon: Calculator,
    popular: false,
  },
  {
    id: 'notify-me',
    name: 'Notify Me',
    description: 'Send notifications via email, Slack, or SMS',
    category: 'notification',
    icon: Bell,
    popular: true,
  },
  {
    id: 'google-sheet-it',
    name: 'Google Sheet It',
    description: 'Sync contact data to Google Sheets',
    category: 'integration',
    icon: FileSpreadsheet,
    popular: true,
  },
  {
    id: 'hook-it',
    name: 'Hook It',
    description: 'Send data to external webhooks',
    category: 'integration',
    icon: Webhook,
    popular: false,
  },
  {
    id: 'merge-it',
    name: 'Merge It',
    description: 'Merge duplicate contacts intelligently',
    category: 'contact',
    icon: Users,
    popular: false,
  },
]

export default function HelpersPage() {
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [searchQuery, setSearchQuery] = useState('')

  const filteredHelpers = helpers.filter((helper) => {
    const matchesCategory = selectedCategory === 'all' || helper.category === selectedCategory
    const matchesSearch =
      helper.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      helper.description.toLowerCase().includes(searchQuery.toLowerCase())
    return matchesCategory && matchesSearch
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Helpers</h1>
          <p className="text-muted-foreground">
            Configure and manage your CRM automation helpers
          </p>
        </div>
        <Link
          href="/helpers/new"
          className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" />
          New Helper
        </Link>
      </div>

      {/* Search and Filters */}
      <div className="flex gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search helpers..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-10 w-full rounded-md border border-input bg-background pl-10 pr-4 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
        <button className="inline-flex items-center gap-2 rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent">
          <Filter className="h-4 w-4" />
          Filters
        </button>
      </div>

      {/* Category Tabs */}
      <div className="flex gap-2 overflow-x-auto pb-2">
        {categories.map((category) => (
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
                selectedCategory === category.id ? 'bg-primary-foreground/20' : 'bg-background'
              )}
            >
              {category.count}
            </span>
          </button>
        ))}
      </div>

      {/* Helpers Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {filteredHelpers.map((helper) => (
          <Link
            key={helper.id}
            href={`/helpers/${helper.id}`}
            className="group relative rounded-lg border bg-card p-6 transition-all hover:border-primary hover:shadow-md"
          >
            {helper.popular && (
              <span className="absolute right-4 top-4 rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                Popular
              </span>
            )}
            <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
              <helper.icon className="h-6 w-6 text-primary" />
            </div>
            <h3 className="mb-1 font-semibold group-hover:text-primary">{helper.name}</h3>
            <p className="text-sm text-muted-foreground">{helper.description}</p>
          </Link>
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
