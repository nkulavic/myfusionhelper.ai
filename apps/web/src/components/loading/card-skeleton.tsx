import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent, CardHeader } from '@/components/ui/card'

interface CardSkeletonProps {
  count?: number
  showHeader?: boolean
}

export function CardSkeleton({ count = 1, showHeader = true }: CardSkeletonProps) {
  return (
    <>
      {Array.from({ length: count }).map((_, i) => (
        <Card key={i}>
          {showHeader && (
            <CardHeader>
              <Skeleton className="h-5 w-[200px]" />
              <Skeleton className="h-4 w-[300px]" />
            </CardHeader>
          )}
          <CardContent>
            <div className="space-y-2">
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-3/4" />
            </div>
          </CardContent>
        </Card>
      ))}
    </>
  )
}
