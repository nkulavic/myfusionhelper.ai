'use client'

import { useState } from 'react'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useStudioStore } from '@/lib/stores/studio-store'
import type { Dashboard, Widget, WidgetSize } from '@/lib/stores/studio-store'
import { useUpdateDashboard } from '@/lib/hooks/use-studio'
import { WidgetCard } from './widget-card'
import { DateRangeFilter } from './date-range-filter'
import { AddWidgetModal } from './add-widget-modal'

interface DashboardCanvasProps {
  dashboard: Dashboard
}

let _counter = 0
function uid(): string {
  _counter += 1
  return `wdg-${Date.now().toString(36)}-${_counter.toString(36)}-${Math.random().toString(36).slice(2, 8)}`
}

export function DashboardCanvas({ dashboard }: DashboardCanvasProps) {
  const [modalOpen, setModalOpen] = useState(false)
  const { activeDateRange, setDateRange } = useStudioStore()
  const updateDashboard = useUpdateDashboard()

  const sortedWidgets = [...dashboard.widgets].sort((a, b) => a.order - b.order)

  const saveWidgets = (newWidgets: Widget[]) => {
    updateDashboard.mutate({
      id: dashboard.id,
      data: { widgets: newWidgets },
    })
  }

  const handleAddWidget = (widget: Omit<Widget, 'id' | 'order'>) => {
    const newWidget: Widget = {
      ...widget,
      id: uid(),
      order: dashboard.widgets.length,
    }
    saveWidgets([...dashboard.widgets, newWidget])
  }

  const handleRemoveWidget = (widgetId: string) => {
    const filtered = dashboard.widgets
      .filter((w) => w.id !== widgetId)
      .map((w, i) => ({ ...w, order: i }))
    saveWidgets(filtered)
  }

  const handleMoveUp = (widgetId: string) => {
    const ids = sortedWidgets.map((w) => w.id)
    const idx = ids.indexOf(widgetId)
    if (idx <= 0) return
    ;[ids[idx - 1], ids[idx]] = [ids[idx], ids[idx - 1]]
    reorderByIds(ids)
  }

  const handleMoveDown = (widgetId: string) => {
    const ids = sortedWidgets.map((w) => w.id)
    const idx = ids.indexOf(widgetId)
    if (idx < 0 || idx >= ids.length - 1) return
    ;[ids[idx], ids[idx + 1]] = [ids[idx + 1], ids[idx]]
    reorderByIds(ids)
  }

  const reorderByIds = (widgetIds: string[]) => {
    const widgetMap = new Map(dashboard.widgets.map((w) => [w.id, w]))
    const reordered = widgetIds
      .map((id, i) => {
        const w = widgetMap.get(id)
        return w ? { ...w, order: i } : null
      })
      .filter(Boolean) as Widget[]
    saveWidgets(reordered)
  }

  const handleResize = (widgetId: string, size: WidgetSize) => {
    const updated = dashboard.widgets.map((w) =>
      w.id === widgetId ? { ...w, size } : w,
    )
    saveWidgets(updated)
  }

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">{dashboard.name}</h2>
          {dashboard.description && (
            <p className="text-sm text-muted-foreground">{dashboard.description}</p>
          )}
        </div>
        <div className="flex items-center gap-3">
          <DateRangeFilter value={activeDateRange} onChange={setDateRange} />
          <Button onClick={() => setModalOpen(true)} size="sm">
            <Plus className="mr-1.5 h-4 w-4" />
            Add Widget
          </Button>
        </div>
      </div>

      {/* Widget Grid */}
      {sortedWidgets.length > 0 ? (
        <div className="grid grid-cols-12 gap-4">
          {sortedWidgets.map((widget, i) => (
            <WidgetCard
              key={widget.id}
              widget={widget}
              dateRange={activeDateRange}
              onRemove={() => handleRemoveWidget(widget.id)}
              onMoveUp={() => handleMoveUp(widget.id)}
              onMoveDown={() => handleMoveDown(widget.id)}
              onResize={(size) => handleResize(widget.id, size)}
              isFirst={i === 0}
              isLast={i === sortedWidgets.length - 1}
            />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-20 text-center">
          <Plus className="mb-3 h-10 w-10 text-muted-foreground/40" />
          <h3 className="mb-1 font-semibold">No widgets yet</h3>
          <p className="mb-4 text-sm text-muted-foreground">
            Add your first chart or scorecard to start building this dashboard.
          </p>
          <Button onClick={() => setModalOpen(true)} size="sm">
            <Plus className="mr-1.5 h-4 w-4" />
            Add Widget
          </Button>
        </div>
      )}

      <AddWidgetModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onAdd={handleAddWidget}
      />
    </div>
  )
}
