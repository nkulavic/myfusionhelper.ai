'use client'

import { useState, useMemo } from 'react'
import {
  Search,
  Plus,
  Blocks,
  ToggleLeft,
  ToggleRight,
  Trash2,
  Settings,
  CheckCircle,
  AlertTriangle,
  ChevronRight,
  MoreVertical,
  Play,
  Loader2,
  XCircle,
  ExternalLink,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import Link from 'next/link'
import { useHelpers, useHelperTypes, useUpdateHelper, useDeleteHelper, useExecuteHelper } from '@/lib/hooks/use-helpers'
import { helpersCatalog } from '@/lib/helpers-catalog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import type { Helper } from '@myfusionhelper/types'

interface MyHelpersListProps {
  onSelectHelper: (id: string) => void
  onNewHelper: () => void
}

export function MyHelpersList({ onSelectHelper, onNewHelper }: MyHelpersListProps) {
  const { data: helpers, isLoading, error } = useHelpers()
  const { data: helperTypesData } = useHelperTypes()
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'inactive'>('all')
  const [categoryFilter, setCategoryFilter] = useState('all')
  const updateHelper = useUpdateHelper()
  const deleteHelper = useDeleteHelper()

  // Build a lookup map: type id -> display name (backend first, then static)
  const typeNameMap = useMemo(() => {
    const map = new Map<string, string>()
    for (const h of helpersCatalog) {
      map.set(h.id, h.name)
    }
    if (helperTypesData?.types) {
      for (const t of helperTypesData.types) {
        map.set(t.type, t.name)
      }
    }
    return map
  }, [helperTypesData])

  const filteredHelpers = useMemo(() => {
    if (!helpers) return []
    return helpers.filter((helper) => {
      const matchesSearch =
        searchQuery === '' ||
        helper.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        helper.helperType.toLowerCase().includes(searchQuery.toLowerCase())
      const matchesStatus =
        statusFilter === 'all' ||
        (statusFilter === 'active' && helper.enabled) ||
        (statusFilter === 'inactive' && !helper.enabled)
      const matchesCategory =
        categoryFilter === 'all' || helper.category === categoryFilter
      return matchesSearch && matchesStatus && matchesCategory
    })
  }, [helpers, searchQuery, statusFilter, categoryFilter])

  const stats = useMemo(() => {
    if (!helpers) return { total: 0, active: 0, inactive: 0 }
    return {
      total: helpers.length,
      active: helpers.filter((h) => h.enabled).length,
      inactive: helpers.filter((h) => !h.enabled).length,
    }
  }, [helpers])

  const categories = useMemo(() => {
    if (!helpers) return []
    const cats = new Set(helpers.map((h) => h.category))
    return Array.from(cats).sort()
  }, [helpers])

  if (isLoading) {
    return <MyHelpersListSkeleton />
  }

  if (error) {
    return (
      <div className="space-y-6">
        <MyHelpersHeader onNewHelper={onNewHelper} total={0} />
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
          <AlertTriangle className="mb-4 h-12 w-12 text-muted-foreground/50" />
          <h3 className="mb-1 font-semibold">Unable to load helpers</h3>
          <p className="text-sm text-muted-foreground">
            {error instanceof Error ? error.message : 'Please try again later.'}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <MyHelpersHeader onNewHelper={onNewHelper} total={stats.total} />

      {/* Stats Row */}
      <div className="animate-stagger-in grid gap-4 sm:grid-cols-3">
        <div className="card-hover rounded-lg border bg-card p-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">Total Helpers</p>
            <Blocks className="h-4 w-4 text-muted-foreground" />
          </div>
          <p className="mt-1 text-2xl font-bold">{stats.total}</p>
        </div>
        <div className="card-hover rounded-lg border bg-card p-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">Active</p>
            <CheckCircle className="h-4 w-4 text-success" />
          </div>
          <p className="mt-1 text-2xl font-bold text-success">{stats.active}</p>
        </div>
        <div className="card-hover rounded-lg border bg-card p-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">Inactive</p>
            <ToggleLeft className="h-4 w-4 text-muted-foreground" />
          </div>
          <p className="mt-1 text-2xl font-bold text-muted-foreground">{stats.inactive}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search your helpers..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as 'all' | 'active' | 'inactive')}
          className="h-10 rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <option value="all">All Status</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
        {categories.length > 1 && (
          <select
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.target.value)}
            className="h-10 rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <option value="all">All Categories</option>
            {categories.map((cat) => (
              <option key={cat} value={cat}>
                {cat.charAt(0).toUpperCase() + cat.slice(1)}
              </option>
            ))}
          </select>
        )}
      </div>

      {/* Helpers List */}
      {filteredHelpers.length > 0 ? (
        <div className="animate-stagger-in space-y-3">
          {filteredHelpers.map((helper) => (
            <HelperRow
              key={helper.helperId}
              helper={helper}
              typeName={typeNameMap.get(helper.helperType)}
              onSelect={() => onSelectHelper(helper.helperId)}
              onToggle={() => {
                updateHelper.mutate({
                  id: helper.helperId,
                  input: { enabled: !helper.enabled },
                })
              }}
              onDelete={() => deleteHelper.mutate(helper.helperId)}
              isToggling={updateHelper.isPending}
              isDeleting={deleteHelper.isPending}
            />
          ))}
        </div>
      ) : helpers && helpers.length === 0 ? (
        <EmptyState onNewHelper={onNewHelper} />
      ) : (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
          <Search className="mb-4 h-12 w-12 text-muted-foreground/50" />
          <h3 className="mb-1 font-semibold">No matching helpers</h3>
          <p className="text-sm text-muted-foreground">
            Try adjusting your search or filters
          </p>
        </div>
      )}
    </div>
  )
}

