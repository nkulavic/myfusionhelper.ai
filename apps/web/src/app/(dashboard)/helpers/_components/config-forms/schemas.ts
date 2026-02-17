import { z } from 'zod'

// ─── Shared field helpers ───────────────────────────────
const optionalString = z.string().optional().default('')
const requiredString = z.string().min(1, 'Required')
const optionalField = z.string().optional().default('')
const requiredField = z.string().min(1, 'Select a field')

// ─── CONTACT HELPERS ────────────────────────────────────

export const copyItSchema = z.object({
  sourceField: requiredField,
  targetField: requiredField,
  overwrite: z.boolean().optional().default(true),
})
export type CopyItConfig = z.infer<typeof copyItSchema>

export const clearItSchema = z.object({
  fields: z.array(z.string()).min(1, 'Select at least one field'),
})
export type ClearItConfig = z.infer<typeof clearItSchema>

export const tagItSchema = z.object({
  action: z.enum(['apply', 'remove']).default('apply'),
  tagIds: z.array(z.string()).min(1, 'Select at least one tag'),
})
export type TagItConfig = z.infer<typeof tagItSchema>

export const foundItSchema = z.object({
  checkField: requiredField,
  foundTagId: optionalString,
  notFoundTagId: optionalString,
  foundGoal: optionalString,
  notFoundGoal: optionalString,
})
export type FoundItConfig = z.infer<typeof foundItSchema>

export const mergeItSchema = z.object({
  sourceFields: z.array(z.string()).min(2, 'Select at least two fields'),
  targetField: requiredField,
  separator: z.string().optional().default(' '),
  skipEmpty: z.boolean().optional().default(true),
})
export type MergeItConfig = z.infer<typeof mergeItSchema>

export const combineItSchema = z.object({
  sourceFields: z.array(z.string()).min(2, 'Select at least two fields'),
  targetField: requiredField,
  separator: z.string().optional().default(' '),
  skipEmpty: z.boolean().optional().default(true),
})
export type CombineItConfig = z.infer<typeof combineItSchema>

export const assignItSchema = z.object({
  ownerId: requiredString,
  ownerField: z.string().optional().default('owner_id'),
})
export type AssignItConfig = z.infer<typeof assignItSchema>

export const ownItSchema = z.object({
  ownerId: requiredString,
})
export type OwnItConfig = z.infer<typeof ownItSchema>

export const nameParseItSchema = z.object({
  sourceField: requiredField,
  firstNameField: z.string().optional().default('first_name'),
  lastNameField: z.string().optional().default('last_name'),
  suffixField: optionalField,
})
export type NameParseItConfig = z.infer<typeof nameParseItSchema>

export const noteItSchema = z.object({
  subject: requiredString,
  body: requiredString,
  noteType: z.string().optional().default('general'),
})
export type NoteItConfig = z.infer<typeof noteItSchema>

export const defaultToFieldSchema = z.object({
  default: requiredString,
  toField: requiredField,
})
export type DefaultToFieldConfig = z.infer<typeof defaultToFieldSchema>

export const companyLinkSchema = z.object({
  companyField: requiredField,
})
export type CompanyLinkConfig = z.infer<typeof companyLinkSchema>

export const optInSchema = z.object({
  emailField: z.string().optional().default('email'),
  reason: optionalString,
})
export type OptInConfig = z.infer<typeof optInSchema>

export const optOutSchema = z.object({
  emailField: z.string().optional().default('email'),
  reason: optionalString,
})
export type OptOutConfig = z.infer<typeof optOutSchema>

export const snapshotItSchema = z.object({
  includeTags: z.boolean().optional().default(true),
  includeCustomFields: z.boolean().optional().default(true),
})
export type SnapshotItConfig = z.infer<typeof snapshotItSchema>

export const fieldToFieldSchema = z.object({
  mappings: z
    .array(
      z.object({
        source: z.string(),
        target: z.string(),
      })
    )
    .min(1, 'Add at least one mapping'),
  overwrite: z.boolean().optional().default(true),
})
export type FieldToFieldConfig = z.infer<typeof fieldToFieldSchema>

