'use client'

// Schema: see schemas.ts > trelloItSchema
import { FormTextField, FormTextArea } from './form-fields'
import type { ConfigFormProps } from './types'

export function TrelloItForm({ config, onChange, disabled }: ConfigFormProps) {
  const boardId = (config.boardId as string) || ''
  const listId = (config.listId as string) || ''
  const cardTitle = (config.cardTitle as string) || ''
  const cardDescription = (config.cardDescription as string) || ''
  const checklist = (config.checklist as string) || ''

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3">
        <FormTextField
          label="Board ID"
          placeholder="Trello board ID"
          value={boardId}
          onChange={(e) => updateConfig({ boardId: e.target.value })}
          disabled={disabled}
        />
        <FormTextField
          label="List ID"
          placeholder="Trello list ID"
          value={listId}
          onChange={(e) => updateConfig({ listId: e.target.value })}
          disabled={disabled}
        />
      </div>

      <FormTextField
        label="Card Title"
        placeholder="Card title (supports merge fields)"
        value={cardTitle}
        onChange={(e) => updateConfig({ cardTitle: e.target.value })}
        disabled={disabled}
      />

      <FormTextArea
        label="Card Description"
        placeholder="Card description..."
        value={cardDescription}
        onChange={(e) => updateConfig({ cardDescription: e.target.value })}
        disabled={disabled}
        rows={3}
      />

      <FormTextArea
        label="Checklist Items"
        placeholder="One item per line"
        value={checklist}
        onChange={(e) => updateConfig({ checklist: e.target.value })}
        disabled={disabled}
        rows={3}
        description="Enter one checklist item per line."
      />
    </div>
  )
}
