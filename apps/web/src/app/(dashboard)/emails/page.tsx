'use client'

import { useState, useMemo } from 'react'
import Link from 'next/link'
import {
  Mail,
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
  CheckCircle,
  Loader2,
  MailOpen,
  MousePointerClick,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { useEmails, useSendEmail, useDeleteEmail } from '@/lib/hooks/use-emails'

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

const statusFilters = ['all', 'sent', 'scheduled', 'draft', 'failed'] as const

export default function EmailsPage() {
  const [activeTab, setActiveTab] = useState<'compose' | 'sent' | 'templates'>('compose')
  const [tone, setTone] = useState('professional')
  const [to, setTo] = useState('')
  const [subject, setSubject] = useState('')
  const [body, setBody] = useState('')
  const [aiPrompt, setAiPrompt] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')

  const { data: emails, isLoading: emailsLoading, error: emailsError } = useEmails()
  const sendEmail = useSendEmail()
  const deleteEmail = useDeleteEmail()

  const filteredEmails = useMemo(() => {
    if (!emails) return []
    return emails.filter((email) => {
      if (statusFilter !== 'all' && email.status !== statusFilter) return false
      if (searchQuery) {
        const q = searchQuery.toLowerCase()
        return (
          email.subject.toLowerCase().includes(q) ||
          email.to.toLowerCase().includes(q)
        )
      }
      return true
    })
  }, [emails, statusFilter, searchQuery])

  const handleSendEmail = () => {
    if (!subject.trim() || !to.trim()) return
    sendEmail.mutate(
      { to, subject, body },
      {
        onSuccess: () => {
          setTo('')
          setSubject('')
          setBody('')
          setActiveTab('sent')
        },
      }
    )
  }

  const handleDeleteEmail = (id: string) => {
    deleteEmail.mutate(id)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Emails</h1>
          <p className="text-muted-foreground">AI-powered email composer and templates</p>
        </div>
        <Link href="/emails/templates">
          <Button variant="outline">
            <FileText className="h-4 w-4" />
            Templates
          </Button>
        </Link>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 rounded-lg border bg-muted p-1">
        {(['compose', 'sent', 'templates'] as const).map((tab) => (
          <Button
            key={tab}
            variant="ghost"
            onClick={() => setActiveTab(tab)}
            className={cn(
              'flex-1',
              activeTab === tab
                ? 'bg-background shadow-sm hover:bg-background'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            {tab === 'compose' ? 'Compose' : tab === 'sent' ? 'Sent & Drafts' : 'Templates'}
          </Button>
        ))}
      </div>

      {activeTab === 'compose' && (
        <div className="grid gap-6 lg:grid-cols-3">
          {/* Email Composer */}
          <div className="lg:col-span-2 space-y-4">
            <div className="rounded-lg border bg-card">
              {/* To field */}
              <div className="border-b px-4 py-3">
                <Input
                  type="text"
                  placeholder="To (segment, contact, or email)..."
                  value={to}
                  onChange={(e) => setTo(e.target.value)}
                  className="border-0 bg-transparent shadow-none focus-visible:ring-0 focus-visible:ring-offset-0"
                />
              </div>

              {/* Subject */}
              <div className="border-b px-4 py-3">
                <Input
                  type="text"
                  placeholder="Email subject..."
                  value={subject}
                  onChange={(e) => setSubject(e.target.value)}
                  className="border-0 bg-transparent text-lg font-medium shadow-none focus-visible:ring-0 focus-visible:ring-offset-0"
                />
              </div>

              {/* Toolbar */}
              <div className="flex items-center gap-1 border-b px-4 py-2">
                <Button variant="ghost" size="icon" className="h-8 w-8" aria-label="Bold">
                  <Bold className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon" className="h-8 w-8" aria-label="Italic">
                  <Italic className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon" className="h-8 w-8" aria-label="Insert link">
                  <Link2 className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon" className="h-8 w-8" aria-label="Bulleted list">
                  <List className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon" className="h-8 w-8" aria-label="Align text">
                  <AlignLeft className="h-4 w-4" />
                </Button>
                <div className="mx-2 h-5 w-px bg-border" />
                <Button variant="ghost" size="sm" className="text-primary hover:bg-primary/10 hover:text-primary">
                  <Sparkles className="h-3 w-3" />
                  AI Improve
                </Button>
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
                  <Button variant="outline" size="sm">
                    <Clock className="h-3 w-3" />
                    Schedule
                  </Button>
                  <Button variant="outline" size="sm">
                    <Copy className="h-3 w-3" />
                    Save as Template
                  </Button>
                </div>
                <Button
                  onClick={handleSendEmail}
                  disabled={!subject.trim() || !to.trim() || sendEmail.isPending}
                >
                  {sendEmail.isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Send className="h-4 w-4" />
                  )}
                  {sendEmail.isPending ? 'Sending...' : 'Send Email'}
                </Button>
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
                className="mb-3 w-full rounded-md border bg-background p-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
              />

              {/* Tone */}
              <div className="mb-3">
                <p className="mb-1.5 text-xs font-medium text-muted-foreground">Tone</p>
                <div className="flex flex-wrap gap-1.5">
                  {toneOptions.map((t) => (
                    <Button
                      key={t.id}
                      variant="ghost"
                      size="sm"
                      onClick={() => setTone(t.id)}
                      className={cn(
                        'h-auto rounded-full px-2.5 py-1 text-xs',
                        tone === t.id
                          ? 'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground'
                          : 'text-muted-foreground hover:text-foreground'
                      )}
                    >
                      {t.label}
                    </Button>
                  ))}
                </div>
              </div>

              <Button className="w-full">
                <Wand2 className="h-4 w-4" />
                Generate Email
              </Button>
            </div>

            {/* Subject Line Generator */}
            <div className="rounded-lg border bg-card p-4">
              <h3 className="mb-2 font-semibold text-sm">Subject Line Ideas</h3>
              <Button variant="outline" size="sm" className="mb-3 w-full">
                <Sparkles className="h-3 w-3" />
                Generate Subject Lines
              </Button>
              <div className="space-y-2">
                {[
                  'Unlock your CRM potential today',
                  "Don't miss out: exclusive automation tips",
                  '{{first_name}}, your weekly report is ready',
                ].map((s, i) => (
                  <Button
                    key={i}
                    variant="ghost"
                    onClick={() => setSubject(s)}
                    className="h-auto w-full justify-start whitespace-normal px-3 py-2 text-left text-xs font-normal"
                  >
                    {s}
                  </Button>
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
                  <Button
                    key={pt.token}
                    variant="secondary"
                    size="sm"
                    onClick={() => setBody((prev) => prev + pt.token)}
                    className="h-auto px-2 py-1 font-mono text-xs"
                  >
                    {pt.token}
                  </Button>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'sent' && (
        <div className="space-y-4">
          {/* Filters */}
          <div className="flex flex-wrap gap-3">
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                type="text"
                placeholder="Search emails..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
            <div className="flex gap-1 rounded-md border bg-background p-1">
              {statusFilters.map((sf) => (
                <Button
                  key={sf}
                  variant="ghost"
                  size="sm"
                  onClick={() => setStatusFilter(sf)}
                  className={cn(
                    'h-auto px-3 py-1 text-xs capitalize',
                    statusFilter === sf
                      ? 'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground'
                      : 'text-muted-foreground hover:text-foreground'
                  )}
                >
                  {sf}
                </Button>
              ))}
            </div>
          </div>

          {/* Email list */}
          {emailsLoading ? (
            <div className="divide-y rounded-lg border bg-card">
              {[1, 2, 3].map((i) => (
                <div key={i} className="flex items-center gap-4 px-5 py-4">
                  <Skeleton className="h-10 w-10 rounded-lg" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-4 w-48" />
                    <Skeleton className="h-3 w-32" />
                  </div>
                  <Skeleton className="h-5 w-16 rounded-full" />
                  <Skeleton className="h-4 w-20" />
                </div>
              ))}
            </div>
          ) : filteredEmails.length > 0 ? (
            <div className="divide-y rounded-lg border bg-card">
              {filteredEmails.map((email) => (
                <div key={email.id} className="group flex items-center gap-4 px-5 py-4">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                    <Mail className="h-5 w-5 text-primary" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{email.subject}</p>
                    <p className="text-sm text-muted-foreground">To: {email.to}</p>
                  </div>
                  <div className="flex items-center gap-4 flex-shrink-0">
                    <span
                      className={cn(
                        'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium',
                        email.status === 'sent' && 'bg-success/10 text-success',
                        email.status === 'scheduled' && 'bg-info/10 text-info',
                        email.status === 'draft' && 'bg-muted text-muted-foreground',
                        email.status === 'failed' && 'bg-destructive/10 text-destructive'
                      )}
                    >
                      {email.status === 'sent' && <CheckCircle className="h-3 w-3" />}
                      {email.status === 'scheduled' && <Clock className="h-3 w-3" />}
                      {email.status}
                    </span>
                    {email.status === 'sent' && (
                      <div className="flex items-center gap-3 text-right">
                        <span className="flex items-center gap-1 text-xs text-muted-foreground">
                          <MailOpen className="h-3 w-3" />
                          {email.opens}
                        </span>
                        <span className="flex items-center gap-1 text-xs text-muted-foreground">
                          <MousePointerClick className="h-3 w-3" />
                          {email.clicks}
                        </span>
                      </div>
                    )}
                    <div className="text-xs text-muted-foreground">
                      {email.sentAt
                        ? new Date(email.sentAt).toLocaleDateString()
                        : email.scheduledAt
                        ? new Date(email.scheduledAt).toLocaleDateString()
                        : 'Draft'}
                    </div>
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 opacity-0 transition-opacity group-hover:opacity-100"
                          aria-label="Delete email"
                        >
                          <Trash2 className="h-3.5 w-3.5 text-destructive" />
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Delete email</AlertDialogTitle>
                          <AlertDialogDescription>
                            Are you sure you want to delete &quot;{email.subject}&quot;? This action cannot be undone.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction
                            onClick={() => handleDeleteEmail(email.id)}
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                          >
                            Delete
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
              <Mail className="mb-4 h-12 w-12 text-muted-foreground/50" />
              <h3 className="mb-1 font-semibold">No emails found</h3>
              <p className="text-sm text-muted-foreground">
                {emailsError
                  ? 'Unable to load emails. The emails endpoint may not be available yet.'
                  : searchQuery || statusFilter !== 'all'
                  ? 'No emails match your current filters.'
                  : 'Compose and send your first email to get started.'}
              </p>
              {!emailsError && !searchQuery && statusFilter === 'all' && (
                <Button
                  className="mt-4"
                  onClick={() => setActiveTab('compose')}
                >
                  <Send className="h-4 w-4" />
                  Compose Email
                </Button>
              )}
            </div>
          )}
        </div>
      )}

      {activeTab === 'templates' && (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
          <FileText className="mb-4 h-12 w-12 text-muted-foreground/50" />
          <h3 className="mb-1 text-lg font-semibold">Email Templates</h3>
          <p className="mt-2 text-sm text-muted-foreground">
            Manage your email templates for consistent communication.
          </p>
          <Link href="/emails/templates">
            <Button className="mt-4">
              Go to Templates
              <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
        </div>
      )}
    </div>
  )
}