export const moveItSchema = z.object({
  sourceField: requiredField,
  targetField: requiredField,
  preserve: z.boolean().optional().default(false),
})
export type MoveItConfig = z.infer<typeof moveItSchema>

export const contactUpdaterSchema = z.object({
  fields: z.record(z.string(), z.any()).refine((val) => Object.keys(val).length > 0, {
    message: 'At least one field mapping is required',
  }),
  secondaryContactIds: z.array(z.string()).optional().default([]),
})
export type ContactUpdaterConfig = z.infer<typeof contactUpdaterSchema>

// ─── DATA HELPERS ───────────────────────────────────────

export const formatItSchema = z.object({
  field: requiredField,
  format: z
    .enum([
      'uppercase',
      'lowercase',
      'title_case',
      'trim',
      'trim_uppercase',
      'trim_lowercase',
      'trim_title_case',
    ])
    .default('title_case'),
  targetField: optionalField,
})
export type FormatItConfig = z.infer<typeof formatItSchema>

export const mathItSchema = z.object({
  field: requiredField,
  operation: z
    .enum(['add', 'subtract', 'multiply', 'divide', 'round', 'ceil', 'floor', 'abs', 'percent'])
    .default('add'),
  operand: z.number().optional(),
  targetField: optionalField,
  decimalPlaces: z.number().min(0).max(10).optional().default(2),
})
export type MathItConfig = z.infer<typeof mathItSchema>

export const advanceMathSchema = z.object({
  sourceField: requiredField,
  operation: z
    .enum(['power', 'sqrt', 'abs', 'round', 'ceil', 'floor', 'min', 'max'])
    .default('sqrt'),
  operand: z.number().optional(),
  secondField: optionalField,
  targetField: requiredField,
})
export type AdvanceMathConfig = z.infer<typeof advanceMathSchema>

export const dateCalcSchema = z.object({
  field: requiredField,
  operation: z
    .enum([
      'add_days',
      'subtract_days',
      'add_months',
      'subtract_months',
      'add_years',
      'subtract_years',
      'set_now',
      'diff_days',
      'format',
    ])
    .default('add_days'),
  amount: z.number().optional().default(0),
  targetField: optionalField,
  compareField: optionalField,
  outputFormat: z.string().optional().default('2006-01-02'),
})
export type DateCalcConfig = z.infer<typeof dateCalcSchema>

export const textItSchema = z.object({
  field: requiredField,
  operation: z
    .enum([
      'prepend',
      'append',
      'replace',
      'remove',
      'truncate',
      'extract_email_domain',
      'extract_numbers',
      'slug',
      'reverse',
    ])
    .default('replace'),
  value: optionalString,
  replaceWith: optionalString,
  maxLength: z.number().optional(),
  targetField: optionalField,
})
export type TextItConfig = z.infer<typeof textItSchema>

export const splitItSchema = z.object({
  mode: z.enum(['tag', 'goal']).default('tag'),
  optionA: requiredString,
  optionB: requiredString,
  stateField: requiredField,
})
export type SplitItConfig = z.infer<typeof splitItSchema>

export const whenIsItSchema = z.object({
  sourceField: requiredField,
  targetField: requiredField,
  fromTimezone: z.string().optional().default('UTC'),
  toTimezone: requiredString,
  outputFormat: requiredString,
})
export type WhenIsItConfig = z.infer<typeof whenIsItSchema>

export const wordCountItSchema = z.object({
  sourceField: requiredField,
  targetField: requiredField,
  countType: z.enum(['words', 'characters']).default('words'),
})
export type WordCountItConfig = z.infer<typeof wordCountItSchema>

export const passwordItSchema = z.object({
  targetField: requiredField,
  length: z.number().min(4).max(128).optional().default(12),
  includeSpecial: z.boolean().optional().default(true),
  overwrite: z.boolean().optional().default(false),
})
export type PasswordItConfig = z.infer<typeof passwordItSchema>

