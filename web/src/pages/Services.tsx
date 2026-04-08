import { useCallback, useEffect, useState } from 'react'
import {
  listServices, createService, deleteService,
  startService, stopService, getSystemStats
} from '../api'
import type { Service, ServiceType, ContainerStats } from '../types'
import { timeAgo } from '../lib/utils'
import {
  Database, Play, Square, Trash2, Plus, X,
  Cpu, HardDrive, Wifi
} from 'lucide-react'

const SERVICE_TYPES: { type: ServiceType; label: string; icon: string }[] = [
  { type: 'postgres', label: 'PostgreSQL', icon: '🐘' },
  { type: 'redis', label: 'Redis', icon: '🔴' },
  { type: 'mysql', label: 'MySQL', icon: '🐬' },
  { type: 'mongo', label: 'MongoDB', icon: '🍃' },
  { type: 'minio', label: 'MinIO', icon: '📦' },
]

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}

export function Services() {
  const [services, setServices] = useState<Service[]>([])
  const [stats, setStats] = useState<ContainerStats[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [name, setName] = useState('')
  const [type, setType] = useState<ServiceType>('postgres')
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      const [svcs, st] = await Promise.all([
        listServices(),
        getSystemStats().catch(() => []),
      ])
      setServices(svcs)
      setStats(st)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  // Auto-refresh stats every 10s
  useEffect(() => {
    const iv = setInterval(() => {
      getSystemStats().then(setStats).catch(() => {})
    }, 10000)
    return () => clearInterval(iv)
  }, [])

  const handleCreate = async () => {
    if (!name.trim()) return
    await createService(name.trim(), type)
    setShowCreate(false)
    setName('')
    await load()
  }

  const handleStart = async (id: string) => {
    setActionLoading(id)
    try {
      await startService(id)
      await new Promise(r => setTimeout(r, 1500))
      await load()
    } finally {
      setActionLoading(null)
    }
  }

  const handleStop = async (id: string) => {
    setActionLoading(id)
    try {
      await stopService(id)
      await new Promise(r => setTimeout(r, 1500))
      await load()
    } finally {
      setActionLoading(null)
    }
  }

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Delete service "${name}"? This cannot be undone.`)) return
    await deleteService(id)
    await load()
  }

  const getServiceStats = (svc: Service): ContainerStats | undefined => {
    const containerName = 'mypaas-svc-' + svc.name
    return stats.find((s) => s.name === containerName)
  }

  if (loading) {
    return <div className="flex items-center justify-center h-full text-gray-500">Loading...</div>
  }

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Database className="w-6 h-6 text-accent" /> Services
          </h1>
          <p className="text-sm text-gray-500 mt-1">Managed databases and caches</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-lg text-sm font-medium hover:bg-accent-hover transition-colors"
        >
          <Plus className="w-4 h-4" /> New Service
        </button>
      </div>

      {/* Create dialog */}
      {showCreate && (
        <div className="bg-surface-50 border border-surface-300 rounded-lg p-4 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-medium">Create Service</h3>
            <button onClick={() => setShowCreate(false)} className="text-gray-500 hover:text-gray-300">
              <X className="w-4 h-4" />
            </button>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-gray-400 mb-1">Name</label>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="my-database"
                className="w-full px-3 py-2 bg-surface border border-surface-300 rounded-md text-sm focus:outline-none focus:ring-1 focus:ring-accent"
              />
            </div>
            <div>
              <label className="block text-sm text-gray-400 mb-1">Type</label>
              <select
                value={type}
                onChange={(e) => setType(e.target.value as ServiceType)}
                className="w-full px-3 py-2 bg-surface border border-surface-300 rounded-md text-sm focus:outline-none focus:ring-1 focus:ring-accent"
              >
                {SERVICE_TYPES.map((st) => (
                  <option key={st.type} value={st.type}>{st.icon} {st.label}</option>
                ))}
              </select>
            </div>
          </div>
          <button
            onClick={handleCreate}
            disabled={!name.trim()}
            className="px-4 py-2 bg-accent text-white rounded-md text-sm font-medium hover:bg-accent-hover disabled:opacity-50 transition-colors"
          >
            Create
          </button>
        </div>
      )}

      {/* System Stats Summary */}
      {stats.length > 0 && (
        <div className="grid grid-cols-3 gap-4">
          <div className="bg-surface-50 border border-surface-300 rounded-lg p-3">
            <div className="flex items-center gap-2 text-sm text-gray-500 mb-1">
              <Cpu className="w-4 h-4" /> Total CPU
            </div>
            <div className="text-lg font-bold">
              {stats.reduce((sum, s) => sum + s.cpu_percent, 0).toFixed(1)}%
            </div>
          </div>
          <div className="bg-surface-50 border border-surface-300 rounded-lg p-3">
            <div className="flex items-center gap-2 text-sm text-gray-500 mb-1">
              <HardDrive className="w-4 h-4" /> Total Memory
            </div>
            <div className="text-lg font-bold">
              {formatBytes(stats.reduce((sum, s) => sum + s.mem_usage, 0))}
            </div>
          </div>
          <div className="bg-surface-50 border border-surface-300 rounded-lg p-3">
            <div className="flex items-center gap-2 text-sm text-gray-500 mb-1">
              <Wifi className="w-4 h-4" /> Containers
            </div>
            <div className="text-lg font-bold">{stats.length}</div>
          </div>
        </div>
      )}

      {/* Service list */}
      {services.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          <Database className="w-12 h-12 mx-auto mb-3 opacity-30" />
          <p>No services yet. Create a database or cache to get started.</p>
        </div>
      ) : (
        <div className="space-y-3">
          {services.map((svc) => {
            const svcType = SERVICE_TYPES.find((s) => s.type === svc.type)
            const svcStats = getServiceStats(svc)
            return (
              <div key={svc.id} className="bg-surface-50 border border-surface-300 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <span className="text-xl">{svcType?.icon ?? '📦'}</span>
                    <div>
                      <div className="font-medium">{svc.name}</div>
                      <div className="text-xs text-gray-500 flex items-center gap-2">
                        <span>{svcType?.label ?? svc.type}</span>
                        <span>·</span>
                        <span className="font-mono">{svc.image}</span>
                        <span>·</span>
                        <span>{timeAgo(svc.created_at)}</span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    {/* Status badge */}
                    <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                      svc.status === 'running'
                        ? 'bg-success/20 text-success'
                        : svc.status === 'error'
                        ? 'bg-danger/20 text-danger'
                        : 'bg-gray-700 text-gray-400'
                    }`}>
                      {svc.status}
                    </span>

                    {/* Stats if running */}
                    {svcStats && (
                      <span className="text-xs text-gray-500 font-mono">
                        {svcStats.cpu_percent.toFixed(1)}% · {formatBytes(svcStats.mem_usage)}
                      </span>
                    )}

                    {/* Actions */}
                    {svc.status !== 'running' ? (
                      <button
                        onClick={() => handleStart(svc.id)}
                        disabled={actionLoading === svc.id}
                        className="p-1.5 text-success hover:bg-success/10 rounded transition-colors disabled:opacity-50"
                        title="Start"
                      >
                        <Play className="w-4 h-4" />
                      </button>
                    ) : (
                      <button
                        onClick={() => handleStop(svc.id)}
                        disabled={actionLoading === svc.id}
                        className="p-1.5 text-warning hover:bg-warning/10 rounded transition-colors disabled:opacity-50"
                        title="Stop"
                      >
                        <Square className="w-4 h-4" />
                      </button>
                    )}
                    <button
                      onClick={() => handleDelete(svc.id, svc.name)}
                      className="p-1.5 text-danger hover:bg-danger/10 rounded transition-colors"
                      title="Delete"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
