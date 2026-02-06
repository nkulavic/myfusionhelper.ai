'use client'

import { useState, useMemo } from 'react'
import Link from 'next/link'
import {
  ArrowLeft,
  Plus,
  Copy,
  Trash2,
  Pencil,
  Search,
  Sparkles,
  Mail,
  Star,
  Loader2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
import {
  useEmailTemplates,
  useCreateTemplate,
  useUpdateTemplate,
  useDeleteTemplate,
} from '@/lib/hooks/use-emails'
import type { EmailTemplate } from '@/lib/api/emails'

const categories = ['all', 'onboarding', 'sales', 'billing', 'marketing', 'events']

export default function EmailTemplatesPage() {
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('all')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<EmailTemplate | null>(null)
  const [formName, setFormName] = useState('')
  const [formCategory, setFormCategory] = useState('onboarding')
  const [formSubject, setFormSubject] = useState('')
  const [formBody, setFormBody] = useState('')

  const { data: templates, isLoading, error } = useEmailTemplates()
  const createTemplate = useCreateTemplate()
  const updateTemplate = useUpdateTemplate()
  const deleteTemplate = useDeleteTemplate()

  const filtered = useMemo(() => {
    if (!templates) return []
    return templates.filter((t) => {
      if (category !== 'all' && t.category !== category) return false
      if (search) {
        const q = search.toLowerCase()
        return (
          t.name.toLowerCase().includes(q) ||
          t.subject.toLowerCase().includes(q)
        )
      }
      return true
    })
  }, [templates, category, search])

  const openCreateDialog = () => {
    setEditingTemplate(null)
    setFormName('')
    setFormCategory('onboarding')
    setFormSubject('')
    setFormBody('')
    setDialogOpen(true)
  }

  const openEditDialog = (template: EmailTemplate) => {
    setEditingTemplate(template)
    setFormName(template.name)
    setFormCategory(template.category)
    setFormSubject(template.subject)
    setFormBody(template.body)
    setDialogOpen(true)
  }

  const handleSave = () => {
    if (!formName.trim() || !formSubject.trim()) return

    if (editingTemplate) {
      updateTemplate.mutate(
        {
          id: editingTemplate.id,
          input: {
            name: formName,
            category: formCategory,
            subject: formSubject,
            body: formBody,
          },
        },
        { onSuccess: () => setDialogOpen(false) }
      )
    } else {
      createTemplate.mutate(
        {
          name: formName,
          category: formCategory,
          subject: formSubject,
          body: formBody,
        },
        { onSuccess: () => setDialogOpen(false) }
      )
    }
  }

  const handleToggleStar = (template: EmailTemplate) => {
    updateTemplate.mutate({
      id: template.id,
      input: { isStarred: !template.isStarred },
    })
  }

  const handleDelete = (id: string) => {
    deleteTemplate.mutate(id)
  }

  const isSaving = createTemplate.isPending || updateTemplate.isPending

  return (
    <div className="animate-fade-in-up space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href="/emails">
          <Button variant="ghost" size="icon" aria-label="Back to emails">
            <ArrowLeft className="h-5 w-5" />
          </Button>
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">Email Templates</h1>
          <p className="text-muted-foreground">Reusable email templates with AI generation</p>
        </div>
        <Button variant="outline">
          <Sparkles className="h-4 w-4" />
          AI Generate
        </Button>
        <Button onClick={openCreateDialog}>
          <Plus className="h-4 w-4" />
          New Template
        </Button>
      </div>

      {/* Search and Filters */}
      <div className="flex flex-wrap gap-3">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search templates..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="flex gap-1 rounded-md border bg-background p-1">
          {categories.map((cat) => (
            <Button
              key={cat}
              variant="ghost"
              size="sm"
              onClick={() => setCategory(cat)}
              className={cn(
                'h-auto px-3 py-1 text-xs capitalize',
                category === cat
                  ? 'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              {cat}
            </Button>
          ))}
        </div>
      </div>

      {/* Template Grid */}
      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="rounded-lg border bg-card p-5">
              <div className="mb-3 flex items-start justify-between">
                <div className="flex items-center gap-2">
                  <Skeleton className="h-9 w-9 rounded-lg" />
                  <div className="space-y-1">
                    <Skeleton className="h-4 w-24" />
                    <Skeleton className="h-3 w-16" />
                  </div>
                </div>
                <Skeleton className="h-6 w-6 rounded" />
              </div>
              <Skeleton className="mb-1 h-3 w-16" />
              <Skeleton className="mb-2 h-4 w-full" />
              <Skeleton className="h-3 w-3/4" />
              <div className="mt-4 flex items-center justify-between">
                <Skeleton className="h-3 w-20" />
              </div>
            </div>
          ))}
        </div>
      ) : filtered.length > 0 ? (
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
                <Button
                  variant="ghost"
                  size="icon"
                  className={cn(
                    'h-8 w-8',
                    template.isStarred ? 'text-warning' : 'text-muted-foreground'
                  )}
                  onClick={() => handleToggleStar(template)}
                  aria-label={template.isStarred ? 'Unstar template' : 'Star template'}
                >
                  <Star className="h-4 w-4" fill={template.isStarred ? 'currentColor' : 'none'} />
                </Button>
              </div>

              <p className="mb-1 text-xs font-medium text-muted-foreground">Subject:</p>
              <p className="mb-2 text-sm font-mono truncate">{template.subject}</p>

              <p className="text-xs text-muted-foreground line-clamp-2">{template.body}</p>

              <div className="mt-4 flex items-center justify-between">
                <span className="text-xs text-muted-foreground">
                  Used {template.usageCount.toLocaleString()} times
                </span>
                <div className="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => openEditDialog(template)}
                    aria-label="Edit template"
                  >
                    <Pencil className="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    aria-label="Duplicate template"
                  >
                    <Copy className="h-3.5 w-3.5" />
                  </Button>
                  <AlertDialog>
                    <AlertDialogTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 text-destructive"
                        aria-label="Delete template"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>Delete template</AlertDialogTitle>
                        <AlertDialogDescription>
                          Are you sure you want to delete &quot;{template.name}&quot;? This action cannot be undone.
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                          onClick={() => handleDelete(template.id)}
                          className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        >
                          Delete
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
          <Mail className="mb-4 h-12 w-12 text-muted-foreground/50" />
          <h3 className="mb-1 font-semibold">No templates found</h3>
          <p className="text-sm text-muted-foreground">
            {error
              ? 'Unable to load templates. The endpoint may not be available yet.'
              : search || category !== 'all'
              ? 'No templates match your current filters.'
              : 'Create your first email template to get started.'}
          </p>
          {!error && !search && category === 'all' && (
            <Button className="mt-4" onClick={openCreateDialog}>
              <Plus className="h-4 w-4" />
              Create Template
            </Button>
          )}
        </div>
      )}

      {/* Create/Edit Template Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[550px]">
          <DialogHeader>
            <DialogTitle>
              {editingTemplate ? 'Edit Template' : 'Create Template'}
            </DialogTitle>
            <DialogDescription>
              {editingTemplate
                ? 'Update the template details below.'
                : 'Fill in the details for your new email template.'}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="template-name">Name</Label>
              <Input
                id="template-name"
                placeholder="e.g. Welcome Email"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="template-category">Category</Label>
              <select
                id="template-category"
                value={formCategory}
                onChange={(e) => setFormCategory(e.target.value)}
                className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                {categories.filter((c) => c !== 'all').map((cat) => (
                  <option key={cat} value={cat}>
                    {cat.charAt(0).toUpperCase() + cat.slice(1)}
                  </option>
                ))}
              </select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="template-subject">Subject</Label>
              <Input
                id="template-subject"
                placeholder="e.g. Welcome to {{company}}, {{first_name}}!"
                value={formSubject}
                onChange={(e) => setFormSubject(e.target.value)}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="template-body">Body</Label>
              <textarea
                id="template-body"
                rows={6}
                placeholder="Write your email template body..."
                value={formBody}
                onChange={(e) => setFormBody(e.target.value)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleSave}
              disabled={!formName.trim() || !formSubject.trim() || isSaving}
            >
              {isSaving && <Loader2 className="h-4 w-4 animate-spin" />}
              {editingTemplate ? 'Save Changes' : 'Create Template'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
