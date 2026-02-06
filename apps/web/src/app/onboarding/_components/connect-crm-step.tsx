'use client'

import { useState } from 'react'
import { Shield, Key, ExternalLink, CheckCircle, Loader2, ArrowLeft } from 'lucide-react'
import { crmPlatforms, type CRMPlatform } from '@/lib/crm-platforms'
import { PlatformLogo } from '@/components/platform-logo'
import { useStartOAuth, useCreateConnection, useConnections } from '@/lib/hooks/use-connections'

interface ConnectCRMStepProps {
  onNext: () => void
  onBack: () => void
  onSkip: () => void
}

export function ConnectCRMStep({ onNext, onBack, onSkip }: ConnectCRMStepProps) {
  const [selectedPlatform, setSelectedPlatform] = useState<CRMPlatform | null>(null)
  const [connectionName, setConnectionName] = useState('')
  const [apiKeyInput, setApiKeyInput] = useState('')
  const [apiUrlInput, setApiUrlInput] = useState('')
  const [appIdInput, setAppIdInput] = useState('')

  const { data: connections } = useConnections()
  const startOAuth = useStartOAuth()
  const createConnection = useCreateConnection()

  const hasConnection = connections && connections.length > 0

  const handleConnect = async () => {
    if (!selectedPlatform) return

    if (selectedPlatform.authType === 'oauth2') {
      startOAuth.mutate(selectedPlatform.id, {
        onSuccess: (res) => {
          if (res.data?.url) {
            window.location.href = res.data.url
          }
        },
      })
    } else {
      createConnection.mutate(
        {
          platformId: selectedPlatform.id,
          input: {
            name: connectionName || `${selectedPlatform.name} Connection`,
            credentials: {
              apiKey: apiKeyInput || undefined,
              apiUrl: apiUrlInput || undefined,
              appId: appIdInput || undefined,
            },
          },
        },
        {
          onSuccess: () => {
            setSelectedPlatform(null)
            setConnectionName('')
            setApiKeyInput('')
            setApiUrlInput('')
            setAppIdInput('')
          },
        }
      )
    }
  }

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h2 className="text-2xl font-bold">Connect Your CRM</h2>
        <p className="mt-1 text-muted-foreground">
          {hasConnection
            ? 'Great, you have a connection! Add more or continue to the next step.'
            : 'Choose your CRM platform to get started. You can add more later.'}
        </p>
      </div>

      {/* Show existing connections */}
      {hasConnection && (
        <div className="space-y-2">
          {connections.map((conn) => {
            const platform = crmPlatforms.find((p) => p.id === conn.platformId)
            return (
              <div
                key={conn.connectionId}
                className="flex items-center gap-3 rounded-lg border border-success/30 bg-success/5 p-3"
              >
                <CheckCircle className="h-5 w-5 text-success" />
                {platform && <PlatformLogo platform={platform} size={28} />}
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
        <div className="grid gap-3 sm:grid-cols-2">
          {crmPlatforms.map((platform) => {
            const isConnected = connections?.some((c) => c.platformId === platform.id)
            return (
              <button
                key={platform.id}
                onClick={() => setSelectedPlatform(platform)}
                className="flex items-center gap-3 rounded-lg border bg-card p-4 text-left transition-all hover:border-primary hover:shadow-sm"
              >
                <PlatformLogo platform={platform} size={40} />
                <div className="flex-1">
                  <h3 className="font-medium">{platform.name}</h3>
                  <div className="flex items-center gap-1 text-xs text-muted-foreground">
                    {platform.authType === 'oauth2' ? (
                      <>
                        <Shield className="h-3 w-3" /> OAuth 2.0
                      </>
                    ) : (
                      <>
                        <Key className="h-3 w-3" /> API Key
                      </>
                    )}
                  </div>
                </div>
                {isConnected && <CheckCircle className="h-4 w-4 text-success" />}
              </button>
            )
          })}
        </div>
      ) : (
        <div className="space-y-4 rounded-lg border bg-card p-5">
          <div className="flex items-center gap-3">
            <button
              onClick={() => setSelectedPlatform(null)}
              className="rounded-md p-1.5 hover:bg-accent"
            >
              <ArrowLeft className="h-4 w-4" />
            </button>
            <PlatformLogo platform={selectedPlatform} size={36} />
            <div>
              <h3 className="font-medium">{selectedPlatform.name}</h3>
              <p className="text-xs text-muted-foreground">
                {selectedPlatform.authType === 'oauth2'
                  ? 'Connects via secure OAuth 2.0'
                  : 'Connects via API Key'}
              </p>
            </div>
          </div>

          {selectedPlatform.authType === 'api_key' && (
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
              {selectedPlatform.id === 'activecampaign' && (
                <div>
                  <label className="mb-1.5 block text-sm font-medium">Account URL</label>
                  <input
                    type="url"
                    value={apiUrlInput}
                    onChange={(e) => setApiUrlInput(e.target.value)}
                    placeholder="https://yourname.api-us1.com"
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  />
                </div>
              )}
              {selectedPlatform.id === 'ontraport' && (
                <div>
                  <label className="mb-1.5 block text-sm font-medium">App ID</label>
                  <input
                    type="text"
                    value={appIdInput}
                    onChange={(e) => setAppIdInput(e.target.value)}
                    placeholder="Your Ontraport App ID"
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  />
                </div>
              )}
              <div>
                <label className="mb-1.5 block text-sm font-medium">API Key</label>
                <input
                  type="password"
                  value={apiKeyInput}
                  onChange={(e) => setApiKeyInput(e.target.value)}
                  placeholder="Enter your API key"
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                />
                <p className="mt-1 text-xs text-muted-foreground">
                  Your API key is encrypted and stored securely
                </p>
              </div>
            </div>
          )}

          <button
            onClick={handleConnect}
            disabled={createConnection.isPending || startOAuth.isPending}
            className="inline-flex w-full items-center justify-center gap-2 rounded-md bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            {(createConnection.isPending || startOAuth.isPending) && (
              <Loader2 className="h-4 w-4 animate-spin" />
            )}
            {selectedPlatform.authType === 'oauth2' ? (
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
          </button>

          {selectedPlatform.authType === 'oauth2' && (
            <p className="text-center text-xs text-muted-foreground">
              You&apos;ll be redirected to {selectedPlatform.name} to authorize access.
            </p>
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
