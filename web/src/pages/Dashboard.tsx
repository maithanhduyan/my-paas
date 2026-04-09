import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listProjects, getHealth } from '../api'
import type { Project, HealthResponse } from '../types'
import { StatusBadge } from '../components/StatusBadge'
import { timeAgo } from '../lib/utils'
import { GitBranch, Plus, Server, Activity, Search, LayoutGrid, List } from 'lucide-react'

type StatusFilter = 'all' | 'healthy' | 'failed' | 'building'
type ViewMode = 'grid' | 'list'

export function Dashboard() {
  const [projects, setProjects] = useState<Project[]>([])
  const [health, setHealth] = useState<HealthResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [viewMode, setViewMode] = useState<ViewMode>('grid')

  useEffect(() => {
    Promise.all([
      listProjects().then(setProjects).catch(() => setProjects([])),
      getHealth().then(setHealth).catch(() => null),
    ]).finally(() => setLoading(false))
  }, [])

  const filtered = projects.filter(p => {
    const matchSearch = !search || p.name.toLowerCase().includes(search.toLowerCase()) ||
      (p.provider && p.provider.toLowerCase().includes(search.toLowerCase()))
    const matchStatus = statusFilter === 'all' || p.status === statusFilter
    return matchSearch && matchStatus
  })

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
          <h1 className="text-2xl font-bold">Projects</h1>
          <p className="text-sm text-gray-500 mt-1">
            {projects.length} project{projects.length !== 1 ? 's' : ''} deployed
          </p>
        </div>
        <Link
          to="/projects/new"
          className="flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-lg text-sm font-medium hover:bg-accent-hover transition-colors"
        >
          <Plus className="w-4 h-4" /> New
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

      {/* Search & Filter */}
      {projects.length > 0 && (
        <div className="flex flex-col sm:flex-row gap-3">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
            <input
              type="text"
              value={search}
              onChange={e => setSearch(e.target.value)}
              placeholder="Search projects..."
              className="w-full pl-9 pr-3 py-2 bg-surface-50 border border-surface-300 rounded-lg text-sm
                         placeholder-gray-600 focus:outline-none focus:border-accent transition-colors"
            />
          </div>
          <div className="flex gap-1.5">
            {(['all', 'healthy', 'failed', 'building'] as StatusFilter[]).map(f => (
              <button key={f} onClick={() => setStatusFilter(f)}
                className={`px-3 py-1.5 text-xs rounded-md font-medium transition-colors capitalize ${
                  statusFilter === f
                    ? 'bg-accent/15 text-accent-hover border border-accent/30'
                    : 'bg-surface-50 text-gray-500 border border-surface-300 hover:text-gray-300 hover:border-surface-200'
                }`}>
                {f}
              </button>
            ))}
          </div>
          {/* View mode toggle */}
          <div className="flex gap-0.5 bg-surface-50 border border-surface-300 rounded-lg p-0.5">
            <button
              onClick={() => setViewMode('grid')}
              className={`p-1.5 rounded-md transition-colors ${
                viewMode === 'grid' ? 'bg-surface-200 text-white' : 'text-gray-500 hover:text-gray-300'
              }`}
            >
              <LayoutGrid className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`p-1.5 rounded-md transition-colors ${
                viewMode === 'list' ? 'bg-surface-200 text-white' : 'text-gray-500 hover:text-gray-300'
              }`}
            >
              <List className="w-4 h-4" />
            </button>
          </div>
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
      ) : filtered.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          No projects match your search.
        </div>
      ) : viewMode === 'grid' ? (
        /* ─── Grid view (Railway-style cards) ─── */
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {filtered.map((p) => (
            <Link
              key={p.id}
              to={`/projects/${p.id}`}
              className="group p-5 bg-surface-50 border border-surface-300 rounded-xl
                         hover:border-accent/40 hover:shadow-lg hover:shadow-accent/5
                         transition-all duration-200"
            >
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-2.5 min-w-0">
                  <div
                    className="w-10 h-10 rounded-lg flex items-center justify-center text-lg shrink-0"
                    style={{ background: providerBg(p.provider) }}
                  >
                    {providerEmoji(p.provider)}
                  </div>
                  <div className="min-w-0">
                    <div className="font-semibold text-sm truncate group-hover:text-accent-hover transition-colors">
                      {p.name}
                    </div>
                    {p.provider && (
                      <div className="text-[11px] text-gray-500 mt-0.5">
                        {p.provider}{p.framework ? ` / ${p.framework}` : ''}
                      </div>
                    )}
                  </div>
                </div>
              </div>

              <div className="flex items-center justify-between">
                <StatusBadge status={p.status} />
                <div className="flex items-center gap-2 text-[11px] text-gray-500">
                  {p.git_url && (
                    <span className="flex items-center gap-1">
                      <GitBranch className="w-3 h-3" />
                      {p.branch || 'main'}
                    </span>
                  )}
                  <span>{timeAgo(p.updated_at)}</span>
                </div>
              </div>
            </Link>
          ))}
        </div>
      ) : (
        /* ─── List view (compact rows) ─── */
        <div className="grid gap-2">
          {filtered.map((p) => (
            <Link
              key={p.id}
              to={`/projects/${p.id}`}
              className="group flex items-center gap-4 p-4 bg-surface-50 border border-surface-300 rounded-lg
                         hover:border-accent/40 hover:bg-surface-100 hover:shadow-lg hover:shadow-accent/5
                         transition-all duration-200"
            >
              <div className="w-10 h-10 rounded-lg flex items-center justify-center text-lg shrink-0"
                   style={{ background: providerBg(p.provider) }}>
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

function providerBg(provider: string): string {
  const map: Record<string, string> = {
    node: '#22c55e15',
    go: '#00ADD815',
    python: '#3572A515',
    rust: '#DEA58415',
    php: '#777BB415',
    java: '#ED8B0015',
    staticfile: '#6366f115',
  }
  return map[provider] ?? '#6366f110'
}
