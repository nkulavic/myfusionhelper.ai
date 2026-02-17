'use client'

// Schema: see schemas.ts > trelloItSchema
import { FormTextField, FormTextArea, InfoBanner } from './form-fields'
import { TagPicker } from '@/components/tag-picker'
import type { ConfigFormProps } from './types'

export function TrelloItForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const boardId = (config.board_id as string) || ''
  const listId = (config.list_id as string) || ''
  const cardNameTemplate = (config.card_name_template as string) || ''
  const cardDescTemplate = (config.card_description_template as string) || ''
  const applyTag = (config.apply_tag as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <InfoBanner>
        Create Trello cards with contact data. Use placeholders like {'{first_name}'}, {'{last_name}'}
        , {'{email}'}, {'{phone}'}, {'{company}'} in templates.
      </InfoBanner>

      <div className="grid grid-cols-2 gap-3">
        <FormTextField
          label="Board ID *"
          placeholder="Trello board ID"
          value={boardId}
          onChange={(e) => updateConfig({ board_id: e.target.value })}
          disabled={disabled}
          description="The Trello board ID where cards will be created"
        />
        <FormTextField
          label="List ID *"
          placeholder="Trello list ID"
          value={listId}
          onChange={(e) => updateConfig({ list_id: e.target.value })}
          disabled={disabled}
          description="The list ID within the board"
        />
      </div>

      <FormTextField
        label="Card Name Template *"
        placeholder="New Lead: {first_name} {last_name}"
        value={cardNameTemplate}
        onChange={(e) => updateConfig({ card_name_template: e.target.value })}
        disabled={disabled}
        description="Template for card title (supports {first_name}, {last_name}, {email}, {phone}, {company})"
      />

      <FormTextArea
        label="Card Description Template (Optional)"
        placeholder="Contact: {first_name} {last_name}\nEmail: {email}\nPhone: {phone}"
        value={cardDescTemplate}
        onChange={(e) => updateConfig({ card_description_template: e.target.value })}
        disabled={disabled}
        rows={4}
        description="Template for card description (supports the same placeholders)"
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Apply Tag After Creation (Optional)</label>
        <TagPicker
          platformId={platformId ?? ''}
          connectionId={connectionId ?? ''}
          value={applyTag ? [applyTag] : []}
          onChange={(value) => updateConfig({ apply_tag: value[0] || '' })}
          placeholder="Tag to apply after card creation..."
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Tag to apply to the contact after the Trello card is created
        </p>
      </div>
    </div>
  )
}