export const phoneLookupSchema = z.object({
  phoneField: requiredField,
  countryCode: z.string().optional().default('US'),
  validGoal: optionalString,
  invalidGoal: optionalString,
  emptyGoal: optionalString,
  saveFormattedTo: optionalField,
})
export type PhoneLookupConfig = z.infer<typeof phoneLookupSchema>

export const ipLocationSchema = z.object({
  ipField: requiredField,
  cityField: optionalField,
  stateField: optionalField,
  countryField: optionalField,
  zipField: optionalField,
})
export type IpLocationConfig = z.infer<typeof ipLocationSchema>

export const ipNotificationsSchema = z.object({
  ip_address: requiredField,
  match_countries: z.array(z.string()).optional().default([]),
  match_regions: z.array(z.string()).optional().default([]),
  apply_tag: optionalString,
  save_location_to: optionalField,
})
export type IpNotificationsConfig = z.infer<typeof ipNotificationsSchema>

export const ipRedirectsSchema = z.object({
  ip_address: requiredField,
  country_urls: z.record(z.string(), z.string()).optional().default({}),
  default_url: optionalString,
  save_redirect_to: optionalField,
})
export type IpRedirectsConfig = z.infer<typeof ipRedirectsSchema>

export const getRecordSchema = z.object({
  type: z
    .enum(['invoice', 'job', 'subscription', 'lead', 'creditcard', 'payment'])
    .default('invoice'),
  fromField: requiredField,
  toField: requiredField,
})
export type GetRecordConfig = z.infer<typeof getRecordSchema>

export const lastClickSchema = z.object({
  emailField: z.string().optional().default('Email'),
  saveTo: requiredField,
})
export type LastClickConfig = z.infer<typeof lastClickSchema>

export const lastOpenSchema = z.object({
  emailField: z.string().optional().default('Email'),
  saveTo: requiredField,
})
export type LastOpenConfig = z.infer<typeof lastOpenSchema>

export const lastSendSchema = z.object({
  emailField: z.string().optional().default('Email'),
  saveTo: requiredField,
})
export type LastSendConfig = z.infer<typeof lastSendSchema>

// ─── TAGGING HELPERS ────────────────────────────────────

export const scoreItSchema = z.object({
  rules: z
    .array(
      z.object({
        tagId: z.string(),
        hasTag: z.boolean().optional().default(true),
        points: z.number().optional().default(1),
      })
    )
    .min(1, 'Add at least one rule'),
  targetField: requiredField,
})
export type ScoreItConfig = z.infer<typeof scoreItSchema>

export const groupItSchema = z.object({
  field: requiredField,
  tagPrefix: requiredString,
})
export type GroupItConfig = z.infer<typeof groupItSchema>

export const countTagsSchema = z.object({
  targetField: requiredField,
  category: optionalString,
})
export type CountTagsConfig = z.infer<typeof countTagsSchema>

export const countItTagsSchema = z.object({
  tagId: requiredString,
  threshold: z.number().optional(),
  thresholdMetTag: optionalString,
  thresholdNotMetTag: optionalString,
  saveCountTo: optionalField,
  savePositionTo: optionalField,
})
export type CountItTagsConfig = z.infer<typeof countItTagsSchema>

export const clearTagsSchema = z.object({
  mode: z.enum(['all', 'specific', 'prefix', 'category']).default('all'),
  tagIds: z.array(z.string()).optional().default([]),
  prefix: optionalString,
  category: optionalString,
})
export type ClearTagsConfig = z.infer<typeof clearTagsSchema>

// ─── AUTOMATION HELPERS ─────────────────────────────────

export const triggerItSchema = z.object({
  automationId: requiredString,
})
export type TriggerItConfig = z.infer<typeof triggerItSchema>

