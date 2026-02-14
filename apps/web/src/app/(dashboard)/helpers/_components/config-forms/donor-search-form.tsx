'use client'

// Schema: see schemas.ts > donorSearchSchema
import { FieldPicker } from '@/components/field-picker'
import { TagPicker } from '@/components/tag-picker'
import { InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

export function DonorSearchForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const dsRatingField = (config.ds_rating_field as string) || ''
  const dsProfileLinkField = (config.ds_profile_link_field as string) || ''
  const applyTag = (config.apply_tag as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <InfoBanner>
        Look up donor profiles and wealth screening via DonorLead.net API using the contact&apos;s
        name, email, and phone. Saves DS_Rating and ProfileLink to custom fields.
      </InfoBanner>

      <div className="grid gap-2">
        <label className="text-sm font-medium">DS Rating Field (Optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={dsRatingField}
          onChange={(value) => updateConfig({ ds_rating_field: value })}
          placeholder="Select field to save DS_Rating..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Custom field to store the donor&apos;s philanthropic rating
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Profile Link Field (Optional)</label>
        <FieldPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={dsProfileLinkField}
          onChange={(value) => updateConfig({ ds_profile_link_field: value })}
          placeholder="Select field to save profile URL..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Custom field to store the donor&apos;s profile link
        </p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Apply Tag on Success (Optional)</label>
        <TagPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={applyTag ? [applyTag] : []}
          onChange={(value) => updateConfig({ apply_tag: value[0] || '' })}
          placeholder="Tag to apply when donor found..."
          disabled={disabled}
          multiple={false}
        />
        <p className="text-xs text-muted-foreground">
          Tag to apply after a successful donor search
        </p>
      </div>
    </div>
  )
}
