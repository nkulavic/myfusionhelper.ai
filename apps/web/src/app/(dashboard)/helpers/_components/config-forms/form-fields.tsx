'use client'

import * as React from 'react'
import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Checkbox } from '@/components/ui/checkbox'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

// ─── Types ──────────────────────────────────────────────

interface SelectOption {
  value: string
  label: string
  description?: string
}

// ─── FormSelect ─────────────────────────────────────────

interface FormSelectProps {
  label: string
  description?: string
  value: string
  onValueChange: (value: string) => void
  options: SelectOption[]
  placeholder?: string
  disabled?: boolean
  error?: string
  className?: string
}

export function FormSelect({
  label,
  description,
  value,
  onValueChange,
  options,
  placeholder,
  disabled,
  error,
  className,
}: FormSelectProps) {
  return (
    <div className={cn('grid gap-2', className)}>
      <Label>{label}</Label>
      <Select value={value} onValueChange={onValueChange} disabled={disabled}>
        <SelectTrigger>
          <SelectValue placeholder={placeholder ?? 'Select...'} />
        </SelectTrigger>
        <SelectContent>
          {options.map((opt) => (
            <SelectItem key={opt.value} value={opt.value}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {description && <p className="text-xs text-muted-foreground">{description}</p>}
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}

// ─── FormSwitch ─────────────────────────────────────────

interface FormSwitchProps {
  label: string
  description?: string
  checked: boolean
  onCheckedChange: (checked: boolean) => void
  disabled?: boolean
  className?: string
}

export function FormSwitch({
  label,
  description,
  checked,
  onCheckedChange,
  disabled,
  className,
}: FormSwitchProps) {
  const id = React.useId()
  return (
    <div className={cn('flex items-center justify-between gap-4 rounded-lg border p-3', className)}>
      <div className="space-y-0.5">
        <Label htmlFor={id} className="text-sm font-medium cursor-pointer">
          {label}
        </Label>
        {description && (
          <p className="text-xs text-muted-foreground">{description}</p>
        )}
      </div>
      <Switch id={id} checked={checked} onCheckedChange={onCheckedChange} disabled={disabled} />
    </div>
  )
}

// ─── FormCheckbox ───────────────────────────────────────

interface FormCheckboxProps {
  label: string
  description?: string
  checked: boolean
  onCheckedChange: (checked: boolean) => void
  disabled?: boolean
  className?: string
}

export function FormCheckbox({
  label,
  description,
  checked,
  onCheckedChange,
  disabled,
  className,
}: FormCheckboxProps) {
  const id = React.useId()
  return (
    <div className={cn('flex items-start gap-3', className)}>
      <Checkbox
        id={id}
        checked={checked}
        onCheckedChange={(v) => onCheckedChange(v === true)}
        disabled={disabled}
        className="mt-0.5"
      />
      <div className="grid gap-0.5 leading-none">
        <Label htmlFor={id} className="text-sm font-normal cursor-pointer">
          {label}
        </Label>
        {description && (
          <p className="text-xs text-muted-foreground">{description}</p>
        )}
      </div>
    </div>
  )
}

// ─── FormRadioGroup ─────────────────────────────────────

interface FormRadioGroupProps {
  label: string
  description?: string
  value: string
  onValueChange: (value: string) => void
  options: SelectOption[]
  disabled?: boolean
  orientation?: 'horizontal' | 'vertical'
  error?: string
  className?: string
}

export function FormRadioGroup({
  label,
  description,
  value,
  onValueChange,
  options,
  disabled,
  orientation = 'horizontal',
  error,
  className,
}: FormRadioGroupProps) {
  return (
    <div className={cn('grid gap-2', className)}>
      <Label>{label}</Label>
      {description && <p className="text-xs text-muted-foreground">{description}</p>}
      <RadioGroup
        value={value}
        onValueChange={onValueChange}
        disabled={disabled}
        className={cn(
          orientation === 'horizontal' ? 'flex flex-wrap gap-4' : 'grid gap-2'
        )}
      >
        {options.map((opt) => (
          <div key={opt.value} className="flex items-center gap-2">
            <RadioGroupItem value={opt.value} id={`radio-${opt.value}`} />
            <Label htmlFor={`radio-${opt.value}`} className="text-sm font-normal cursor-pointer">
              {opt.label}
            </Label>
          </div>
        ))}
      </RadioGroup>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}

// ─── FormTextField ───────────────────────────────────────

interface FormTextFieldProps extends React.ComponentProps<typeof Input> {
  label: string
  description?: string
  error?: string
}

export function FormTextField({
  label,
  description,
  error,
  id: externalId,
  className,
  ...inputProps
}: FormTextFieldProps) {
  const generatedId = React.useId()
  const id = externalId ?? generatedId
  return (
    <div className={cn('grid gap-2', className)}>
      <Label htmlFor={id}>{label}</Label>
      <Input id={id} {...inputProps} />
      {description && <p className="text-xs text-muted-foreground">{description}</p>}
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}

// ─── FormTextArea ───────────────────────────────────────

interface FormTextAreaProps extends React.ComponentProps<typeof Textarea> {
  label: string
  description?: string
  error?: string
}

export function FormTextArea({
  label,
  description,
  error,
  id: externalId,
  className,
  ...textareaProps
}: FormTextAreaProps) {
  const generatedId = React.useId()
  const id = externalId ?? generatedId
  return (
    <div className={cn('grid gap-2', className)}>
      <Label htmlFor={id}>{label}</Label>
      <Textarea id={id} {...textareaProps} />
      {description && <p className="text-xs text-muted-foreground">{description}</p>}
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}

// ─── FormColorPicker ────────────────────────────────────

interface FormColorPickerProps {
  label: string
  value: string
  onChange: (value: string) => void
  disabled?: boolean
  className?: string
}

export function FormColorPicker({
  label,
  value,
  onChange,
  disabled,
  className,
}: FormColorPickerProps) {
  const id = React.useId()
  return (
    <div className={cn('grid gap-2', className)}>
      <Label htmlFor={id}>{label}</Label>
      <div className="flex items-center gap-2">
        <Input
          id={id}
          type="color"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          className="h-10 w-14 cursor-pointer p-1"
        />
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          className="flex-1 font-mono text-xs"
          maxLength={7}
        />
      </div>
    </div>
  )
}

// ─── DynamicList ────────────────────────────────────────

interface DynamicListProps<T> {
  label: string
  description?: string
  items: T[]
  onItemsChange: (items: T[]) => void
  renderItem: (item: T, index: number) => React.ReactNode
  renderAddForm: (onAdd: (item: T) => void) => React.ReactNode
  disabled?: boolean
  className?: string
}

export function DynamicList<T>({
  label,
  description,
  items,
  onItemsChange,
  renderItem,
  renderAddForm,
  disabled,
  className,
}: DynamicListProps<T>) {
  const handleAdd = (item: T) => {
    onItemsChange([...items, item])
  }

  const handleRemove = (index: number) => {
    onItemsChange(items.filter((_, i) => i !== index))
  }

  return (
    <div className={cn('grid gap-2', className)}>
      <Label>{label}</Label>
      {description && <p className="text-xs text-muted-foreground">{description}</p>}
      {!disabled && renderAddForm(handleAdd)}
      {items.length > 0 && (
        <div className="space-y-1.5 mt-1">
          {items.map((item, i) => (
            <div key={i} className="flex items-center gap-2 rounded-md bg-muted px-3 py-2 text-xs">
              <div className="flex-1">{renderItem(item, i)}</div>
              {!disabled && (
                <button
                  type="button"
                  onClick={() => handleRemove(i)}
                  className="rounded p-0.5 hover:bg-accent"
                  aria-label="Remove item"
                >
                  <X className="h-3 w-3" />
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

// ─── AddItemRow (helper for DynamicList) ────────────────

interface AddItemRowProps {
  children: React.ReactNode
  onAdd: () => void
  disabled?: boolean
  canAdd?: boolean
}

export function AddItemRow({ children, onAdd, disabled, canAdd = true }: AddItemRowProps) {
  return (
    <div className="flex gap-2">
      {children}
      <Button
        type="button"
        variant="outline"
        size="icon"
        onClick={onAdd}
        disabled={disabled || !canAdd}
        aria-label="Add item"
      >
        <Plus className="h-4 w-4" />
      </Button>
    </div>
  )
}

// ─── InfoBanner ─────────────────────────────────────────

interface InfoBannerProps {
  children: React.ReactNode
  className?: string
}

export function InfoBanner({ children, className }: InfoBannerProps) {
  return (
    <div className={cn('rounded-md border border-dashed bg-muted/50 p-3', className)}>
      <p className="text-xs text-muted-foreground">{children}</p>
    </div>
  )
}
