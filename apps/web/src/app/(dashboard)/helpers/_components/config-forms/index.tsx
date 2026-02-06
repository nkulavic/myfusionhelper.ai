'use client'

import { TagItForm } from './tag-it-form'
import { ClearTagsForm } from './clear-tags-form'
import { CopyItForm } from './copy-it-form'
import { FormatItForm } from './format-it-form'
import { HookItForm } from './hook-it-form'
import { MathItForm } from './math-it-form'
import { NotifyMeForm } from './notify-me-form'
import { ScoreItForm } from './score-it-form'
import { TriggerItForm } from './trigger-it-form'
import { DateCalcForm } from './date-calc-form'
import { TextItForm } from './text-it-form'
import { FoundItForm } from './found-it-form'
import { MergeItForm } from './merge-it-form'
import { SplitItForm } from './split-it-form'
import { ChainItForm } from './chain-it-form'
import { SlackItForm } from './slack-it-form'
import { AssignItForm } from './assign-it-form'
import type { ConfigFormProps } from './types'

export type { ConfigFormProps } from './types'

// Map helper types to their config form components.
// Helper types without a registered form will fall back to the JSON editor.
const formRegistry: Record<string, React.ComponentType<ConfigFormProps>> = {
  // Contact helpers
  tag_it: TagItForm,
  clear_tags: ClearTagsForm,
  copy_it: CopyItForm,
  field_to_field: CopyItForm,
  move_it: CopyItForm,
  found_it: FoundItForm,
  merge_it: MergeItForm,
  assign_it: AssignItForm,
  own_it: AssignItForm,
  // Data helpers
  format_it: FormatItForm,
  math_it: MathItForm,
  date_calc: DateCalcForm,
  text_it: TextItForm,
  split_it: SplitItForm,
  // Tagging helpers
  score_it: ScoreItForm,
  // Automation helpers
  trigger_it: TriggerItForm,
  chain_it: ChainItForm,
  // Integration helpers
  hook_it: HookItForm,
  slack_it: SlackItForm,
  // Notification helpers
  notify_me: NotifyMeForm,
}

export function getConfigForm(helperType: string): React.ComponentType<ConfigFormProps> | null {
  return formRegistry[helperType] ?? null
}

export function hasConfigForm(helperType: string): boolean {
  return helperType in formRegistry
}
