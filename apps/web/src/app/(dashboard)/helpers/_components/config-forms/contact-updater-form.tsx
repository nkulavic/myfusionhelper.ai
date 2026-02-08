'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { X, Plus } from 'lucide-react'
import { FormTextField, InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

export function ContactUpdaterForm({ config, onChange, disabled }: ConfigFormProps) {
  const fields = (config.fields as Record<string, any>) || {}
  const secondaryContactIds = (config.secondaryContactIds as string[]) || []

  const [newFieldName, setNewFieldName] = useState('')
  const [newFieldValue, setNewFieldValue] = useState('')
  const [newContactId, setNewContactId] = useState('')

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  const addField = () => {
    if (!newFieldName.trim() || !newFieldValue.trim()) return
    updateConfig({
      fields: { ...fields, [newFieldName.trim()]: newFieldValue.trim() },
    })
    setNewFieldName('')
    setNewFieldValue('')
  }

  const removeField = (fieldName: string) => {
    const updatedFields = { ...fields }
    delete updatedFields[fieldName]
    updateConfig({ fields: updatedFields })
  }

  const updateFieldValue = (fieldName: string, value: string) => {
    updateConfig({
      fields: { ...fields, [fieldName]: value },
    })
  }

  const addSecondaryContact = () => {
    if (!newContactId.trim()) return
    updateConfig({
      secondaryContactIds: [...secondaryContactIds, newContactId.trim()],
    })
    setNewContactId('')
  }

  const removeSecondaryContact = (index: number) => {
    updateConfig({
      secondaryContactIds: secondaryContactIds.filter((_, i) => i !== index),
    })
  }

  return (
    <div className="space-y-6">
      <InfoBanner>
        Updates arbitrary contact fields with specified values. Optionally trigger goals for
        secondary contacts.
      </InfoBanner>

      <div className="space-y-4">
        <div>
          <Label className="text-sm font-medium">Field Mappings</Label>
          <p className="text-xs text-muted-foreground mb-3">
            Define which fields to update and their values
          </p>

          {/* Existing field mappings */}
          {Object.entries(fields).length > 0 && (
            <div className="space-y-2 mb-3">
              {Object.entries(fields).map(([fieldName, value]) => (
                <div key={fieldName} className="flex gap-2 items-center">
                  <Input
                    value={fieldName}
                    disabled
                    className="flex-1 bg-muted"
                    placeholder="Field name"
                  />
                  <Input
                    value={String(value)}
                    onChange={(e) => updateFieldValue(fieldName, e.target.value)}
                    disabled={disabled}
                    className="flex-1"
                    placeholder="Field value"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => removeField(fieldName)}
                    disabled={disabled}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          )}

          {/* Add new field mapping */}
          <div className="flex gap-2 items-end">
            <div className="flex-1">
              <Input
                value={newFieldName}
                onChange={(e) => setNewFieldName(e.target.value)}
                disabled={disabled}
                placeholder="Field name (e.g., status, score)"
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    addField()
                  }
                }}
              />
            </div>
            <div className="flex-1">
              <Input
                value={newFieldValue}
                onChange={(e) => setNewFieldValue(e.target.value)}
                disabled={disabled}
                placeholder="Field value"
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    addField()
                  }
                }}
              />
            </div>
            <Button
              type="button"
              variant="outline"
              size="icon"
              onClick={addField}
              disabled={disabled || !newFieldName.trim() || !newFieldValue.trim()}
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Secondary contact IDs */}
        <div>
          <Label className="text-sm font-medium">Secondary Contact IDs (Optional)</Label>
          <p className="text-xs text-muted-foreground mb-3">
            Additional contact IDs to trigger the &apos;contact_updated&apos; goal for
          </p>

          {secondaryContactIds.length > 0 && (
            <div className="space-y-2 mb-3">
              {secondaryContactIds.map((contactId, index) => (
                <div key={index} className="flex gap-2 items-center">
                  <Input value={contactId} disabled className="flex-1 bg-muted" />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => removeSecondaryContact(index)}
                    disabled={disabled}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          )}

          <div className="flex gap-2">
            <Input
              value={newContactId}
              onChange={(e) => setNewContactId(e.target.value)}
              disabled={disabled}
              placeholder="Contact ID"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  addSecondaryContact()
                }
              }}
            />
            <Button
              type="button"
              variant="outline"
              size="icon"
              onClick={addSecondaryContact}
              disabled={disabled || !newContactId.trim()}
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
