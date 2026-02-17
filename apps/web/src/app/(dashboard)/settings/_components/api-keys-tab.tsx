'use client'

import { useState } from 'react'
import { Copy, Loader2, Plus } from 'lucide-react'
import { useAPIKeys, useCreateAPIKey, useRevokeAPIKey } from '@/lib/hooks/use-settings'
import { usePlanLimits } from '@/lib/hooks/use-plan-limits'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

export function APIKeysTab() {
  const { data: apiKeys, isLoading } = useAPIKeys()
  const createKey = useCreateAPIKey()
  const revokeKey = useRevokeAPIKey()
  const { canCreate: canCreateResource, getUsage, getLimit } = usePlanLimits()
  const [newKeyName, setNewKeyName] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [newKey, setNewKey] = useState<string | null>(null)
  const atApiKeyLimit = !canCreateResource('apiKeys')

  const handleCreate = () => {
    if (!newKeyName.trim()) return
    createKey.mutate(
      { name: newKeyName, permissions: ['execute_helpers'] },
      {
        onSuccess: (res) => {
          if (res.data && 'key' in res.data) {
            setNewKey(res.data.key)
          }
          setNewKeyName('')
          setShowCreate(false)
        },
      }
    )
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-lg">API Keys</CardTitle>
              <CardDescription>
                Use API keys to authenticate helper executions from your CRM automations.
              </CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xs text-muted-foreground">
                {getUsage('apiKeys')} / {getLimit('apiKeys')} keys
              </span>
              <Button onClick={() => setShowCreate(!showCreate)} disabled={atApiKeyLimit}>
                <Plus className="h-4 w-4" />
                Create New Key
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {showCreate && (
            <div className="flex items-center gap-2 rounded-md border p-4">
              <Input
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                placeholder="Key name (e.g. Production)"
                className="flex-1"
              />
              <Button onClick={handleCreate} disabled={createKey.isPending || !newKeyName.trim()}>
                {createKey.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                Create
              </Button>
              <Button variant="outline" onClick={() => setShowCreate(false)}>
                Cancel
              </Button>
            </div>
          )}

          {newKey && (
            <div className="rounded-md border border-success/30 bg-success/10 p-4">
              <p className="mb-2 text-sm font-medium text-success">
                API key created! Copy it now -- you won&apos;t be able to see it again.
              </p>
              <div className="flex items-center gap-2">
                <code className="flex-1 rounded bg-background p-2 font-mono text-sm">{newKey}</code>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    navigator.clipboard.writeText(newKey)
                    setNewKey(null)
                  }}
                >
                  <Copy className="h-3 w-3" />
                  Copy
                </Button>
              </div>
            </div>
          )}

          {isLoading ? (
            <div className="space-y-3">
              {[1, 2].map((i) => (
                <div key={i} className="flex items-center justify-between rounded-md border p-4">
                  <div>
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="mt-1 h-3 w-40" />
                  </div>
                  <Skeleton className="h-8 w-20" />
                </div>
              ))}
            </div>
          ) : apiKeys && apiKeys.length > 0 ? (
            <div className="space-y-3">
              {apiKeys.map((key) => (
                <div key={key.keyId} className="flex items-center justify-between rounded-md border p-4">
                  <div>
                    <p className="font-medium">{key.name}</p>
                    <p className="font-mono text-sm text-muted-foreground">{key.keyPrefix}...</p>
                    <p className="text-xs text-muted-foreground">
                      Created {new Date(key.createdAt).toLocaleDateString()}
                      {key.lastUsedAt &&
                        ` -- Last used ${new Date(key.lastUsedAt).toLocaleDateString()}`}
                    </p>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-destructive hover:bg-destructive/10"
                    onClick={() => revokeKey.mutate(key.keyId)}
                    disabled={revokeKey.isPending}
                  >
                    Revoke
                  </Button>
                </div>
              ))}
            </div>
          ) : (
            <div className="py-8 text-center text-sm text-muted-foreground">
              No API keys yet. Create one to start executing helpers via API.
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
