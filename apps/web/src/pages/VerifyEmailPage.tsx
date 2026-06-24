import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { api } from '../lib/api'
import { ThemeToggle } from '../components/ThemeToggle'

export function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [message, setMessage] = useState('')

  useEffect(() => {
    if (!token) {
      setStatus('error')
      setMessage('Missing verification token')
      return
    }
    api
      .get(`/auth/verify?token=${token}`)
      .then(() => {
        setStatus('success')
        setMessage('Email verified successfully. You can now log in.')
      })
      .catch((err) => {
        setStatus('error')
        setMessage(err.response?.data?.error || 'Verification failed')
      })
  }, [token])

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-obsidian-base px-4">
      <div className="absolute right-4 top-4">
        <ThemeToggle />
      </div>
      <div className="w-full max-w-md border border-obsidian-border-dim bg-obsidian-surface p-8 text-center">
        <h1 className="mb-4 font-mono text-2xl font-bold tracking-tight text-obsidian-text-primary">
          <span className="text-obsidian-accent">&gt;</span> Email Verification
        </h1>
        {status === 'loading' && <p className="font-mono text-sm text-obsidian-text-secondary">Verifying...</p>}
        {status === 'success' && (
          <>
            <p className="mb-4 font-mono text-sm text-obsidian-positive">{message}</p>
            <Link
              to="/login"
              className="font-mono text-sm text-obsidian-accent hover:underline"
            >
              Go to Login
            </Link>
          </>
        )}
        {status === 'error' && (
          <>
            <p className="mb-4 font-mono text-sm text-obsidian-negative">{message}</p>
            <Link
              to="/login"
              className="font-mono text-sm text-obsidian-accent hover:underline"
            >
              Go to Login
            </Link>
          </>
        )}
      </div>
    </div>
  )
}
