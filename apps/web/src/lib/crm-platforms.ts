export interface CRMPlatform {
  id: string
  name: string
  color: string
  accent: string
  initial: string
  logo: string | null
  authType: 'oauth2' | 'api_key'
  apiBaseUrl: string
  capabilities: string[]
}

const HIDDEN_PLATFORMS: string[] = []

export const crmPlatforms: CRMPlatform[] = [
  {
    id: 'keap',
    name: 'Keap',
    color: '#2AB27B',
    accent: '#e6f7ef',
    initial: 'K',
    logo: '/images/platforms/keap.png',
    authType: 'oauth2',
    apiBaseUrl: 'https://api.infusionsoft.com',
    capabilities: ['Contacts', 'Tags', 'Custom Fields', 'Automations', 'Goals', 'Deals'],
  },
  {
    id: 'gohighlevel',
    name: 'GoHighLevel',
    color: '#4285F4',
    accent: '#e8f0fe',
    initial: 'G',
    logo: '/images/platforms/gohighlevel.png',
    authType: 'api_key',
    apiBaseUrl: 'https://services.leadconnectorhq.com',
    capabilities: ['Contacts', 'Tags', 'Custom Fields', 'Workflows', 'Pipelines', 'SMS'],
  },
  {
    id: 'activecampaign',
    name: 'ActiveCampaign',
    color: '#356AE6',
    accent: '#ebeffe',
    initial: 'A',
    logo: '/images/platforms/activecampaign.png',
    authType: 'api_key',
    apiBaseUrl: '',
    capabilities: ['Contacts', 'Tags', 'Custom Fields', 'Automations', 'Deals', 'Campaigns'],
  },
  {
    id: 'ontraport',
    name: 'Ontraport',
    color: '#6C4DC4',
    accent: '#f0ebfa',
    initial: 'O',
    logo: '/images/platforms/ontraport.png',
    authType: 'api_key',
    apiBaseUrl: 'https://api.ontraport.com',
    capabilities: ['Contacts', 'Tags', 'Custom Fields', 'Sequences', 'Tasks'],
  },
  {
    id: 'hubspot',
    name: 'HubSpot',
    color: '#FF7A59',
    accent: '#fff0ec',
    initial: 'H',
    logo: '/images/platforms/hubspot.png',
    authType: 'api_key',
    apiBaseUrl: 'https://api.hubapi.com',
    capabilities: ['Contacts', 'Deals', 'Lists', 'Workflows', 'Custom Properties', 'Pipelines'],
  },
  {
    id: 'stripe',
    name: 'Stripe',
    color: '#635BFF',
    accent: '#eeecff',
    initial: 'S',
    logo: '/images/platforms/stripe.png',
    authType: 'api_key',
    apiBaseUrl: 'https://api.stripe.com',
    capabilities: ['Customers', 'Charges', 'Subscriptions', 'Invoices', 'Products', 'Payments'],
  },
]

export const activeCRMPlatforms = crmPlatforms.filter((p) => !HIDDEN_PLATFORMS.includes(p.id))

export function getCRMPlatform(id: string): CRMPlatform | undefined {
  return crmPlatforms.find((p) => p.id === id)
}
