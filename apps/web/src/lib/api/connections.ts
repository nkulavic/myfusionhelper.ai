import { apiClient } from './client'
import type { PlatformConnection, CRMPlatform } from '@myfusionhelper/types'

export interface CreateConnectionInput {
  platformId: string
  name: string
  credentials?: {
    apiKey?: string
    apiUrl?: string
    appId?: string
  }
}

export interface UpdateConnectionInput {
  name?: string
  credentials?: {
    apiKey?: string
    apiUrl?: string
    appId?: string
  }
}

export interface PlatformDefinition {
  id: string
  name: string
  category: string
  authType: 'oauth2' | 'api_key'
  oauthConfig?: {
    authorizationUrl: string
    tokenUrl: string
    scopes: string[]
  }
  apiBaseUrl: string
  rateLimit: {
    requestsPerSecond: number
    dailyLimit: number
  }
  capabilities: string[]
}

export const connectionsApi = {
  // List ALL connections across all platforms
  list: () =>
    apiClient.get<PlatformConnection[]>('/platform-connections'),

  // List connections scoped to a specific platform
  listByPlatform: (platformId: string) =>
    apiClient.get<PlatformConnection[]>(
      `/platforms/${platformId}/connections`
    ),

  // Get a single connection
  get: (platformId: string, connectionId: string) =>
    apiClient.get<PlatformConnection>(
      `/platforms/${platformId}/connections/${connectionId}`
    ),

  // Create a connection under a platform
  create: (platformId: string, input: Omit<CreateConnectionInput, 'platformId'>) =>
    apiClient.post<PlatformConnection>(
      `/platforms/${platformId}/connections`,
      input
    ),

  // Update a connection
  update: (platformId: string, connectionId: string, input: UpdateConnectionInput) =>
    apiClient.put<PlatformConnection>(
      `/platforms/${platformId}/connections/${connectionId}`,
      input
    ),

  // Delete a connection
  delete: (platformId: string, connectionId: string) =>
    apiClient.delete<void>(
      `/platforms/${platformId}/connections/${connectionId}`
    ),

  // Test a connection
  test: (platformId: string, connectionId: string) =>
    apiClient.post<{ status: string; message?: string }>(
      `/platforms/${platformId}/connections/${connectionId}/test`
    ),

  // Start OAuth flow for a platform
  startOAuth: (platformId: string) =>
    apiClient.post<{ url: string }>(
      `/platforms/${platformId}/oauth/start`
    ),

  // List available platforms
  listPlatforms: () =>
    apiClient.get<PlatformDefinition[]>('/platforms'),

  // Get a single platform definition
  getPlatform: (platformId: string) =>
    apiClient.get<PlatformDefinition>(`/platforms/${platformId}`),
}
