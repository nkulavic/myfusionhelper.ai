/**
 * Chat API Client
 *
 * Handles chat conversations and messages with the Go backend chat service.
 * Supports SSE streaming for real-time responses.
 */

import { apiClient } from './client'

export interface ChatConversation {
  conversationId: string
  title: string
  status: 'active' | 'archived'
  createdAt: string
  updatedAt: string
  messageCount: number
}

export interface ChatMessage {
  messageId: string
  conversationId: string
  sequence: number
  role: 'user' | 'assistant'
  content: string
  toolCalls?: Array<{
    id: string
    type: string
    function: {
      name: string
      arguments: string
    }
  }>
  toolResults?: Array<{
    toolCallId: string
    result: string
  }>
  createdAt: string
}

export interface CreateConversationInput {
  title?: string
}

export interface SendMessageInput {
  content: string
}

export interface StreamChatResponse {
  type: 'content' | 'tool_call' | 'tool_result' | 'done'
  content?: string
  toolCall?: {
    id: string
    type: string
    function: {
      name: string
      arguments: string
    }
  }
  toolResult?: {
    toolCallId: string
    result: string
  }
  done?: boolean
}

/**
 * Create a new conversation
 */
export async function createConversation(
  input: CreateConversationInput = {}
): Promise<{ conversationId: string; title: string; status: string; createdAt: string }> {
  const response = await apiClient.post<{
    conversationId: string
    title: string
    status: string
    createdAt: string
  }>('/chat/conversations', input)
  return response.data!
}

/**
 * List all conversations for the current user
 */
export async function listConversations(): Promise<ChatConversation[]> {
  const response = await apiClient.get<{ conversations: ChatConversation[] }>(
    '/chat/conversations'
  )
  return response.data?.conversations || []
}

/**
 * Get a specific conversation with its messages
 */
export async function getConversation(conversationId: string): Promise<{
  conversation: ChatConversation
  messages: ChatMessage[]
}> {
  const response = await apiClient.get<{
    conversation: ChatConversation
    messages: ChatMessage[]
  }>(`/chat/conversations/${conversationId}`)
  return response.data!
}

/**
 * Delete a conversation (soft delete)
 */
export async function deleteConversation(conversationId: string): Promise<void> {
  await apiClient.delete(`/chat/conversations/${conversationId}`)
}

/**
 * Get messages for a conversation
 */
export async function getMessages(conversationId: string): Promise<ChatMessage[]> {
  const response = await apiClient.get<{
    conversationId: string
    messages: ChatMessage[]
  }>(`/chat/conversations/${conversationId}/messages`)
  return response.data?.messages || []
}

/**
 * Send a message and stream the response via SSE
 *
 * @param conversationId - The conversation ID
 * @param input - The message input
 * @param onChunk - Callback for each SSE chunk
 * @param signal - AbortSignal for cancellation
 */
export async function sendMessage(
  conversationId: string,
  input: SendMessageInput,
  onChunk: (chunk: StreamChatResponse) => void,
  signal?: AbortSignal
): Promise<void> {
  const token = localStorage.getItem('mfh_access_token')
  if (!token) {
    throw new Error('Not authenticated')
  }

  const baseURL = process.env.NEXT_PUBLIC_API_URL || ''
  const response = await fetch(`${baseURL}/chat/conversations/${conversationId}/messages`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({ content: input.content }),
    signal,
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Failed to send message' }))
    throw new Error(error.error || 'Failed to send message')
  }

  const reader = response.body?.getReader()
  if (!reader) {
    throw new Error('No response body')
  }

  const decoder = new TextDecoder()
  let buffer = ''

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (!line.trim() || !line.startsWith('data: ')) continue

        const data = line.slice(6)
        if (data === '[DONE]') {
          onChunk({ type: 'done', done: true })
          return
        }

        try {
          const chunk = JSON.parse(data) as StreamChatResponse
          onChunk(chunk)
        } catch (e) {
          console.error('Failed to parse SSE chunk:', e, data)
        }
      }
    }
  } finally {
    reader.releaseLock()
  }
}

export const chatApi = {
  createConversation,
  listConversations,
  getConversation,
  deleteConversation,
  getMessages,
  sendMessage,
}
