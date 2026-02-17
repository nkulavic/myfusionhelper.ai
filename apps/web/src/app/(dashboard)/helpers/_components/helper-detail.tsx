'use client'

import { useMemo, useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import {
  ArrowLeft,
  Play,
  Settings,
  Trash2,
  ToggleLeft,
  ToggleRight,
  Clock,
  CheckCircle,
  XCircle,
  Copy,
  AlertTriangle,
  Blocks,
  ExternalLink,
  Loader2,
  Pencil,
  Save,
  X,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { helpersCatalog } from '@/lib/helpers-catalog'
import { useHelper, useHelperType, useExecutions, useDeleteHelper, useUpdateHelper, useExecuteHelper } from '@/lib/hooks/use-helpers'
import { getConfigForm, hasConfigForm } from './config-forms'
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
import { CRMBadges } from '@/components/crm-badges'
import { useConnections } from '@/lib/hooks/use-connections'
import type { PlatformConnection } from '@myfusionhelper/types'
import { ScheduleConfig } from './schedule-config'

interface HelperDetailProps {
  helperId: string
  onBack?: () => void
}

function DetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Skeleton className="h-8 w-8" />
        <div>
          <Skeleton className="h-7 w-48" />
          <Skeleton className="mt-1 h-4 w-64" />
        </div>
      </div>
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <div className="grid grid-cols-3 gap-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="rounded-lg border bg-card p-4">
                <Skeleton className="h-3 w-20" />
                <Skeleton className="mt-2 h-6 w-16" />
              </div>
            ))}
          </div>
          <Skeleton className="h-48 rounded-lg" />
        </div>
        <div className="space-y-4">
          <Skeleton className="h-48 rounded-lg" />
          <Skeleton className="h-32 rounded-lg" />
        </div>
      </div>
    </div>
  )
}