export const actionItSchema = z.object({
  automationIds: z.array(z.string()).min(1, 'Add at least one automation'),
})
export type ActionItConfig = z.infer<typeof actionItSchema>

export const chainItSchema = z.object({
  helpers: z.array(z.string()).min(1, 'Add at least one helper'),
})
export type ChainItConfig = z.infer<typeof chainItSchema>

export const dripItSchema = z.object({
  steps: z.array(z.string()).min(1, 'Add at least one step'),
  stateField: requiredField,
})
export type DripItConfig = z.infer<typeof dripItSchema>

export const goalItSchema = z.object({
  goalName: requiredString,
  integration: z.string().optional().default('mfh'),
})
export type GoalItConfig = z.infer<typeof goalItSchema>

export const stageItSchema = z.object({
  basicMatch: requiredString,
  toStage: requiredString,
  opportunityCount: z.enum(['first', 'all']).optional().default('first'),
  foundGoal: optionalString,
  notFoundGoal: optionalString,
})
export type StageItConfig = z.infer<typeof stageItSchema>

export const timezoneTriggersSchema = z.object({
  day: requiredString,
  time: requiredString,
  saveTimeZone: optionalField,
  saveLatLng: optionalField,
  saveTimeZoneOffset: optionalField,
  triggerGoal: optionalString,
  failedGoal: optionalString,
})
export type TimezoneTriggersConfig = z.infer<typeof timezoneTriggersSchema>

// ─── INTEGRATION HELPERS ────────────────────────────────

export const hookItSchema = z.object({
  hookAction: optionalString,
  goalPrefix: optionalString,
  actions: z
    .array(
      z.object({
        event: z.string(),
        goalName: z.string(),
      })
    )
    .optional()
    .default([]),
})
export type HookItConfig = z.infer<typeof hookItSchema>

export const mailItSchema = z.object({
  toField: z.string().optional().default('Email'),
  subjectTemplate: requiredString,
  bodyTemplate: requiredString,
  fromName: requiredString,
  fromEmail: z.string().email('Must be a valid email'),
  replyTo: optionalString,
  contentType: z.enum(['text/plain', 'text/html']).optional().default('text/html'),
})
export type MailItConfig = z.infer<typeof mailItSchema>

export const slackItSchema = z.object({
  webhook: z.string().url('Must be a valid URL'),
  message: requiredString,
  username: requiredString,
  channel: optionalString,
  iconEmoji: optionalString,
})
export type SlackItConfig = z.infer<typeof slackItSchema>

export const twilioSmsSchema = z.object({
  accountSid: requiredString,
  authToken: requiredString,
  fromNumber: requiredString,
  messageTemplate: requiredString,
  toField: z.string().optional().default('Phone1'),
})
export type TwilioSmsConfig = z.infer<typeof twilioSmsSchema>

export const calendlySchema = z.object({
  eventTypeUri: requiredString,
  apiToken: requiredString,
  emailField: z.string().optional().default('Email'),
  nameField: z.string().optional().default('FirstName'),
  resultField: optionalField,
})
export type CalendlyConfig = z.infer<typeof calendlySchema>

export const emailValidateSchema = z.object({
  emailField: z.string().optional().default('Email'),
  resultField: requiredField,
  checkMx: z.boolean().optional().default(true),
  validGoal: optionalString,
  invalidGoal: optionalString,
})
export type EmailValidateConfig = z.infer<typeof emailValidateSchema>

export const excelItSchema = z.object({
  fields: z.array(z.string()).min(1, 'Select at least one field'),
  format: z.enum(['csv', 'xlsx']).optional().default('csv'),
  delimiter: z.string().optional().default(','),
  includeHeaders: z.boolean().optional().default(true),
})
export type ExcelItConfig = z.infer<typeof excelItSchema>

export const googleSheetSchema = z.object({
  spreadsheetId: requiredString,
  sheetId: requiredString,
  googleAccountId: requiredString,
  searchData: optionalString,
  translate: z.string().optional().default('false'),
  fields: z.array(z.string()).optional().default([]),
  mode: z.enum(['replace', 'append']).optional().default('replace'),
})
export type GoogleSheetConfig = z.infer<typeof googleSheetSchema>

