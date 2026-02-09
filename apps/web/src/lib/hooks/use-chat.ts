/**
 * React Query hooks for chat API
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  chatApi,
  type ChatConversation,
  type ChatMessage,
  type CreateConversationInput,
  type SendMessageInput,
  type StreamChatResponse,
} from '../api/chat'
import { useState, useCallback, useRef } from 'react'

/**
 * Query hook: List all conversations
 */
export function useConversations() {
  return useQuery({
    queryKey: ['conversations'],
    queryFn: () => chatApi.listConversations(),
  })
}

/**
 * Query hook: Get a specific conversation with messages
 */
export function useConversation(conversationId: string | null) {
  return useQuery({
    queryKey: ['conversation', conversationId],
    queryFn: () => {
      if (!conversationId) {
        return Promise.resolve({ conversation: null, messages: [] })
      }
      return chatApi.getConversation(conversationId)
    },
    enabled: !!conversationId,
  })
}

/**
 * Query hook: Get messages for a conversation
 */
export function useMessages(conversationId: string | null) {
  return useQuery({
    queryKey: ['messages', conversationId],
    queryFn: () => {
      if (!conversationId) return Promise.resolve([])
      return chatApi.getMessages(conversationId)
    },
    enabled: !!conversationId,
  })
}

/**
 * Mutation hook: Create a new conversation
 */
export function useCreateConversation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateConversationInput) => chatApi.createConversation(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversations'] })
    },
  })
}

/**
 * Mutation hook: Delete a conversation
 */
export function useDeleteConversation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (conversationId: string) => chatApi.deleteConversation(conversationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversations'] })
    },
  })
}

/**
 * Hook for streaming chat messages
 *
 * Usage:
 * ```tsx
 * const { sendMessage, isStreaming, streamedContent } = useStreamChat(conversationId)
 *
 * const handleSend = async (text: string) => {
 *   await sendMessage({ content: text })
 * }
 * ```
 */
export function useStreamChat(conversationId: string | null) {
  const queryClient = useQueryClient()
  const [isStreaming, setIsStreaming] = useState(false)
  const [streamedContent, setStreamedContent] = useState('')
  const [toolCalls, setToolCalls] = useState<StreamChatResponse['toolCall'][]>([])
  const [error, setError] = useState<string | null>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

  const sendMessage = useCallback(
    async (input: SendMessageInput) => {
      if (!conversationId) {
        setError('No conversation selected')
        return
      }

      setIsStreaming(true)
      setStreamedContent('')
      setToolCalls([])
      setError(null)

      abortControllerRef.current = new AbortController()

      try {
        await chatApi.sendMessage(
          conversationId,
          input,
          (chunk) => {
            if (chunk.type === 'content' && chunk.content) {
              setStreamedContent((prev) => prev + chunk.content)
            } else if (chunk.type === 'tool_call' && chunk.toolCall) {
              setToolCalls((prev) => [...prev, chunk.toolCall!])
            } else if (chunk.type === 'done') {
              setIsStreaming(false)
              // Invalidate messages to refetch from server
              queryClient.invalidateQueries({ queryKey: ['messages', conversationId] })
              queryClient.invalidateQueries({ queryKey: ['conversation', conversationId] })
            }
          },
          abortControllerRef.current.signal
        )
      } catch (err) {
        if (err instanceof Error && err.name === 'AbortError') {
          // User cancelled
          setError('Cancelled')
        } else {
          setError(err instanceof Error ? err.message : 'Failed to send message')
        }
        setIsStreaming(false)
      }
    },
    [conversationId, queryClient]
  )

  const cancel = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      abortControllerRef.current = null
    }
  }, [])

  return {
    sendMessage,
    cancel,
    isStreaming,
    streamedContent,
    toolCalls,
    error,
  }
}
