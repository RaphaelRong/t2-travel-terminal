import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../lib/api'
import { ThemeToggle } from '../components/ThemeToggle'

export function RegisterPage() {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess('')
    setLoading(true)
    try {
      await api.post('/auth/register', { email, password, name })
      setSuccess('Registration successful. Please check your email to verify your account.')
      setTimeout(() => navigate('/login'), 3000)
    } catch (err) {
      setError((err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Registration failed')
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
            <span className="text-obsidian-accent">&gt;</span> Register
          </h1>
          <p className="mt-1 font-mono text-xs text-obsidian-text-tertiary">T2 — Travel Terminal</p>
        </div>
        {error && (
          <p className="mb-4 border border-obsidian-negative-dim bg-obsidian-negative-dim/20 p-2 font-mono text-xs text-obsidian-negative">
            {error}
          </p>
        )}
        {success && (
          <p className="mb-4 border border-obsidian-positive-dim bg-obsidian-positive-dim/20 p-2 font-mono text-xs text-obsidian-positive">
            {success}
          </p>
        )}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
              Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-accent"
            />
          </div>
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
              minLength={8}
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
          >
            {loading ? 'Registering...' : 'Register'}
          </button>
        </form>
        <p className="mt-4 text-center font-mono text-xs text-obsidian-text-tertiary">
          Already have an account?{' '}
          <Link to="/login" className="text-obsidian-accent hover:underline">
            Login
          </Link>
        </p>
      </div>
    </div>
  )
}
