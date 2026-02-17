'use client'

import { useEffect, useRef, useState, type FormEvent } from 'react'
import { Sparkles, Send, Loader2, MessageSquare, Trash2, Plus } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useConversations,
  useMessages,
  useCreateConversation,
  useDeleteConversation,
  useStreamChat,
} from '@/lib/hooks/use-chat'
import type { ChatMessage } from '@/lib/api/chat'

interface AIChatPanelProps {
  onClose: () => void
}

export function AIChatPanel({ onClose }: AIChatPanelProps) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [input, setInput] = useState('')
  const [selectedConversationId, setSelectedConversationId] = useState<string | null>(null)
  const [showConversations, setShowConversations] = useState(false)

  // Queries
  const { data: conversations = [], isLoading: loadingConversations } = useConversations()
  const { data: messagesData, isLoading: loadingMessages } = useMessages(selectedConversationId)
  const messages: ChatMessage[] = messagesData || []

  // Mutations
  const createConversation = useCreateConversation()
  const deleteConversation = useDeleteConversation()

  // Streaming
  const { sendMessage, isStreaming, streamedContent, toolCalls, error } =
    useStreamChat(selectedConversationId)

  // Auto-select first conversation or create new one
  useEffect(() => {
    if (!selectedConversationId && conversations.length > 0) {
      setSelectedConversationId(conversations[0].conversationId)
    }
  }, [conversations, selectedConversationId])

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages, streamedContent])

  const handleCreateConversation = async () => {
    try {
      const result = await createConversation.mutateAsync({ title: 'New Conversation' })
      setSelectedConversationId(result.conversationId)
      setShowConversations(false)
    } catch (err) {
      console.error('Failed to create conversation:', err)
    }
  }

  const handleDeleteConversation = async (conversationId: string) => {
    try {
      await deleteConversation.mutateAsync(conversationId)
      if (selectedConversationId === conversationId) {
        setSelectedConversationId(null)
      }
    } catch (err) {
      console.error('Failed to delete conversation:', err)
    }
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isStreaming) return

    // Create conversation if none selected
    if (!selectedConversationId) {
      try {
        const result = await createConversation.mutateAsync({ title: input.slice(0, 50) })
        setSelectedConversationId(result.conversationId)
        // Send message after conversation is created
        await sendMessage({ content: input })
        setInput('')
      } catch (err) {
        console.error('Failed to create conversation:', err)
      }
    } else {
      await sendMessage({ content: input })
      setInput('')
    }
  }

  const currentConversation = conversations.find(
    (c) => c.conversationId === selectedConversationId
  )

  return (
    <div className="fixed bottom-24 right-6 z-50 flex h-[520px] w-[400px] flex-col rounded-xl border bg-card shadow-2xl">
      {/* Header */}
      <div className="flex items-center justify-between border-b px-4 py-3">
        <div className="flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-primary" />
          <div>
            <h3 className="text-sm font-semibold">AI Assistant</h3>
            <p className="text-[10px] text-muted-foreground">
              {currentConversation?.title || 'No conversation'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={() => setShowConversations(!showConversations)}
            className="rounded-md p-1 hover:bg-accent"
            title="Conversations"
          >
            <MessageSquare className="h-4 w-4" />
          </button>
          <button onClick={onClose} className="rounded-md p-1 hover:bg-accent">
            <span className="text-lg leading-none">&times;</span>
          </button>
        </div>
      </div>

      {/* Conversations Sidebar (toggled) */}
      {showConversations && (
        <div className="border-b bg-muted/50 p-2">
          <div className="mb-2 flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">Conversations</span>
            <button
              onClick={handleCreateConversation}
              disabled={createConversation.isPending}
              className="rounded-md p-1 hover:bg-accent disabled:opacity-50"
              title="New conversation"
            >
              <Plus className="h-3 w-3" />
            </button>
          </div>
          <div className="max-h-32 space-y-1 overflow-y-auto">
            {loadingConversations ? (
              <div className="flex items-center justify-center py-4">
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              </div>
            ) : conversations.length === 0 ? (
              <p className="py-2 text-center text-xs text-muted-foreground">
                No conversations yet
              </p>
            ) : (
              conversations.map((conv) => (
                <div
                  key={conv.conversationId}
                  className={cn(
                    'flex items-center justify-between rounded-md px-2 py-1.5 text-xs hover:bg-accent',
                    conv.conversationId === selectedConversationId && 'bg-accent'
                  )}
                >
                  <button
                    onClick={() => {
                      setSelectedConversationId(conv.conversationId)
                      setShowConversations(false)
                    }}
                    className="flex-1 truncate text-left"
                  >
                    {conv.title}
                  </button>
                  <button
                    onClick={() => handleDeleteConversation(conv.conversationId)}
                    disabled={deleteConversation.isPending}
                    className="rounded-md p-0.5 hover:bg-destructive/10 hover:text-destructive disabled:opacity-50"
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                </div>
              ))
            )}
          </div>
        </div>
      )}

      {/* Messages */}
      <div ref={scrollRef} className="flex-1 space-y-4 overflow-y-auto p-4">
        {loadingMessages ? (
          <div className="flex h-full items-center justify-center">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : messages.length === 0 && !streamedContent ? (
          <div className="flex gap-3">
            <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-primary text-xs text-primary-foreground">
              AI
            </div>
            <div className="rounded-lg bg-muted px-3 py-2 text-sm">
              <p>Hi! I can help you with your CRM data and automations. Try asking me:</p>
              <ul className="mt-2 space-y-1 text-muted-foreground">
                <li>&bull; &quot;Show my Keap contacts&quot;</li>
                <li>&bull; &quot;What helpers work with HubSpot?&quot;</li>
                <li>&bull; &quot;Tag all contacts with email john@example.com as VIP&quot;</li>
              </ul>
            </div>
          </div>
        ) : (
          <>
            {messages.map((msg) => (
              <div
                key={msg.messageId}
                className={cn('flex gap-3', msg.role === 'user' ? 'flex-row-reverse' : '')}
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
                    msg.role === 'user' ? 'bg-primary text-primary-foreground' : 'bg-muted'
                  )}
                >
                  <p className="whitespace-pre-wrap break-words">{msg.content}</p>
                  {msg.toolCalls && msg.toolCalls.length > 0 && (
                    <div className="mt-2 space-y-1 border-t pt-2 text-xs opacity-75">
                      {msg.toolCalls.map((tc, i) => (
                        <div key={i}>
                          🔧 {tc.function.name}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            ))}
            {/* Streaming message */}
            {(isStreaming || streamedContent) && (
              <div className="flex gap-3">
                <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-muted text-xs text-muted-foreground">
                  AI
                </div>
                <div className="max-w-[280px] rounded-lg bg-muted px-3 py-2 text-sm">
                  {streamedContent ? (
                    <>
                      <p className="whitespace-pre-wrap break-words">{streamedContent}</p>
                      {toolCalls.length > 0 && (
                        <div className="mt-2 space-y-1 border-t pt-2 text-xs opacity-75">
                          {toolCalls.map((tc, i) => (
                            <div key={i}>
                              🔧 {tc?.function?.name}
                            </div>
                          ))}
                        </div>
                      )}
                      {isStreaming && <span className="animate-pulse">▊</span>}
                    </>
                  ) : (
                    <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                  )}
                </div>
              </div>
            )}
          </>
        )}
        {error && (
          <div className="rounded-md bg-destructive/10 px-3 py-2 text-xs text-destructive">
            {error}
          </div>
        )}
      </div>

      {/* Input */}
      <div className="border-t p-3">
        <form onSubmit={handleSubmit} className="flex gap-2">
          <input
            type="text"
            placeholder="Ask about your CRM data..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            disabled={isStreaming}
            className="flex-1 rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
          />
          <button
            type="submit"
            disabled={isStreaming || !input.trim()}
            className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            <Send className="h-4 w-4" />
          </button>
        </form>
      </div>
    </div>
  )
}