export const zoomWebinarSchema = z.object({
  webinarId: requiredString,
  apiKey: requiredString,
  apiSecret: requiredString,
  nameField: z.string().optional().default('FirstName'),
  emailField: z.string().optional().default('Email'),
  customQuestions: z.array(z.record(z.string(), z.string())).optional().default([]),
})
export type ZoomWebinarConfig = z.infer<typeof zoomWebinarSchema>

export const zoomMeetingSchema = z.object({
  action: z.enum(['create', 'register']).default('create'),
  userId: optionalString,
  meetingId: optionalString,
  topic: optionalString,
  startTime: optionalString,
  duration: z.number().min(1).optional().default(60),
  timezone: z.string().optional().default('UTC'),
  password: optionalString,
  nameField: z.string().optional().default('FirstName'),
  emailField: z.string().optional().default('Email'),
  saveJoinUrlTo: optionalField,
  registrationType: z.number().min(1).max(3).optional().default(1),
  autoRecording: z.enum(['none', 'local', 'cloud']).optional().default('none'),
})
export type ZoomMeetingConfig = z.infer<typeof zoomMeetingSchema>

export const emailFieldSchema = z.object({
  emailField: z.enum(['Email', 'EmailAddress2', 'EmailAddress3']).default('Email'),
  saveTo: requiredField,
})
export type EmailFieldConfig = z.infer<typeof emailFieldSchema>

export const gotowebinarSchema = z.object({
  organizer_key: requiredString,
  webinar_key: requiredString,
  apply_tag: optionalString,
})
export type GotoWebinarConfig = z.infer<typeof gotowebinarSchema>

export const everwebinarSchema = z.object({
  webinar_id: requiredString,
  schedule: optionalString,
  apply_tag: optionalString,
})
export type EverWebinarConfig = z.infer<typeof everwebinarSchema>

export const webinarJamSchema = z.object({
  webinar_id: requiredString,
  schedule: optionalString,
  apply_tag: optionalString,
})
export type WebinarJamConfig = z.infer<typeof webinarJamSchema>

// ─── NOTIFICATION HELPERS ───────────────────────────────

export const notifyMeSchema = z.object({
  channel: z.enum(['email', 'slack', 'webhook']).default('email'),
  subject: optionalString,
  message: requiredString,
  recipient: optionalString,
  includeFields: z.array(z.string()).optional().default([]),
})
export type NotifyMeConfig = z.infer<typeof notifyMeSchema>

export const emailEngagementSchema = z.object({
  engagementType: z.enum(['opens', 'clicks', 'sends', 'all']).optional().default('all'),
  lookbackDays: z.number().optional().default(30),
  thresholds: z
    .object({
      highlyEngaged: z.number().optional().default(5),
      engaged: z.number().optional().default(2),
    })
    .optional(),
  tags: z
    .object({
      highlyEngagedTag: optionalString,
      engagedTag: optionalString,
      disengagedTag: optionalString,
    })
    .optional(),
  scoreField: optionalField,
  lastEngagementField: optionalField,
})
export type EmailEngagementConfig = z.infer<typeof emailEngagementSchema>

// ─── ANALYTICS HELPERS ──────────────────────────────────

