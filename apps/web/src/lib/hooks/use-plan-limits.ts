import { useBillingInfo } from './use-settings'
import { PLAN_CONFIGS, type PlanId } from '@/lib/plan-constants'

type Resource = 'helpers' | 'connections' | 'apiKeys' | 'teamMembers' | 'executions'

const RESOURCE_TO_LIMIT_KEY: Record<Resource, string> = {
  helpers: 'maxHelpers',
  connections: 'maxConnections',
  apiKeys: 'maxApiKeys',
  teamMembers: 'maxTeamMembers',
  executions: 'maxExecutions',
}

const RESOURCE_TO_USAGE_KEY: Record<Resource, string> = {
  helpers: 'helpers',
  connections: 'connections',
  apiKeys: 'apiKeys',
  teamMembers: 'teamMembers',
  executions: 'monthlyExecutions',
}

export function usePlanLimits() {
  const { data: billing, isLoading } = useBillingInfo()

  const plan = (billing?.plan ?? 'trial') as PlanId
  const planConfig = PLAN_CONFIGS[plan] ?? PLAN_CONFIGS.trial

  const limits = billing?.limits ?? {
    maxHelpers: planConfig.maxHelpers,
    maxConnections: planConfig.maxConnections,
    maxApiKeys: planConfig.maxApiKeys,
    maxTeamMembers: planConfig.maxTeamMembers,
    maxExecutions: planConfig.maxExecutions,
    webhooksEnabled: false,
  }

  const usage = billing?.usage ?? {
    helpers: 0,
    connections: 0,
    apiKeys: 0,
    teamMembers: 0,
    monthlyExecutions: 0,
    monthlyApiRequests: 0,
  }

  function getLimit(resource: Resource): number {
    const key = RESOURCE_TO_LIMIT_KEY[resource]
    return (limits as unknown as Record<string, number>)[key] ?? 0
  }

  function getUsage(resource: Resource): number {
    const key = RESOURCE_TO_USAGE_KEY[resource]
    return (usage as Record<string, number>)[key] ?? 0
  }

  function isAtLimit(resource: Resource): boolean {
    return getUsage(resource) >= getLimit(resource)
  }

  function percentUsed(resource: Resource): number {
    const limit = getLimit(resource)
    if (limit === 0) return 0
    return Math.min(100, Math.round((getUsage(resource) / limit) * 100))
  }

  const isTrialing = billing?.isTrialing ?? false
  const isTrialExpired = billing?.trialExpired ?? false
  const daysRemaining = billing?.daysRemaining ?? 0

  function canCreate(resource: Resource): boolean {
    if (isTrialExpired) return false
    return getUsage(resource) < getLimit(resource)
  }

  return {
    plan,
    planConfig,
    limits,
    usage,
    billing,
    isLoading,
    isAtLimit,
    percentUsed,
    canCreate,
    getLimit,
    getUsage,
    isSandbox: plan === 'free',
    isTrialing,
    isTrialExpired,
    daysRemaining,
    isPaid: plan !== 'free' && plan !== 'trial',
  }
}
