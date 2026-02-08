'use client'

// Contact helpers
import { TagItForm } from './tag-it-form'
import { ClearTagsForm } from './clear-tags-form'
import { ClearItForm } from './clear-it-form'
import { CopyItForm } from './copy-it-form'
import { CombineItForm } from './combine-it-form'
import { AssignItForm } from './assign-it-form'
import { FoundItForm } from './found-it-form'
import { MergeItForm } from './merge-it-form'
import { NameParseItForm } from './name-parse-it-form'
import { NoteItForm } from './note-it-form'
import { DefaultToFieldForm } from './default-to-field-form'
import { OptInForm } from './opt-in-form'
import { SnapshotItForm } from './snapshot-it-form'
import { ContactUpdaterForm } from './contact-updater-form'

// Data helpers
import { FormatItForm } from './format-it-form'
import { MathItForm } from './math-it-form'
import { DateCalcForm } from './date-calc-form'
import { TextItForm } from './text-it-form'
import { SplitItForm } from './split-it-form'
import { PasswordItForm } from './password-it-form'
import { WordCountItForm } from './word-count-it-form'
import { WhenIsItForm } from './when-is-it-form'
import { GetRecordForm } from './get-record-form'
import { EmailFieldForm } from './email-field-form'
import { CountdownTimerForm } from './countdown-timer-form'
import { QuoteItForm } from './quote-it-form'
import { OrderItForm } from './order-it-form'
import { QueryItForm } from './query-it-form'
import { SearchItForm } from './search-it-form'

// Tagging helpers
import { ScoreItForm } from './score-it-form'
import { GroupItForm } from './group-it-form'
import { CountTagsForm } from './count-tags-form'

// Automation helpers
import { TriggerItForm } from './trigger-it-form'
import { ChainItForm } from './chain-it-form'
import { GoalItForm } from './goal-it-form'
import { ActionItForm } from './action-it-form'
import { DripItForm } from './drip-it-form'
import { StageItForm } from './stage-it-form'
import { TimezoneTriggersForm } from './timezone-triggers-form'
import { RouteItForm } from './route-it-form'
import { VideoTriggerForm } from './video-trigger-form'

// Integration helpers
import { HookItForm } from './hook-it-form'
import { SlackItForm } from './slack-it-form'
import { PhoneLookupForm } from './phone-lookup-form'
import { SpreadsheetForm } from './spreadsheet-form'
import { MailItForm } from './mail-it-form'
import { EmailScrubForm } from './email-scrub-form'
import { EmailValidateForm } from './email-validate-form'
import { EmailAttachItForm } from './email-attach-it-form'
import { WebinarForm } from './webinar-form'
import { CalendlyForm } from './calendly-form'
import { CloudStorageForm } from './cloud-storage-form'
import { FacebookAudienceForm } from './facebook-audience-form'
import { FacebookLeadsForm } from './facebook-leads-form'
import { StripeHooksForm } from './stripe-hooks-form'
import { TrelloItForm } from './trello-it-form'
import { DonorSearchForm } from './donor-search-form'
import { UploadItForm } from './upload-it-form'

// HubSpot helpers
import { HubspotDealStagerForm } from './hubspot-deal-stager-form'
import { HubspotListSyncForm } from './hubspot-list-sync-form'
import { HubspotPropertyMapperForm } from './hubspot-property-mapper-form'
import { HubspotWorkflowTriggerForm } from './hubspot-workflow-trigger-form'

// Notification helpers
import { NotifyMeForm } from './notify-me-form'
import { TwilioSmsForm } from './twilio-sms-form'
import { EmailEngagementForm } from './email-engagement-form'

// Analytics helpers
import { RfmForm } from './rfm-form'
import { ClvForm } from './clv-form'
import { IpLocationForm } from './ip-location-form'

import type { ConfigFormProps } from './types'

export type { ConfigFormProps } from './types'