export const rfmSchema = z.object({
  options: z.object({
    recencyCalculation: z
      .object({
        '1_2_threshold': z.number().optional(),
        '2_3_threshold': z.number().optional(),
        '3_4_threshold': z.number().optional(),
        '4_5_threshold': z.number().optional(),
      })
      .optional(),
    frequencyCalculation: z
      .object({
        '1_2_threshold': z.number().optional(),
        '2_3_threshold': z.number().optional(),
        '3_4_threshold': z.number().optional(),
        '4_5_threshold': z.number().optional(),
      })
      .optional(),
    monetaryCalculation: z
      .object({
        '1_2_threshold': z.number().optional(),
        '2_3_threshold': z.number().optional(),
        '3_4_threshold': z.number().optional(),
        '4_5_threshold': z.number().optional(),
      })
      .optional(),
    saveData: z
      .object({
        recencyScore: optionalField,
        frequencyScore: optionalField,
        monetaryScore: optionalField,
        rfmCompositeScore: optionalField,
        totalOrderValue: optionalField,
        averageOrderValue: optionalField,
        totalOrderCount: optionalField,
        firstOrderDate: optionalField,
        firstOrderValue: optionalField,
        lastOrderDate: optionalField,
        lastOrderValue: optionalField,
        daysSinceLastOrder: optionalField,
      })
      .optional(),
  }),
})
export type RfmConfig = z.infer<typeof rfmSchema>

export const clvSchema = z.object({
  lcvTotalOrders: optionalField,
  lcvTotalSpend: optionalField,
  lcvAverageOrder: optionalField,
  lcvTotalDue: optionalField,
  includeZero: z.string().optional().default('No'),
})
export type ClvConfig = z.infer<typeof clvSchema>

// ─── FRONTEND-ONLY SCHEMAS (no Go backend) ──────────────
// These schemas are for helpers that exist in the catalog but
// don't have Go backend implementations yet.

export const countdownTimerSchema = z.object({
  timerType: z.enum(['standard', 'contact_field', 'evergreen']).default('standard'),
  endTime: optionalString,
  contactField: optionalField,
  addDays: z.number().min(0).optional().default(0),
  addHours: z.number().min(0).max(23).optional().default(0),
  addMinutes: z.number().min(0).max(59).optional().default(0),
  backgroundColor: z.string().optional().default('#000000'),
  digitColor: z.string().optional().default('#FFFFFF'),
  labelColor: z.string().optional().default('#CCCCCC'),
  transparentBg: z.boolean().optional().default(false),
})
export type CountdownTimerConfig = z.infer<typeof countdownTimerSchema>

export const quoteItSchema = z.object({
  category: z
    .enum(['inspire', 'management', 'sports', 'life', 'funny', 'love', 'art', 'students'])
    .default('inspire'),
  format: z.enum(['single_line', 'multi_line', 'multi_field']).default('single_line'),
  targetField: optionalField,
  quoteField: optionalField,
  authorField: optionalField,
})
export type QuoteItConfig = z.infer<typeof quoteItSchema>

export const orderItSchema = z.object({
  product_id: z.number().min(1, 'Product ID is required'),
  quantity: z.number().min(1).optional().default(1),
  promo_codes: z.array(z.string()).optional().default([]),
  apply_tag: optionalString,
})
export type OrderItConfig = z.infer<typeof orderItSchema>

export const queryItSchema = z.object({
  savedSearchId: requiredString,
  actionType: z.enum(['tag', 'goal', 'helper']).default('tag'),
  goalName: optionalString,
  actionTags: z.array(z.string()).optional().default([]),
  batchSize: z.number().min(1).max(1000).optional().default(100),
})
export type QueryItConfig = z.infer<typeof queryItSchema>

export const searchItSchema = z.object({
  savedSearchId: requiredString,
  resultField: optionalField,
  matchTags: z.array(z.string()).optional().default([]),
  goalName: optionalString,
})
export type SearchItConfig = z.infer<typeof searchItSchema>

export const routeItSchema = z.object({
  routes: z
    .array(
      z.object({
        label: z.string(),
        redirectUrl: z.string().url('Must be a valid URL'),
      })
    )
    .optional()
    .default([]),
  fallbackUrl: optionalString,
  saveToField: optionalString,
  applyTag: optionalString,
})
export type RouteItConfig = z.infer<typeof routeItSchema>

