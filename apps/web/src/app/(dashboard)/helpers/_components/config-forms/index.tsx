'use client'

import { TagItForm } from './tag-it-form'
import { ClearTagsForm } from './clear-tags-form'
import { CopyItForm } from './copy-it-form'
import { FormatItForm } from './format-it-form'
import { HookItForm } from './hook-it-form'
import { MathItForm } from './math-it-form'
import { NotifyMeForm } from './notify-me-form'
import type { ConfigFormProps } from './types'

export type { ConfigFormProps } from './types'

// Map helper types to their config form components.
// Helper types without a registered form will fall back to the JSON editor.
const formRegistry: Record<string, React.ComponentType<ConfigFormProps>> = {
  tag_it: TagItForm,
  clear_tags: ClearTagsForm,
  copy_it: CopyItForm,
  field_to_field: CopyItForm, // reuses copy_it form (same field-to-field pattern)
  move_it: CopyItForm,
  format_it: FormatItForm,
  math_it: MathItForm,
  hook_it: HookItForm,
  notify_me: NotifyMeForm,
}

export function getConfigForm(helperType: string): React.ComponentType<ConfigFormProps> | null {
  return formRegistry[helperType] ?? null
}

export function hasConfigForm(helperType: string): boolean {
  return helperType in formRegistry
}
