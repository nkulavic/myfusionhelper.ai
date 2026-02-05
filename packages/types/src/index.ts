// User & Account types
// Field names match Go backend JSON (after snake_case → camelCase transform)
export interface User {
  userId: string
  cognitoUserId?: string
  email: string
  name: string
  phoneNumber?: string
  company?: string
  status: string
  currentAccountId: string
  createdAt: string
  updatedAt?: string
}

export interface Account {
  accountId: string
  ownerUserId: string
  createdByUserId?: string
  name: string
  company?: string
  plan: 'free' | 'pro' | 'business'
  status: 'active' | 'suspended' | 'cancelled'
  stripeCustomerId?: string
  createdAt: string
  updatedAt?: string
}

export interface UserAccount {
  userId: string
  accountId: string
  role: 'owner' | 'admin' | 'member' | 'viewer'
  permissions: UserPermissions
}

export interface UserPermissions {
  canManageHelpers: boolean
  canExecuteHelpers: boolean
  canManageConnections: boolean
  canManageTeam: boolean
  canManageBilling: boolean
  canViewAnalytics: boolean
  canManageAPIKeys: boolean
}

// CRM Platform types
export type CRMPlatform = 'keap' | 'gohighlevel' | 'activecampaign' | 'ontraport' | 'hubspot' | 'generic'

export interface PlatformConnection {
  id: string
  accountId: string
  platform: CRMPlatform
  name: string
  status: 'active' | 'disconnected' | 'error'
  credentials: {
    accessToken?: string
    refreshToken?: string
    expiresAt?: string
  }
  metadata?: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

// Helper types
export type HelperCategory =
  | 'contact'
  | 'data'
  | 'tagging'
  | 'automation'
  | 'integration'
  | 'notification'
  | 'analytics'

export interface Helper {
  id: string
  accountId: string
  name: string
  description: string
  type: string
  category: HelperCategory
  config: Record<string, unknown>
  configSchema?: Record<string, unknown>
  connectionId: string
  status: 'active' | 'deleted'
  enabled: boolean
  executionCount: number
  lastExecutedAt?: string
  createdAt: string
  updatedAt: string
}

export interface HelperExecution {
  id: string
  helperId: string
  accountId: string
  userId?: string
  connectionId?: string
  contactId?: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  triggerType?: string
  input: Record<string, unknown>
  output?: Record<string, unknown>
  error?: string
  createdAt: string
  startedAt: string
  completedAt?: string
  durationMs?: number
}

export interface HelperTypeDefinition {
  type: string
  name: string
  category: HelperCategory
  description: string
  requiresCrm: boolean
  supportedCrms: string[]
  configSchema: Record<string, unknown>
}

// API Key types
export interface APIKey {
  id: string
  accountId: string
  createdBy: string
  name: string
  keyPrefix: string
  permissions: string[]
  status: 'active' | 'revoked'
  lastUsedAt?: string
  createdAt: string
  expiresAt?: string
}

// Normalized CRM data
export interface NormalizedContact {
  id: string
  firstName: string
  lastName: string
  email: string
  phone?: string
  company?: string
  tags: string[]
  customFields: Record<string, unknown>
  sourceCRM: CRMPlatform
  sourceId: string
  createdAt: string
  updatedAt: string
}

// API Response types
export interface APIResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
  }
  meta?: {
    page?: number
    limit?: number
    total?: number
    nextToken?: string
    hasMore?: boolean
  }
}