// Map helper types to their config form components.
// Helper types without a registered form will fall back to the JSON editor.
const formRegistry: Record<string, React.ComponentType<ConfigFormProps>> = {
  // ========== Contact helpers ==========
  tag_it: TagItForm,
  clear_tags: ClearTagsForm,
  clear_it: ClearItForm,
  copy_it: CopyItForm,
  field_to_field: CopyItForm,
  move_it: CopyItForm,
  found_it: FoundItForm,
  merge_it: MergeItForm,
  combine_it: CombineItForm,
  assign_it: AssignItForm,
  own_it: AssignItForm,
  name_parse_it: NameParseItForm,
  note_it: NoteItForm,
  default_to_field: DefaultToFieldForm,
  company_link: DefaultToFieldForm,
  opt_in: OptInForm,
  opt_out: OptInForm, // same config as opt-in
  snapshot_it: SnapshotItForm,
  contact_updater: ContactUpdaterForm,

  // ========== Data helpers ==========
  format_it: FormatItForm,
  math_it: MathItForm,
  date_calc: DateCalcForm,
  text_it: TextItForm,
  split_it: SplitItForm,
  split_it_basic: SplitItForm,
  password_it: PasswordItForm,
  word_count_it: WordCountItForm,
  when_is_it: WhenIsItForm,
  get_the_first: GetRecordForm,
  get_the_last: GetRecordForm,
  countdown_timer: CountdownTimerForm,
  quote_it: QuoteItForm,
  order_it: OrderItForm,
  query_it_basic: QueryItForm,
  search_it: SearchItForm,

  // ========== Tagging helpers ==========
  score_it: ScoreItForm,
  group_it: GroupItForm,
  count_tags: CountTagsForm,
  count_it_tags: CountTagsForm,

  // ========== Automation helpers ==========
  trigger_it: TriggerItForm,
  chain_it: ChainItForm,
  goal_it: GoalItForm,
  action_it: ActionItForm,
  drip_it: DripItForm,
  stage_it: StageItForm,
  timezone_triggers: TimezoneTriggersForm,
  route_it: RouteItForm,
  video_trigger_it: VideoTriggerForm,

  // ========== Integration helpers ==========
  hook_it: HookItForm,
  slack_it: SlackItForm,
  phone_lookup: PhoneLookupForm,
  google_sheet_it: SpreadsheetForm,
  excel_it: SpreadsheetForm,
  mail_it: MailItForm,
  email_scrub_it: EmailScrubForm,
  email_validate_it: EmailValidateForm,
  email_scrub_validate: EmailValidateForm,
  email_attach_it: EmailAttachItForm,
  zoom_webinar: WebinarForm,
  zoom_meeting: WebinarForm,
  zoom_webinar_oauth: WebinarForm,
  gotowebinar: WebinarForm,
  everwebinar: WebinarForm,
  webinar_jam: WebinarForm,
  calendly_it: CalendlyForm,
  dropbox_it: CloudStorageForm,
  google_drive_it: CloudStorageForm,
  facebook_custom_audience: FacebookAudienceForm,
  facebook_lead_ads: FacebookLeadsForm,
  stripe_hooks: StripeHooksForm,
  trello_it: TrelloItForm,
  donor_search: DonorSearchForm,
  upload_it: UploadItForm,

  // ========== HubSpot helpers ==========
  hubspot_deal_stager: HubspotDealStagerForm,
  hubspot_list_sync: HubspotListSyncForm,
  hubspot_property_mapper: HubspotPropertyMapperForm,
  hubspot_workflow_trigger: HubspotWorkflowTriggerForm,

  // ========== Notification helpers ==========
  notify_me: NotifyMeForm,
  twilio_sms: TwilioSmsForm,
  email_engagement: EmailEngagementForm,

  // ========== Email engagement (last_click, last_open, last_send share same form) ==========
  last_click_it: EmailFieldForm,
  last_open_it: EmailFieldForm,
  last_send_it: EmailFieldForm,

  // ========== Analytics helpers ==========
  rfm_calculation: RfmForm,
  customer_lifetime_value: ClvForm,
  ip_location: IpLocationForm,
}

export function getConfigForm(helperType: string): React.ComponentType<ConfigFormProps> | null {
  return formRegistry[helperType] ?? null
}

export function hasConfigForm(helperType: string): boolean {
  return helperType in formRegistry
}
