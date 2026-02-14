import type { CatalogObjectType } from '../api/data-explorer'

// ---------------------------------------------------------------------------
// Interfaces
// ---------------------------------------------------------------------------

export interface Contact {
  id: string
  firstName: string
  lastName: string
  email: string
  phone: string
  company: string
  title: string
  tags: string[]
  status: 'active' | 'inactive' | 'unsubscribed'
  source: 'Web Form' | 'Import' | 'API' | 'Manual' | 'Referral'
  createdAt: string
  lastActivity: string
  score: number
  customFields: {
    leadSource: string
    industry: string
    annualRevenue: number
    preferredContact: string
  }
}

export interface Tag {
  id: string
  name: string
  category: 'status' | 'interest' | 'source' | 'lifecycle'
  contactCount: number
  createdAt: string
}

export interface CustomField {
  id: string
  name: string
  fieldKey: string
  type: 'text' | 'number' | 'date' | 'dropdown' | 'checkbox' | 'email' | 'phone'
  options?: string[]
  required: boolean
}

export interface Deal {
  id: string
  name: string
  stage: 'Discovery' | 'Proposal' | 'Negotiation' | 'Closed Won' | 'Closed Lost'
  value: number
  probability: number
  contactId: string
  contactName: string
  expectedCloseDate: string
  createdAt: string
  owner: string
}

// ---------------------------------------------------------------------------
// Deterministic pseudo-random number generator (mulberry32)
// ---------------------------------------------------------------------------

