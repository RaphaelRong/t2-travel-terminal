import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { ThemeToggle } from './ThemeToggle'

const baseNavItems = [
  { to: '/', label: 'Globe' },
  { to: '/playground', label: 'Playground' },
  { to: '/projects', label: 'Projects' },
  { to: '/members', label: 'Members' },
  { to: '/plans', label: 'Plans' },
  { to: '/profile', label: 'Profile' },
]

export function Layout() {
  const logout = useAuthStore((s) => s.logout)
  const role = useAuthStore((s) => s.role)
  const navigate = useNavigate()
  const location = useLocation()

  const navItems = role === 'superadmin'
    ? [...baseNavItems, { to: '/admin/plans', label: 'Admin' }]
    : baseNavItems

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const isActive = (path: string) => location.pathname === path

  return (
    <div className="flex min-h-screen flex-col bg-obsidian-base">
      <header className="border-b border-obsidian-border-dim bg-obsidian-raised">
        <div className="mx-auto flex h-12 max-w-7xl items-center justify-between px-4">
          <div className="flex items-center gap-6">
            <Link to="/" className="flex items-center gap-2 font-mono text-sm font-semibold tracking-tight text-obsidian-accent">
              <span className="inline-block h-2 w-2 rounded-full bg-obsidian-positive animate-pulse" />
              T2 — Travel Terminal
            </Link>
            <nav className="hidden items-center gap-1 md:flex">
              {navItems.map((item) => (
                <Link
                  key={item.to}
                  to={item.to}
                  className={`rounded px-3 py-1.5 font-mono text-xs uppercase tracking-wide transition-colors ${
                    isActive(item.to)
                      ? 'bg-obsidian-highlight text-obsidian-text-primary'
                      : 'text-obsidian-text-secondary hover:bg-obsidian-highlight hover:text-obsidian-text-primary'
                  }`}
                >
                  {item.label}
                </Link>
              ))}
            </nav>
          </div>
          <div className="flex items-center gap-2">
            <ThemeToggle />
            <button
              onClick={handleLogout}
              className="rounded border border-obsidian-border-dim bg-obsidian-surface px-3 py-1.5 font-mono text-xs text-obsidian-text-secondary hover:border-obsidian-border-med hover:text-obsidian-text-primary"
            >
              Logout
            </button>
          </div>
        </div>
      </header>

      <div className="border-b border-obsidian-border-dim bg-obsidian-surface">
        <div className="mx-auto flex h-8 max-w-7xl items-center justify-between px-4">
          <div className="flex items-center gap-2 font-mono text-xs text-obsidian-text-tertiary">
            <span className="text-obsidian-accent">&gt;_</span>
            <span>Ready</span>
            <span className="text-obsidian-border-bright">|</span>
            <span>session:active</span>
          </div>
          <div className="font-mono text-xs text-obsidian-text-tertiary">
            {new Date().toLocaleTimeString()}
          </div>
        </div>
      </div>

      <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-6">
        <Outlet />
      </main>

      <footer className="border-t border-obsidian-border-dim bg-obsidian-raised py-2">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 font-mono text-[10px] uppercase tracking-wider text-obsidian-text-dim">
          <span>Status: Online</span>
          <span>API: /api</span>
        </div>
      </footer>
    </div>
  )
}
