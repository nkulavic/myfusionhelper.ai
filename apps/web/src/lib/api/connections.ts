import { apiClient } from './client'
import type { PlatformConnection } from '@myfusionhelper/types'

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
  platformId: string
  name: string
  slug: string
  category: string
  description: string
  status: string
  version: string
  logoUrl: string
  documentationUrl: string
  oauth?: {
    authUrl: string
    tokenUrl: string
    userInfoUrl: string
    scopes: string[]
    responseType: string
  }
  apiConfig: {
    baseUrl: string
    authType: string
    testEndpoint: string
    rateLimits: {
      requestsPerSecond: number
      requestsPerMinute: number
      requestsPerHour: number
      burstLimit: number
    }
    requiredHeaders: Record<string, string>
    version: string
  }
  capabilities: string[]
  createdAt: string
  updatedAt: string
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
