import { useTheme } from '../hooks/useTheme'

export function ThemeToggle() {
  const { theme, toggleTheme } = useTheme()

  return (
    <button
      onClick={toggleTheme}
      className="rounded border border-obsidian-border-dim bg-obsidian-surface px-3 py-1.5 font-mono text-xs text-obsidian-text-secondary transition-colors hover:border-obsidian-border-med hover:text-obsidian-text-primary"
      title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
    >
      {theme === 'dark' ? '☀' : '☾'}
    </button>
  )
}
