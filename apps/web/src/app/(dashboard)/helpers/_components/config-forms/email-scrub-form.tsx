'use client'

// Schema: see schemas.ts > emailScrubSchema

import { Label } from '@/components/ui/label'
import { FieldPicker } from '@/components/field-picker'
import { TagPicker } from '@/components/tag-picker'
import { FormSelect } from './form-fields'
import type { ConfigFormProps } from './types'

const validationLevelOptions = [
  { value: 'syntax', label: 'Syntax Only' },
  { value: 'domain', label: 'Syntax + Domain Check' },
  { value: 'full', label: 'Full Verification' },
]

export function EmailScrubForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const emailField = (config.emailField as string) || ''
  const resultField = (config.resultField as string) || ''
  const validationLevel = (config.validationLevel as string) || 'full'
  const validTags = (config.validTags as string[]) || []
  const invalidTags = (config.invalidTags as string[]) || []

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label>Email Field</Label>
        <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={emailField} onChange={(value) => updateConfig({ emailField: value })} placeholder="Select email field..." disabled={disabled} />
      </div>
      <FormSelect
        label="Validation Level"
        value={validationLevel}
        onValueChange={(value) => updateConfig({ validationLevel: value })}
        options={validationLevelOptions}
        disabled={disabled}
      />
      <div className="grid gap-2">
        <Label>Result Field</Label>
        <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={resultField} onChange={(value) => updateConfig({ resultField: value })} placeholder="Store validation result..." disabled={disabled} />
      </div>
      <div className="grid gap-2">
        <Label>Valid Email Tags</Label>
        <TagPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={validTags} onChange={(value) => updateConfig({ validTags: value })} placeholder="Tags for valid emails..." disabled={disabled} />
      </div>
      <div className="grid gap-2">
        <Label>Invalid Email Tags</Label>
        <TagPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={invalidTags} onChange={(value) => updateConfig({ invalidTags: value })} placeholder="Tags for invalid emails..." disabled={disabled} />
      </div>
    </div>
  )
}
