import { type LucideIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

interface EmptyStateProps {
  icon?: LucideIcon
  title: string
  description?: string
  action?: {
    label: string
    onClick: () => void
  }
  compact?: boolean
}

export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
  compact = false,
}: EmptyStateProps) {
  if (compact) {
    return (
      <div className="flex flex-col items-center justify-center py-8 text-center">
        {Icon && <Icon className="mb-2 h-8 w-8 text-muted-foreground" />}
        <p className="text-sm font-medium text-muted-foreground">{title}</p>
        {description && <p className="mt-1 text-xs text-muted-foreground">{description}</p>}
        {action && (
          <Button onClick={action.onClick} variant="outline" size="sm" className="mt-4">
            {action.label}
          </Button>
        )}
      </div>
    )
  }

  return (
    <Card>
      <CardContent className="flex min-h-[400px] flex-col items-center justify-center py-12 text-center">
        {Icon && <Icon className="mb-4 h-12 w-12 text-muted-foreground" />}
        <h3 className="text-lg font-semibold">{title}</h3>
        {description && <p className="mt-2 text-sm text-muted-foreground">{description}</p>}
        {action && (
          <Button onClick={action.onClick} className="mt-6">
            {action.label}
          </Button>
        )}
      </CardContent>
    </Card>
  )
}
