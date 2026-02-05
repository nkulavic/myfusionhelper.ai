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
  getCategoryCounts,
  type HelperDefinition,
} from '@/lib/helpers-catalog'
import { useCreateHelper } from '@/lib/hooks/use-helpers'
import { useConnections } from '@/lib/hooks/use-connections'

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

  const categoryCounts = useMemo(() => getCategoryCounts(), [])

  const availableHelpers = useMemo(() => {
    return helpersCatalog
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
  }, [selectedCategory, searchQuery])

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
          const id = res.data?.id
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
        <button
          onClick={handleBack}
          className="rounded-md p-2 hover:bg-accent"
        >
          <ArrowLeft className="h-4 w-4" />
        </button>
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
            <input
              type="text"
              placeholder="Search available helpers..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="h-10 w-full rounded-md border border-input bg-background pl-10 pr-4 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
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
          <input
            type="text"
            value={helperName}
            onChange={(e) => setHelperName(e.target.value)}
            placeholder="Give your helper a name..."
            className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
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
                connections.map((conn) => (
                  <option key={conn.id} value={conn.id}>
                    {conn.name || conn.platform} ({conn.status})
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
          <button
            onClick={onCancel}
            className="rounded-md border border-input px-4 py-2 text-sm font-medium hover:bg-accent"
          >
            Cancel
          </button>
          <button
            onClick={onSubmit}
            disabled={isSubmitting || !helperName.trim()}
            className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
            {isSubmitting ? 'Creating...' : 'Create Helper'}
          </button>
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

    default:
      return (
        <div className="rounded-md border border-dashed p-6 text-center">
          <p className="text-sm text-muted-foreground">
            Configuration fields will be dynamically generated based on the helper schema.
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            This helper supports JSON configuration which can be edited after creation.
          </p>
        </div>
      )
  }
}
