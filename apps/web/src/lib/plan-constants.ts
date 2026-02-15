// Shared plan configuration — single source of truth for all pricing displays.
// Must match backend/golang/internal/billing/plans.go

export type PlanId = 'free' | 'start' | 'grow' | 'deliver'

export interface PlanConfig {
  id: PlanId
  name: string
  description: string
  monthlyPrice: number
  annualPrice: number // total per year
  annualMonthlyPrice: number // annualPrice / 12 (displayed as "per month")
  maxHelpers: number
  maxConnections: number
  maxExecutions: number
  maxApiKeys: number
  maxTeamMembers: number
  overageRate: number // cost per execution beyond included (0 = blocked)
}

export const PLAN_CONFIGS: Record<PlanId, PlanConfig> = {
  free: {
    id: 'free',
    name: 'Sandbox',
    description: 'Free tier with limited access',
    monthlyPrice: 0,
    annualPrice: 0,
    annualMonthlyPrice: 0,
    maxHelpers: 1,
    maxConnections: 1,
    maxExecutions: 100,
    maxApiKeys: 1,
    maxTeamMembers: 1,
    overageRate: 0,
  },
  start: {
    id: 'start',
    name: 'Start',
    description: 'For solopreneurs getting started',
    monthlyPrice: 39,
    annualPrice: 396,
    annualMonthlyPrice: 33,
    maxHelpers: 10,
    maxConnections: 2,
    maxExecutions: 5000,
    maxApiKeys: 2,
    maxTeamMembers: 2,
    overageRate: 0.01,
  },
  grow: {
    id: 'grow',
    name: 'Grow',
    description: 'For growing businesses',
    monthlyPrice: 59,
    annualPrice: 588,
    annualMonthlyPrice: 49,
    maxHelpers: 50,
    maxConnections: 5,
    maxExecutions: 25000,
    maxApiKeys: 10,
    maxTeamMembers: 5,
    overageRate: 0.008,
  },
  deliver: {
    id: 'deliver',
    name: 'Deliver',
    description: 'For teams that need it all',
    monthlyPrice: 79,
    annualPrice: 792,
    annualMonthlyPrice: 66,
    maxHelpers: 999999,
    maxConnections: 20,
    maxExecutions: 100000,
    maxApiKeys: 100,
    maxTeamMembers: 100,
    overageRate: 0.005,
  },
}

export const PAID_PLAN_IDS: PlanId[] = ['start', 'grow', 'deliver']

/** Feature lists for plan cards (displayed as bullet points). */
export function getPlanFeatures(
  planId: PlanId,
  billingPeriod: 'monthly' | 'annual' = 'monthly'
): string[] {
  const plan = PLAN_CONFIGS[planId]
  if (planId === 'free') {
    return [
      `${plan.maxHelpers} active helper`,
      `${plan.maxConnections} CRM connection`,
      `${formatLimit(plan.maxExecutions)} executions/mo`,
      `${plan.maxApiKeys} API key`,
    ]
  }
  const price =
    billingPeriod === 'annual'
      ? `$${plan.annualMonthlyPrice}/mo (billed yearly)`
      : `$${plan.monthlyPrice}/mo`
  return [
    `${formatLimit(plan.maxHelpers)} active helpers`,
    `${plan.maxConnections} CRM connections`,
    `${formatLimit(plan.maxExecutions)} executions/mo`,
    `${plan.maxApiKeys} API keys`,
    `${plan.maxTeamMembers} team members`,
    plan.overageRate > 0
      ? `$${plan.overageRate}/execution overage`
      : 'Hard limit on executions',
  ]
}

/** Comparison table rows for landing page. */
export const COMPARISON_ROWS = [
  {
    label: 'Active Helpers',
    start: '10',
    grow: '50',
    deliver: 'Unlimited',
  },
  {
    label: 'CRM Connections',
    start: '2',
    grow: '5',
    deliver: '20',
  },
  {
    label: 'Executions / mo',
    start: '5,000',
    grow: '25,000',
    deliver: '100,000',
  },
  {
    label: 'API Keys',
    start: '2',
    grow: '10',
    deliver: '100',
  },
  {
    label: 'Team Members',
    start: '2',
    grow: '5',
    deliver: '100',
  },
  {
    label: 'Support',
    start: 'Email',
    grow: 'Priority',
    deliver: 'Dedicated',
  },
]

/** Returns human-readable plan label. */
export function getPlanLabel(plan: string): string {
  return PLAN_CONFIGS[plan as PlanId]?.name ?? 'Sandbox'
}

/** Returns true if the limit value represents "unlimited". */
export function isUnlimited(value: number): boolean {
  return value >= 999999
}

/** Formats a numeric limit for display. */
export function formatLimit(value: number): string {
  if (isUnlimited(value)) return 'Unlimited'
  if (value >= 1000) return `${(value / 1000).toLocaleString('en-US')}k`
  return value.toLocaleString('en-US')
}

/** Returns the annual savings percentage. */
export function getAnnualSavingsPercent(planId: PlanId): number {
  const plan = PLAN_CONFIGS[planId]
  if (plan.monthlyPrice === 0) return 0
  const yearlyAtMonthly = plan.monthlyPrice * 12
  return Math.round(((yearlyAtMonthly - plan.annualPrice) / yearlyAtMonthly) * 100)
}
