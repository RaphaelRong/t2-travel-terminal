import { Link, Outlet, useLocation } from 'react-router-dom'

const tabs = [
  { to: '/admin/plans', label: 'Plans' },
  { to: '/admin/projects', label: 'Projects' },
  { to: '/admin/users', label: 'Users' },
]

export function AdminPage() {
  const location = useLocation()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-mono text-2xl font-bold tracking-tight text-obsidian-text-primary">
          <span className="text-obsidian-accent">&gt;</span> Admin
        </h1>
        <p className="mt-1 font-mono text-sm text-obsidian-text-secondary">
          System management for SuperAdmin
        </p>
      </div>

      <nav className="flex gap-2 border-b border-obsidian-border-dim">
        {tabs.map((tab) => (
          <Link
            key={tab.to}
            to={tab.to}
            className={`border-b-2 px-4 py-2 font-mono text-sm transition-colors ${
              location.pathname.startsWith(tab.to)
                ? 'border-obsidian-accent text-obsidian-text-primary'
                : 'border-transparent text-obsidian-text-secondary hover:text-obsidian-text-primary'
            }`}
          >
            {tab.label}
          </Link>
        ))}
      </nav>

      <Outlet />
    </div>
  )
}
