'use client'

import { Database } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useDataExplorerStore } from '@/lib/stores/data-explorer-store'
import { PlatformOverview } from './platform-overview'
import { ConnectionOverview } from './connection-overview'
import { RecordDetail } from './record-detail'
import { NLQueryBar } from './nl-query-bar'
import { DataTable } from './data-table'

interface ContentPreviewProps {
  className?: string
}

// ---------------------------------------------------------------------------
// WelcomeView — shown when no data source is selected
// ---------------------------------------------------------------------------

function WelcomeView() {
  return (
    <div className="flex flex-col items-center justify-center h-full gap-4 text-center px-6">
      <div className="rounded-full bg-muted p-4">
        <Database className="h-10 w-10 text-muted-foreground" />
      </div>
      <div className="space-y-2 max-w-md">
        <h2 className="text-xl font-semibold text-foreground">
          Select a data source
        </h2>
        <p className="text-sm text-muted-foreground leading-relaxed">
          Choose a platform, connection, or object type from the sidebar to
          start exploring your CRM data. You can browse records, run natural
          language queries, and export results.
        </p>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// ObjectTypeView — data table with NL query bar
// ---------------------------------------------------------------------------

function ObjectTypeView() {
  return (
    <div className="flex flex-col h-full gap-4 p-6">
      <NLQueryBar />
      <div className="flex-1 min-h-0">
        <DataTable />
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// ContentPreview — router component
// ---------------------------------------------------------------------------

export function ContentPreview({ className }: ContentPreviewProps) {
  const { selection } = useDataExplorerStore()

  const renderContent = () => {
    switch (selection.level) {
      case 'platform':
        return <PlatformOverview />
      case 'connection':
        return <ConnectionOverview />
      case 'objectType':
        return <ObjectTypeView />
      case 'record':
        return <RecordDetail />
      case 'none':
      default:
        return <WelcomeView />
    }
  }

  return (
    <div className={cn('flex flex-col h-full min-h-0', className)}>
      {renderContent()}
    </div>
  )
}
