import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Box } from 'lucide-react'

export function Login() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [isSetup, setIsSetup] = useState(false)
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('mypaas_token')
    fetch('/api/auth/status', {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
      .then((r) => r.json())
      .then((data) => {
        if (data.authenticated) {
          navigate('/')
          return
        }
        setIsSetup(data.setup_required)
      })
      .catch(() => {})
      .finally(() => setChecking(false))
  }, [navigate])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!username.trim() || !password.trim()) {
      setError('Username and password required')
      return
    }
    setLoading(true)
    setError('')
    try {
      const endpoint = isSetup ? '/api/auth/setup' : '/api/auth/login'
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: username.trim(), password }),
      })
      const data = await res.json()
      if (!res.ok) {
        setError(data.error || 'Authentication failed')
        return
      }
      localStorage.setItem('mypaas_token', data.token)
      navigate('/')
    } catch {
      setError('Connection failed')
    } finally {
      setLoading(false)
    }
  }

  if (checking) {
    return (
      <div className="flex items-center justify-center h-screen bg-surface">
        <div className="text-gray-500">Loading...</div>
      </div>
    )
  }

  return (
    <div className="flex items-center justify-center h-screen bg-surface">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <Box className="w-12 h-12 text-accent mx-auto mb-3" />
          <h1 className="text-2xl font-bold">My PaaS</h1>
          <p className="text-sm text-gray-500 mt-1">
            {isSetup ? 'Create your admin account' : 'Sign in to continue'}
          </p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="bg-surface-50 border border-surface-300 rounded-lg p-6 space-y-4"
        >
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoComplete="username"
              className="w-full px-3 py-2 bg-surface border border-surface-300 rounded-md text-sm
                         focus:outline-none focus:ring-1 focus:ring-accent"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete={isSetup ? 'new-password' : 'current-password'}
              className="w-full px-3 py-2 bg-surface border border-surface-300 rounded-md text-sm
                         focus:outline-none focus:ring-1 focus:ring-accent"
            />
          </div>

          {error && (
            <div className="text-sm text-danger bg-danger/10 border border-danger/20 rounded px-3 py-2">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full px-4 py-2.5 bg-accent text-white rounded-md text-sm font-medium
                       hover:bg-accent-hover disabled:opacity-50 transition-colors"
          >
            {loading ? 'Please wait...' : isSetup ? 'Create Account' : 'Sign In'}
          </button>
        </form>
      </div>
    </div>
  )
}