function MyHelpersHeader({
  onNewHelper,
  total,
}: {
  onNewHelper: () => void
  total: number
}) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-2xl font-bold">My Helpers</h1>
        <p className="text-muted-foreground">
          {total > 0
            ? `${total} configured helper${total !== 1 ? 's' : ''}`
            : 'Set up automation helpers for your CRM'}
        </p>
      </div>
      <Button onClick={onNewHelper}>
        <Plus className="h-4 w-4" />
        New Helper
      </Button>
    </div>
  )
}

function HelperRow({
  helper,
  typeName,
  onSelect,
  onToggle,
  onDelete,
  isToggling,
  isDeleting,
}: {
  helper: Helper
  typeName?: string
  onSelect: () => void
  onToggle: () => void
  onDelete: () => void
  isToggling: boolean
  isDeleting: boolean
}) {
  const [showActions, setShowActions] = useState(false)
  const [showRunDialog, setShowRunDialog] = useState(false)
  const template = helpersCatalog.find((h) => h.id === helper.helperType)

  return (
    <>
      <div
        className={cn(
          'group relative flex items-center gap-4 rounded-lg border bg-card p-4 transition-all card-hover',
          !helper.enabled && 'opacity-70'
        )}
      >
        {/* Icon */}
        <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary/10">
          {template ? (
            <template.icon className="h-5 w-5 text-primary" />
          ) : (
            <Blocks className="h-5 w-5 text-primary" />
          )}
        </div>

        {/* Content */}
        <button
          onClick={onSelect}
          className="flex flex-1 items-center gap-4 text-left min-w-0"
        >
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <h3 className="text-sm font-semibold truncate group-hover:text-primary">
                {helper.name}
              </h3>
              <span
                className={cn(
                  'rounded-full px-2 py-0.5 text-[10px] font-medium',
                  helper.enabled
                    ? 'bg-success/10 text-success'
                    : 'bg-muted text-muted-foreground'
                )}
              >
                {helper.enabled ? 'Active' : 'Inactive'}
              </span>
            </div>
            <div className="mt-0.5 flex items-center gap-3 text-xs text-muted-foreground">
              <span className="capitalize">{helper.category}</span>
              <span className="text-border">|</span>
              <span>{typeName || template?.name || helper.helperType}</span>
              {helper.executionCount > 0 && (
                <>
                  <span className="text-border">|</span>
                  <span>{helper.executionCount.toLocaleString()} runs</span>
                </>
              )}
            </div>
          </div>
          <ChevronRight className="h-4 w-4 flex-shrink-0 text-muted-foreground/50 group-hover:text-primary" />
        </button>

        {/* Quick Actions */}
        <div className="flex items-center gap-1">
          <button
            onClick={() => setShowRunDialog(true)}
            className="rounded-md p-2 text-muted-foreground hover:bg-primary/10 hover:text-primary disabled:opacity-50"
            title="Run helper"
          >
            <Play className="h-4 w-4" />
          </button>
          <button
            onClick={onToggle}
            disabled={isToggling}
            className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-foreground disabled:opacity-50"
            title={helper.enabled ? 'Disable' : 'Enable'}
          >
            {helper.enabled ? (
              <ToggleRight className="h-4 w-4 text-success" />
            ) : (
              <ToggleLeft className="h-4 w-4" />
            )}
          </button>
          <div className="relative">
            <button
              onClick={() => setShowActions(!showActions)}
              className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-foreground"
            >
              <MoreVertical className="h-4 w-4" />
            </button>
            {showActions && (
              <>
                <div
                  className="fixed inset-0 z-10"
                  onClick={() => setShowActions(false)}
                />
                <div className="absolute right-0 top-full z-20 mt-1 w-40 rounded-md border bg-popover py-1 shadow-md">
                  <button
                    onClick={() => {
                      setShowRunDialog(true)
                      setShowActions(false)
                    }}
                    className="flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-accent"
                  >
                    <Play className="h-3.5 w-3.5" />
                    Run
                  </button>
                  <button
                    onClick={() => {
                      onSelect()
                      setShowActions(false)
                    }}
                    className="flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-accent"
                  >
                    <Settings className="h-3.5 w-3.5" />
                    Configure
                  </button>
                  <button
                    onClick={() => {
                      onDelete()
                      setShowActions(false)
                    }}
                    disabled={isDeleting}
                    className="flex w-full items-center gap-2 px-3 py-2 text-sm text-destructive hover:bg-destructive/10 disabled:opacity-50"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                    {isDeleting ? 'Deleting...' : 'Delete'}
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      </div>

      <RunHelperDialog
        helperId={helper.helperId}
        helperName={helper.name}
        open={showRunDialog}
        onOpenChange={setShowRunDialog}
      />
    </>
  )
}

// ---------------------------------------------------------------------------
// Run Helper Dialog (shared quick-execute UI)
// ---------------------------------------------------------------------------
function RunHelperDialog({
  helperId,
  helperName,
  open,
  onOpenChange,
}: {
  helperId: string
  helperName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const executeHelper = useExecuteHelper()
  const [contactId, setContactId] = useState('')
  const [executionId, setExecutionId] = useState<string | null>(null)

  const handleRun = () => {
    if (!contactId.trim()) return
    setExecutionId(null)
    executeHelper.mutate(
      { id: helperId, input: { contactId: contactId.trim() } },
      {
        onSuccess: (res) => {
          setExecutionId(res.data?.executionId ?? null)
        },
      }
    )
  }

  const handleClose = (nextOpen: boolean) => {
    if (!nextOpen) {
      setContactId('')
      setExecutionId(null)
      executeHelper.reset()
    }
    onOpenChange(nextOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Run Helper</DialogTitle>
          <DialogDescription>
            Execute &ldquo;{helperName}&rdquo; against a specific contact.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="run-contact-id">Contact ID</Label>
            <Input
              id="run-contact-id"
              value={contactId}
              onChange={(e) => setContactId(e.target.value)}
              placeholder="Enter CRM contact ID..."
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleRun()
              }}
            />
          </div>

          {executeHelper.isSuccess && executionId && (
            <div className="flex items-start gap-2 rounded-md border border-success/30 bg-success/10 p-3">
              <CheckCircle className="mt-0.5 h-4 w-4 shrink-0 text-success" />
              <div className="flex-1 text-sm">
                <p className="font-medium text-success">Execution started</p>
                <Link
                  href={`/executions/${executionId}`}
                  className="mt-1 inline-flex items-center gap-1 text-xs text-primary hover:underline"
                >
                  View execution details
                  <ExternalLink className="h-3 w-3" />
                </Link>
              </div>
            </div>
          )}

          {executeHelper.isError && (
            <div className="flex items-start gap-2 rounded-md border border-destructive/30 bg-destructive/10 p-3">
              <XCircle className="mt-0.5 h-4 w-4 shrink-0 text-destructive" />
              <div className="flex-1 text-sm">
                <p className="font-medium text-destructive">Execution failed</p>
                <p className="mt-0.5 text-xs text-destructive/80">
                  {executeHelper.error instanceof Error
                    ? executeHelper.error.message
                    : 'Something went wrong. Please try again.'}
                </p>
              </div>
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => handleClose(false)}>
            {executionId ? 'Close' : 'Cancel'}
          </Button>
          <Button
            onClick={handleRun}
            disabled={!contactId.trim() || executeHelper.isPending}
          >
            {executeHelper.isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Play className="h-4 w-4" />
            )}
            {executeHelper.isPending
              ? 'Running...'
              : executeHelper.isError
                ? 'Retry'
                : 'Run Now'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function EmptyState({ onNewHelper }: { onNewHelper: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
      <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
        <Blocks className="h-8 w-8 text-primary" />
      </div>
      <h3 className="mb-1 text-lg font-semibold">No helpers configured yet</h3>
      <p className="mb-6 max-w-sm text-sm text-muted-foreground">
        Helpers are small automation units that extend your CRM — tagging, formatting, scoring, syncing, and more. Browse 62 pre-built helpers and configure your first one in minutes.
      </p>
      <Button onClick={onNewHelper}>
        <Plus className="h-4 w-4" />
        Browse 62 Helpers
      </Button>
    </div>
  )
}

function MyHelpersListSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <div className="flex items-center justify-between">
        <div>
          <div className="h-7 w-32 rounded bg-muted" />
          <div className="mt-1 h-4 w-64 rounded bg-muted" />
        </div>
        <div className="h-10 w-32 rounded-md bg-muted" />
      </div>
      <div className="grid gap-4 sm:grid-cols-3">
        {[1, 2, 3].map((i) => (
          <div key={i} className="rounded-lg border bg-card p-4">
            <Skeleton className="h-4 w-20" />
            <Skeleton className="mt-2 h-7 w-16" />
          </div>
        ))}
      </div>
      <div className="h-10 w-full rounded-md bg-muted" />
      <div className="space-y-3">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="flex items-center gap-4 rounded-lg border bg-card p-4">
            <Skeleton className="h-10 w-10 rounded-lg" />
            <div className="flex-1">
              <Skeleton className="h-4 w-48" />
              <Skeleton className="mt-2 h-3 w-32" />
            </div>
            <Skeleton className="h-8 w-8 rounded-md" />
          </div>
        ))}
      </div>
    </div>
  )
}