export const routeItByDaySchema = z.object({
  dayRoutes: z.record(z.string(), z.string().url('Must be a valid URL')),
  fallbackUrl: optionalString,
  timezone: z.string().optional().default('UTC'),
  saveToField: optionalString,
  applyTag: optionalString,
})
export type RouteItByDayConfig = z.infer<typeof routeItByDaySchema>

export const routeItByTimeSchema = z.object({
  timeRoutes: z
    .array(
      z.object({
        startTime: z.string().regex(/^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/, 'Must be HH:MM format'),
        endTime: z.string().regex(/^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/, 'Must be HH:MM format'),
        url: z.string().url('Must be a valid URL'),
        label: optionalString,
      })
    )
    .min(1, 'At least one time route is required'),
  fallbackUrl: optionalString,
  timezone: z.string().optional().default('UTC'),
  saveToField: optionalString,
  applyTag: optionalString,
})
export type RouteItByTimeConfig = z.infer<typeof routeItByTimeSchema>

export const routeItByCustomSchema = z.object({
  fieldName: requiredField,
  valueRoutes: z.record(z.string(), z.string().url('Must be a valid URL')),
  fallbackUrl: optionalString,
  saveToField: optionalString,
  applyTag: optionalString,
})
export type RouteItByCustomConfig = z.infer<typeof routeItByCustomSchema>

export const videoTriggerSchema = z.object({
  videoSource: z.enum(['youtube', 'wistia', 'vimeo']).default('youtube'),
  videoId: requiredString,
  embedType: z.enum(['inline', 'popover']).default('inline'),
  youtubeSize: z.string().optional().default('640x360'),
  autoplay: z.boolean().optional().default(false),
  showControls: z.boolean().optional().default(true),
})
export type VideoTriggerConfig = z.infer<typeof videoTriggerSchema>

export const spreadsheetSchema = z.object({
  sheetUrl: requiredString,
  sheetName: z.string().optional().default('Sheet1'),
  action: z.enum(['append', 'update', 'read']).default('append'),
  columnMappings: z
    .array(
      z.object({
        column: z.string(),
        field: z.string(),
      })
    )
    .optional()
    .default([]),
})
export type SpreadsheetConfig = z.infer<typeof spreadsheetSchema>

export const emailScrubSchema = z.object({
  emailField: requiredField,
  validationLevel: z.enum(['syntax', 'domain', 'full']).default('full'),
  resultField: optionalField,
  validTags: z.array(z.string()).optional().default([]),
  invalidTags: z.array(z.string()).optional().default([]),
})
export type EmailScrubConfig = z.infer<typeof emailScrubSchema>

export const emailAttachItSchema = z.object({
  authorizedSenders: z
    .array(z.string().email('Must be valid email'))
    .min(1, 'Add at least one sender'),
  goalName: optionalString,
})
export type EmailAttachItConfig = z.infer<typeof emailAttachItSchema>

export const webinarSchema = z.object({
  webinarId: requiredString,
  scheduleId: optionalString,
  firstNameField: optionalField,
  lastNameField: optionalField,
  emailField: optionalField,
  phoneField: optionalField,
  joinUrlField: optionalField,
  registeredTags: z.array(z.string()).optional().default([]),
  attendedTags: z.array(z.string()).optional().default([]),
})
export type WebinarConfig = z.infer<typeof webinarSchema>

export const zoomWebinarAbsenteeSchema = z.object({
  webinarId: optionalString,
  tagPrefix: z.string().optional().default('Webinar'),
  applyNoShowTag: z.boolean().optional().default(true),
  applyRegisteredTag: z.boolean().optional().default(true),
  customTags: z.array(z.string()).optional().default([]),
})
export type ZoomWebinarAbsenteeConfig = z.infer<typeof zoomWebinarAbsenteeSchema>

