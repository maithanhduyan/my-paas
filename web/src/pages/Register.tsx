import { useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { registerWithInvite } from '../api'
import { Box } from 'lucide-react'

export function Register() {
  const { token } = useParams<{ token: string }>()
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!username.trim() || !password.trim()) {
      setError('Username and password required')
      return
    }
    if (password !== confirmPassword) {
      setError('Passwords do not match')
      return
    }
    if (password.length < 6) {
      setError('Password must be at least 6 characters')
      return
    }
    if (!token) {
      setError('Invalid invitation link')
      return
    }
    setLoading(true)
    setError('')
    try {
      const data = await registerWithInvite(token, username.trim(), password)
      if (data.token) {
        localStorage.setItem('mypaas_token', data.token)
        navigate('/')
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Registration failed'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center h-screen bg-surface">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <Box className="w-12 h-12 text-accent mx-auto mb-3" />
          <h1 className="text-2xl font-bold">My PaaS</h1>
          <p className="text-sm text-gray-500 mt-1">Create your account</p>
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
              autoComplete="new-password"
              className="w-full px-3 py-2 bg-surface border border-surface-300 rounded-md text-sm
                         focus:outline-none focus:ring-1 focus:ring-accent"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">Confirm Password</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              autoComplete="new-password"
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
            {loading ? 'Please wait...' : 'Create Account'}
          </button>

          <p className="text-center text-xs text-gray-500">
            Already have an account?{' '}
            <Link to="/login" className="text-accent hover:underline">Sign in</Link>
          </p>
        </form>
      </div>
    </div>
  )
}
