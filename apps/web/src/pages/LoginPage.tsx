import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { api } from '../lib/api'
import { useAuthStore } from '../store/authStore'
import { ThemeToggle } from '../components/ThemeToggle'

export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const setToken = useAuthStore((s) => s.setToken)
  const setSuperAdmin = useAuthStore((s) => s.setSuperAdmin)
  const setRole = useAuthStore((s) => s.setRole)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await api.post('/auth/login', { email, password })
      const role = res.data.role ?? 'free_user'
      setToken(res.data.access_token)
      setSuperAdmin(res.data.is_superadmin === true)
      setRole(role)
      navigate(searchParams.get('redirect') || '/playground')
    } catch (err) {
      setError((err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-obsidian-base px-4">
      <div className="absolute right-4 top-4">
        <ThemeToggle />
      </div>
      <div className="w-full max-w-md border border-obsidian-border-dim bg-obsidian-surface p-8">
        <div className="mb-6 text-center">
          <h1 className="font-mono text-2xl font-bold tracking-tight text-obsidian-text-primary">
            <span className="text-obsidian-accent">&gt;</span> Login
          </h1>
          <p className="mt-1 font-mono text-xs text-obsidian-text-tertiary">T2 — Travel Terminal</p>
        </div>
        {error && (
          <p className="mb-4 border border-obsidian-negative-dim bg-obsidian-negative-dim/20 p-2 font-mono text-xs text-obsidian-negative">
            {error}
          </p>
        )}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-accent"
              required
            />
          </div>
          <div>
            <label className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
              Password
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-accent"
              required
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
          >
            {loading ? 'Authenticating...' : 'Login'}
          </button>
        </form>
        <p className="mt-4 text-center font-mono text-xs text-obsidian-text-tertiary">
          No account?{' '}
          <Link to="/register" className="text-obsidian-accent hover:underline">
            Register
          </Link>
        </p>
      </div>
    </div>
  )
}
