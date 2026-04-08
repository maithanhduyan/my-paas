import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createProject } from '../api'
import { GitBranch, Globe, ArrowLeft } from 'lucide-react'
import { Link } from 'react-router-dom'

export function NewProject() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [gitUrl, setGitUrl] = useState('')
  const [branch, setBranch] = useState('main')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) { setError('Project name is required'); return }
    if (!gitUrl.trim()) { setError('Git URL is required'); return }

    setLoading(true)
    setError('')
    try {
      const project = await createProject({ name: name.trim(), git_url: gitUrl.trim(), branch })
      navigate(`/projects/${project.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create project')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-6 max-w-lg mx-auto">
      <Link to="/" className="inline-flex items-center gap-1 text-sm text-gray-400 hover:text-gray-200 mb-6">
        <ArrowLeft className="w-4 h-4" /> Back to Dashboard
      </Link>

      <h1 className="text-2xl font-bold mb-6">New Project</h1>

      <form onSubmit={submit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">Project Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="my-awesome-app"
            className="w-full px-3 py-2 bg-surface-100 border border-surface-300 rounded-lg text-sm
                       focus:outline-none focus:border-accent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">
            <Globe className="w-4 h-4 inline mr-1" />
            Git Repository URL
          </label>
          <input
            type="text"
            value={gitUrl}
            onChange={(e) => setGitUrl(e.target.value)}
            placeholder="https://github.com/user/repo.git"
            className="w-full px-3 py-2 bg-surface-100 border border-surface-300 rounded-lg text-sm
                       focus:outline-none focus:border-accent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">
            <GitBranch className="w-4 h-4 inline mr-1" />
            Branch
          </label>
          <input
            type="text"
            value={branch}
            onChange={(e) => setBranch(e.target.value)}
            placeholder="main"
            className="w-full px-3 py-2 bg-surface-100 border border-surface-300 rounded-lg text-sm
                       focus:outline-none focus:border-accent"
          />
        </div>

        {error && (
          <div className="text-sm text-danger bg-danger/10 border border-danger/20 rounded-lg px-3 py-2">
            {error}
          </div>
        )}

        <button
          type="submit"
          disabled={loading}
          className="w-full px-4 py-2.5 bg-accent text-white rounded-lg text-sm font-medium
                     hover:bg-accent-hover disabled:opacity-50 transition-colors"
        >
          {loading ? 'Creating...' : 'Create Project'}
        </button>
      </form>
    </div>
  )
}
