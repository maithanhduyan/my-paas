import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listProjects, getHealth } from '../api'
import type { Project, HealthResponse } from '../types'
import { StatusBadge } from '../components/StatusBadge'
import { timeAgo } from '../lib/utils'
import { GitBranch, Plus, Server, Activity } from 'lucide-react'

export function Dashboard() {
  const [projects, setProjects] = useState<Project[]>([])
  const [health, setHealth] = useState<HealthResponse | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      listProjects().then(setProjects).catch(() => setProjects([])),
      getHealth().then(setHealth).catch(() => null),
    ]).finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-500">Loading...</div>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-5xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Dashboard</h1>
          <p className="text-sm text-gray-500 mt-1">
            {projects.length} project{projects.length !== 1 ? 's' : ''} deployed
          </p>
        </div>
        <Link
          to="/projects/new"
          className="flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-lg text-sm font-medium hover:bg-accent-hover transition-colors"
        >
          <Plus className="w-4 h-4" /> New Project
        </Link>
      </div>

      {/* Status bar */}
      {health && (
        <div className="flex gap-4 text-xs text-gray-400">
          <span className="flex items-center gap-1">
            <Server className="w-3.5 h-3.5" />
            Docker: <span className="text-success">{health.docker}</span>
          </span>
          <span className="flex items-center gap-1">
            <Activity className="w-3.5 h-3.5" />
            Go: {health.go}
          </span>
        </div>
      )}

      {/* Project list */}
      {projects.length === 0 ? (
        <div className="text-center py-16 text-gray-500">
          <Server className="w-12 h-12 mx-auto mb-3 opacity-30" />
          <p className="text-lg">No projects yet</p>
          <p className="text-sm mt-1">Create your first project to get started.</p>
          <Link
            to="/projects/new"
            className="inline-flex items-center gap-2 mt-4 px-4 py-2 bg-accent text-white rounded-lg text-sm hover:bg-accent-hover"
          >
            <Plus className="w-4 h-4" /> New Project
          </Link>
        </div>
      ) : (
        <div className="grid gap-3">
          {projects.map((p) => (
            <Link
              key={p.id}
              to={`/projects/${p.id}`}
              className="flex items-center gap-4 p-4 bg-surface-50 border border-surface-300 rounded-lg hover:border-surface-200 hover:bg-surface-100 transition-colors"
            >
              <div className="w-10 h-10 bg-surface-200 rounded-lg flex items-center justify-center text-lg">
                {providerEmoji(p.provider)}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium truncate">{p.name}</span>
                  <StatusBadge status={p.status} />
                </div>
                <div className="flex items-center gap-3 mt-0.5 text-xs text-gray-500">
                  {p.provider && <span>{p.provider}{p.framework ? ` / ${p.framework}` : ''}</span>}
                  {p.git_url && (
                    <span className="flex items-center gap-1 truncate">
                      <GitBranch className="w-3 h-3" />
                      {p.branch || 'main'}
                    </span>
                  )}
                </div>
              </div>
              <div className="text-xs text-gray-500 shrink-0">
                {timeAgo(p.updated_at)}
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

function providerEmoji(provider: string): string {
  const map: Record<string, string> = {
    node: '🟢',
    go: '🔵',
    python: '🐍',
    rust: '🦀',
    php: '🐘',
    java: '☕',
    staticfile: '📄',
  }
  return map[provider] ?? '📦'
}
