'use client'

import { useState } from 'react'
import Link from 'next/link'
import {
  ArrowLeft,
  Plus,
  FileText,
  Copy,
  Trash2,
  Edit,
  Search,
  Sparkles,
  Mail,
  Star,
} from 'lucide-react'

const templates = [
  {
    id: 'tpl_001',
    name: 'Welcome Email',
    category: 'onboarding',
    subject: 'Welcome to {{company}}, {{first_name}}!',
    preview: 'Hi {{first_name}}, we are thrilled to have you on board...',
    usageCount: 1_247,
    isStarred: true,
    createdAt: '2025-12-15',
  },
  {
    id: 'tpl_002',
    name: 'Follow-up After Demo',
    category: 'sales',
    subject: 'Great connecting today, {{first_name}}',
    preview: 'Thank you for taking the time to see our demo today...',
    usageCount: 856,
    isStarred: true,
    createdAt: '2026-01-05',
  },
  {
    id: 'tpl_003',
    name: 'Payment Reminder',
    category: 'billing',
    subject: 'Friendly reminder: Invoice due soon',
    preview: 'Hi {{first_name}}, this is a quick reminder that...',
    usageCount: 423,
    isStarred: false,
    createdAt: '2026-01-10',
  },
  {
    id: 'tpl_004',
    name: 'Re-engagement Campaign',
    category: 'marketing',
    subject: 'We miss you, {{first_name}}!',
    preview: "It's been a while since we last connected. We wanted...",
    usageCount: 312,
    isStarred: false,
    createdAt: '2026-01-20',
  },
  {
    id: 'tpl_005',
    name: 'Monthly Newsletter',
    category: 'marketing',
    subject: 'Your {{company}} Monthly Update',
    preview: "Here's what happened this month and what's coming next...",
    usageCount: 2_100,
    isStarred: true,
    createdAt: '2025-11-01',
  },
  {
    id: 'tpl_006',
    name: 'Webinar Invitation',
    category: 'events',
    subject: "You're invited: {{event_name}}",
    preview: 'Join us for an exclusive webinar on automation strategies...',
    usageCount: 645,
    isStarred: false,
    createdAt: '2026-02-01',
  },
]

const categories = ['all', 'onboarding', 'sales', 'billing', 'marketing', 'events']

export default function EmailTemplatesPage() {
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('all')

  const filtered = templates.filter((t) => {
    if (category !== 'all' && t.category !== category) return false
    if (search && !t.name.toLowerCase().includes(search.toLowerCase()) && !t.subject.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href="/emails" className="rounded-md p-2 hover:bg-accent">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">Email Templates</h1>
          <p className="text-muted-foreground">Reusable email templates with AI generation</p>
        </div>
        <button className="inline-flex items-center gap-2 rounded-md border px-4 py-2 text-sm font-medium hover:bg-accent">
          <Sparkles className="h-4 w-4" />
          AI Generate
        </button>
        <button className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
          <Plus className="h-4 w-4" />
          New Template
        </button>
      </div>

      {/* Search and Filters */}
      <div className="flex gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search templates..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-md border bg-background py-2 pl-9 pr-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>
        <div className="flex gap-1 rounded-md border bg-background p-1">
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setCategory(cat)}
              className={`rounded px-3 py-1 text-xs font-medium capitalize transition-colors ${
                category === cat
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      {/* Template Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {filtered.map((template) => (
          <div
            key={template.id}
            className="group rounded-lg border bg-card p-5 transition-colors hover:border-primary"
          >
            <div className="mb-3 flex items-start justify-between">
              <div className="flex items-center gap-2">
                <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10">
                  <Mail className="h-4 w-4 text-primary" />
                </div>
                <div>
                  <h3 className="font-semibold text-sm">{template.name}</h3>
                  <span className="text-xs capitalize text-muted-foreground">{template.category}</span>
                </div>
              </div>
              <button className={`rounded p-1 hover:bg-accent ${template.isStarred ? 'text-warning' : 'text-muted-foreground'}`}>
                <Star className="h-4 w-4" fill={template.isStarred ? 'currentColor' : 'none'} />
              </button>
            </div>

            <p className="mb-1 text-xs font-medium text-muted-foreground">Subject:</p>
            <p className="mb-2 text-sm font-mono truncate">{template.subject}</p>

            <p className="text-xs text-muted-foreground line-clamp-2">{template.preview}</p>

            <div className="mt-4 flex items-center justify-between">
              <span className="text-xs text-muted-foreground">
                Used {template.usageCount.toLocaleString()} times
              </span>
              <div className="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                <button className="rounded p-1 hover:bg-accent" title="Edit">
                  <Edit className="h-3.5 w-3.5" />
                </button>
                <button className="rounded p-1 hover:bg-accent" title="Duplicate">
                  <Copy className="h-3.5 w-3.5" />
                </button>
                <button className="rounded p-1 hover:bg-accent text-destructive" title="Delete">
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
