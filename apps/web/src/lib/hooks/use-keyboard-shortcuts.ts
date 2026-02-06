import { useEffect } from 'react'
import { useUIStore } from '@/lib/stores/ui-store'

/**
 * Global keyboard shortcuts for the dashboard.
 *
 * - Cmd+K / Ctrl+K: Toggle command palette
 * - Cmd+/ / Ctrl+/: Toggle AI chat panel
 * - /: Focus search input on list pages (when not in an input)
 */
export function useGlobalKeyboardShortcuts() {
  const { setCommandPaletteOpen, commandPaletteOpen, setAIChatOpen, aiChatOpen } = useUIStore()

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      const target = e.target as HTMLElement
      const isInInput =
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.tagName === 'SELECT' ||
        target.isContentEditable

      // Cmd+K / Ctrl+K: Toggle command palette
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setCommandPaletteOpen(!commandPaletteOpen)
        return
      }

      // Cmd+/ / Ctrl+/: Toggle AI chat
      if ((e.metaKey || e.ctrlKey) && e.key === '/') {
        e.preventDefault()
        setAIChatOpen(!aiChatOpen)
        return
      }

      // / : Focus search input (when not already in an input)
      if (e.key === '/' && !isInInput && !commandPaletteOpen) {
        const searchInput = document.querySelector<HTMLInputElement>(
          'input[type="text"][placeholder*="Search"], input[type="text"][placeholder*="search"]'
        )
        if (searchInput) {
          e.preventDefault()
          searchInput.focus()
          return
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [commandPaletteOpen, aiChatOpen, setCommandPaletteOpen, setAIChatOpen])
}
