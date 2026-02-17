'use client'

import { useState } from 'react'
import { Loader2, CheckCircle, XCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useAISettingsStore,
  groqModels,
  modelLabels,
} from '@/lib/stores/ai-settings-store'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export function AITab() {
  const { preferredModel, setPreferredModel } = useAISettingsStore()
  const [testStatus, setTestStatus] = useState<'idle' | 'testing' | 'success' | 'error'>('idle')
  const [testError, setTestError] = useState('')

  const handleTest = async () => {
    setTestStatus('testing')
    setTestError('')
    try {
      const res = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          messages: [
            { role: 'user', content: 'Say "Connection successful!" in 3 words or less.' },
          ],
          model: preferredModel,
        }),
      })

      if (!res.ok) {
        const text = await res.text()
        throw new Error(text || `HTTP ${res.status}`)
      }

      setTestStatus('success')
    } catch (err) {
      setTestStatus('error')
      setTestError(err instanceof Error ? err.message : 'Connection failed')
    }
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">AI Assistant Configuration</CardTitle>
          <CardDescription>
            Groq is included free for all users. Choose your preferred model for the AI assistant.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Model Selection */}
          <div className="space-y-2">
            <Label htmlFor="ai-model">Model</Label>
            <select
              id="ai-model"
              value={preferredModel}
              onChange={(e) => setPreferredModel(e.target.value)}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            >
              {groqModels.map((model) => (
                <option key={model} value={model}>
                  {modelLabels[model] || model}
                </option>
              ))}
            </select>
          </div>

          {/* Test Connection */}
          <div className="flex items-center gap-3 pt-2">
            <Button onClick={handleTest} disabled={testStatus === 'testing'}>
              {testStatus === 'testing' && <Loader2 className="h-4 w-4 animate-spin" />}
              Test Connection
            </Button>
            {testStatus === 'success' && (
              <span className="inline-flex items-center gap-1.5 text-sm text-success">
                <CheckCircle className="h-4 w-4" />
                Connected successfully
              </span>
            )}
            {testStatus === 'error' && (
              <span className="inline-flex items-center gap-1.5 text-sm text-destructive">
                <XCircle className="h-4 w-4" />
                {testError}
              </span>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
