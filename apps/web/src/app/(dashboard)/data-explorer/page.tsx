'use client'

import { useCallback, useRef, useState } from 'react'
import { GripVertical, PanelLeftClose, PanelLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useDataExplorerStore } from '@/lib/stores/data-explorer-store'
import { HierarchicalNav } from '@/components/data-explorer/hierarchical-nav'
import { ContentPreview } from '@/components/data-explorer/content-preview'

const MIN_SIDEBAR_WIDTH = 200
const MAX_SIDEBAR_WIDTH = 600

export default function DataExplorerPage() {
  const {
    sidebarOpen,
    sidebarWidth,
    setSidebarOpen,
    setSidebarWidth,
  } = useDataExplorerStore()

  const [isDragging, setIsDragging] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  // -- Resize handler --------------------------------------------------------

  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault()
      setIsDragging(true)

      const startX = e.clientX
      const startWidth = sidebarWidth

      const handleMouseMove = (moveEvent: MouseEvent) => {
        const containerLeft = containerRef.current?.getBoundingClientRect().left ?? 0
        const newWidth = Math.min(
          MAX_SIDEBAR_WIDTH,
          Math.max(MIN_SIDEBAR_WIDTH, startWidth + (moveEvent.clientX - startX))
        )
        // Only update if the mouse is within the container bounds
        if (moveEvent.clientX > containerLeft + MIN_SIDEBAR_WIDTH) {
          setSidebarWidth(newWidth)
        }
      }

      const handleMouseUp = () => {
        setIsDragging(false)
        document.removeEventListener('mousemove', handleMouseMove)
        document.removeEventListener('mouseup', handleMouseUp)
      }

      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
    },
    [sidebarWidth, setSidebarWidth]
  )

  return (
    <div
      ref={containerRef}
      className="-m-6 flex h-[calc(100vh-3.5rem)] overflow-hidden"
    >
      {/* Sidebar */}
      {sidebarOpen && (
        <div
          className="flex shrink-0 border-r bg-card"
          style={{ width: sidebarWidth }}
        >
          {/* Nav content */}
          <div className="flex flex-1 flex-col min-w-0">
            {/* Sidebar header */}
            <div className="flex h-12 items-center justify-between border-b px-3">
              <h2 className="text-sm font-semibold">Data Explorer</h2>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                onClick={() => setSidebarOpen(false)}
                title="Close sidebar"
              >
                <PanelLeftClose className="h-4 w-4" />
              </Button>
            </div>

            {/* Tree nav */}
            <div className="flex-1 min-h-0 overflow-hidden">
              <HierarchicalNav />
            </div>
          </div>

          {/* Resize handle */}
          <div
            className={cn(
              'flex w-1.5 cursor-col-resize items-center justify-center transition-colors hover:bg-primary/10',
              isDragging && 'bg-primary/20'
            )}
            onMouseDown={handleMouseDown}
          >
            <GripVertical className="h-4 w-4 text-muted-foreground/50" />
          </div>
        </div>
      )}

      {/* Main content area */}
      <div className="flex flex-1 flex-col min-w-0">
        {/* Toggle sidebar button (when closed) */}
        {!sidebarOpen && (
          <div className="flex h-12 items-center border-b px-3">
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => setSidebarOpen(true)}
              title="Open sidebar"
            >
              <PanelLeft className="h-4 w-4" />
            </Button>
            <span className="ml-2 text-sm font-semibold">Data Explorer</span>
          </div>
        )}

        {/* Content preview (router for selection level) */}
        <div className="flex-1 min-h-0 overflow-auto">
          <ContentPreview />
        </div>
      </div>
    </div>
  )
}
