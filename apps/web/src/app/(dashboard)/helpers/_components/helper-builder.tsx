'use client'

import { useState, useMemo, useRef } from 'react'
import {
  ArrowLeft,
  Search,
  ChevronRight,
  Sparkles,
  Info,
  Loader2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  helpersCatalog,
  categoryInfo,
  type HelperDefinition,
} from '@/lib/helpers-catalog'
import { useCreateHelper, useHelperTypes } from '@/lib/hooks/use-helpers'
import { useConnections } from '@/lib/hooks/use-connections'
import type { PlatformConnection } from '@myfusionhelper/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

interface HelperBuilderProps {
  initialType?: string
  onBack: () => void
  onCreated: (id: string) => void
}

export function HelperBuilder({ initialType, onBack, onCreated }: HelperBuilderProps) {
  const initialHelper = initialType
    ? helpersCatalog.find((h) => h.id === initialType && h.status === 'available') ?? null
    : null

  const [selectedCategory, setSelectedCategory] = useState('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedHelper, setSelectedHelper] = useState<HelperDefinition | null>(initialHelper)
  const [helperName, setHelperName] = useState(initialHelper ? `My ${initialHelper.name}` : '')
  const [selectedConnection, setSelectedConnection] = useState('')
  const [step, setStep] = useState<'select' | 'configure'>(initialHelper ? 'configure' : 'select')
  const configRef = useRef<Record<string, unknown>>({})
  const createHelper = useCreateHelper()
  const { data: backendTypes } = useHelperTypes()

  // Merge backend types with static catalog (backend takes priority)
  const allHelpers = useMemo(() => {
    const staticMap = new Map(helpersCatalog.map((h) => [h.id, h]))

    if (backendTypes?.types && backendTypes.types.length > 0) {
      const seen = new Set<string>()
      const merged: HelperDefinition[] = []

      for (const bt of backendTypes.types) {
        seen.add(bt.type)
        const staticEntry = staticMap.get(bt.type)
        merged.push({
          id: bt.type,
          name: bt.name,
          description: bt.description,
          category: bt.category as HelperDefinition['category'],
          requiresCRM: bt.requiresCrm,
          supportedCRMs: bt.supportedCrms ?? [],
          icon: staticEntry?.icon ?? Sparkles,
          popular: staticEntry?.popular ?? false,
          status: staticEntry?.status ?? 'available',
        })
      }

      // Add static-only entries (e.g. coming_soon)
      for (const entry of helpersCatalog) {
        if (!seen.has(entry.id)) {
          merged.push(entry)
        }
      }

      return merged
    }

    return helpersCatalog
  }, [backendTypes])

  const categoryCounts = useMemo(() => {
    const counts: Record<string, number> = { all: allHelpers.filter((h) => h.status === 'available').length }
    for (const h of allHelpers) {
      if (h.status === 'available') {
        counts[h.category] = (counts[h.category] || 0) + 1
      }
    }
    return counts
  }, [allHelpers])

  const availableHelpers = useMemo(() => {
    return allHelpers
      .filter((h) => h.status === 'available')
      .filter((helper) => {
        const matchesCategory =
          selectedCategory === 'all' || helper.category === selectedCategory
        const matchesSearch =
          searchQuery === '' ||
          helper.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
          helper.description.toLowerCase().includes(searchQuery.toLowerCase())
        return matchesCategory && matchesSearch
      })
  }, [allHelpers, selectedCategory, searchQuery])

  const handleSelectHelper = (helper: HelperDefinition) => {
    setSelectedHelper(helper)
    setHelperName(`My ${helper.name}`)
    setStep('configure')
  }

  const handleBack = () => {
    onBack()
  }

  const handleSubmit = () => {
    if (!selectedHelper) return
    createHelper.mutate(
      {
        name: helperName,
        helperType: selectedHelper.id,
        category: selectedHelper.category,
        connectionId: selectedConnection,
        config: configRef.current,
      },
      {
        onSuccess: (res) => {
          const id = res.data?.helperId
          if (id) {
            onCreated(id)
          } else {
            onBack()
          }
        },
      }
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={handleBack}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-2xl font-bold">
            {step === 'select' ? 'Create New Helper' : 'Configure Helper'}
          </h1>
          <p className="text-muted-foreground">
            {step === 'select'
              ? 'Choose a helper type to get started'
              : `Setting up ${selectedHelper?.name}`}
          </p>
        </div>
      </div>

      {/* Step Indicator */}
      <div className="flex items-center gap-3">
        <div
          className={cn(
            'flex items-center gap-2 rounded-full px-3 py-1 text-sm font-medium',
            step === 'select'
              ? 'bg-primary text-primary-foreground'
              : 'bg-muted text-muted-foreground'
          )}
        >
          <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary-foreground/20 text-xs">
            1
          </span>
          Select Type
        </div>
        <ChevronRight className="h-4 w-4 text-muted-foreground" />
        <div
          className={cn(
            'flex items-center gap-2 rounded-full px-3 py-1 text-sm font-medium',
            step === 'configure'
              ? 'bg-primary text-primary-foreground'
              : 'bg-muted text-muted-foreground'
          )}
        >
          <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary-foreground/20 text-xs">
            2
          </span>
          Configure
        </div>
      </div>

      {createHelper.error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
          {createHelper.error instanceof Error ? createHelper.error.message : 'Failed to create helper'}
        </div>
      )}

      {step === 'select' ? (
        <>
          {/* Search */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              type="text"
              placeholder="Search available helpers..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
              autoFocus
            />
          </div>

          {/* Category Tabs */}
          <div className="flex flex-wrap gap-2">
            {categoryInfo.map((category) => {
              const count = categoryCounts[category.id] || 0
              return (
                <button
                  key={category.id}
                  onClick={() => setSelectedCategory(category.id)}
                  className={cn(
                    'inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium transition-colors',
                    selectedCategory === category.id
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted text-muted-foreground hover:bg-muted/80'
                  )}
                >
                  {category.name}
                  <span
                    className={cn(
                      'rounded-full px-1.5 py-0.5 text-[10px]',
                      selectedCategory === category.id
                        ? 'bg-primary-foreground/20'
                        : 'bg-background'
                    )}
                  >
                    {count}
                  </span>
                </button>
              )
            })}
          </div>

          {/* Helper Selection Grid */}
          <div className="grid gap-3 sm:grid-cols-2">
            {availableHelpers.map((helper) => (
              <button
                key={helper.id}
                onClick={() => handleSelectHelper(helper)}
                className="group flex items-start gap-4 rounded-lg border bg-card p-4 text-left transition-all hover:border-primary hover:shadow-sm"
              >
                <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary/10">
                  <helper.icon className="h-5 w-5 text-primary" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-semibold group-hover:text-primary">
                      {helper.name}
                    </h3>
                    {helper.popular && (
                      <Sparkles className="h-3 w-3 text-primary" />
                    )}
                  </div>
                  <p className="mt-0.5 text-xs text-muted-foreground leading-relaxed">
                    {helper.description}
                  </p>
                  {helper.supportedCRMs.length > 0 && (
                    <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground/60">
                      {helper.supportedCRMs.join(', ')} only
                    </p>
                  )}
                </div>
                <ChevronRight className="h-4 w-4 flex-shrink-0 text-muted-foreground/50 group-hover:text-primary" />
              </button>
            ))}
          </div>

          {availableHelpers.length === 0 && (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Search className="mb-4 h-12 w-12 text-muted-foreground/50" />
              <h3 className="mb-1 font-semibold">No helpers found</h3>
              <p className="text-sm text-muted-foreground">
                Try a different search term or category
              </p>
            </div>
          )}
        </>
      ) : (
        selectedHelper && (
          <HelperConfigForm
            helper={selectedHelper}
            helperName={helperName}
            setHelperName={setHelperName}
            selectedConnection={selectedConnection}
            setSelectedConnection={setSelectedConnection}
            configRef={configRef}
            onSubmit={handleSubmit}
            onCancel={onBack}
            isSubmitting={createHelper.isPending}
          />
        )
      )}
    </div>
  )
}

function HelperConfigForm({
  helper,
  helperName,
  setHelperName,
  selectedConnection,
  setSelectedConnection,
  configRef,
  onSubmit,
  onCancel,
  isSubmitting,
}: {
  helper: HelperDefinition
  helperName: string
  setHelperName: (name: string) => void
  selectedConnection: string
  setSelectedConnection: (id: string) => void
  configRef: React.MutableRefObject<Record<string, unknown>>
  onSubmit: () => void
  onCancel: () => void
  isSubmitting: boolean
}) {
  const { data: connections } = useConnections()

  return (
    <div className="space-y-6">
      {/* Selected Helper Info */}
      <div className="flex items-start gap-4 rounded-lg border bg-primary/5 p-4">
        <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary/10">
          <helper.icon className="h-5 w-5 text-primary" />
        </div>
        <div>
          <h3 className="font-semibold">{helper.name}</h3>
          <p className="text-sm text-muted-foreground">{helper.description}</p>
        </div>
      </div>

      {/* Basic Info */}
      <div className="rounded-lg border bg-card p-5 space-y-4">
        <h3 className="font-semibold">Basic Information</h3>

        <div className="space-y-2">
          <label className="text-sm font-medium">Helper Name</label>
          <Input
            type="text"
            value={helperName}
            onChange={(e) => setHelperName(e.target.value)}
            placeholder="Give your helper a name..."
          />
        </div>

        {helper.requiresCRM && (
          <div className="space-y-2">
            <label className="text-sm font-medium">CRM Connection</label>
            <select
              value={selectedConnection}
              onChange={(e) => setSelectedConnection(e.target.value)}
              className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              <option value="">Select a connection...</option>
              {connections && connections.length > 0 ? (
                connections.map((conn: PlatformConnection) => (
                  <option key={conn.connectionId} value={conn.connectionId}>
                    {conn.name || conn.platformId} ({conn.status})
                  </option>
                ))
              ) : (
                <option value="" disabled>
                  No connections found — add one in Connections
                </option>
              )}
            </select>
            <p className="text-xs text-muted-foreground">
              This helper requires a CRM connection to access contact data
            </p>
          </div>
        )}
      </div>

      {/* Dynamic Config Based on Helper Type */}
      <div className="rounded-lg border bg-card p-5 space-y-4">
        <h3 className="font-semibold">Configuration</h3>
        <HelperTypeConfig helper={helper} configRef={configRef} />
      </div>

      {/* Actions */}
      <div className="flex items-center justify-between rounded-lg border bg-card p-5">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Info className="h-4 w-4" />
          <span>You can edit configuration after creation</span>
        </div>
        <div className="flex gap-3">
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button
            onClick={onSubmit}
            disabled={isSubmitting || !helperName.trim()}
          >
            {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
            {isSubmitting ? 'Creating...' : 'Create Helper'}
          </Button>
        </div>
      </div>
    </div>
  )
}

function HelperTypeConfig({
  helper,
  configRef,
}: {
  helper: HelperDefinition
  configRef: React.MutableRefObject<Record<string, unknown>>
}) {
  const inputClass =
    'h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring'
  const selectClass =
    'h-10 w-full rounded-md border border-input bg-background px-3 text-sm'

  const setConfig = (key: string, value: unknown) => {
    configRef.current = { ...configRef.current, [key]: value }
  }

  switch (helper.id) {
    case 'tag_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Action</label>
            <select
              className={selectClass}
              defaultValue="apply"
              onChange={(e) => setConfig('action', e.target.value)}
            >
              <option value="apply">Apply Tags</option>
              <option value="remove">Remove Tags</option>
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Tag IDs</label>
            <input
              type="text"
              placeholder="Enter tag IDs separated by commas..."
              className={inputClass}
              onChange={(e) => setConfig('tag_ids', e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Comma-separated list of tag IDs to apply or remove
            </p>
          </div>
        </div>
      )

    case 'copy_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Source Field</label>
            <input
              type="text"
              placeholder="e.g., first_name, email, custom_field_1"
              className={inputClass}
              onChange={(e) => setConfig('source_field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Target Field</label>
            <input
              type="text"
              placeholder="e.g., custom_field_2"
              className={inputClass}
              onChange={(e) => setConfig('target_field', e.target.value)}
            />
          </div>
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="overwrite"
              defaultChecked
              className="rounded"
              onChange={(e) => setConfig('overwrite', e.target.checked)}
            />
            <label htmlFor="overwrite" className="text-sm">
              Overwrite existing target value
            </label>
          </div>
        </div>
      )

    case 'format_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Field</label>
            <input
              type="text"
              placeholder="Field key to format..."
              className={inputClass}
              onChange={(e) => setConfig('field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Format</label>
            <select
              className={selectClass}
              defaultValue="uppercase"
              onChange={(e) => setConfig('format', e.target.value)}
            >
              <option value="uppercase">UPPERCASE</option>
              <option value="lowercase">lowercase</option>
              <option value="title_case">Title Case</option>
              <option value="trim">Trim Whitespace</option>
            </select>
          </div>
        </div>
      )

    case 'math_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Field</label>
            <input
              type="text"
              placeholder="Field key to perform math on..."
              className={inputClass}
              onChange={(e) => setConfig('field', e.target.value)}
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Operation</label>
              <select
                className={selectClass}
                defaultValue="add"
                onChange={(e) => setConfig('operation', e.target.value)}
              >
                <option value="add">Add (+)</option>
                <option value="subtract">Subtract (-)</option>
                <option value="multiply">Multiply (x)</option>
                <option value="divide">Divide (/)</option>
                <option value="round">Round</option>
                <option value="abs">Absolute Value</option>
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Value</label>
              <input
                type="number"
                placeholder="0"
                className={inputClass}
                onChange={(e) => setConfig('value', Number(e.target.value))}
              />
            </div>
          </div>
        </div>
      )

    case 'date_calc':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Operation</label>
            <select
              className={selectClass}
              defaultValue="add_days"
              onChange={(e) => setConfig('operation', e.target.value)}
            >
              <option value="add_days">Add Days</option>
              <option value="subtract_days">Subtract Days</option>
              <option value="add_months">Add Months</option>
              <option value="subtract_months">Subtract Months</option>
              <option value="diff_days">Difference in Days</option>
              <option value="set_now">Set to Now</option>
              <option value="format">Format Date</option>
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Date Field</label>
            <input
              type="text"
              placeholder="Field containing the date..."
              className={inputClass}
              onChange={(e) => setConfig('date_field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Amount</label>
            <input
              type="number"
              placeholder="Number of days/months/years"
              className={inputClass}
              onChange={(e) => setConfig('amount', Number(e.target.value))}
            />
          </div>
        </div>
      )

    case 'notify_me':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Channel</label>
            <select
              className={selectClass}
              defaultValue="email"
              onChange={(e) => setConfig('channel', e.target.value)}
            >
              <option value="email">Email</option>
              <option value="slack">Slack</option>
              <option value="webhook">Webhook</option>
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Subject</label>
            <input
              type="text"
              placeholder="Notification subject (supports {{field_name}} merge)"
              className={inputClass}
              onChange={(e) => setConfig('subject', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Message</label>
            <textarea
              rows={3}
              placeholder="Notification message (supports {{field_name}} merge)..."
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              onChange={(e) => setConfig('message', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Recipient</label>
            <input
              type="text"
              placeholder="Email address, Slack channel, or webhook URL"
              className={inputClass}
              onChange={(e) => setConfig('recipient', e.target.value)}
            />
          </div>
        </div>
      )

    case 'score_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Target Field</label>
            <input
              type="text"
              placeholder="Field to store the calculated score..."
              className={inputClass}
              onChange={(e) => setConfig('target_field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Scoring Rules</label>
            <p className="text-xs text-muted-foreground">
              Add rules to calculate the score based on tag presence
            </p>
            <div className="rounded-md border p-3 space-y-3">
              <div className="flex items-center gap-3">
                <select className="h-9 flex-1 rounded-md border border-input bg-background px-2 text-sm">
                  <option value="has_tag">Has Tag</option>
                  <option value="no_tag">Does Not Have Tag</option>
                </select>
                <input
                  type="text"
                  placeholder="Tag ID"
                  className="h-9 w-32 rounded-md border border-input bg-background px-2 text-sm"
                />
                <input
                  type="number"
                  placeholder="Points"
                  className="h-9 w-20 rounded-md border border-input bg-background px-2 text-sm"
                />
              </div>
            </div>
            <button type="button" className="text-xs text-primary hover:underline">
              + Add Rule
            </button>
          </div>
        </div>
      )

    case 'clear_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Fields to Clear</label>
            <input
              type="text"
              placeholder="Field keys separated by commas (e.g., custom_1, custom_2)"
              className={inputClass}
              onChange={(e) => setConfig('fields', e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Comma-separated list of field keys to clear to empty
            </p>
          </div>
        </div>
      )

    case 'merge_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Source Fields</label>
            <input
              type="text"
              placeholder="Field keys separated by commas (e.g., first_name, last_name)"
              className={inputClass}
              onChange={(e) => setConfig('source_fields', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Target Field</label>
            <input
              type="text"
              placeholder="Field to store the merged value"
              className={inputClass}
              onChange={(e) => setConfig('target_field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Separator</label>
            <input
              type="text"
              placeholder="e.g., a space ' ' or comma ','"
              defaultValue=" "
              className={inputClass}
              onChange={(e) => setConfig('separator', e.target.value)}
            />
          </div>
        </div>
      )

    case 'text_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Field</label>
            <input
              type="text"
              placeholder="Field key to manipulate..."
              className={inputClass}
              onChange={(e) => setConfig('field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Operation</label>
            <select
              className={selectClass}
              defaultValue="prepend"
              onChange={(e) => setConfig('operation', e.target.value)}
            >
              <option value="prepend">Prepend Text</option>
              <option value="append">Append Text</option>
              <option value="replace">Find & Replace</option>
              <option value="extract">Extract (regex)</option>
              <option value="truncate">Truncate</option>
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Value</label>
            <input
              type="text"
              placeholder="Text to prepend/append, or search pattern..."
              className={inputClass}
              onChange={(e) => setConfig('value', e.target.value)}
            />
          </div>
        </div>
      )

    case 'found_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Field to Check</label>
            <input
              type="text"
              placeholder="Field key to check if populated..."
              className={inputClass}
              onChange={(e) => setConfig('field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Found Tag ID</label>
            <input
              type="text"
              placeholder="Tag to apply if field is populated"
              className={inputClass}
              onChange={(e) => setConfig('found_tag_id', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Not Found Tag ID</label>
            <input
              type="text"
              placeholder="Tag to apply if field is empty (optional)"
              className={inputClass}
              onChange={(e) => setConfig('not_found_tag_id', e.target.value)}
            />
          </div>
        </div>
      )

    case 'trigger_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Automation Type</label>
            <select
              className={selectClass}
              defaultValue="campaign"
              onChange={(e) => setConfig('automation_type', e.target.value)}
            >
              <option value="campaign">Campaign Sequence</option>
              <option value="workflow">Workflow</option>
              <option value="goal">Campaign Goal</option>
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Automation ID</label>
            <input
              type="text"
              placeholder="ID of the campaign, workflow, or goal"
              className={inputClass}
              onChange={(e) => setConfig('automation_id', e.target.value)}
            />
          </div>
        </div>
      )

    case 'hook_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Webhook URL</label>
            <input
              type="url"
              placeholder="https://example.com/webhook"
              className={inputClass}
              onChange={(e) => setConfig('url', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">HTTP Method</label>
            <select
              className={selectClass}
              defaultValue="POST"
              onChange={(e) => setConfig('method', e.target.value)}
            >
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="PATCH">PATCH</option>
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Include Fields</label>
            <input
              type="text"
              placeholder="Fields to send (comma-separated, or 'all')"
              defaultValue="all"
              className={inputClass}
              onChange={(e) => setConfig('fields', e.target.value)}
            />
          </div>
        </div>
      )

    case 'slack_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Channel</label>
            <input
              type="text"
              placeholder="#general or @username"
              className={inputClass}
              onChange={(e) => setConfig('channel', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Message Template</label>
            <textarea
              rows={3}
              placeholder="Message template (supports {{field_name}} merge)..."
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              onChange={(e) => setConfig('message', e.target.value)}
            />
          </div>
        </div>
      )

    case 'google_sheet_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Spreadsheet ID</label>
            <input
              type="text"
              placeholder="Google Sheets spreadsheet ID"
              className={inputClass}
              onChange={(e) => setConfig('spreadsheet_id', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Sheet Name</label>
            <input
              type="text"
              placeholder="e.g., Sheet1"
              defaultValue="Sheet1"
              className={inputClass}
              onChange={(e) => setConfig('sheet_name', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Fields to Sync</label>
            <input
              type="text"
              placeholder="Comma-separated field keys (or 'all')"
              defaultValue="all"
              className={inputClass}
              onChange={(e) => setConfig('fields', e.target.value)}
            />
          </div>
        </div>
      )

    case 'assign_it':
    case 'own_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Assignment Mode</label>
            <select
              className={selectClass}
              defaultValue="specific"
              onChange={(e) => setConfig('mode', e.target.value)}
            >
              <option value="specific">Specific Owner</option>
              <option value="round_robin">Round Robin</option>
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Owner ID(s)</label>
            <input
              type="text"
              placeholder="Owner user ID (or comma-separated for round-robin)"
              className={inputClass}
              onChange={(e) => setConfig('owner_ids', e.target.value)}
            />
          </div>
        </div>
      )

    case 'split_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Group A Tag ID</label>
            <input
              type="text"
              placeholder="Tag to apply for group A"
              className={inputClass}
              onChange={(e) => setConfig('group_a_tag', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Group B Tag ID</label>
            <input
              type="text"
              placeholder="Tag to apply for group B"
              className={inputClass}
              onChange={(e) => setConfig('group_b_tag', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Split Ratio</label>
            <select
              className={selectClass}
              defaultValue="50"
              onChange={(e) => setConfig('ratio', Number(e.target.value))}
            >
              <option value="50">50/50</option>
              <option value="60">60/40</option>
              <option value="70">70/30</option>
              <option value="80">80/20</option>
              <option value="90">90/10</option>
            </select>
          </div>
        </div>
      )

    case 'chain_it':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Helper IDs to Chain</label>
            <textarea
              rows={3}
              placeholder="Enter helper IDs, one per line, in execution order..."
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              onChange={(e) =>
                setConfig(
                  'helper_ids',
                  e.target.value.split('\n').filter(Boolean)
                )
              }
            />
            <p className="text-xs text-muted-foreground">
              Helpers will execute in sequence. If one fails, the chain stops.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="stop_on_error"
              defaultChecked
              className="rounded"
              onChange={(e) => setConfig('stop_on_error', e.target.checked)}
            />
            <label htmlFor="stop_on_error" className="text-sm">
              Stop chain on first error
            </label>
          </div>
        </div>
      )

    case 'route_it_by_custom':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Custom Field Name</label>
            <input
              type="text"
              placeholder="e.g., ProductInterest, LeadType"
              className={inputClass}
              onChange={(e) => setConfig('field_name', e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              The custom field to check for routing decisions
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Value Routes (JSON)</label>
            <textarea
              rows={5}
              placeholder={'{\n  "premium": "https://example.com/premium",\n  "standard": "https://example.com/standard"\n}'}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              onChange={(e) => {
                try {
                  const parsed = JSON.parse(e.target.value)
                  setConfig('value_routes', parsed)
                } catch {
                  // Invalid JSON, ignore
                }
              }}
            />
            <p className="text-xs text-muted-foreground">
              Map of field values to redirect URLs
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Fallback URL</label>
            <input
              type="url"
              placeholder="https://example.com/default"
              className={inputClass}
              onChange={(e) => setConfig('fallback_url', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Save URL To Field (Optional)</label>
            <input
              type="text"
              placeholder="e.g., RedirectURL"
              className={inputClass}
              onChange={(e) => setConfig('save_to_field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Apply Tag ID (Optional)</label>
            <input
              type="text"
              placeholder="Tag ID to apply when routing"
              className={inputClass}
              onChange={(e) => setConfig('apply_tag', e.target.value)}
            />
          </div>
        </div>
      )

    case 'route_it_by_day':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Day Routes (JSON)</label>
            <textarea
              rows={7}
              placeholder={'{\n  "Monday": "https://example.com/monday",\n  "Tuesday": "https://example.com/tuesday",\n  "Wednesday": "https://example.com/wednesday"\n}'}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              onChange={(e) => {
                try {
                  const parsed = JSON.parse(e.target.value)
                  setConfig('day_routes', parsed)
                } catch {
                  // Invalid JSON, ignore
                }
              }}
            />
            <p className="text-xs text-muted-foreground">
              Map of day names to redirect URLs (e.g., Monday, Tuesday)
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Timezone</label>
            <input
              type="text"
              placeholder="America/New_York"
              defaultValue="UTC"
              className={inputClass}
              onChange={(e) => setConfig('timezone', e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              IANA timezone name (e.g., America/New_York, Europe/London)
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Fallback URL</label>
            <input
              type="url"
              placeholder="https://example.com/default"
              className={inputClass}
              onChange={(e) => setConfig('fallback_url', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Save URL To Field (Optional)</label>
            <input
              type="text"
              placeholder="e.g., RedirectURL"
              className={inputClass}
              onChange={(e) => setConfig('save_to_field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Apply Tag ID (Optional)</label>
            <input
              type="text"
              placeholder="Tag ID to apply when routing"
              className={inputClass}
              onChange={(e) => setConfig('apply_tag', e.target.value)}
            />
          </div>
        </div>
      )

    case 'route_it_by_time':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Time Routes (JSON Array)</label>
            <textarea
              rows={8}
              placeholder={'[\n  {\n    "start_time": "09:00",\n    "end_time": "17:00",\n    "url": "https://example.com/business-hours",\n    "label": "Business Hours"\n  }\n]'}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              onChange={(e) => {
                try {
                  const parsed = JSON.parse(e.target.value)
                  setConfig('time_routes', parsed)
                } catch {
                  // Invalid JSON, ignore
                }
              }}
            />
            <p className="text-xs text-muted-foreground">
              Array of time ranges with start_time, end_time (HH:MM format, 24-hour), url, and optional label
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Timezone</label>
            <input
              type="text"
              placeholder="America/New_York"
              defaultValue="UTC"
              className={inputClass}
              onChange={(e) => setConfig('timezone', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Fallback URL</label>
            <input
              type="url"
              placeholder="https://example.com/after-hours"
              className={inputClass}
              onChange={(e) => setConfig('fallback_url', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Save URL To Field (Optional)</label>
            <input
              type="text"
              placeholder="e.g., RedirectURL"
              className={inputClass}
              onChange={(e) => setConfig('save_to_field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Apply Tag ID (Optional)</label>
            <input
              type="text"
              placeholder="Tag ID to apply when routing"
              className={inputClass}
              onChange={(e) => setConfig('apply_tag', e.target.value)}
            />
          </div>
        </div>
      )

    case 'route_it_geo':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">IP Address</label>
            <input
              type="text"
              placeholder="e.g., {{ContactIP}} or 8.8.8.8"
              className={inputClass}
              onChange={(e) => setConfig('ip_address', e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              IP address to geolocate (supports merge fields)
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Country Routes (JSON, Optional)</label>
            <textarea
              rows={4}
              placeholder={'{\n  "US": "https://example.com/us",\n  "CA": "https://example.com/canada"\n}'}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              onChange={(e) => {
                try {
                  const parsed = JSON.parse(e.target.value)
                  setConfig('country_routes', parsed)
                } catch {
                  // Invalid JSON, ignore
                }
              }}
            />
            <p className="text-xs text-muted-foreground">
              Map of country codes (e.g., US, CA, GB) to redirect URLs
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Region Routes (JSON, Optional)</label>
            <textarea
              rows={4}
              placeholder={'{\n  "California": "https://example.com/ca",\n  "Texas": "https://example.com/tx"\n}'}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              onChange={(e) => {
                try {
                  const parsed = JSON.parse(e.target.value)
                  setConfig('region_routes', parsed)
                } catch {
                  // Invalid JSON, ignore
                }
              }}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Fallback URL</label>
            <input
              type="url"
              placeholder="https://example.com/default"
              className={inputClass}
              onChange={(e) => setConfig('fallback_url', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Save URL To Field (Optional)</label>
            <input
              type="text"
              placeholder="e.g., RedirectURL"
              className={inputClass}
              onChange={(e) => setConfig('save_to_field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Save Location To Field (Optional)</label>
            <input
              type="text"
              placeholder="e.g., Location"
              className={inputClass}
              onChange={(e) => setConfig('save_location_to', e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Save the formatted location string (e.g., "San Francisco, California, USA")
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Apply Tag ID (Optional)</label>
            <input
              type="text"
              placeholder="Tag ID to apply when routing"
              className={inputClass}
              onChange={(e) => setConfig('apply_tag', e.target.value)}
            />
          </div>
        </div>
      )

    case 'route_it_score':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Score Field</label>
            <input
              type="text"
              placeholder="Score"
              defaultValue="Score"
              className={inputClass}
              onChange={(e) => setConfig('score_field', e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Contact field containing the lead score (numeric)
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Score Ranges (JSON Array)</label>
            <textarea
              rows={10}
              placeholder={'[\n  {\n    "label": "Hot Lead",\n    "min_score": 80,\n    "redirect_url": "https://example.com/hot-leads"\n  },\n  {\n    "label": "Warm Lead",\n    "min_score": 50,\n    "max_score": 79,\n    "redirect_url": "https://example.com/warm-leads"\n  }\n]'}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              onChange={(e) => {
                try {
                  const parsed = JSON.parse(e.target.value)
                  setConfig('score_ranges', parsed)
                } catch {
                  // Invalid JSON, ignore
                }
              }}
            />
            <p className="text-xs text-muted-foreground">
              Array of score ranges with optional label, min_score, max_score, and redirect_url
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Fallback URL</label>
            <input
              type="url"
              placeholder="https://example.com/default"
              className={inputClass}
              onChange={(e) => setConfig('fallback_url', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Save URL To Field (Optional)</label>
            <input
              type="text"
              placeholder="e.g., RedirectURL"
              className={inputClass}
              onChange={(e) => setConfig('save_to_field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Apply Tag ID (Optional)</label>
            <input
              type="text"
              placeholder="Tag ID to apply when routing"
              className={inputClass}
              onChange={(e) => setConfig('apply_tag', e.target.value)}
            />
          </div>
        </div>
      )

    case 'route_it_source':
      return (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Source Field</label>
            <input
              type="text"
              placeholder="LeadSource"
              defaultValue="LeadSource"
              className={inputClass}
              onChange={(e) => setConfig('source_field', e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Contact field containing the traffic source (e.g., LeadSource, utm_source)
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Source Routes (JSON)</label>
            <textarea
              rows={6}
              placeholder={'{\n  "google": "https://example.com/google",\n  "facebook": "https://example.com/facebook",\n  "direct": "https://example.com/direct"\n}'}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              onChange={(e) => {
                try {
                  const parsed = JSON.parse(e.target.value)
                  setConfig('source_routes', parsed)
                } catch {
                  // Invalid JSON, ignore
                }
              }}
            />
            <p className="text-xs text-muted-foreground">
              Map of source values to redirect URLs
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Match Mode</label>
            <select
              className={selectClass}
              defaultValue="exact"
              onChange={(e) => setConfig('match_mode', e.target.value)}
            >
              <option value="exact">Exact Match</option>
              <option value="contains">Contains</option>
              <option value="starts_with">Starts With</option>
            </select>
          </div>
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="case_sensitive"
              className="rounded"
              onChange={(e) => setConfig('case_sensitive', e.target.checked)}
            />
            <label htmlFor="case_sensitive" className="text-sm">
              Case sensitive matching
            </label>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Fallback URL</label>
            <input
              type="url"
              placeholder="https://example.com/default"
              className={inputClass}
              onChange={(e) => setConfig('fallback_url', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Save URL To Field (Optional)</label>
            <input
              type="text"
              placeholder="e.g., RedirectURL"
              className={inputClass}
              onChange={(e) => setConfig('save_to_field', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Apply Tag ID (Optional)</label>
            <input
              type="text"
              placeholder="Tag ID to apply when routing"
              className={inputClass}
              onChange={(e) => setConfig('apply_tag', e.target.value)}
            />
          </div>
        </div>
      )

    default:
      return <DynamicSchemaForm helper={helper} configRef={configRef} />
  }
}

/**
 * Fallback form that generates inputs dynamically based on the helper type.
 * For helpers without a custom config form, this provides a structured
 * set of common fields based on the helper's category.
 */
function DynamicSchemaForm({
  helper,
  configRef,
}: {
  helper: HelperDefinition
  configRef: React.MutableRefObject<Record<string, unknown>>
}) {
  const inputClass =
    'h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring'

  const setConfig = (key: string, value: unknown) => {
    configRef.current = { ...configRef.current, [key]: value }
  }

  // Generate common config fields based on category
  const commonFields: Array<{
    key: string
    label: string
    placeholder: string
    type: 'text' | 'number' | 'textarea'
  }> = []

  if (helper.category === 'contact' || helper.category === 'data') {
    commonFields.push(
      { key: 'field', label: 'Field', placeholder: 'CRM field key (e.g., custom_field_1)', type: 'text' },
      { key: 'target_field', label: 'Target Field', placeholder: 'Target field key (optional)', type: 'text' }
    )
  }

  if (helper.category === 'tagging') {
    commonFields.push(
      { key: 'tag_ids', label: 'Tag IDs', placeholder: 'Comma-separated tag IDs', type: 'text' }
    )
  }

  if (helper.category === 'automation') {
    commonFields.push(
      { key: 'automation_id', label: 'Automation ID', placeholder: 'ID of the automation to trigger', type: 'text' }
    )
  }

  if (helper.category === 'integration') {
    commonFields.push(
      { key: 'url', label: 'URL / Endpoint', placeholder: 'Service URL or endpoint', type: 'text' }
    )
  }

  if (helper.category === 'notification') {
    commonFields.push(
      { key: 'recipient', label: 'Recipient', placeholder: 'Email, channel, or destination', type: 'text' },
      { key: 'message', label: 'Message Template', placeholder: 'Message (supports {{field_name}} merge fields)', type: 'textarea' }
    )
  }

  if (helper.category === 'analytics') {
    commonFields.push(
      { key: 'target_field', label: 'Target Field', placeholder: 'Field to store the result', type: 'text' }
    )
  }

  if (commonFields.length === 0) {
    return (
      <div className="rounded-md border border-dashed p-6 text-center">
        <p className="text-sm text-muted-foreground">
          This helper can be configured with JSON after creation.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <p className="text-xs text-muted-foreground">
        Common configuration for {helper.category} helpers. Additional settings can be edited as JSON after creation.
      </p>
      {commonFields.map((field) => (
        <div key={field.key} className="space-y-2">
          <label className="text-sm font-medium">{field.label}</label>
          {field.type === 'textarea' ? (
            <textarea
              rows={3}
              placeholder={field.placeholder}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              onChange={(e) => setConfig(field.key, e.target.value)}
            />
          ) : (
            <input
              type={field.type}
              placeholder={field.placeholder}
              className={inputClass}
              onChange={(e) =>
                setConfig(
                  field.key,
                  field.type === 'number' ? Number(e.target.value) : e.target.value
                )
              }
            />
          )}
        </div>
      ))}
    </div>
  )
}
