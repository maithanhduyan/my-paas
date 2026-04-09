import { useState, useCallback, useEffect } from 'react'
import { listServices, linkService, unlinkService, startService, stopService, createService } from '../api'
import type { Project, Service, ServiceType } from '../types'
import { useToast } from './ui/Toast'
import { Modal } from './ui/Modal'
import {
  Database, Plus, Link2, Unlink, Play, Square,
  GripVertical, Package, Loader2, ArrowRight,
} from 'lucide-react'

const SERVICE_ICONS: Record<string, string> = {
  postgres: '🐘', mysql: '🐬', redis: '⚡', mongo: '🍃', minio: '🪣',
}

const SERVICE_COLORS: Record<string, string> = {
  postgres: '#336791', mysql: '#4479A1', redis: '#DC382D', mongo: '#47A248', minio: '#C72C48',
}

interface Props {
  project: Project
  onUpdate: () => void
}

export function ProjectArchitecture({ project, onUpdate }: Props) {
  const [services, setServices] = useState<Service[]>([])
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const [showAddService, setShowAddService] = useState(false)
  const [dragOverApp, setDragOverApp] = useState(false)
  const { toast } = useToast()

  // Linked = services whose name starts with project prefix (convention from marketplace template)
  // We'll also keep track by storing in local state
  const [linkedIds, setLinkedIds] = useState<Set<string>>(new Set())

  const load = useCallback(async () => {
    try {
      const allServices = await listServices()
      setServices(allServices)
      // Heuristic: a service is linked if its name starts with the project name prefix
      // (this aligns with the marketplace template naming convention)
      const projectBase = project.name.replace(/-app$/, '').replace(/-wordpress-app$/, '')
      const linked = new Set<string>(
        allServices
          .filter((s: Service) => s.name.startsWith(projectBase))
          .map((s: Service) => s.id)
      )
      setLinkedIds(linked)
    } catch {
      setServices([])
    } finally {
      setLoading(false)
    }
  }, [project.name])

  useEffect(() => { load() }, [load])

  const linkedServices = services.filter(s => linkedIds.has(s.id))
  const unlinkedServices = services.filter(s => !linkedIds.has(s.id))

  const handleLink = async (serviceId: string) => {
    setActionLoading(serviceId)
    try {
      await linkService(serviceId, project.id)
      setLinkedIds(prev => new Set([...prev, serviceId]))
      toast({ type: 'success', title: 'Service linked', description: 'Environment variables have been injected.' })
      onUpdate()
    } catch (e: any) {
      toast({ type: 'error', title: 'Link failed', description: e.message })
    } finally {
      setActionLoading(null)
    }
  }

  const handleUnlink = async (serviceId: string) => {
    setActionLoading(serviceId)
    try {
      await unlinkService(serviceId, project.id)
      setLinkedIds(prev => {
        const next = new Set(prev)
        next.delete(serviceId)
        return next
      })
      toast({ type: 'success', title: 'Service unlinked' })
      onUpdate()
    } catch (e: any) {
      toast({ type: 'error', title: 'Unlink failed', description: e.message })
    } finally {
      setActionLoading(null)
    }
  }

  const handleStartService = async (serviceId: string) => {
    setActionLoading(serviceId)
    try {
      await startService(serviceId)
      toast({ type: 'success', title: 'Service started' })
      await load()
    } catch (e: any) {
      toast({ type: 'error', title: 'Start failed', description: e.message })
    } finally {
      setActionLoading(null)
    }
  }

  const handleStopService = async (serviceId: string) => {
    setActionLoading(serviceId)
    try {
      await stopService(serviceId)
      toast({ type: 'success', title: 'Service stopped' })
      await load()
    } catch (e: any) {
      toast({ type: 'error', title: 'Stop failed', description: e.message })
    } finally {
      setActionLoading(null)
    }
  }

  // Drag & drop from unlinked list to the architecture area
  const handleDragStart = (e: React.DragEvent, serviceId: string) => {
    e.dataTransfer.setData('text/plain', serviceId)
    e.dataTransfer.effectAllowed = 'link'
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOverApp(false)
    const serviceId = e.dataTransfer.getData('text/plain')
    if (serviceId && !linkedIds.has(serviceId)) {
      handleLink(serviceId)
    }
  }

  if (loading) {
    return <div className="text-center py-8 text-gray-500">Loading architecture...</div>
  }

  return (
    <div className="space-y-6">
      {/* Visual Architecture */}
      <div
        className={`
          relative bg-surface-50 border-2 border-dashed rounded-2xl p-6 min-h-[300px]
          transition-colors
          ${dragOverApp ? 'border-accent bg-accent/5' : 'border-surface-300'}
        `}
        onDragOver={(e) => { e.preventDefault(); setDragOverApp(true) }}
        onDragLeave={() => setDragOverApp(false)}
        onDrop={handleDrop}
      >
        {/* Drop hint */}
        {dragOverApp && (
          <div className="absolute inset-0 flex items-center justify-center z-10 pointer-events-none">
            <div className="px-4 py-2 bg-accent/20 border border-accent/40 rounded-xl text-sm text-accent font-medium animate-pulse">
              Drop to link service
            </div>
          </div>
        )}

        <div className="flex items-start gap-6 flex-wrap">
          {/* App node (center / main) */}
          <div className="flex flex-col items-center gap-4">
            <div className="w-48 bg-surface border border-accent/30 rounded-xl shadow-lg shadow-accent/10 overflow-hidden">
              <div className="h-1 bg-gradient-to-r from-accent to-accent/60" />
              <div className="p-4">
                <div className="flex items-center gap-2 mb-2">
                  <div className="w-8 h-8 bg-accent/10 rounded-lg flex items-center justify-center">
                    <Package className="w-4 h-4 text-accent" />
                  </div>
                  <div className="min-w-0">
                    <div className="text-sm font-semibold truncate">{project.name}</div>
                    <div className="text-[10px] text-gray-500">Application</div>
                  </div>
                </div>
                <div className="flex items-center gap-1.5 mt-2">
                  <span className={`w-1.5 h-1.5 rounded-full ${
                    project.status === 'healthy' ? 'bg-success' :
                    project.status === 'failed' ? 'bg-danger' :
                    project.status === 'building' ? 'bg-blue-400 animate-pulse' :
                    'bg-gray-500'
                  }`} />
                  <span className="text-[11px] text-gray-500 capitalize">{project.status}</span>
                </div>
              </div>
            </div>
            <span className="text-[10px] text-gray-600 uppercase tracking-wider font-semibold">App Container</span>
          </div>

          {/* Connection lines + Linked services */}
          {linkedServices.length > 0 && (
            <div className="flex flex-col gap-3 flex-1 min-w-[200px]">
              <div className="text-[10px] text-gray-600 uppercase tracking-wider font-semibold mb-1 flex items-center gap-2">
                <ArrowRight className="w-3 h-3" /> Linked Services
              </div>
              {linkedServices.map(svc => (
                <ServiceCard
                  key={svc.id}
                  service={svc}
                  linked
                  loading={actionLoading === svc.id}
                  onUnlink={() => handleUnlink(svc.id)}
                  onStart={() => handleStartService(svc.id)}
                  onStop={() => handleStopService(svc.id)}
                />
              ))}
            </div>
          )}

          {/* Empty state */}
          {linkedServices.length === 0 && (
            <div className="flex-1 flex items-center justify-center min-h-[200px]">
              <div className="text-center text-gray-600">
                <Database className="w-8 h-8 mx-auto mb-2 opacity-40" />
                <p className="text-sm">No linked services</p>
                <p className="text-xs mt-1">Drag a service here or click "Add Service"</p>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Available (unlinked) services */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">
            Available Services ({unlinkedServices.length})
          </h4>
          <button
            onClick={() => setShowAddService(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-accent text-white rounded-lg hover:bg-accent-hover transition-colors"
          >
            <Plus className="w-3.5 h-3.5" /> New Service
          </button>
        </div>
        {unlinkedServices.length === 0 ? (
          <p className="text-sm text-gray-600">All services are linked or none exist. Create one to get started.</p>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
            {unlinkedServices.map(svc => (
              <div
                key={svc.id}
                draggable
                onDragStart={(e) => handleDragStart(e, svc.id)}
                className="flex items-center gap-2 px-3 py-2 bg-surface-50 border border-surface-300 rounded-lg
                           cursor-grab active:cursor-grabbing hover:border-accent/40 transition-colors group"
              >
                <GripVertical className="w-3 h-3 text-gray-600 group-hover:text-gray-400" />
                <span className="text-base">{SERVICE_ICONS[svc.type] || '📦'}</span>
                <div className="min-w-0 flex-1">
                  <div className="text-xs font-medium truncate">{svc.name}</div>
                  <div className="text-[10px] text-gray-600">{svc.type} · {svc.status}</div>
                </div>
                <button
                  onClick={() => handleLink(svc.id)}
                  disabled={actionLoading === svc.id}
                  className="p-1 text-gray-500 hover:text-accent transition-colors"
                  title="Link to project"
                >
                  {actionLoading === svc.id ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <Link2 className="w-3.5 h-3.5" />
                  )}
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create service modal */}
      <CreateServiceModal
        open={showAddService}
        onClose={() => setShowAddService(false)}
        onCreated={async (svcId) => {
          await load()
          // auto-link the newly created service
          handleLink(svcId)
        }}
      />
    </div>
  )
}

/* ─── Service Card (linked) ─── */
function ServiceCard({ service, linked, loading, onUnlink, onStart, onStop }: {
  service: Service
  linked: boolean
  loading: boolean
  onUnlink: () => void
  onStart: () => void
  onStop: () => void
}) {
  const color = SERVICE_COLORS[service.type] || '#6366f1'
  const icon = SERVICE_ICONS[service.type] || '📦'
  const isRunning = service.status === 'running'

  return (
    <div className="flex items-center gap-3 bg-surface border border-surface-300 rounded-xl p-3 hover:border-surface-200 transition-colors">
      <div
        className="w-9 h-9 rounded-lg flex items-center justify-center text-lg shrink-0"
        style={{ background: `${color}15` }}
      >
        {icon}
      </div>
      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium truncate">{service.name}</div>
        <div className="flex items-center gap-1.5 mt-0.5">
          <span className={`w-1.5 h-1.5 rounded-full ${isRunning ? 'bg-success' : 'bg-gray-500'}`} />
          <span className="text-[11px] text-gray-500">{service.type} · {service.image}</span>
        </div>
      </div>
      <div className="flex items-center gap-1 shrink-0">
        {loading ? (
          <Loader2 className="w-4 h-4 text-gray-500 animate-spin" />
        ) : (
          <>
            {isRunning ? (
              <button onClick={onStop} className="p-1.5 text-warning hover:bg-warning/10 rounded transition-colors" title="Stop">
                <Square className="w-3.5 h-3.5" />
              </button>
            ) : (
              <button onClick={onStart} className="p-1.5 text-success hover:bg-success/10 rounded transition-colors" title="Start">
                <Play className="w-3.5 h-3.5" />
              </button>
            )}
            {linked && (
              <button onClick={onUnlink} className="p-1.5 text-gray-500 hover:text-danger hover:bg-danger/10 rounded transition-colors" title="Unlink">
                <Unlink className="w-3.5 h-3.5" />
              </button>
            )}
          </>
        )}
      </div>
    </div>
  )
}

/* ─── Create Service Modal ─── */
const SVC_TYPES: { type: ServiceType; label: string; icon: string }[] = [
  { type: 'postgres', label: 'PostgreSQL', icon: '🐘' },
  { type: 'mysql', label: 'MySQL', icon: '🐬' },
  { type: 'redis', label: 'Redis', icon: '⚡' },
  { type: 'mongo', label: 'MongoDB', icon: '🍃' },
  { type: 'minio', label: 'MinIO', icon: '🪣' },
]

function CreateServiceModal({ open, onClose, onCreated }: {
  open: boolean
  onClose: () => void
  onCreated: (serviceId: string) => void
}) {
  const [name, setName] = useState('')
  const [type, setType] = useState<ServiceType>('postgres')
  const [creating, setCreating] = useState(false)
  const { toast } = useToast()

  const handleCreate = async () => {
    if (!name.trim()) return
    setCreating(true)
    try {
      const svc = await createService(name.trim(), type)
      // Auto-start the service
      await startService(svc.id)
      toast({ type: 'success', title: `${type} service created & started` })
      onCreated(svc.id)
      onClose()
      setName('')
    } catch (e: any) {
      toast({ type: 'error', title: 'Create failed', description: e.message })
    } finally {
      setCreating(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="New Service" description="Create a database or cache and link it to your project.">
      <div className="space-y-4">
        <div>
          <label className="block text-xs font-medium text-gray-400 mb-1.5">Service Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="input-field"
            placeholder="my-database"
            autoFocus
            onKeyDown={(e) => { if (e.key === 'Enter') handleCreate() }}
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-400 mb-1.5">Type</label>
          <div className="grid grid-cols-3 gap-2">
            {SVC_TYPES.map(t => (
              <button
                key={t.type}
                onClick={() => setType(t.type)}
                className={`flex items-center gap-2 px-3 py-2 rounded-lg border text-sm transition-colors ${
                  type === t.type
                    ? 'border-accent bg-accent/10 text-accent'
                    : 'border-surface-300 text-gray-400 hover:border-surface-200'
                }`}
              >
                <span>{t.icon}</span>
                <span>{t.label}</span>
              </button>
            ))}
          </div>
        </div>
        <div className="flex gap-2 justify-end pt-2">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-gray-400 hover:text-gray-200 rounded-lg hover:bg-surface-200 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleCreate}
            disabled={!name.trim() || creating}
            className="flex items-center gap-2 px-5 py-2 bg-accent text-white rounded-lg text-sm font-medium hover:bg-accent-hover disabled:opacity-50 transition-colors"
          >
            {creating ? <Loader2 className="w-4 h-4 animate-spin" /> : <Plus className="w-4 h-4" />}
            Create & Link
          </button>
        </div>
      </div>
    </Modal>
  )
}
