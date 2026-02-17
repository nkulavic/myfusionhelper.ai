'use client'

import { useState } from 'react'
import { Key, CheckCircle, Loader2, ArrowLeft, Shield, ExternalLink } from 'lucide-react'
import { PlatformLogo } from '@/components/platform-logo'
import {
  useCreateConnection,
  useConnections,
  useStartOAuth,
  usePlatforms,
} from '@/lib/hooks/use-connections'
import type { PlatformDefinition } from '@/lib/api/connections'
import type { PlatformConnection } from '@myfusionhelper/types'

interface ConnectCRMStepProps {
  onNext: () => void
  onBack: () => void
  onSkip: () => void
}

export function ConnectCRMStep({ onNext, onBack, onSkip }: ConnectCRMStepProps) {
  const [selectedPlatform, setSelectedPlatform] = useState<PlatformDefinition | null>(null)
  const [connectionName, setConnectionName] = useState('')
  const [credentialValues, setCredentialValues] = useState<Record<string, string>>({})

  const { data: connections } = useConnections()
  const { data: platforms, isLoading: platformsLoading } = usePlatforms()
  const createConnection = useCreateConnection()
  const startOAuth = useStartOAuth()

  const hasConnection = connections && connections.length > 0

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
      return
    }

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
          setSelectedPlatform(null)
          setConnectionName('')
          setCredentialValues({})
        },
      }
    )
  }

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h2 className="text-2xl font-bold">Connect Your CRM</h2>
        <p className="mt-1 text-muted-foreground">
          {hasConnection
            ? 'Great, you are connected! Add more platforms or continue to the next step.'
            : 'Choose your CRM platform to get started. Your credentials are encrypted and we only request the minimum permissions needed. You can add more platforms later.'}
        </p>
      </div>

      {/* Show existing connections */}
      {hasConnection && (
        <div className="space-y-2">
          {connections.map((conn: PlatformConnection) => {
            const platform = platforms?.find((p: PlatformDefinition) => p.platformId === conn.platformId || p.slug === conn.platformId)
            return (
              <div
                key={conn.connectionId}
                className="flex items-center gap-3 rounded-lg border border-success/30 bg-success/5 p-3"
              >
                <CheckCircle className="h-5 w-5 text-success" />
                {platform && <PlatformLogo definition={platform} size={28} />}
                <div className="flex-1">
                  <p className="text-sm font-medium">{conn.name}</p>
                  <p className="text-xs text-muted-foreground">{platform?.name || conn.platformId}</p>
                </div>
                <span className="rounded-full bg-success/10 px-2 py-0.5 text-xs font-medium text-success">
                  Connected
                </span>
              </div>
            )
          })}
        </div>
      )}

      {/* Platform selection or form */}
      {!selectedPlatform ? (
        platformsLoading ? (
          <div className="grid gap-3 sm:grid-cols-2">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="flex items-center gap-3 rounded-lg border bg-card p-4">
                <div className="h-10 w-10 animate-pulse rounded-lg bg-muted" />
                <div className="flex-1">
                  <div className="h-4 w-24 animate-pulse rounded bg-muted" />
                  <div className="mt-1 h-3 w-16 animate-pulse rounded bg-muted" />
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {platforms?.filter((p: PlatformDefinition) => p.types?.includes('crm')).map((platform: PlatformDefinition) => {
              const isConnected = connections?.some(
                (c: PlatformConnection) => c.platformId === platform.platformId || c.platformId === platform.slug
              )
              return (
                <button
                  key={platform.platformId}
                  onClick={() => setSelectedPlatform(platform)}
                  className="flex items-center gap-3 rounded-lg border bg-card p-4 text-left transition-all hover:border-primary hover:shadow-sm"
                >
                  <PlatformLogo definition={platform} size={40} />
                  <div className="flex-1">
                    <h3 className="font-medium">{platform.name}</h3>
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
                      {platform.apiConfig.authType === 'oauth2' ? (
                        <><Shield className="h-3 w-3" /> OAuth 2.0</>
                      ) : (
                        <><Key className="h-3 w-3" /> API Key</>
                      )}
                    </div>
                  </div>
                  {isConnected && <CheckCircle className="h-4 w-4 text-success" />}
                </button>
              )
            })}
          </div>
        )
      ) : (
        <div className="space-y-4 rounded-lg border bg-card p-5">
          <div className="flex items-center gap-3">
            <button
              onClick={() => { setSelectedPlatform(null); setCredentialValues({}) }}
              className="rounded-md p-1.5 hover:bg-accent"
            >
              <ArrowLeft className="h-4 w-4" />
            </button>
            <PlatformLogo definition={selectedPlatform} size={36} />
            <div>
              <h3 className="font-medium">{selectedPlatform.name}</h3>
              <p className="text-xs text-muted-foreground">
                {selectedPlatform.apiConfig.authType === 'oauth2' ? 'Connects via OAuth 2.0' : 'Connects via API Key'}
              </p>
            </div>
          </div>

          {selectedPlatform.apiConfig.authType === 'oauth2' ? (
            <>
              <p className="text-sm text-muted-foreground">
                You&apos;ll be redirected to {selectedPlatform.name} to authorize access.
                We only request the minimum permissions needed.
              </p>
              <button
                onClick={handleConnect}
                disabled={startOAuth.isPending}
                className="inline-flex w-full items-center justify-center gap-2 rounded-md bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
              >
                {startOAuth.isPending && (
                  <Loader2 className="h-4 w-4 animate-spin" />
                )}
                <ExternalLink className="h-4 w-4" />
                Connect with {selectedPlatform.name}
              </button>
            </>
          ) : (
            <>
              <div className="space-y-3">
                <div>
                  <label className="mb-1.5 block text-sm font-medium">Connection Name</label>
                  <input
                    type="text"
                    value={connectionName}
                    onChange={(e) => setConnectionName(e.target.value)}
                    placeholder={`e.g. ${selectedPlatform.name} (Production)`}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  />
                </div>
                {selectedPlatform.credentialFields?.map((field) => (
                  <div key={field.key}>
                    <label className="mb-1.5 block text-sm font-medium">{field.label}</label>
                    <input
                      type={field.inputType === 'password' ? 'password' : 'text'}
                      value={credentialValues[field.key] || ''}
                      onChange={(e) =>
                        setCredentialValues((prev) => ({ ...prev, [field.key]: e.target.value }))
                      }
                      placeholder={field.placeholder}
                      className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    />
                    {field.hint && (
                      <p className="mt-1 text-xs text-muted-foreground">{field.hint}</p>
                    )}
                  </div>
                ))}
              </div>

              <button
                onClick={handleConnect}
                disabled={createConnection.isPending}
                className="inline-flex w-full items-center justify-center gap-2 rounded-md bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
              >
                {createConnection.isPending && (
                  <Loader2 className="h-4 w-4 animate-spin" />
                )}
                <Key className="h-4 w-4" />
                Save Connection
              </button>
            </>
          )}
        </div>
      )}

      {/* Navigation */}
      <div className="flex items-center justify-between pt-2">
        <button
          onClick={onBack}
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Back
        </button>
        <div className="flex items-center gap-3">
          <button
            onClick={onSkip}
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            Skip for now
          </button>
          <button
            onClick={onNext}
            className="inline-flex items-center gap-2 rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            {hasConnection ? 'Continue' : 'Skip & Continue'}
          </button>
        </div>
      </div>
    </div>
  )
}
