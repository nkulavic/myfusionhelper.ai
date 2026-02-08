// User & Account types
// Field names match Go backend JSON (after snake_case → camelCase transform)
export interface User {
  userId: string
  cognitoUserId: string
  email: string
  name: string
  phoneNumber?: string
  company?: string
  status: string
  currentAccountId: string
  notificationPreferences?: NotificationPreferences
  createdAt: string
  updatedAt: string
  lastLoginAt?: string
}

export interface NotificationPreferences {
  executionFailures: boolean
  connectionIssues: boolean
  usageAlerts: boolean
  weeklySummary: boolean
  newFeatures: boolean
  teamActivity: boolean
  realtimeStatus: boolean
  aiInsights: boolean
  systemMaintenance: boolean
  webhookUrl?: string
}

export interface Account {
  accountId: string
  ownerUserId: string
  createdByUserId: string
  name: string
  company: string
  plan: 'free' | 'pro' | 'business'
  status: 'active' | 'suspended' | 'cancelled'
  stripeCustomerId?: string
  settings: AccountSettings
  usage: AccountUsage
  createdAt: string
  updatedAt: string
}

export interface AccountSettings {
  maxHelpers: number
  maxConnections: number
  maxApiKeys: number
  maxTeamMembers: number
  maxExecutions: number
  webhooksEnabled: boolean
}

export interface AccountUsage {
  helpers: number
  connections: number
  apiKeys: number
  teamMembers: number
  monthlyExecutions: number
  monthlyApiRequests: number
}

export interface UserAccount {
  userId: string
  accountId: string
  role: 'owner' | 'admin' | 'member' | 'viewer'
  status: string
  permissions: UserPermissions
  linkedAt: string
  updatedAt: string
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
  connectionId: string
  accountId: string
  userId: string
  platformId: string
  externalUserId?: string
  externalUserEmail?: string
  externalAppId?: string
  externalAppName?: string
  name: string
  status: 'active' | 'disconnected' | 'error'
  authType: string
  authId?: string
  credentialsMetadata: Record<string, unknown>
  lastConnected?: string
  createdAt: string
  updatedAt: string
  expiresAt?: string
  lastSyncedAt?: string
  syncStatus?: string
  syncRecordCounts?: Record<string, number>
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
  helperId: string
  accountId: string
  createdBy: string
  connectionId?: string
  shortKey?: string
  name: string
  description: string
  helperType: string
  category: HelperCategory
  status: 'active' | 'deleted'
  config: Record<string, unknown>
  configSchema?: Record<string, unknown>
  enabled: boolean
  executionCount: number
  lastExecutedAt?: string
  scheduleEnabled: boolean
  cronExpression?: string
  scheduleRuleArn?: string
  lastScheduledAt?: string
  nextScheduledAt?: string
  createdAt: string
  updatedAt: string
}

export interface HelperExecution {
  executionId: string
  helperId: string
  accountId: string
  userId?: string
  apiKeyId?: string
  connectionId?: string
  contactId?: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  triggerType: string
  input?: Record<string, unknown>
  output?: Record<string, unknown>
  errorMessage?: string
  durationMs: number
  createdAt: string
  startedAt: string
  completedAt?: string
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
  keyId: string
  accountId: string
  createdBy: string
  name: string
  keyHash: string
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

// Auth context types (returned by /auth/status)
export interface AuthContext {
  userId: string
  accountId: string
  email: string
  role: string
  permissions: UserPermissions
  availableAccounts: AccountAccess[]
}

export interface AccountAccess {
  accountId: string
  accountName: string
  role: string
  permissions: UserPermissions
  isCurrent: boolean
}

// Plan types
export interface PlanLimits {
  maxHelpers: number
  maxExecutions: number
  maxConnections: number
  maxTeamMembers: number
  maxApiKeys: number
}

// API Response types
// Backend returns { success: true, message: "...", data: T } on success
// Backend returns { success: false, error: "..." } on failure
export interface APIResponse<T> {
  success: boolean
  message?: string
  data?: T
  error?: string
}