export function HelperDetail({ helperId, onBack }: HelperDetailProps) {
  const router = useRouter()
  const { data: helper, isLoading, error } = useHelper(helperId)
  const { data: helperTypeData } = useHelperType(helper?.helperType ?? '')
  const { data: executions } = useExecutions({ helperId, limit: 10 })
  const deleteHelper = useDeleteHelper()
  const updateHelper = useUpdateHelper()
  const executeHelper = useExecuteHelper()
  const [editingName, setEditingName] = useState(false)
  const [nameValue, setNameValue] = useState('')
  const [editingConfig, setEditingConfig] = useState(false)
  const [configValue, setConfigValue] = useState('')
  const [showJsonEditor, setShowJsonEditor] = useState(false)
  const [testContactId, setTestContactId] = useState('')
  const [showTestRun, setShowTestRun] = useState(false)
  const [copiedEndpoint, setCopiedEndpoint] = useState(false)

  const { data: connections } = useConnections()
  const helperConnection = useMemo(
    () => helper?.connectionId ? connections?.find((c: PlatformConnection) => c.connectionId === helper.connectionId) : undefined,
    [helper?.connectionId, connections]
  )

  // Use backend type data if available, fall back to static catalog
  const helperTemplate = useMemo(
    () => helpersCatalog.find((h) => h.id === helper?.helperType),
    [helper?.helperType]
  )

  const handleBack = () => {
    if (onBack) {
      onBack()
    } else {
      router.push('/helpers')
    }
  }

  if (isLoading) return <DetailSkeleton />

  if (error || !helper) {
    return (
      <div>
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <AlertTriangle className="mb-4 h-12 w-12 text-muted-foreground/50" />
          <h2 className="text-lg font-semibold">Helper not found</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            {error instanceof Error ? error.message : 'The helper could not be loaded.'}
          </p>
          <Button onClick={handleBack} className="mt-4">
            <ArrowLeft className="h-4 w-4" />
            Back to Helpers
          </Button>
        </div>
      </div>
    )
  }

  const isEnabled = helper.enabled

  const handleToggle = () => {
    updateHelper.mutate({
      id: helperId,
      input: { enabled: !isEnabled },
    })
  }

  const handleDelete = () => {
    deleteHelper.mutate(helperId, {
      onSuccess: () => handleBack(),
    })
  }

  const handleSaveName = () => {
    if (!nameValue.trim()) return
    updateHelper.mutate(
      { id: helperId, input: { name: nameValue.trim() } },
      { onSuccess: () => setEditingName(false) }
    )
  }

  const handleSaveConfig = () => {
    try {
      const parsed = JSON.parse(configValue)
      updateHelper.mutate(
        { id: helperId, input: { config: parsed } },
        { onSuccess: () => setEditingConfig(false) }
      )
    } catch {
      // Invalid JSON - don't save
    }
  }

  const handleFormConfigChange = (newConfig: Record<string, unknown>) => {
    updateHelper.mutate({ id: helperId, input: { config: newConfig } })
  }

  const ConfigForm = helper ? getConfigForm(helper.helperType) : null

  const [lastExecutionId, setLastExecutionId] = useState<string | null>(null)

  const handleTestRun = () => {
    if (!testContactId.trim()) return
    setLastExecutionId(null)
    executeHelper.mutate(
      { id: helperId, input: { contactId: testContactId.trim() } },
      {
        onSuccess: (res) => {
          setLastExecutionId(res.data?.executionId ?? null)
        },
      }
    )
  }

  const handleCopyEndpoint = () => {
    navigator.clipboard.writeText(`POST /helpers/${helper.helperId}/execute`)
    setCopiedEndpoint(true)
    setTimeout(() => setCopiedEndpoint(false), 2000)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={handleBack}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary/10">
            {helperTemplate ? (
              <helperTemplate.icon className="h-5 w-5 text-primary" />
            ) : (
              <Blocks className="h-5 w-5 text-primary" />
            )}
          </div>
          <div>
            <div className="flex items-center gap-3">
              {editingName ? (
                <div className="flex items-center gap-2">
                  <Input
                    type="text"
                    value={nameValue}
                    onChange={(e) => setNameValue(e.target.value)}
                    className="h-8 text-lg font-bold"
                    autoFocus
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') handleSaveName()
                      if (e.key === 'Escape') setEditingName(false)
                    }}
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-success"
                    onClick={handleSaveName}
                    disabled={updateHelper.isPending}
                  >
                    <Save className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => setEditingName(false)}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              ) : (
                <>
                  <h1 className="text-2xl font-bold">{helper.name}</h1>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => {
                      setNameValue(helper.name)
                      setEditingName(true)
                    }}
                    title="Edit name"
                  >
                    <Pencil className="h-3.5 w-3.5" />
                  </Button>
                </>
              )}
              <span
                className={cn(
                  'rounded-full px-2.5 py-0.5 text-xs font-medium',
                  isEnabled
                    ? 'bg-success/10 text-success'
                    : 'bg-muted text-muted-foreground'
                )}
              >
                {isEnabled ? 'Active' : 'Inactive'}
              </span>
            </div>
            <p className="text-sm text-muted-foreground">
              {helper.description || helperTypeData?.description || helperTemplate?.description || 'Custom automation helper'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowTestRun(!showTestRun)}
          >
            <Play className="h-4 w-4" />
            Run Helper
          </Button>
          <Button
            variant={isEnabled ? 'outline' : 'default'}
            size="sm"
            onClick={handleToggle}
            disabled={updateHelper.isPending}
          >
            {isEnabled ? (
              <>
                <ToggleRight className="h-4 w-4" />
                Disable
              </>
            ) : (
              <>
                <ToggleLeft className="h-4 w-4" />
                Enable
              </>
            )}
          </Button>
        </div>
      </div>

      {/* Test Run Panel */}
      {showTestRun && (
        <div className="rounded-lg border border-primary/20 bg-primary/5 p-4">
          <h3 className="text-sm font-semibold mb-3">Run Helper</h3>
          <div className="flex gap-3">
            <Input
              type="text"
              value={testContactId}
              onChange={(e) => setTestContactId(e.target.value)}
              placeholder="Enter a Contact ID to test with..."
              className="flex-1"
            />
            <Button
              onClick={handleTestRun}
              disabled={!testContactId.trim() || executeHelper.isPending}
            >
              {executeHelper.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Play className="h-4 w-4" />
              )}
              {executeHelper.isPending ? 'Running...' : 'Execute'}
            </Button>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setShowTestRun(false)}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
          {executeHelper.isSuccess && (
            <div className="mt-3 flex items-center gap-3 rounded-md border border-success/30 bg-success/10 p-3">
              <CheckCircle className="h-4 w-4 shrink-0 text-success" />
              <p className="flex-1 text-xs text-success">Execution started successfully.</p>
              {lastExecutionId && (
                <Link
                  href={`/executions/${lastExecutionId}`}
                  className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline"
                >
                  View details
                  <ExternalLink className="h-3 w-3" />
                </Link>
              )}
            </div>
          )}
          {executeHelper.error && (
            <div className="mt-3 flex items-center gap-3 rounded-md border border-destructive/30 bg-destructive/10 p-3">
              <XCircle className="h-4 w-4 shrink-0 text-destructive" />
              <p className="flex-1 text-xs text-destructive">
                {executeHelper.error instanceof Error ? executeHelper.error.message : 'Execution failed'}
              </p>
              <Button
                variant="outline"
                size="sm"
                className="h-7 text-xs"
                onClick={handleTestRun}
                disabled={executeHelper.isPending}
              >
                Retry
              </Button>
            </div>
          )}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Stats */}
          <div className="grid grid-cols-3 gap-4">
            <div className="rounded-lg border bg-card p-4">
              <p className="text-xs text-muted-foreground">Type</p>
              <p className="mt-1 text-sm font-bold">{helperTypeData?.name || helperTemplate?.name || helper.helperType}</p>
            </div>
            <div className="rounded-lg border bg-card p-4">
              <p className="text-xs text-muted-foreground">Category</p>
              <p className="mt-1 text-sm font-bold capitalize">{helper.category}</p>
            </div>
            <div className="rounded-lg border bg-card p-4">
              <p className="text-xs text-muted-foreground">Executions</p>
              <p className="mt-1 text-sm font-bold">{helper.executionCount?.toLocaleString() || '0'}</p>
            </div>
          </div>

          {/* CRM Support */}
          {(helperTypeData || helperTemplate) && (
            <div className="rounded-lg border bg-card p-5">
              <h2 className="text-sm font-semibold mb-3">Supported Platforms</h2>
              <CRMBadges crmIds={helperTypeData?.supportedCrms ?? helperTemplate?.supportedCRMs ?? []} />
            </div>
          )}

          {/* Configuration */}
          <div className="rounded-lg border bg-card">
            <div className="flex items-center justify-between border-b px-5 py-4">
              <div className="flex items-center gap-2">
                <h2 className="font-semibold">Configuration</h2>
                {updateHelper.isPending && (
                  <Loader2 className="h-3.5 w-3.5 animate-spin text-muted-foreground" />
                )}
              </div>
              <div className="flex items-center gap-2">
                {ConfigForm && !showJsonEditor && !editingConfig && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      setConfigValue(JSON.stringify(helper.config, null, 2))
                      setShowJsonEditor(true)
                    }}
                    className="text-xs text-muted-foreground"
                  >
                    JSON
                  </Button>
                )}
                {showJsonEditor && !editingConfig && ConfigForm && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowJsonEditor(false)}
                    className="text-xs text-muted-foreground"
                  >
                    Form
                  </Button>
                )}
                {(showJsonEditor || !ConfigForm) && !editingConfig && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setConfigValue(JSON.stringify(helper.config, null, 2))
                      setEditingConfig(true)
                    }}
                  >
                    <Pencil className="h-3 w-3" />
                    Edit JSON
                  </Button>
                )}
                {editingConfig && (
                  <div className="flex items-center gap-2">
                    <Button
                      size="sm"
                      onClick={handleSaveConfig}
                      disabled={updateHelper.isPending}
                    >
                      {updateHelper.isPending ? (
                        <Loader2 className="h-3 w-3 animate-spin" />
                      ) : (
                        <Save className="h-3 w-3" />
                      )}
                      Save
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setEditingConfig(false)}
                    >
                      Cancel
                    </Button>
                  </div>
                )}
              </div>
            </div>
            <div className="p-5">
              {editingConfig ? (
                <textarea
                  value={configValue}
                  onChange={(e) => setConfigValue(e.target.value)}
                  rows={Math.max(6, configValue.split('\n').length + 1)}
                  className="w-full rounded-md border border-input bg-muted p-4 text-sm font-mono resize-y focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  spellCheck={false}
                />
              ) : ConfigForm && !showJsonEditor ? (
                <ConfigForm
                  config={helper.config || {}}
                  onChange={handleFormConfigChange}
                  platformId={helperConnection?.platformId}
                  connectionId={helper.connectionId}
                />
              ) : (
                <pre className="rounded-md bg-muted p-4 text-sm font-mono overflow-x-auto">
                  {helper.config && Object.keys(helper.config).length > 0
                    ? JSON.stringify(helper.config, null, 2)
                    : '{\n  // No configuration set\n}'}
                </pre>
              )}
            </div>
          </div>

          {/* Recent Executions */}
          <div className="rounded-lg border bg-card">
            <div className="flex items-center justify-between border-b px-5 py-4">
              <h2 className="font-semibold">Recent Executions</h2>
              <Link
                href={`/executions?helper=${helperId}`}
                className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
              >
                View all
                <ExternalLink className="h-3 w-3" />
              </Link>
            </div>
            {executions && executions.length > 0 ? (
              <div className="divide-y">
                {executions.map((exec) => (
                  <Link
                    key={exec.executionId}
                    href={`/executions/${exec.executionId}`}
                    className="flex items-center gap-4 px-5 py-3 hover:bg-accent/50 transition-colors"
                  >
                    <div className="flex-shrink-0">
                      {exec.status === 'completed' ? (
                        <CheckCircle className="h-4 w-4 text-success" />
                      ) : exec.status === 'failed' ? (
                        <XCircle className="h-4 w-4 text-destructive" />
                      ) : (
                        <Clock className="h-4 w-4 animate-spin text-info" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-mono truncate">
                        {exec.contactId || 'No contact'}
                      </p>
                    </div>
                    <p className="text-xs font-mono text-muted-foreground">
                      {exec.durationMs ? `${exec.durationMs}ms` : '-'}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(exec.startedAt).toLocaleString()}
                    </p>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="py-8 text-center text-sm text-muted-foreground">
                No executions yet. Use the Test Run button above to try it out.
              </div>
            )}
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-4">
          {/* Details */}
          <div className="rounded-lg border bg-card p-5 space-y-4">
            <h3 className="font-semibold">Details</h3>
            <div className="space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Helper ID</span>
                <span className="font-mono text-xs truncate max-w-[140px]" title={helper.helperId}>
                  {helper.helperId}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Connection</span>
                <span className="font-mono text-xs truncate max-w-[140px]" title={helper.connectionId || 'None'}>
                  {helper.connectionId || 'None'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Created</span>
                <span className="font-medium">
                  {new Date(helper.createdAt).toLocaleDateString()}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Updated</span>
                <span className="font-medium">
                  {new Date(helper.updatedAt).toLocaleDateString()}
                </span>
              </div>
              {helper.lastExecutedAt && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Last Run</span>
                  <span className="font-medium">
                    {new Date(helper.lastExecutedAt).toLocaleDateString()}
                  </span>
                </div>
              )}
            </div>
          </div>

          {/* API Endpoint */}
          <div className="rounded-lg border bg-card p-5 space-y-3">
            <h3 className="font-semibold">API Endpoint</h3>
            <div className="rounded-md bg-muted p-3">
              <p className="text-[11px] font-mono text-muted-foreground break-all">
                POST /helpers/{helper.helperId}/execute
              </p>
            </div>
            <Button
              variant="link"
              size="sm"
              className="h-auto p-0 text-xs"
              onClick={handleCopyEndpoint}
            >
              <Copy className="h-3 w-3" />
              {copiedEndpoint ? 'Copied!' : 'Copy endpoint'}
            </Button>
          </div>

          {/* Schedule */}
          <ScheduleConfig helper={helper} />

          {/* Config Schema */}
          {helper.configSchema && Object.keys(helper.configSchema).length > 0 && (
            <div className="rounded-lg border bg-card p-5 space-y-3">
              <h3 className="font-semibold">Config Schema</h3>
              <pre className="rounded-md bg-muted p-3 text-[10px] font-mono overflow-x-auto max-h-48">
                {JSON.stringify(helper.configSchema, null, 2)}
              </pre>
            </div>
          )}

          {/* Danger Zone */}
          <div className="rounded-lg border border-destructive/30 bg-card p-5 space-y-3">
            <h3 className="font-semibold text-destructive">Danger Zone</h3>
            <p className="text-xs text-muted-foreground">
              Deleting a helper is permanent and will stop all future executions.
            </p>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button
                  variant="outline"
                  className="w-full border-destructive/30 text-destructive hover:bg-destructive/10 hover:text-destructive"
                  disabled={deleteHelper.isPending}
                >
                  <Trash2 className="h-4 w-4" />
                  {deleteHelper.isPending ? 'Deleting...' : 'Delete Helper'}
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete Helper</AlertDialogTitle>
                  <AlertDialogDescription>
                    Are you sure you want to delete &ldquo;{helper.name}&rdquo;? This action is
                    permanent and will stop all future executions of this helper.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={handleDelete}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  >
                    Delete Helper
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </div>
    </div>
  )
}
