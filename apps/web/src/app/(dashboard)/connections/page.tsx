'use client'

import { useState } from 'react'
import {
  Plus,
  CheckCircle,
  XCircle,
  AlertCircle,
  ExternalLink,
  Trash2,
  RefreshCw,
  Shield,
  Key,
  Clock,
  ChevronRight,
  ArrowLeft,
  Loader2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useConnections,
  usePlatforms,
  useCreateConnection,
  useDeleteConnection,
  useTestConnection,
  useStartOAuth,
} from '@/lib/hooks/use-connections'
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
import type { PlatformDefinition } from '@/lib/api/connections'
import { PlatformLogo } from '@/components/platform-logo'

type ViewState = 'list' | 'add' | 'detail'

export default function ConnectionsPage() {
  const [view, setView] = useState<ViewState>('list')
  const [selectedPlatform, setSelectedPlatform] = useState<PlatformDefinition | null>(null)
  const [selectedConnectionId, setSelectedConnectionId] = useState<string | null>(null)
  const [connectionName, setConnectionName] = useState('')
  const [credentialValues, setCredentialValues] = useState<Record<string, string>>({})

  const { data: connections, isLoading: connectionsLoading } = useConnections()
  const { data: platforms, isLoading: platformsLoading } = usePlatforms()
  const createConnection = useCreateConnection()
  const deleteConnection = useDeleteConnection()
  const testConnection = useTestConnection()
  const startOAuth = useStartOAuth()

  const selectedConnection = connections?.find((c) => c.connectionId === selectedConnectionId)

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="h-5 w-5 text-success" />
      case 'error':
        return <XCircle className="h-5 w-5 text-destructive" />
      case 'disconnected':
        return <AlertCircle className="h-5 w-5 text-warning" />
      default:
        return <AlertCircle className="h-5 w-5 text-warning" />
    }
  }

  const getStatusBadge = (status: string) => {
    const styles: Record<string, string> = {
      active: 'bg-success/10 text-success',
      error: 'bg-destructive/10 text-destructive',
      disconnected: 'bg-warning/10 text-warning',
    }
    return (
      <span className={cn('rounded-full px-2 py-0.5 text-xs font-medium capitalize', styles[status] || styles.disconnected)}>
        {status}
      </span>
    )
  }

  const getPlatformInfo = (platformId: string) =>
    platforms?.find((p) => p.platformId === platformId || p.slug === platformId)

  const handleConnect = async () => {
    if (!selectedPlatform) return

    if (selectedPlatform.apiConfig.authType === 'oauth2') {
      startOAuth.mutate(selectedPlatform.platformId, {
        onSuccess: (res) => {
          if (res.data?.url) {
            window.location.href = res.data.url
          }
        },
      })
    } else {
      createConnection.mutate(
        {
          platformId: selectedPlatform.platformId,
          input: {
            name: connectionName || `${selectedPlatform.name} Connection`,
            credentials: {
              apiKey: credentialValues.api_key || undefined,
              apiUrl: credentialValues.api_url || undefined,
              appId: credentialValues.app_id || undefined,
            },
          },
        },
        {
          onSuccess: () => {
            setView('list')
            setSelectedPlatform(null)
            setConnectionName('')
            setCredentialValues({})
          },
        }
      )
    }
  }

  const handleDelete = (platformId: string, connectionId: string) => {
    deleteConnection.mutate(
      { platformId, connectionId },
      { onSuccess: () => { setView('list'); setSelectedConnectionId(null) } }
    )
  }

  const handleTest = (platformId: string, connectionId: string) => {
    testConnection.mutate({ platformId, connectionId })
  }

  // Add Connection Flow
  if (view === 'add') {
    return (
      <div className="animate-slide-in-right mx-auto max-w-2xl space-y-6">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={() => { setView('list'); setSelectedPlatform(null); setCredentialValues({}) }}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold">Add Connection</h1>
            <p className="text-muted-foreground">
              {selectedPlatform ? `Connect to ${selectedPlatform.name}` : 'Select a platform to connect'}
            </p>
          </div>
        </div>

        {!selectedPlatform ? (
          platformsLoading ? (
            <div className="grid gap-4 sm:grid-cols-2">
              {[1, 2, 3, 4].map((i) => (
                <div key={i} className="flex flex-col items-start rounded-lg border bg-card p-5">
                  <Skeleton className="mb-3 h-12 w-12 rounded-lg" />
                  <Skeleton className="mb-2 h-5 w-32" />
                  <Skeleton className="h-4 w-24" />
                </div>
              ))}
            </div>
          ) : platforms && platforms.length > 0 ? (
            <>
              {/* CRM Platforms */}
              {platforms.filter((p) => p.types?.includes('crm')).length > 0 && (
                <div className="space-y-3">
                  <div>
                    <h3 className="text-sm font-semibold text-muted-foreground">Primary CRM Platforms</h3>
                    <p className="text-xs text-muted-foreground/70">Connect your main CRM platform</p>
                  </div>
                  <div className="animate-stagger-in grid gap-4 sm:grid-cols-2">
                    {platforms
                      .filter((p) => p.types?.includes('crm'))
                      .map((platform) => (
                        <button
                          key={platform.platformId}
                          onClick={() => setSelectedPlatform(platform)}
                          className="card-hover flex flex-col items-start rounded-lg border bg-card p-5 text-left transition-all hover:border-primary active:scale-[0.98]"
                        >
                          <PlatformLogo definition={platform} size={48} className="mb-3" />
                          <h3 className="mb-1 font-semibold">{platform.name}</h3>
                          <div className="mb-3 flex flex-wrap gap-1">
                            {platform.capabilities.slice(0, 3).map((cap) => (
                              <span key={cap} className="rounded bg-muted px-1.5 py-0.5 text-[10px] capitalize">
                                {cap}
                              </span>
                            ))}
                          </div>
                          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            {platform.apiConfig.authType === 'oauth2' ? (
                              <><Shield className="h-3 w-3" /> OAuth 2.0</>
                            ) : (
                              <><Key className="h-3 w-3" /> API Key</>
                            )}
                          </div>
                        </button>
                      ))}
                  </div>
                </div>
              )}

              {/* Integration Platforms */}
              {platforms.filter((p) => p.types?.includes('integration')).length > 0 && (
                <div className="space-y-3">
                  <div>
                    <h3 className="text-sm font-semibold text-muted-foreground">Integration Platforms</h3>
                    <p className="text-xs text-muted-foreground/70">Additional services for helper automation</p>
                  </div>
                  <div className="animate-stagger-in grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    {platforms
                      .filter((p) => p.types?.includes('integration'))
                      .map((platform) => (
                        <button
                          key={platform.platformId}
                          onClick={() => setSelectedPlatform(platform)}
                          className="card-hover flex flex-col items-start rounded-lg border bg-card p-5 text-left transition-all hover:border-primary active:scale-[0.98]"
                        >
                          <PlatformLogo definition={platform} size={48} className="mb-3" />
                          <h3 className="mb-1 font-semibold">{platform.name}</h3>
                          <div className="mb-3 flex flex-wrap gap-1">
                            {platform.capabilities.slice(0, 3).map((cap) => (
                              <span key={cap} className="rounded bg-muted px-1.5 py-0.5 text-[10px] capitalize">
                                {cap}
                              </span>
                            ))}
                          </div>
                          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            {platform.apiConfig.authType === 'oauth2' ? (
                              <><Shield className="h-3 w-3" /> OAuth 2.0</>
                            ) : (
                              <><Key className="h-3 w-3" /> API Key</>
                            )}
                          </div>
                        </button>
                      ))}
                  </div>
                </div>
              )}
            </>
          ) : (
            <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
              <AlertCircle className="mb-4 h-12 w-12 text-muted-foreground/50" />
              <h3 className="mb-1 font-semibold">Unable to load platforms</h3>
              <p className="text-sm text-muted-foreground">
                Please check your connection and try again.
              </p>
            </div>
          )
        ) : (
          <div className="space-y-6">
            <div className="rounded-lg border bg-card p-5">
              <div className="flex items-center gap-4">
                <PlatformLogo definition={selectedPlatform} size={48} />
                <div>
                  <h3 className="font-semibold">{selectedPlatform.name}</h3>
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    {selectedPlatform.apiConfig.authType === 'oauth2' ? (
                      <><Shield className="h-3 w-3" /> OAuth 2.0 Authentication</>
                    ) : (
                      <><Key className="h-3 w-3" /> API Key Authentication</>
                    )}
                  </div>
                </div>
              </div>
              <div className="mt-4 flex flex-wrap gap-1.5">
                {selectedPlatform.capabilities.map((cap) => (
                  <span key={cap} className="rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium capitalize">
                    {cap}
                  </span>
                ))}
              </div>
            </div>

            <div className="rounded-lg border bg-card p-5 space-y-4">
              <div>
                <label className="mb-2 block text-sm font-medium">Connection Name</label>
                <Input
                  type="text"
                  value={connectionName}
                  onChange={(e) => setConnectionName(e.target.value)}
                  placeholder={`e.g. ${selectedPlatform.name} (Production)`}
                />
              </div>

              {selectedPlatform.apiConfig.authType === 'api_key' &&
                selectedPlatform.credentialFields?.map((field) => (
                  <div key={field.key}>
                    <label className="mb-2 block text-sm font-medium">{field.label}</label>
                    <Input
                      type={field.inputType === 'password' ? 'password' : 'text'}
                      value={credentialValues[field.key] || ''}
                      onChange={(e) =>
                        setCredentialValues((prev) => ({ ...prev, [field.key]: e.target.value }))
                      }
                      placeholder={field.placeholder}
                      className="font-mono"
                    />
                    {field.hint && (
                      <p className="mt-1.5 text-xs text-muted-foreground">{field.hint}</p>
                    )}
                  </div>
                ))}
            </div>

            <div className="flex items-center justify-between">
              <Button
                variant="link"
                onClick={() => { setSelectedPlatform(null); setCredentialValues({}) }}
                className="text-muted-foreground hover:text-foreground"
              >
                Choose a different platform
              </Button>
              <Button
                onClick={handleConnect}
                disabled={createConnection.isPending || startOAuth.isPending}
              >
                {(createConnection.isPending || startOAuth.isPending) && (
                  <Loader2 className="h-4 w-4 animate-spin" />
                )}
                {selectedPlatform.apiConfig.authType === 'oauth2' ? (
                  <>
                    <ExternalLink className="h-4 w-4" />
                    Connect with {selectedPlatform.name}
                  </>
                ) : (
                  <>
                    <Key className="h-4 w-4" />
                    Save Connection
                  </>
                )}
              </Button>
            </div>

            {createConnection.isError && (
              <p className="text-center text-sm text-destructive">
                {createConnection.error instanceof Error
                  ? createConnection.error.message
                  : 'Failed to create connection'}
              </p>
            )}

            {selectedPlatform.apiConfig.authType === 'oauth2' && (
              <p className="text-center text-xs text-muted-foreground">
                You&apos;ll be redirected to {selectedPlatform.name} to authorize access.
                We only request the minimum permissions needed.
              </p>
            )}
          </div>
        )}
      </div>
    )
  }

  // Connection Detail View
  if (view === 'detail' && selectedConnection) {
    const platform = getPlatformInfo(selectedConnection.platformId)
    return (
      <div className="mx-auto max-w-3xl space-y-6">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={() => { setView('list'); setSelectedConnectionId(null) }}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div className="flex-1">
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold">{selectedConnection.name}</h1>
              {getStatusBadge(selectedConnection.status)}
            </div>
            <p className="text-sm text-muted-foreground">{platform?.name} connection</p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleTest(selectedConnection.platformId, selectedConnection.connectionId)}
              disabled={testConnection.isPending}
            >
              {testConnection.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <RefreshCw className="h-4 w-4" />
              )}
              Test Connection
            </Button>
            {selectedConnection.status === 'error' && platform?.apiConfig.authType === 'oauth2' && (
              <Button
                size="sm"
                onClick={() => startOAuth.mutate(selectedConnection.platformId, {
                  onSuccess: (res) => { if (res.data?.url) window.location.href = res.data.url },
                })}
              >
                <ExternalLink className="h-4 w-4" />
                Re-authorize
              </Button>
            )}
          </div>
        </div>

        {testConnection.isSuccess && (
          <div className="animate-scale-in flex items-start gap-3 rounded-lg border border-success/30 bg-success/10 p-4">
            <CheckCircle className="mt-0.5 h-5 w-5 animate-success-check text-success" />
            <p className="text-sm text-success">
              Connection test passed successfully.
            </p>
          </div>
        )}

        {testConnection.isError && (
          <div className="animate-scale-in flex items-start gap-3 rounded-lg border border-destructive/30 bg-destructive/10 p-4">
            <XCircle className="mt-0.5 h-5 w-5 text-destructive" />
            <div>
              <p className="text-sm font-medium text-destructive">Connection test failed</p>
              <p className="text-sm text-destructive/80">
                {testConnection.error instanceof Error ? testConnection.error.message : 'Unknown error'}
              </p>
            </div>
          </div>
        )}

        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2 space-y-6">
            <div className="rounded-lg border bg-card p-5 space-y-3">
              <h3 className="font-semibold">Connection Details</h3>
              <div className="space-y-3 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Platform</span>
                  <span className="font-medium">{platform?.name}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Auth Type</span>
                  <span className="font-medium capitalize">{platform?.apiConfig.authType === 'oauth2' ? 'OAuth 2.0' : 'API Key'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Status</span>
                  {getStatusBadge(selectedConnection.status)}
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Created</span>
                  <span className="font-medium">{new Date(selectedConnection.createdAt).toLocaleDateString()}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Updated</span>
                  <span className="font-medium">{new Date(selectedConnection.updatedAt).toLocaleDateString()}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Connection ID</span>
                  <span className="font-mono text-xs">{selectedConnection.connectionId}</span>
                </div>
              </div>
            </div>

            {platform && (
              <div className="rounded-lg border bg-card p-5 space-y-3">
                <h3 className="font-semibold">Platform Capabilities</h3>
                <div className="flex flex-wrap gap-2">
                  {platform.capabilities.map((cap) => (
                    <span
                      key={cap}
                      className="inline-flex items-center gap-1.5 rounded-full bg-muted px-2.5 py-1 text-xs font-medium capitalize"
                    >
                      <CheckCircle className="h-3 w-3 text-success" />
                      {cap}
                    </span>
                  ))}
                </div>
                {platform.apiConfig.rateLimits && (
                  <div className="mt-2 space-y-1 text-xs text-muted-foreground">
                    <p>{platform.apiConfig.rateLimits.requestsPerSecond} requests/sec &middot; {platform.apiConfig.rateLimits.requestsPerHour.toLocaleString()} hourly limit</p>
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="space-y-4">
            <div className="rounded-lg border bg-card p-5 space-y-3">
              <h3 className="font-semibold">Actions</h3>
              <div className="space-y-2">
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant="ghost"
                      className="w-full justify-start text-destructive hover:bg-destructive/10 hover:text-destructive"
                      disabled={deleteConnection.isPending}
                    >
                      <Trash2 className="h-4 w-4" />
                      {deleteConnection.isPending ? 'Deleting...' : 'Delete Connection'}
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Delete Connection</AlertDialogTitle>
                      <AlertDialogDescription>
                        Are you sure you want to delete &ldquo;{selectedConnection.name}&rdquo;? This will
                        remove all associated data and cannot be undone. Any helpers using this
                        connection will stop working.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Cancel</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={() => handleDelete(selectedConnection.platformId, selectedConnection.connectionId)}
                        className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                      >
                        Delete Connection
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  // Main List View
  return (
    <div className="animate-fade-in-up space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Connections</h1>
          <p className="text-muted-foreground">Connect and manage your CRM platforms</p>
        </div>
        <Button onClick={() => setView('add')}>
          <Plus className="h-4 w-4" />
          Add Connection
        </Button>
      </div>

      {/* Active Connections */}
      <div>
        <h2 className="mb-4 text-lg font-semibold">
          Your Connections
          {connections && (
            <span className="ml-2 text-sm font-normal text-muted-foreground">
              ({connections.length})
            </span>
          )}
        </h2>
        {connectionsLoading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex items-center gap-4 rounded-lg border bg-card p-4">
                <Skeleton className="h-12 w-12 rounded-lg" />
                <div className="flex-1">
                  <Skeleton className="h-5 w-40" />
                  <Skeleton className="mt-1 h-4 w-32" />
                </div>
                <Skeleton className="h-5 w-5" />
              </div>
            ))}
          </div>
        ) : connections && connections.length > 0 ? (
          <div className="animate-stagger-in space-y-3">
            {connections.map((connection) => {
              const platform = getPlatformInfo(connection.platformId)
              return (
                <button
                  key={connection.connectionId}
                  onClick={() => { setSelectedConnectionId(connection.connectionId); setView('detail') }}
                  className="flex w-full items-center justify-between rounded-lg border bg-card p-4 text-left transition-all hover:border-primary/50 hover:shadow-sm"
                >
                  <div className="flex items-center gap-4">
                    {platform ? (
                      <PlatformLogo definition={platform} size={48} />
                    ) : (
                      <div className="flex h-12 w-12 items-center justify-center rounded-lg text-xl font-bold text-white bg-primary">
                        ?
                      </div>
                    )}
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="font-semibold">{connection.name}</h3>
                        {getStatusBadge(connection.status)}
                      </div>
                      <div className="flex items-center gap-3 text-sm text-muted-foreground">
                        <span>{platform?.name || connection.platformId}</span>
                        <span className="text-border">&middot;</span>
                        <span className="flex items-center gap-1">
                          <Clock className="h-3 w-3" />
                          {new Date(connection.updatedAt).toLocaleDateString()}
                        </span>
                        {platform && (
                          <>
                            <span className="text-border">&middot;</span>
                            <span>{platform.capabilities.length} features</span>
                          </>
                        )}
                      </div>
                    </div>
                  </div>
                  <ChevronRight className="h-5 w-5 text-muted-foreground" />
                </button>
              )
            })}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
            <Plus className="mb-4 h-12 w-12 text-muted-foreground/50" />
            <h3 className="mb-1 font-semibold">No connections yet</h3>
            <p className="mb-4 max-w-sm text-sm text-muted-foreground">
              Connect a platform to start using helpers. Your credentials are encrypted and stored securely.
            </p>
            <Button onClick={() => setView('add')}>
              <Plus className="h-4 w-4" />
              Add Your First Connection
            </Button>
          </div>
        )}
      </div>

      {/* Available Platforms */}
      {platforms && platforms.length > 0 && (
        <div className="space-y-8">
          {/* CRM Platforms */}
          {platforms.filter((p) => p.types?.includes('crm')).length > 0 && (
            <div>
              <h2 className="mb-4 text-lg font-semibold">Primary CRM Platforms</h2>
              <div className="animate-stagger-in grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {platforms
                  .filter((p) => p.types?.includes('crm'))
                  .map((platform) => {
                    const connectedCount = connections?.filter((c) => c.platformId === platform.platformId).length || 0
                    return (
                      <div
                        key={platform.platformId}
                        className="card-hover relative overflow-hidden rounded-lg border bg-card p-4 transition-all hover:border-primary/50"
                      >
                        <div
                          className="absolute inset-x-0 top-0 h-[3px]"
                          style={{ backgroundColor: platform.displayConfig?.color || 'hsl(var(--primary))' }}
                        />
                        <PlatformLogo definition={platform} size={48} className="mb-3" />
                        <h3 className="font-semibold">{platform.name}</h3>
                        <div className="mb-3 flex items-center gap-1.5 text-[10px] uppercase tracking-wider text-muted-foreground/70">
                          {platform.apiConfig.authType === 'oauth2' ? 'OAuth 2.0' : 'API Key'}
                          {connectedCount > 0 && (
                            <span className="ml-1 rounded-full bg-success/10 px-1.5 py-0.5 text-success normal-case">
                              {connectedCount} connected
                            </span>
                          )}
                        </div>
                        <Button
                          variant="outline"
                          className="w-full"
                          onClick={() => { setSelectedPlatform(platform); setView('add') }}
                        >
                          <ExternalLink className="h-4 w-4" />
                          Connect
                        </Button>
                      </div>
                    )
                  })}
              </div>
            </div>
          )}

          {/* Integration Platforms */}
          {platforms.filter((p) => p.types?.includes('integration')).length > 0 && (
            <div>
              <h2 className="mb-4 text-lg font-semibold">Integration Platforms</h2>
              <p className="mb-4 text-sm text-muted-foreground">
                Connect additional services to extend helper automation capabilities
              </p>
              <div className="animate-stagger-in grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                {platforms
                  .filter((p) => p.types?.includes('integration'))
                  .map((platform) => {
                    const connectedCount = connections?.filter((c) => c.platformId === platform.platformId).length || 0
                    return (
                      <div
                        key={platform.platformId}
                        className="card-hover relative overflow-hidden rounded-lg border bg-card p-4 transition-all hover:border-primary/50"
                      >
                        <div
                          className="absolute inset-x-0 top-0 h-[3px]"
                          style={{ backgroundColor: platform.displayConfig?.color || 'hsl(var(--primary))' }}
                        />
                        <PlatformLogo definition={platform} size={48} className="mb-3" />
                        <h3 className="font-semibold">{platform.name}</h3>
                        <div className="mb-3 flex items-center gap-1.5 text-[10px] uppercase tracking-wider text-muted-foreground/70">
                          {platform.apiConfig.authType === 'oauth2' ? 'OAuth 2.0' : 'API Key'}
                          {connectedCount > 0 && (
                            <span className="ml-1 rounded-full bg-success/10 px-1.5 py-0.5 text-success normal-case">
                              {connectedCount} connected
                            </span>
                          )}
                        </div>
                        <Button
                          variant="outline"
                          className="w-full"
                          onClick={() => { setSelectedPlatform(platform); setView('add') }}
                        >
                          <ExternalLink className="h-4 w-4" />
                          Connect
                        </Button>
                      </div>
                    )
                  })}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
