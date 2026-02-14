'use client'

import { Moon, Sun } from 'lucide-react'
import { useTheme } from 'next-themes'
import { useEffect, useState } from 'react'

export function ThemeToggle({ variant = 'default' }: { variant?: 'default' | 'sidebar' }) {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  const isSidebar = variant === 'sidebar'
  const hoverClass = isSidebar ? 'hover:bg-sidebar-accent' : 'hover:bg-accent'
  const iconClass = isSidebar ? 'text-sidebar-muted-foreground' : 'text-muted-foreground'

  if (!mounted) {
    return (
      <button className={`rounded-md p-2 ${hoverClass}`} aria-label="Toggle theme">
        <Sun className={`h-4 w-4 ${iconClass}`} />
      </button>
    )
  }

  return (
    <button
      onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
      className={`rounded-md p-2 ${hoverClass}`}
      aria-label="Toggle theme"
      title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
    >
      {theme === 'dark' ? (
        <Sun className={`h-4 w-4 ${iconClass}`} />
      ) : (
        <Moon className={`h-4 w-4 ${iconClass}`} />
      )}
    </button>
  )
}
