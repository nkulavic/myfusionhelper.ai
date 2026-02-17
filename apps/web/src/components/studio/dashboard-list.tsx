'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Plus, BarChart3, Trash2, MoreVertical, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useDashboards, useCreateDashboard, useDeleteDashboard } from '@/lib/hooks/use-studio'
import { mapDashboard } from '@/lib/api/studio'
import { useRouter } from 'next/navigation'

export function DashboardList() {
  const { data: dashboards = [], isLoading } = useDashboards()
  const createDashboard = useCreateDashboard()
  const deleteDashboard = useDeleteDashboard()
  const [createOpen, setCreateOpen] = useState(false)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const router = useRouter()

  const handleCreate = () => {
    if (!name.trim()) return
    createDashboard.mutate(
      { name: name.trim(), description: description.trim() || undefined },
      {
        onSuccess: (res) => {
          const dashboard = res.data ? mapDashboard(res.data) : null
          setName('')
          setDescription('')
          setCreateOpen(false)
          if (dashboard) router.push(`/studio/${dashboard.id}`)
        },
      },
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Studio</h1>
          <p className="text-muted-foreground">Build custom data dashboards</p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-1.5 h-4 w-4" />
          New Dashboard
        </Button>
      </div>

      {dashboards.length > 0 ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {dashboards.map((dashboard) => (
            <div
              key={dashboard.id}
              className="group relative rounded-lg border bg-card transition-colors hover:border-primary"
            >
              <Link href={`/studio/${dashboard.id}`} className="block p-5">
                <div className="mb-3 flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                    <BarChart3 className="h-5 w-5 text-primary" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <h3 className="font-semibold truncate">{dashboard.name}</h3>
                    {dashboard.description && (
                      <p className="text-xs text-muted-foreground truncate">
                        {dashboard.description}
                      </p>
                    )}
                  </div>
                </div>
                <div className="flex items-center justify-between text-xs text-muted-foreground">
                  <span>{dashboard.widgets.length} widget{dashboard.widgets.length !== 1 ? 's' : ''}</span>
                  <span>
                    Updated{' '}
                    {new Date(dashboard.updatedAt).toLocaleDateString('en-US', {
                      month: 'short',
                      day: 'numeric',
                    })}
                  </span>
                </div>
              </Link>

              <div className="absolute right-2 top-2 opacity-0 transition-opacity group-hover:opacity-100">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-7 w-7">
                      <MoreVertical className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem
                      onClick={(e) => {
                        e.preventDefault()
                        deleteDashboard.mutate(dashboard.id)
                      }}
                      className="text-destructive"
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>
          ))}

          {/* Create New Card */}
          <button
            onClick={() => setCreateOpen(true)}
            className="flex flex-col items-center justify-center gap-2 rounded-lg border border-dashed p-5 text-muted-foreground transition-colors hover:border-primary hover:text-foreground"
          >
            <Plus className="h-8 w-8" />
            <span className="text-sm font-medium">Create New Dashboard</span>
          </button>
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-20 text-center">
          <BarChart3 className="mb-4 h-12 w-12 text-muted-foreground/40" />
          <h3 className="mb-1 font-semibold">No dashboards yet</h3>
          <p className="mb-4 max-w-sm text-sm text-muted-foreground">
            Create your first dashboard and add charts, scorecards, and tables to visualize your CRM
            data.
          </p>
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="mr-1.5 h-4 w-4" />
            Create Dashboard
          </Button>
        </div>
      )}

      {/* Create Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>New Dashboard</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Name</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Sales Overview"
                autoFocus
                onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
              />
            </div>
            <div className="space-y-2">
              <Label>Description (optional)</Label>
              <Input
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="e.g. Monthly sales metrics and pipeline"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!name.trim() || createDashboard.isPending}
            >
              {createDashboard.isPending ? (
                <Loader2 className="mr-1.5 h-4 w-4 animate-spin" />
              ) : null}
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
