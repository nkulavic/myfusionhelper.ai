'use client'

import { useState, useCallback, useMemo } from 'react'
import { Sparkles, Search, X, Info } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { useDataExplorerStore } from '@/lib/stores/data-explorer-store'

const exampleQueriesByType: Record<string, string[]> = {
  contacts: [
    'Show VIP contacts',
    'Contacts with gmail emails',
    'Active contacts from referrals',
  ],
  deals: [
    'Deals over $10,000',
    'Deals closing this month',
    'Closed won deals',
  ],
  tags: [
    'Tags with more than 100 contacts',
    'Lifecycle tags',
  ],
}

const defaultExamples = ['Show all records', 'Recent records']

export function NLQueryBar() {
  const {
    selection,
    nlQuery,
    generatedDescription,
    setNLQuery,
    setGeneratedDescription,
  } = useDataExplorerStore()

  const [inputValue, setInputValue] = useState(nlQuery ?? '')

  const examples = useMemo(() => {
    const objectType = selection.objectType?.toLowerCase() ?? ''
    return exampleQueriesByType[objectType] ?? defaultExamples
  }, [selection.objectType])

  const handleSubmit = useCallback(() => {
    const trimmed = inputValue.trim()
    if (!trimmed) return
    setNLQuery(trimmed)
  }, [inputValue, setNLQuery])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === 'Enter') {
        e.preventDefault()
        handleSubmit()
      }
    },
    [handleSubmit]
  )

  const handleExampleClick = useCallback(
    (example: string) => {
      setInputValue(example)
      setNLQuery(example)
    },
    [setNLQuery]
  )

  const handleClearFilter = useCallback(() => {
    setInputValue('')
    setNLQuery(null)
    setGeneratedDescription(null)
  }, [setNLQuery, setGeneratedDescription])

  return (
    <div className="space-y-2">
      <Card className="p-4 space-y-3">
        {/* Input row */}
        <div className="flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-primary shrink-0" />
          <Input
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Ask about your data..."
            className="flex-1"
          />
          <Button
            size="sm"
            onClick={handleSubmit}
            disabled={!inputValue.trim()}
          >
            <Search className="h-4 w-4 mr-1" />
            Search
          </Button>
        </div>

        {/* Example chips */}
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-xs text-muted-foreground">Examples:</span>
          {examples.map((example) => (
            <Badge
              key={example}
              variant="secondary"
              className="cursor-pointer hover:bg-secondary/80 transition-colors text-xs"
              onClick={() => handleExampleClick(example)}
            >
              {example}
            </Badge>
          ))}
        </div>
      </Card>

      {/* Generated description banner */}
      {generatedDescription && (
        <Card
          className={cn(
            'p-3 flex items-start gap-3',
            'bg-blue-50 border-blue-200 dark:bg-blue-950/30 dark:border-blue-800'
          )}
        >
          <Info className="h-4 w-4 text-blue-600 dark:text-blue-400 shrink-0 mt-0.5" />
          <div className="flex-1 min-w-0">
            <p className="text-sm text-blue-800 dark:text-blue-300">
              <span className="font-medium">Applied:</span>{' '}
              &quot;{generatedDescription}&quot;
            </p>
          </div>
          <Button
            variant="ghost"
            size="sm"
            className="shrink-0 text-blue-600 hover:text-blue-800 hover:bg-blue-100 dark:text-blue-400 dark:hover:text-blue-200 dark:hover:bg-blue-900/50"
            onClick={handleClearFilter}
          >
            <X className="h-3 w-3 mr-1" />
            Clear NL Filter
          </Button>
        </Card>
      )}
    </div>
  )
}
