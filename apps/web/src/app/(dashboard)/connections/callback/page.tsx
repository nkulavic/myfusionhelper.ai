'use client'

import { Suspense, useEffect, useState } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { CheckCircle, XCircle, Loader2 } from 'lucide-react'

type CallbackState = 'processing' | 'success' | 'error'

function OAuthCallbackContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const [state, setState] = useState<CallbackState>('processing')
  const [message, setMessage] = useState('')
  const [platform, setPlatform] = useState('')

  useEffect(() => {
    const code = searchParams.get('code')
    const stateParam = searchParams.get('state')
    const error = searchParams.get('error')
    const errorDescription = searchParams.get('error_description')

    if (error) {
      setState('error')
      setMessage(errorDescription || error || 'OAuth authorization was denied.')
      return
    }

    if (!code || !stateParam) {
      setState('error')
      setMessage('Missing required OAuth parameters (code or state).')
      return
    }

    // Parse the state to get platform info
    try {
      const stateData = JSON.parse(atob(stateParam))
      setPlatform(stateData.platform || 'CRM')
    } catch {
      setPlatform('CRM')
    }

    // Exchange the code for tokens via our API
    async function exchangeCode(oauthCode: string, oauthState: string) {
      try {
        const apiBase = process.env.NEXT_PUBLIC_API_URL || 'https://api.myfusionhelper.ai'
        // The backend handles OAuth callback as a GET redirect from the CRM provider
        // We forward the code and state to complete the exchange
        const callbackUrl = new URL(`${apiBase}/platforms/oauth/callback`)
        callbackUrl.searchParams.set('code', oauthCode)
        callbackUrl.searchParams.set('state', oauthState)
        const response = await fetch(callbackUrl.toString(), {
          method: 'GET',
          headers: { 'Content-Type': 'application/json' },
        })

        if (!response.ok) {
          const data = await response.json().catch(() => ({}))
          throw new Error(data.message || `Failed to complete OAuth (${response.status})`)
        }

        setState('success')
        setMessage('Connection established successfully!')

        // Redirect to connections page after brief delay
        setTimeout(() => {
          router.push('/connections')
        }, 2000)
      } catch (err) {
        setState('error')
        setMessage(err instanceof Error ? err.message : 'Failed to complete OAuth exchange.')
      }
    }

    exchangeCode(code, stateParam)
  }, [searchParams, router])

  return (
    <div className="flex min-h-[60vh] items-center justify-center">
      <div className="w-full max-w-md rounded-lg border bg-card p-8 text-center">
        {state === 'processing' && (
          <>
            <Loader2 className="mx-auto h-12 w-12 animate-spin text-primary" />
            <h2 className="mt-4 text-xl font-bold">Connecting {platform}...</h2>
            <p className="mt-2 text-sm text-muted-foreground">
              Please wait while we complete the authorization.
            </p>
          </>
        )}

        {state === 'success' && (
          <>
            <CheckCircle className="mx-auto h-12 w-12 text-success" />
            <h2 className="mt-4 text-xl font-bold">Connected!</h2>
            <p className="mt-2 text-sm text-muted-foreground">{message}</p>
            <p className="mt-4 text-xs text-muted-foreground">Redirecting to connections...</p>
          </>
        )}

        {state === 'error' && (
          <>
            <XCircle className="mx-auto h-12 w-12 text-destructive" />
            <h2 className="mt-4 text-xl font-bold">Connection Failed</h2>
            <p className="mt-2 text-sm text-muted-foreground">{message}</p>
            <div className="mt-6 flex items-center justify-center gap-3">
              <button
                onClick={() => router.push('/connections')}
                className="rounded-md border px-4 py-2 text-sm font-medium hover:bg-accent"
              >
                Back to Connections
              </button>
              <button
                onClick={() => window.location.reload()}
                className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
              >
                Try Again
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}

export default function OAuthCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-[60vh] items-center justify-center">
          <div className="w-full max-w-md rounded-lg border bg-card p-8 text-center">
            <Loader2 className="mx-auto h-12 w-12 animate-spin text-primary" />
            <h2 className="mt-4 text-xl font-bold">Loading...</h2>
          </div>
        </div>
      }
    >
      <OAuthCallbackContent />
    </Suspense>
  )
}
