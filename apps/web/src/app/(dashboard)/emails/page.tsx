'use client'

import { useState } from 'react'
import Link from 'next/link'
import {
  Mail,
  Plus,
  Sparkles,
  Send,
  FileText,
  Copy,
  Trash2,
  Clock,
  ArrowRight,
  Search,
  Bold,
  Italic,
  Link2,
  List,
  AlignLeft,
  Wand2,
} from 'lucide-react'

const toneOptions = [
  { id: 'professional', label: 'Professional' },
  { id: 'friendly', label: 'Friendly' },
  { id: 'urgent', label: 'Urgent' },
  { id: 'casual', label: 'Casual' },
]

const personalizationTokens = [
  { token: '{{first_name}}', label: 'First Name' },
  { token: '{{last_name}}', label: 'Last Name' },
  { token: '{{email}}', label: 'Email' },
  { token: '{{company}}', label: 'Company' },
  { token: '{{phone}}', label: 'Phone' },
]

const recentEmails = [
  {
    id: 'em_001',
    subject: 'Welcome to our community!',
    to: 'New Leads Segment',
    status: 'sent',
    sentAt: '2026-02-04T14:00:00Z',
    opens: 847,
    clicks: 234,
  },
  {
    id: 'em_002',
    subject: 'Your weekly automation report',
    to: 'All Contacts',
    status: 'scheduled',
    sentAt: '2026-02-07T09:00:00Z',
    opens: 0,
    clicks: 0,
  },
  {
    id: 'em_003',
    subject: 'Special offer for loyal customers',
    to: 'VIP Segment',
    status: 'draft',
    sentAt: null,
    opens: 0,
    clicks: 0,
  },
]