function createRng(seed: number) {
  let s = seed | 0
  return () => {
    s = (s + 0x6d2b79f5) | 0
    let t = Math.imul(s ^ (s >>> 15), 1 | s)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

function pick<T>(rng: () => number, arr: readonly T[]): T {
  return arr[Math.floor(rng() * arr.length)]
}

function pickN<T>(rng: () => number, arr: readonly T[], min: number, max: number): T[] {
  const count = min + Math.floor(rng() * (max - min + 1))
  const shuffled = [...arr].sort(() => rng() - 0.5)
  return shuffled.slice(0, count)
}

function randInt(rng: () => number, min: number, max: number): number {
  return min + Math.floor(rng() * (max - min + 1))
}

// ---------------------------------------------------------------------------
// Seed data pools
// ---------------------------------------------------------------------------

const FIRST_NAMES = [
  'James', 'Mary', 'Robert', 'Patricia', 'John', 'Jennifer', 'Michael', 'Linda',
  'David', 'Elizabeth', 'William', 'Barbara', 'Richard', 'Susan', 'Joseph', 'Jessica',
  'Thomas', 'Sarah', 'Christopher', 'Karen', 'Charles', 'Lisa', 'Daniel', 'Nancy',
  'Matthew', 'Betty', 'Anthony', 'Margaret', 'Mark', 'Sandra', 'Donald', 'Ashley',
  'Steven', 'Dorothy', 'Paul', 'Kimberly', 'Andrew', 'Emily', 'Joshua', 'Donna',
  'Kenneth', 'Michelle', 'Kevin', 'Carol', 'Brian', 'Amanda', 'George', 'Melissa',
  'Timothy', 'Deborah', 'Ronald', 'Stephanie', 'Edward', 'Rebecca', 'Jason', 'Sharon',
  'Jeffrey', 'Laura', 'Ryan', 'Cynthia', 'Jacob', 'Kathleen', 'Gary', 'Amy',
  'Nicholas', 'Angela', 'Eric', 'Shirley', 'Jonathan', 'Anna', 'Stephen', 'Brenda',
  'Larry', 'Pamela', 'Justin', 'Emma', 'Scott', 'Nicole', 'Brandon', 'Helen',
  'Benjamin', 'Samantha', 'Samuel', 'Katherine', 'Raymond', 'Christine', 'Gregory', 'Debra',
  'Frank', 'Rachel', 'Alexander', 'Carolyn', 'Patrick', 'Janet', 'Jack', 'Catherine',
  'Dennis', 'Maria', 'Jerry', 'Heather', 'Tyler', 'Diane', 'Aaron', 'Ruth',
  'Jose', 'Julie', 'Adam', 'Olivia', 'Nathan', 'Joyce', 'Henry', 'Virginia',
  'Peter', 'Victoria', 'Zachary', 'Kelly', 'Douglas', 'Lauren', 'Harold', 'Christina',
  'Carl', 'Joan', 'Arthur', 'Evelyn', 'Gerald', 'Judith', 'Roger', 'Megan',
  'Keith', 'Andrea', 'Jeremy', 'Cheryl', 'Terry', 'Hannah', 'Lawrence', 'Jacqueline',
  'Sean', 'Martha', 'Christian', 'Gloria', 'Albert', 'Teresa', 'Joe', 'Ann',
  'Ethan', 'Sara', 'Austin', 'Madison', 'Jesse', 'Frances', 'Willie', 'Kathryn',
  'Billy', 'Janice', 'Bryan', 'Jean', 'Bruce', 'Abigail', 'Jordan', 'Alice',
  'Ralph', 'Judy', 'Roy', 'Sophia', 'Noah', 'Grace', 'Dylan', 'Denise',
  'Eugene', 'Amber', 'Wayne', 'Doris', 'Alan', 'Marilyn', 'Juan', 'Danielle',
  'Louis', 'Beverly', 'Russell', 'Isabella', 'Vincent', 'Theresa', 'Philip', 'Diana',
  'Bobby', 'Natalie', 'Johnny', 'Brittany', 'Bradley', 'Charlotte', 'Liam', 'Marie',
  'Mason', 'Kayla', 'Logan', 'Alexis', 'Oscar', 'Lori', 'Connor', 'Jane',
] as const

const LAST_NAMES = [
  'Smith', 'Johnson', 'Williams', 'Brown', 'Jones', 'Garcia', 'Miller', 'Davis',
  'Rodriguez', 'Martinez', 'Hernandez', 'Lopez', 'Gonzalez', 'Wilson', 'Anderson',
  'Thomas', 'Taylor', 'Moore', 'Jackson', 'Martin', 'Lee', 'Perez', 'Thompson',
  'White', 'Harris', 'Sanchez', 'Clark', 'Ramirez', 'Lewis', 'Robinson', 'Walker',
  'Young', 'Allen', 'King', 'Wright', 'Scott', 'Torres', 'Nguyen', 'Hill',
  'Flores', 'Green', 'Adams', 'Nelson', 'Baker', 'Hall', 'Rivera', 'Campbell',
  'Mitchell', 'Carter', 'Roberts', 'Gomez', 'Phillips', 'Evans', 'Turner', 'Diaz',
  'Parker', 'Cruz', 'Edwards', 'Collins', 'Reyes', 'Stewart', 'Morris', 'Morales',
  'Murphy', 'Cook', 'Rogers', 'Gutierrez', 'Ortiz', 'Morgan', 'Cooper', 'Peterson',
  'Bailey', 'Reed', 'Kelly', 'Howard', 'Ramos', 'Kim', 'Cox', 'Ward',
  'Richardson', 'Watson', 'Brooks', 'Chavez', 'Wood', 'James', 'Bennett', 'Gray',
  'Mendoza', 'Ruiz', 'Hughes', 'Price', 'Alvarez', 'Castillo', 'Sanders', 'Patel',
  'Myers', 'Long', 'Ross', 'Foster', 'Jimenez', 'Powell', 'Jenkins', 'Perry',
  'Russell', 'Sullivan', 'Bell', 'Coleman', 'Butler', 'Henderson', 'Barnes', 'Gonzales',
  'Fisher', 'Vasquez', 'Simmons', 'Griffin', 'McDaniel', 'Stone', 'Ford', 'Chapman',
  'Warren', 'Bishop', 'Larson', 'Kelley', 'Hicks', 'Burke', 'Schneider', 'Hoffman',
  'Fleming', 'Bowman', 'Medina', 'Franklin', 'Garrett', 'Walsh', 'Pearson', 'Dunn',
  'Crawford', 'Weber', 'Harrison', 'Santos', 'Lawrence', 'Keller', 'Hunt', 'Dixon',
  'Pierce', 'Armstrong', 'Elliott', 'Quinn', 'Gallagher', 'Brennan', 'Schultz', 'Stein',
  'Olsen', 'Brady', 'Carr', 'Lane', 'Gordon', 'Wolfe', 'Barton', 'Conner',
] as const

const EMAIL_DOMAINS = [
  'gmail.com', 'yahoo.com', 'outlook.com', 'hotmail.com', 'icloud.com',
  'protonmail.com', 'aol.com', 'mail.com', 'fastmail.com', 'zoho.com',
] as const

const COMPANY_NAMES = [
  'Acme Corporation', 'Stellar Dynamics', 'BrightPath Solutions', 'NexGen Technologies',
  'Pinnacle Group', 'Vertex Software', 'Ironclad Industries', 'Atlas Ventures',
  'Quantum Leap Inc', 'Horizon Digital', 'SilverStone Capital', 'Emerald Health',
  'TrueNorth Analytics', 'Summit Strategies', 'Velocity Media', 'Precision Partners',
  'Catalyst Consulting', 'FusionWorks', 'Apex Global', 'Maverick Solutions',
  'BlueWave Marketing', 'Redwood Financial', 'CoreBridge Systems', 'TitanForge Labs',
  'Sapphire Retail', 'CloudPeak Software', 'RiverBend Advisors', 'SwiftShip Logistics',
  'Broadleaf Enterprises', 'Oakmont Properties', 'IronPeak Mining', 'GreenField Energy',
  'NovaTech Solutions', 'ClearSky Aviation', 'Coastal Commerce', 'PrimeStone Realty',
  'ThunderBolt Electric', 'EvergreenHR', 'DigitalFront Agency', 'SureStep Insurance',
  'Keystone Education', 'Lighthouse Labs', 'Trident Defense', 'Pacific Rim Trading',
  'Northwind Partners', 'Everest Capital', 'MarbleBridge Consulting', 'OptiFlow Systems',
  'GoldLeaf Financial', 'SunRise Healthcare', 'AnchorPoint Legal', 'Cascade Engineering',
  'Falcon Security', 'VantagePoint Media', 'BridgeSpan Group', 'StoneMill Manufacturing',
  'HarborView Hotels', 'BluePrint Architecture', 'SignalFire Ventures', 'CrestLine Pharma',
] as const

const JOB_TITLES = [
  'CEO', 'CTO', 'CFO', 'COO', 'CMO', 'VP of Sales', 'VP of Marketing',
  'VP of Engineering', 'VP of Operations', 'Director of Sales', 'Director of Marketing',
  'Marketing Director', 'Sales Director', 'Engineering Manager', 'Product Manager',
  'Project Manager', 'Account Executive', 'Business Development Manager',
  'Customer Success Manager', 'Operations Manager', 'HR Director', 'IT Director',
  'Creative Director', 'Brand Manager', 'Content Strategist', 'Digital Marketing Manager',
  'Sales Manager', 'Regional Manager', 'General Manager', 'Managing Director',
  'Partner', 'Founder', 'Co-Founder', 'Owner', 'President',
  'Head of Growth', 'Head of Product', 'Head of Partnerships', 'Head of Design',
  'Senior Consultant', 'Principal Consultant', 'Strategy Consultant',
  'Software Engineer', 'Senior Developer', 'Data Analyst', 'Financial Analyst',
  'Marketing Coordinator', 'Sales Representative', 'Office Manager', 'Executive Assistant',
] as const

const TAG_NAMES_POOL = [
  'VIP', 'Newsletter', 'Customer', 'Lead', 'Prospect', 'Hot Lead',
  'Warm Lead', 'Cold Lead', 'Qualified', 'Event Attendee', 'Webinar Attendee',
  'Free Trial', 'Premium', 'Enterprise', 'Partner', 'Referral Source',
  'Do Not Contact', 'Opted In', 'Blog Subscriber', 'Social Media',
  'Influencer', 'Decision Maker', 'Champion', 'Budget Holder',
  'Churned', 'At Risk', 'Renewal Pending', 'Upsell Candidate',
  'Onboarding', 'Active User',
] as const

const LEAD_SOURCES = [
  'Google Ads', 'Facebook Ads', 'LinkedIn', 'Organic Search', 'Direct Traffic',
  'Email Campaign', 'Webinar', 'Trade Show', 'Partner Referral', 'Cold Outreach',
  'Word of Mouth', 'Content Marketing', 'YouTube', 'Podcast', 'Affiliate',
] as const

const INDUSTRIES = [
  'Technology', 'Healthcare', 'Financial Services', 'Real Estate', 'Education',
  'Manufacturing', 'Retail', 'Consulting', 'Marketing & Advertising', 'Legal',
  'Hospitality', 'Construction', 'Nonprofit', 'Energy', 'Transportation',
  'Media & Entertainment', 'Insurance', 'Telecommunications', 'Agriculture', 'Aerospace',
] as const

const PREFERRED_CONTACTS = ['Email', 'Phone', 'SMS', 'LinkedIn', 'In Person'] as const

const STATUSES: Contact['status'][] = ['active', 'inactive', 'unsubscribed']
const STATUS_WEIGHTS = [0.7, 0.2, 0.1]

const SOURCES: Contact['source'][] = ['Web Form', 'Import', 'API', 'Manual', 'Referral']
const SOURCE_WEIGHTS = [0.35, 0.25, 0.15, 0.1, 0.15]

const DEAL_STAGES: Deal['stage'][] = ['Discovery', 'Proposal', 'Negotiation', 'Closed Won', 'Closed Lost']
const DEAL_STAGE_WEIGHTS = [0.25, 0.25, 0.2, 0.2, 0.1]

const SALES_REPS = [
  'Alex Thompson', 'Sarah Chen', 'Marcus Williams', 'Jessica Rivera',
  'David Park', 'Emily Watson', 'Chris Morales', 'Amanda Foster',
  'Ryan O\'Brien', 'Natalie Kim',
] as const

// ---------------------------------------------------------------------------
// Helper: weighted pick
// ---------------------------------------------------------------------------

function weightedPick<T>(rng: () => number, items: T[], weights: number[]): T {
  const r = rng()
  let cumulative = 0
  for (let i = 0; i < items.length; i++) {
    cumulative += weights[i]
    if (r < cumulative) return items[i]
  }
  return items[items.length - 1]
}

// ---------------------------------------------------------------------------
// Helper: deterministic date generation
// ---------------------------------------------------------------------------

const BASE_DATE = new Date('2024-01-15T00:00:00Z').getTime()
const TWO_YEARS_MS = 2 * 365 * 24 * 60 * 60 * 1000
const SIX_MONTHS_MS = 6 * 30 * 24 * 60 * 60 * 1000

function generateCreatedAt(rng: () => number): string {
  const offset = Math.floor(rng() * TWO_YEARS_MS)
  return new Date(BASE_DATE - TWO_YEARS_MS + offset).toISOString()
}

function generateLastActivity(rng: () => number, createdAt: string): string {
  const createdMs = new Date(createdAt).getTime()
  const now = BASE_DATE
  const recentWindow = Math.min(now - createdMs, SIX_MONTHS_MS)
  const offset = Math.floor(rng() * recentWindow)
  return new Date(now - offset).toISOString()
}

function generateFutureDate(rng: () => number): string {
  const daysAhead = 14 + Math.floor(rng() * 180)
  return new Date(BASE_DATE + daysAhead * 24 * 60 * 60 * 1000).toISOString()
}

// ---------------------------------------------------------------------------
// Helper: US phone number
// ---------------------------------------------------------------------------

function generatePhone(rng: () => number): string {
  const area = 200 + Math.floor(rng() * 800)
  const prefix = 200 + Math.floor(rng() * 800)
  const line = 1000 + Math.floor(rng() * 9000)
  return `(${area}) ${prefix}-${line}`
}

// ---------------------------------------------------------------------------
// Contacts generator
// ---------------------------------------------------------------------------

let _contacts: Contact[] | null = null

export function getMockContacts(): Contact[] {
  if (_contacts) return _contacts

  const rng = createRng(42)
  const contacts: Contact[] = []

  for (let i = 1; i <= 200; i++) {
    const firstName = pick(rng, FIRST_NAMES)
    const lastName = pick(rng, LAST_NAMES)
    const emailPrefix = `${firstName.toLowerCase()}.${lastName.toLowerCase()}`
    // Append index to ensure uniqueness
    const domain = pick(rng, EMAIL_DOMAINS)
    const email = i <= 60
      ? `${emailPrefix}@${domain}`
      : `${emailPrefix}${i}@${domain}`

    const createdAt = generateCreatedAt(rng)
    const lastActivity = generateLastActivity(rng, createdAt)
    const company = pick(rng, COMPANY_NAMES)
    const status = weightedPick(rng, STATUSES, STATUS_WEIGHTS)

    const contactTags = pickN(rng, TAG_NAMES_POOL, 1, 5)

    const annualRevenueBase = [
      50000, 100000, 250000, 500000, 750000,
      1000000, 2500000, 5000000, 10000000, 25000000,
    ]

    contacts.push({
      id: String(i),
      firstName,
      lastName,
      email,
      phone: generatePhone(rng),
      company,
      title: pick(rng, JOB_TITLES),
      tags: contactTags,
      status,
      source: weightedPick(rng, SOURCES, SOURCE_WEIGHTS),
      createdAt,
      lastActivity,
      score: randInt(rng, 1, 100),
      customFields: {
        leadSource: pick(rng, LEAD_SOURCES),
        industry: pick(rng, INDUSTRIES),
        annualRevenue: pick(rng, annualRevenueBase),
        preferredContact: pick(rng, PREFERRED_CONTACTS),
      },
    })
  }

  _contacts = contacts
  return contacts
}

// ---------------------------------------------------------------------------
// Tags generator
// ---------------------------------------------------------------------------

const TAG_DEFINITIONS: { name: string; category: Tag['category'] }[] = [
  { name: 'VIP Client', category: 'status' },
  { name: 'Newsletter Subscriber', category: 'interest' },
  { name: 'Hot Lead', category: 'lifecycle' },
  { name: 'Warm Lead', category: 'lifecycle' },
  { name: 'Cold Lead', category: 'lifecycle' },
  { name: 'Customer', category: 'lifecycle' },
  { name: 'Prospect', category: 'lifecycle' },
  { name: 'Event Attendee', category: 'interest' },
  { name: 'Webinar Registrant', category: 'interest' },
  { name: 'Free Trial User', category: 'lifecycle' },
  { name: 'Premium Subscriber', category: 'status' },
  { name: 'Enterprise Client', category: 'status' },
  { name: 'Partner', category: 'status' },
  { name: 'Referral Source', category: 'source' },
  { name: 'Churned', category: 'lifecycle' },
  { name: 'At Risk', category: 'lifecycle' },
  { name: 'Web Form Lead', category: 'source' },
  { name: 'Imported Contact', category: 'source' },
  { name: 'API Synced', category: 'source' },
  { name: 'Manual Entry', category: 'source' },
  { name: 'Facebook Lead', category: 'source' },
  { name: 'Google Ads Lead', category: 'source' },
  { name: 'LinkedIn Contact', category: 'source' },
  { name: 'Opted In - Email', category: 'interest' },
  { name: 'Opted In - SMS', category: 'interest' },
  { name: 'Do Not Contact', category: 'status' },
  { name: 'Upsell Candidate', category: 'lifecycle' },
  { name: 'Renewal Pending', category: 'lifecycle' },
  { name: 'Onboarding', category: 'lifecycle' },
  { name: 'Decision Maker', category: 'status' },
]

let _tags: Tag[] | null = null

export function getMockTags(): Tag[] {
  if (_tags) return _tags

  const rng = createRng(1337)

  const tags: Tag[] = TAG_DEFINITIONS.map((def, i) => {
    const createdAt = generateCreatedAt(rng)
    // Spread contact counts: status/lifecycle tags get more, source tags get moderate
    let countMin = 5
    let countMax = 500
    if (def.category === 'lifecycle') { countMin = 20; countMax = 2000 }
    if (def.category === 'status') { countMin = 10; countMax = 5000 }
    if (def.category === 'source') { countMin = 15; countMax = 1200 }
    if (def.category === 'interest') { countMin = 50; countMax = 3000 }

    return {
      id: String(i + 1),
      name: def.name,
      category: def.category,
      contactCount: randInt(rng, countMin, countMax),
      createdAt,
    }
  })

  _tags = tags
  return tags
}

// ---------------------------------------------------------------------------
// Custom Fields generator
// ---------------------------------------------------------------------------

const CUSTOM_FIELD_DEFINITIONS: Omit<CustomField, 'id'>[] = [
  { name: 'Lead Source', fieldKey: 'lead_source', type: 'dropdown', options: [...LEAD_SOURCES], required: false },
  { name: 'Industry', fieldKey: 'industry', type: 'dropdown', options: [...INDUSTRIES], required: false },
  { name: 'Annual Revenue', fieldKey: 'annual_revenue', type: 'number', required: false },
  { name: 'Preferred Contact Method', fieldKey: 'preferred_contact', type: 'dropdown', options: ['Email', 'Phone', 'SMS', 'LinkedIn', 'In Person'], required: false },
  { name: 'Company Size', fieldKey: 'company_size', type: 'dropdown', options: ['1-10', '11-50', '51-200', '201-500', '501-1000', '1000+'], required: false },
  { name: 'Website', fieldKey: 'website', type: 'text', required: false },
  { name: 'LinkedIn Profile', fieldKey: 'linkedin_profile', type: 'text', required: false },
  { name: 'Date of Birth', fieldKey: 'date_of_birth', type: 'date', required: false },
  { name: 'Subscription Start Date', fieldKey: 'subscription_start_date', type: 'date', required: false },
  { name: 'NPS Score', fieldKey: 'nps_score', type: 'number', required: false },
  { name: 'Opted In Email', fieldKey: 'opted_in_email', type: 'checkbox', required: false },
  { name: 'Opted In SMS', fieldKey: 'opted_in_sms', type: 'checkbox', required: false },
  { name: 'Secondary Email', fieldKey: 'secondary_email', type: 'email', required: false },
  { name: 'Mobile Phone', fieldKey: 'mobile_phone', type: 'phone', required: false },
  { name: 'Notes', fieldKey: 'notes', type: 'text', required: false },
]

let _customFields: CustomField[] | null = null

export function getMockCustomFields(): CustomField[] {
  if (_customFields) return _customFields

  _customFields = CUSTOM_FIELD_DEFINITIONS.map((def, i) => ({
    id: String(i + 1),
    ...def,
  }))

  return _customFields
}

// ---------------------------------------------------------------------------
// Deals generator
// ---------------------------------------------------------------------------

const DEAL_NAME_TEMPLATES = [
  '{company} - Enterprise Plan',
  '{company} - Professional Upgrade',
  '{company} - Annual Contract',
  '{company} - Starter Package',
  '{company} - Custom Integration',
  '{company} - Consulting Engagement',
  '{company} - Platform Migration',
  '{company} - Marketing Suite',
  '{company} - Data Analytics Package',
  '{company} - Support Plan Renewal',
  '{company} - Expansion Deal',
  '{company} - Multi-Year Agreement',
  '{company} - Pilot Program',
  '{company} - Team License',
  '{company} - White Label Partnership',
  '{company} - API Access License',
  '{company} - Training Package',
  '{company} - Managed Services',
  '{company} - Security Audit',
  '{company} - Cloud Migration',
] as const

let _deals: Deal[] | null = null

export function getMockDeals(): Deal[] {
  if (_deals) return _deals

  const rng = createRng(2024)
  const contacts = getMockContacts()
  const deals: Deal[] = []

  for (let i = 0; i < 20; i++) {
    const contact = contacts[Math.floor(rng() * 80)] // pick from first 80 contacts
    const stage = weightedPick(rng, DEAL_STAGES, DEAL_STAGE_WEIGHTS)
    const template = DEAL_NAME_TEMPLATES[i]
    const dealName = template.replace('{company}', contact.company)

    // Value ranges by stage
    let valueMin = 1000
    let valueMax = 50000
    if (stage === 'Proposal' || stage === 'Negotiation') { valueMin = 5000; valueMax = 200000 }
    if (stage === 'Closed Won') { valueMin = 10000; valueMax = 500000 }
    if (stage === 'Closed Lost') { valueMin = 2000; valueMax = 150000 }

    const value = Math.round(randInt(rng, valueMin, valueMax) / 100) * 100

    // Probability aligns with stage
    let probMin = 0
    let probMax = 100
    if (stage === 'Discovery') { probMin = 10; probMax = 30 }
    if (stage === 'Proposal') { probMin = 30; probMax = 60 }
    if (stage === 'Negotiation') { probMin = 50; probMax = 80 }
    if (stage === 'Closed Won') { probMin = 100; probMax = 100 }
    if (stage === 'Closed Lost') { probMin = 0; probMax = 0 }

    const createdAt = generateCreatedAt(rng)

    deals.push({
      id: String(i + 1),
      name: dealName,
      stage,
      value,
      probability: randInt(rng, probMin, probMax),
      contactId: contact.id,
      contactName: `${contact.firstName} ${contact.lastName}`,
      expectedCloseDate: generateFutureDate(rng),
      createdAt,
      owner: pick(rng, SALES_REPS),
    })
  }

  _deals = deals
  return deals
}

// ---------------------------------------------------------------------------
// Catalog generator (object types across all 5 CRM platforms)
// ---------------------------------------------------------------------------

interface PlatformCatalogConfig {
  platformId: string
  platformName: string
  connections: {
    connectionId: string
    connectionName: string
    objects: {
      objectType: string
      label: string
      icon: string
      recordCount: number
    }[]
  }[]
}

const PLATFORM_CATALOGS: PlatformCatalogConfig[] = [
  {
    platformId: 'keap',
    platformName: 'Keap',
    connections: [
      {
        connectionId: 'keap-prod-001',
        connectionName: 'Keap - Production',
        objects: [
          { objectType: 'contacts', label: 'Contacts', icon: 'users', recordCount: 12847 },
          { objectType: 'tags', label: 'Tags', icon: 'tag', recordCount: 156 },
          { objectType: 'custom_fields', label: 'Custom Fields', icon: 'sliders', recordCount: 42 },
          { objectType: 'opportunities', label: 'Opportunities', icon: 'target', recordCount: 389 },
          { objectType: 'orders', label: 'Orders', icon: 'shopping-cart', recordCount: 2341 },
          { objectType: 'products', label: 'Products', icon: 'package', recordCount: 67 },
          { objectType: 'campaigns', label: 'Campaigns', icon: 'megaphone', recordCount: 34 },
          { objectType: 'emails', label: 'Email Templates', icon: 'mail', recordCount: 128 },
        ],
      },
      {
        connectionId: 'keap-sandbox-002',
        connectionName: 'Keap - Sandbox',
        objects: [
          { objectType: 'contacts', label: 'Contacts', icon: 'users', recordCount: 450 },
          { objectType: 'tags', label: 'Tags', icon: 'tag', recordCount: 28 },
          { objectType: 'custom_fields', label: 'Custom Fields', icon: 'sliders', recordCount: 15 },
          { objectType: 'opportunities', label: 'Opportunities', icon: 'target', recordCount: 12 },
        ],
      },
    ],
  },
  {
    platformId: 'gohighlevel',
    platformName: 'GoHighLevel',
    connections: [
      {
        connectionId: 'ghl-agency-001',
        connectionName: 'GHL - Agency Account',
        objects: [
          { objectType: 'contacts', label: 'Contacts', icon: 'users', recordCount: 8934 },
          { objectType: 'tags', label: 'Tags', icon: 'tag', recordCount: 89 },
          { objectType: 'custom_fields', label: 'Custom Fields', icon: 'sliders', recordCount: 37 },
          { objectType: 'opportunities', label: 'Opportunities', icon: 'target', recordCount: 267 },
          { objectType: 'pipelines', label: 'Pipelines', icon: 'git-branch', recordCount: 5 },
          { objectType: 'calendars', label: 'Calendars', icon: 'calendar', recordCount: 12 },
          { objectType: 'workflows', label: 'Workflows', icon: 'workflow', recordCount: 45 },
          { objectType: 'forms', label: 'Forms', icon: 'file-text', recordCount: 23 },
          { objectType: 'conversations', label: 'Conversations', icon: 'message-circle', recordCount: 4521 },
        ],
      },
    ],
  },
  {
    platformId: 'activecampaign',
    platformName: 'ActiveCampaign',
    connections: [
      {
        connectionId: 'ac-main-001',
        connectionName: 'ActiveCampaign - Main',
        objects: [
          { objectType: 'contacts', label: 'Contacts', icon: 'users', recordCount: 23456 },
          { objectType: 'tags', label: 'Tags', icon: 'tag', recordCount: 234 },
          { objectType: 'custom_fields', label: 'Custom Fields', icon: 'sliders', recordCount: 56 },
          { objectType: 'deals', label: 'Deals', icon: 'handshake', recordCount: 892 },
          { objectType: 'lists', label: 'Lists', icon: 'list', recordCount: 18 },
          { objectType: 'automations', label: 'Automations', icon: 'zap', recordCount: 67 },
          { objectType: 'campaigns', label: 'Campaigns', icon: 'megaphone', recordCount: 145 },
          { objectType: 'forms', label: 'Forms', icon: 'file-text', recordCount: 31 },
        ],
      },
    ],
  },
  {
    platformId: 'ontraport',
    platformName: 'Ontraport',
    connections: [
      {
        connectionId: 'op-prod-001',
        connectionName: 'Ontraport - Production',
        objects: [
          { objectType: 'contacts', label: 'Contacts', icon: 'users', recordCount: 6782 },
          { objectType: 'tags', label: 'Tags', icon: 'tag', recordCount: 112 },
          { objectType: 'custom_fields', label: 'Custom Fields', icon: 'sliders', recordCount: 29 },
          { objectType: 'deals', label: 'Deals', icon: 'handshake', recordCount: 178 },
          { objectType: 'sequences', label: 'Sequences', icon: 'layers', recordCount: 24 },
          { objectType: 'tasks', label: 'Tasks', icon: 'check-square', recordCount: 567 },
          { objectType: 'landing_pages', label: 'Landing Pages', icon: 'layout', recordCount: 19 },
        ],
      },
    ],
  },
  {
    platformId: 'hubspot',
    platformName: 'HubSpot',
    connections: [
      {
        connectionId: 'hs-prod-001',
        connectionName: 'HubSpot - Production',
        objects: [
          { objectType: 'contacts', label: 'Contacts', icon: 'users', recordCount: 45123 },
          { objectType: 'companies', label: 'Companies', icon: 'building', recordCount: 8934 },
          { objectType: 'deals', label: 'Deals', icon: 'handshake', recordCount: 2345 },
          { objectType: 'tickets', label: 'Tickets', icon: 'ticket', recordCount: 1234 },
          { objectType: 'custom_properties', label: 'Custom Properties', icon: 'sliders', recordCount: 78 },
          { objectType: 'lists', label: 'Lists', icon: 'list', recordCount: 67 },
          { objectType: 'workflows', label: 'Workflows', icon: 'workflow', recordCount: 89 },
          { objectType: 'email_templates', label: 'Email Templates', icon: 'mail', recordCount: 156 },
          { objectType: 'forms', label: 'Forms', icon: 'file-text', recordCount: 43 },
          { objectType: 'pipelines', label: 'Pipelines', icon: 'git-branch', recordCount: 4 },
        ],
      },
      {
        connectionId: 'hs-dev-002',
        connectionName: 'HubSpot - Dev Portal',
        objects: [
          { objectType: 'contacts', label: 'Contacts', icon: 'users', recordCount: 1200 },
          { objectType: 'companies', label: 'Companies', icon: 'building', recordCount: 340 },
          { objectType: 'deals', label: 'Deals', icon: 'handshake', recordCount: 89 },
          { objectType: 'custom_properties', label: 'Custom Properties', icon: 'sliders', recordCount: 22 },
        ],
      },
    ],
  },
]

let _catalog: CatalogObjectType[] | null = null

export function getMockCatalog(): CatalogObjectType[] {
  if (_catalog) return _catalog

  const sources: CatalogObjectType[] = []

  for (const platform of PLATFORM_CATALOGS) {
    for (const connection of platform.connections) {
      for (const obj of connection.objects) {
        sources.push({
          objectType: obj.objectType,
          label: obj.label,
          icon: obj.icon,
          recordCount: obj.recordCount,
          connectionId: connection.connectionId,
          connectionName: connection.connectionName,
          platformId: platform.platformId,
          platformName: platform.platformName,
        })
      }
    }
  }

  _catalog = sources
  return sources
}
