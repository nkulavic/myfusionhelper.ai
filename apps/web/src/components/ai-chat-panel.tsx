'use client'

import { useEffect, useRef, useState, useMemo, type FormEvent } from 'react'
import { useChat } from '@ai-sdk/react'
import { DefaultChatTransport } from 'ai'
import { Sparkles, Send, Loader2, Settings, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useAISettingsStore,
  providerLabels,
  modelLabels,
} from '@/lib/stores/ai-settings-store'

interface AIChatPanelProps {
  onClose: () => void
}

function getTextContent(msg: { parts?: Array<{ type: string; text?: string }> }): string {
  if (!msg.parts) return ''
  return msg.parts
    .filter((p): p is { type: 'text'; text: string } => p.type === 'text' && typeof p.text === 'string')
    .map((p) => p.text)
    .join('')
}

export function AIChatPanel({ onClose }: AIChatPanelProps) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [input, setInput] = useState('')
  const {
    preferredProvider,
    preferredModel,
    anthropicApiKey,
    openaiApiKey,
  } = useAISettingsStore()

  const needsKey =
    (preferredProvider === 'anthropic' && !anthropicApiKey) ||
    (preferredProvider === 'openai' && !openaiApiKey)

  const transport = useMemo(
    () =>
      new DefaultChatTransport({
        api: '/api/chat',
        body: { provider: preferredProvider, model: preferredModel },
        headers: {
          ...(preferredProvider === 'anthropic' && anthropicApiKey
            ? { 'x-anthropic-key': anthropicApiKey }
            : {}),
          ...(preferredProvider === 'openai' && openaiApiKey
            ? { 'x-openai-key': openaiApiKey }
            : {}),
        },
      }),
    [preferredProvider, preferredModel, anthropicApiKey, openaiApiKey]
  )

  const { messages, sendMessage, status, error } = useChat({ transport })

  const isStreaming = status === 'streaming' || status === 'submitted'

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages])

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!input.trim() || needsKey || isStreaming) return
    sendMessage({ text: input })
    setInput('')
  }

  return (
    <div className="fixed bottom-24 right-6 z-50 flex h-[520px] w-[400px] flex-col rounded-xl border bg-card shadow-2xl">
      {/* Header */}
      <div className="flex items-center justify-between border-b px-4 py-3">
        <div className="flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-primary" />
          <div>
            <h3 className="text-sm font-semibold">AI Assistant</h3>
            <p className="text-[10px] text-muted-foreground">
              {providerLabels[preferredProvider]} &middot;{' '}
              {modelLabels[preferredModel] || preferredModel}
            </p>
          </div>
        </div>
        <button
          onClick={onClose}
          className="rounded-md p-1 hover:bg-accent"
        >
          <span className="text-lg leading-none">&times;</span>
        </button>
      </div>

      {/* Messages */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto p-4 space-y-4">
        {needsKey ? (
          <div className="flex flex-col items-center justify-center h-full gap-3 text-center px-4">
            <AlertCircle className="h-10 w-10 text-warning" />
            <p className="text-sm font-medium">API Key Required</p>
            <p className="text-xs text-muted-foreground">
              {preferredProvider === 'anthropic' ? 'Claude' : 'OpenAI'} requires
              your own API key. Configure it in Settings.
            </p>
            <a
              href="/settings"
              className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground hover:bg-primary/90"
            >
              <Settings className="h-3 w-3" />
              Go to Settings
            </a>
          </div>
        ) : messages.length === 0 ? (
          <div className="flex gap-3">
            <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-primary text-xs text-primary-foreground">
              AI
            </div>
            <div className="rounded-lg bg-muted px-3 py-2 text-sm">
              <p>
                Hi! I can help you with your CRM data and automations. Try
                asking me:
              </p>
              <ul className="mt-2 space-y-1 text-muted-foreground">
                <li>&bull; &quot;How do I set up a tag helper?&quot;</li>
                <li>&bull; &quot;What helpers work with HubSpot?&quot;</li>
                <li>&bull; &quot;Help me automate follow-up emails&quot;</li>
              </ul>
            </div>
          </div>
        ) : (
          messages.map((msg) => {
            const text = getTextContent(msg)
            if (!text) return null
            return (
              <div
                key={msg.id}
                className={cn(
                  'flex gap-3',
                  msg.role === 'user' ? 'flex-row-reverse' : ''
                )}
              >
                <div
                  className={cn(
                    'flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-xs font-medium',
                    msg.role === 'user'
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted text-muted-foreground'
                  )}
                >
                  {msg.role === 'user' ? 'You' : 'AI'}
                </div>
                <div
                  className={cn(
                    'max-w-[280px] rounded-lg px-3 py-2 text-sm',
                    msg.role === 'user'
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted'
                  )}
                >
                  <p className="whitespace-pre-wrap break-words">{text}</p>
                </div>
              </div>
            )
          })
        )}
        {isStreaming && messages[messages.length - 1]?.role !== 'assistant' && (
          <div className="flex gap-3">
            <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-muted text-xs text-muted-foreground">
              AI
            </div>
            <div className="rounded-lg bg-muted px-3 py-2">
              <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
            </div>
          </div>
        )}
        {error && (
          <div className="rounded-md bg-destructive/10 px-3 py-2 text-xs text-destructive">
            {error.message || 'An error occurred. Please try again.'}
          </div>
        )}
      </div>

      {/* Input */}
      <div className="border-t p-3">
        <form onSubmit={handleSubmit} className="flex gap-2">
          <input
            type="text"
            placeholder={
              needsKey
                ? 'Configure API key first...'
                : 'Ask anything about your data...'
            }
            value={input}
            onChange={(e) => setInput(e.target.value)}
            disabled={needsKey || isStreaming}
            className="flex-1 rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
          />
          <button
            type="submit"
            disabled={needsKey || isStreaming || !input.trim()}
            className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            <Send className="h-4 w-4" />
          </button>
        </form>
      </div>
    </div>
  )
}
