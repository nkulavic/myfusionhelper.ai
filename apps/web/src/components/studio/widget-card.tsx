'use client'

import { MoreVertical, Trash2, ArrowUp, ArrowDown, Maximize2, Minimize2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Widget, WidgetSize, DateRangePreset } from '@/lib/stores/studio-store'
import { WidgetRenderer } from './widget-renderer'
import { cn } from '@/lib/utils'

interface WidgetCardProps {
  widget: Widget
  dateRange: DateRangePreset
  onRemove: () => void
  onMoveUp: () => void
  onMoveDown: () => void
  onResize: (size: WidgetSize) => void
  isFirst: boolean
  isLast: boolean
}

const sizeClasses: Record<WidgetSize, string> = {
  sm: 'col-span-12 sm:col-span-6 lg:col-span-3',
  md: 'col-span-12 sm:col-span-6',
  lg: 'col-span-12 lg:col-span-8',
  full: 'col-span-12',
}

const sizeLabels: Record<WidgetSize, string> = {
  sm: 'Small (1/4)',
  md: 'Medium (1/2)',
  lg: 'Large (2/3)',
  full: 'Full Width',
}

export function WidgetCard({
  widget,
  dateRange,
  onRemove,
  onMoveUp,
  onMoveDown,
  onResize,
  isFirst,
  isLast,
}: WidgetCardProps) {
  return (
    <div className={cn('group relative', sizeClasses[widget.size])}>
      <div className="absolute right-2 top-2 z-10 opacity-0 transition-opacity group-hover:opacity-100">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-7 w-7">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={onMoveUp} disabled={isFirst}>
              <ArrowUp className="mr-2 h-4 w-4" />
              Move Up
            </DropdownMenuItem>
            <DropdownMenuItem onClick={onMoveDown} disabled={isLast}>
              <ArrowDown className="mr-2 h-4 w-4" />
              Move Down
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            {(Object.keys(sizeClasses) as WidgetSize[]).map((size) => (
              <DropdownMenuItem
                key={size}
                onClick={() => onResize(size)}
                className={widget.size === size ? 'bg-accent' : ''}
              >
                {size === 'sm' || size === 'md' ? (
                  <Minimize2 className="mr-2 h-4 w-4" />
                ) : (
                  <Maximize2 className="mr-2 h-4 w-4" />
                )}
                {sizeLabels[size]}
              </DropdownMenuItem>
            ))}
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={onRemove} className="text-destructive">
              <Trash2 className="mr-2 h-4 w-4" />
              Remove
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <WidgetRenderer widget={widget} dateRange={dateRange} />
    </div>
  )
}
