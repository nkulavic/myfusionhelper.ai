'use client'

// Schema: see schemas.ts > cloudStorageSchema
import { FieldPicker } from '@/components/field-picker'
import { FormTextField, FormSwitch } from './form-fields'
import type { ConfigFormProps } from './types'

export function CloudStorageForm({ config, onChange, disabled, platformId, connectionId }: ConfigFormProps) {
  const templateFolder = (config.templateFolder as string) || ''
  const folderNameField = (config.folderNameField as string) || ''
  const folderUrlField = (config.folderUrlField as string) || ''
  const shareWithContact = (config.shareWithContact as boolean) ?? false
  const copyTemplateFiles = (config.copyTemplateFiles as boolean) ?? false

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormTextField
        label="Template Folder Path"
        placeholder="/Templates/Client Folder"
        value={templateFolder}
        onChange={(e) => updateConfig({ templateFolder: e.target.value })}
        disabled={disabled}
        description="Path to the template folder to clone for each contact."
      />

      <div className="grid gap-2">
        <label className="text-sm font-medium">Folder Name Field</label>
        <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={folderNameField} onChange={(value) => updateConfig({ folderNameField: value })} placeholder="Field for naming the folder..." disabled={disabled} />
        <p className="text-xs text-muted-foreground">Contact field value used as the folder name.</p>
      </div>

      <div className="grid gap-2">
        <label className="text-sm font-medium">Folder URL Field</label>
        <FieldPicker platformId={platformId ?? ''} connectionId={connectionId ?? ''} value={folderUrlField} onChange={(value) => updateConfig({ folderUrlField: value })} placeholder="Store created folder URL..." disabled={disabled} />
      </div>

      <FormSwitch
        label="Copy template files into new folder"
        checked={copyTemplateFiles}
        onCheckedChange={(v) => updateConfig({ copyTemplateFiles: v })}
        disabled={disabled}
      />

      <FormSwitch
        label="Share folder with contact email"
        checked={shareWithContact}
        onCheckedChange={(v) => updateConfig({ shareWithContact: v })}
        disabled={disabled}
      />
    </div>
  )
}