export default function EmailsPage() {
  const [activeTab, setActiveTab] = useState<'compose' | 'sent' | 'templates'>('compose')
  const [tone, setTone] = useState('professional')
  const [subject, setSubject] = useState('')
  const [body, setBody] = useState('')
  const [aiPrompt, setAiPrompt] = useState('')

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Emails</h1>
          <p className="text-muted-foreground">AI-powered email composer and templates</p>
        </div>
        <Link
          href="/emails/templates"
          className="inline-flex items-center gap-2 rounded-md border px-4 py-2 text-sm font-medium hover:bg-accent"
        >
          <FileText className="h-4 w-4" />
          Templates
        </Link>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 rounded-lg border bg-muted p-1">
        {(['compose', 'sent', 'templates'] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`flex-1 rounded-md px-4 py-2 text-sm font-medium capitalize transition-colors ${
              activeTab === tab
                ? 'bg-background shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {tab === 'compose' ? 'Compose' : tab === 'sent' ? 'Sent & Drafts' : 'Templates'}
          </button>
        ))}
      </div>

      {activeTab === 'compose' && (
        <div className="grid gap-6 lg:grid-cols-3">
          {/* Email Composer */}
          <div className="lg:col-span-2 space-y-4">
            <div className="rounded-lg border bg-card">
              {/* Subject */}
              <div className="border-b px-4 py-3">
                <input
                  type="text"
                  placeholder="Email subject..."
                  value={subject}
                  onChange={(e) => setSubject(e.target.value)}
                  className="w-full bg-transparent text-lg font-medium placeholder:text-muted-foreground focus:outline-none"
                />
              </div>

              {/* Toolbar */}
              <div className="flex items-center gap-1 border-b px-4 py-2">
                <button className="rounded p-1.5 hover:bg-accent"><Bold className="h-4 w-4" /></button>
                <button className="rounded p-1.5 hover:bg-accent"><Italic className="h-4 w-4" /></button>
                <button className="rounded p-1.5 hover:bg-accent"><Link2 className="h-4 w-4" /></button>
                <button className="rounded p-1.5 hover:bg-accent"><List className="h-4 w-4" /></button>
                <button className="rounded p-1.5 hover:bg-accent"><AlignLeft className="h-4 w-4" /></button>
                <div className="mx-2 h-5 w-px bg-border" />
                <button className="inline-flex items-center gap-1 rounded px-2 py-1 text-xs font-medium text-primary hover:bg-primary/10">
                  <Sparkles className="h-3 w-3" />
                  AI Improve
                </button>
              </div>

              {/* Body */}
              <textarea
                rows={14}
                placeholder="Write your email here... Use personalization tokens like {{first_name}} for dynamic content."
                value={body}
                onChange={(e) => setBody(e.target.value)}
                className="w-full resize-none bg-transparent p-4 text-sm placeholder:text-muted-foreground focus:outline-none"
              />

              {/* Actions */}
              <div className="flex items-center justify-between border-t px-4 py-3">
                <div className="flex items-center gap-2">
                  <button className="inline-flex items-center gap-1 rounded-md border px-3 py-1.5 text-xs font-medium hover:bg-accent">
                    <Clock className="h-3 w-3" />
                    Schedule
                  </button>
                  <button className="inline-flex items-center gap-1 rounded-md border px-3 py-1.5 text-xs font-medium hover:bg-accent">
                    <Copy className="h-3 w-3" />
                    Save as Template
                  </button>
                </div>
                <button className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
                  <Send className="h-4 w-4" />
                  Send Email
                </button>
              </div>
            </div>
          </div>

          {/* Sidebar: AI + Tokens */}
          <div className="space-y-4">
            {/* AI Composer */}
            <div className="rounded-lg border bg-card p-4">
              <div className="mb-3 flex items-center gap-2">
                <Sparkles className="h-5 w-5 text-primary" />
                <h3 className="font-semibold">AI Composer</h3>
              </div>
              <p className="mb-3 text-xs text-muted-foreground">
                Describe what you want to say and AI will write it for you.
              </p>
              <textarea
                rows={3}
                placeholder='e.g. "Write a follow-up email for leads who attended our webinar"'
                value={aiPrompt}
                onChange={(e) => setAiPrompt(e.target.value)}
                className="mb-3 w-full rounded-md border bg-background p-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
              />

              {/* Tone */}
              <div className="mb-3">
                <p className="mb-1.5 text-xs font-medium text-muted-foreground">Tone</p>
                <div className="flex flex-wrap gap-1.5">
                  {toneOptions.map((t) => (
                    <button
                      key={t.id}
                      onClick={() => setTone(t.id)}
                      className={`rounded-full px-2.5 py-1 text-xs font-medium transition-colors ${
                        tone === t.id
                          ? 'bg-primary text-primary-foreground'
                          : 'bg-muted text-muted-foreground hover:text-foreground'
                      }`}
                    >
                      {t.label}
                    </button>
                  ))}
                </div>
              </div>

              <button className="inline-flex w-full items-center justify-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
                <Wand2 className="h-4 w-4" />
                Generate Email
              </button>
            </div>

            {/* Subject Line Generator */}
            <div className="rounded-lg border bg-card p-4">
              <h3 className="mb-2 font-semibold text-sm">Subject Line Ideas</h3>
              <button className="mb-3 inline-flex w-full items-center justify-center gap-1 rounded-md border px-3 py-1.5 text-xs font-medium hover:bg-accent">
                <Sparkles className="h-3 w-3" />
                Generate Subject Lines
              </button>
              <div className="space-y-2">
                {[
                  'Unlock your CRM potential today',
                  "Don't miss out: exclusive automation tips",
                  '{{first_name}}, your weekly report is ready',
                ].map((s, i) => (
                  <button
                    key={i}
                    onClick={() => setSubject(s)}
                    className="w-full rounded-md border px-3 py-2 text-left text-xs hover:bg-accent"
                  >
                    {s}
                  </button>
                ))}
              </div>
            </div>

            {/* Personalization Tokens */}
            <div className="rounded-lg border bg-card p-4">
              <h3 className="mb-2 font-semibold text-sm">Personalization</h3>
              <p className="mb-2 text-xs text-muted-foreground">
                Click to insert a token at the cursor position.
              </p>
              <div className="flex flex-wrap gap-1.5">
                {personalizationTokens.map((pt) => (
                  <button
                    key={pt.token}
                    onClick={() => setBody((prev) => prev + pt.token)}
                    className="rounded-md bg-muted px-2 py-1 text-xs font-mono hover:bg-accent"
                  >
                    {pt.token}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'sent' && (
        <div className="divide-y rounded-lg border bg-card">
          {recentEmails.map((email) => (
            <div key={email.id} className="flex items-center gap-4 px-5 py-4">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                <Mail className="h-5 w-5 text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="font-medium truncate">{email.subject}</p>
                <p className="text-sm text-muted-foreground">To: {email.to}</p>
              </div>
              <div className="flex items-center gap-6 flex-shrink-0">
                <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                  email.status === 'sent'
                    ? 'bg-success/10 text-success'
                    : email.status === 'scheduled'
                    ? 'bg-info/10 text-info'
                    : 'bg-muted text-muted-foreground'
                }`}>
                  {email.status}
                </span>
                {email.status === 'sent' && (
                  <div className="text-right">
                    <p className="text-xs text-muted-foreground">{email.opens} opens</p>
                    <p className="text-xs text-muted-foreground">{email.clicks} clicks</p>
                  </div>
                )}
                <div className="text-xs text-muted-foreground">
                  {email.sentAt ? new Date(email.sentAt).toLocaleDateString() : 'Draft'}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {activeTab === 'templates' && (
        <div className="text-center py-12">
          <FileText className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-4 text-lg font-semibold">Email Templates</h3>
          <p className="mt-2 text-sm text-muted-foreground">
            Manage your email templates for consistent communication.
          </p>
          <Link
            href="/emails/templates"
            className="mt-4 inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            Go to Templates
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      )}
    </div>
  )
}