export const zoomWebinarParticipantSchema = z.object({
  webinarId: optionalString,
  attendedPercent: z.number().min(0).max(100).optional().default(0),
  durationMinutes: z.number().min(0).optional().default(0),
  tagPrefix: z.string().optional().default('Webinar'),
  highEngagementThreshold: z.number().min(0).max(100).optional().default(75),
  mediumEngagementThreshold: z.number().min(0).max(100).optional().default(50),
  applyAttendanceTag: z.boolean().optional().default(true),
  applyEngagementTags: z.boolean().optional().default(true),
  applyDurationTag: z.boolean().optional().default(false),
  customTags: z.array(z.string()).optional().default([]),
})
export type ZoomWebinarParticipantConfig = z.infer<typeof zoomWebinarParticipantSchema>

export const cloudStorageSchema = z.object({
  templateFolder: requiredString,
  folderNameField: requiredField,
  folderUrlField: optionalField,
  copyTemplateFiles: z.boolean().optional().default(false),
  shareWithContact: z.boolean().optional().default(false),
})
export type CloudStorageConfig = z.infer<typeof cloudStorageSchema>

export const facebookAudienceSchema = z.object({
  audienceId: requiredString,
  action: z.enum(['add', 'remove']).default('add'),
})
export type FacebookAudienceConfig = z.infer<typeof facebookAudienceSchema>

export const facebookLeadsSchema = z.object({
  pageId: requiredString,
  formId: requiredString,
  fieldMappings: z
    .array(
      z.object({
        fbField: z.string(),
        crmField: z.string(),
      })
    )
    .optional()
    .default([]),
  leadTags: z.array(z.string()).optional().default([]),
})
export type FacebookLeadsConfig = z.infer<typeof facebookLeadsSchema>

export const stripeHooksSchema = z.object({
  selectedEvents: z.array(z.string()).min(1, 'Select at least one event'),
  goalName: optionalString,
  eventTags: z.array(z.string()).optional().default([]),
})
export type StripeHooksConfig = z.infer<typeof stripeHooksSchema>

export const trelloItSchema = z.object({
  board_id: requiredString,
  list_id: requiredString,
  card_name_template: requiredString,
  card_description_template: optionalString,
  apply_tag: optionalString,
  service_connection_ids: z
    .object({
      trello: optionalString,
    })
    .optional()
    .default({ trello: '' }),
})
export type TrelloItConfig = z.infer<typeof trelloItSchema>

export const donorSearchSchema = z.object({
  ds_rating_field: optionalField,
  ds_profile_link_field: optionalField,
  apply_tag: optionalString,
  service_connection_ids: z
    .object({
      donorlead: optionalString,
    })
    .optional()
    .default({ donorlead: '' }),
})
export type DonorSearchConfig = z.infer<typeof donorSearchSchema>

export const uploadItSchema = z.object({
  goal: optionalString,
})
export type UploadItConfig = z.infer<typeof uploadItSchema>

export const hubspotDealStagerSchema = z.object({
  pipelineId: requiredString,
  stageTriggers: z
    .array(
      z.object({
        stageName: z.string(),
        action: z.string(),
      })
    )
    .optional()
    .default([]),
})
export type HubspotDealStagerConfig = z.infer<typeof hubspotDealStagerSchema>

export const hubspotListSyncSchema = z.object({
  listId: requiredString,
  syncDirection: z.enum(['to_list', 'from_list', 'sync']).default('to_list'),
  filterCriteria: optionalString,
})
export type HubspotListSyncConfig = z.infer<typeof hubspotListSyncSchema>

export const hubspotPropertyMapperSchema = z.object({
  objectType: z.enum(['contacts', 'companies', 'deals', 'tickets']).default('contacts'),
  mappings: z
    .array(
      z.object({
        source: z.string(),
        target: z.string(),
      })
    )
    .optional()
    .default([]),
})
export type HubspotPropertyMapperConfig = z.infer<typeof hubspotPropertyMapperSchema>

export const hubspotWorkflowTriggerSchema = z.object({
  workflowId: requiredString,
  enrollmentEmail: optionalString,
})
export type HubspotWorkflowTriggerConfig = z.infer<typeof hubspotWorkflowTriggerSchema>
