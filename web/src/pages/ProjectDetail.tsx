import { useCallback, useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  getProject, listDeployments, getEnvVars,
  triggerDeploy, deleteProject,
  getProjectStats, listDomains,
  listVolumes,
  restartProject, stopProject, startProject,
  unlinkService, startService, stopService,
} from '../api'
import type { Project, Deployment, EnvVar, ContainerStats, Domain, Volume, Service } from '../types'
import { ProjectCanvas } from '../components/flow/ProjectCanvas'
import { ActivitySidebar } from '../components/flow/ActivitySidebar'
import { ServiceDrawer, AppDrawer } from '../components/flow/Drawers'
import { LogViewer } from '../components/LogViewer'
import { useToast } from '../components/ui/Toast'
import {
  ArrowLeft, Rocket, GitBranch, Loader2, Play, Square, RefreshCw,
  Trash2, AlertTriangle, Settings, ChevronRight,
} from 'lucide-react'

/* ─── Delete Confirmation Modal ─── */
function DeleteModal({ projectName, onConfirm, onCancel }: {
  projectName: string
  onConfirm: () => void
  onCancel: () => void
}) {
  const [confirmText, setConfirmText] = useState('')
  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" onClick={onCancel}>
      <div className="bg-white rounded-2xl p-6 max-w-md w-full space-y-4 shadow-2xl" onClick={e => e.stopPropagation()}>
        <div className="flex items-center gap-3 text-red-600">
          <AlertTriangle className="w-5 h-5 shrink-0" />
          <h3 className="font-semibold text-lg">Delete Project</h3>
        </div>
        <p className="text-sm text-gray-500">
          This action <strong className="text-gray-900">cannot be undone</strong>. This will permanently delete
          <strong className="text-gray-900"> {projectName}</strong> and remove the Swarm service.
        </p>
        <div>
          <label className="block text-sm text-gray-500 mb-1.5">
            Type <strong className="text-gray-900 font-mono">{projectName}</strong> to confirm
          </label>
          <input type="text" value={confirmText} onChange={e => setConfirmText(e.target.value)}
            placeholder={projectName} autoFocus
            className="w-full px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-400" />
        </div>
        <div className="flex justify-end gap-3">
          <button onClick={onCancel} className="px-4 py-2 text-sm text-gray-500 hover:text-gray-700 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors">Cancel</button>
          <button onClick={onConfirm} disabled={confirmText !== projectName}
            className="px-4 py-2 text-sm text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-30 disabled:cursor-not-allowed transition-colors">
            Delete this project
          </button>
        </div>
      </div>
    </div>
  )
}

/* ─── Log Viewer Modal ─── */
function LogModal({ deploymentId, projectId, onClose }: {
  deploymentId: string; projectId: string; onClose: () => void
}) {
  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div className="bg-surface-50 border border-surface-300 rounded-2xl w-full max-w-3xl max-h-[80vh] overflow-hidden shadow-2xl"
        onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between px-5 py-3 border-b border-surface-300">
          <h3 className="text-sm font-semibold text-gray-200">Deployment Logs</h3>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-300 text-lg">&times;</button>
        </div>
        <div className="p-4 max-h-[70vh] overflow-y-auto">
          <LogViewer deploymentId={deploymentId} projectId={projectId} />
        </div>
      </div>
    </div>
  )
}

export function ProjectDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const [project, setProject] = useState<Project | null>(null)
  const [deployments, setDeployments] = useState<Deployment[]>([])
  const [envVars, setEnvVars] = useState<EnvVar[]>([])
  const [deploying, setDeploying] = useState(false)
  const [loading, setLoading] = useState(true)
  const [projectStats, setProjectStats] = useState<ContainerStats | null>(null)
  const [domains, setDomains] = useState<Domain[]>([])
  const [volumes, setVolumes] = useState<Volume[]>([])
  const [showDeleteModal, setShowDeleteModal] = useState(false)

  // Side panel states
  const [selectedService, setSelectedService] = useState<Service | null>(null)
  const [showAppDrawer, setShowAppDrawer] = useState(false)
  const [logDeploymentId, setLogDeploymentId] = useState<string | null>(null)

  const load = useCallback(async () => {
    if (!id) return
    try {
      const [p, deps, env] = await Promise.all([
        getProject(id),
        listDeployments(id),
        getEnvVars(id),
      ])
      setProject(p)
      setDeployments(deps)
      setEnvVars(env)
      listDomains(id).then(setDomains).catch(() => {})
      listVolumes(id).then(setVolumes).catch(() => {})
      getProjectStats(id).then((s) => {
        if (s && 'cpu_percent' in s) setProjectStats(s)
      }).catch(() => {})
    } catch {
      navigate('/')
    } finally {
      setLoading(false)
    }
  }, [id, navigate])

  useEffect(() => { load() }, [load])
  // Poll deployments while building
  useEffect(() => {
    if (!deploying || !id) return
    const interval = setInterval(async () => {
      try {
        const deps = await listDeployments(id)
        setDeployments(deps)
        const latest = deps[0]
        if (latest && !['building', 'deploying', 'queued', 'cloning', 'detecting'].includes(latest.status)) {
          setDeploying(false)
          const p = await getProject(id)
          setProject(p)
        }
      } catch { setDeploying(false) }
    }, 3000)
    return () => clearInterval(interval)
  }, [deploying, id])

  const handleDeploy = async () => {
    if (!id) return
    setDeploying(true)
    try {
      const dep = await triggerDeploy(id)
      setLogDeploymentId(dep.id)
      await load()
    } catch {
      setDeploying(false)
    }
  }

  const handleDelete = async () => {
    if (!id) return
    await deleteProject(id)
    navigate('/')
  }

  const handleServiceSelect = (svc: Service | null) => {
    if (svc) {
      setSelectedService(svc)
      setShowAppDrawer(false)
    } else {
      setShowAppDrawer(true)
      setSelectedService(null)
    }
  }

  const handleUnlinkService = async () => {
    if (!selectedService || !id) return
    try {
      await unlinkService(selectedService.id, id)
      toast({ type: 'success', title: 'Service unlinked' })
      setSelectedService(null)
      load()
    } catch (e: any) {
      toast({ type: 'error', title: 'Unlink failed', description: e.message })
    }
  }

  const handleStartService = async () => {
    if (!selectedService) return
    try {
      await startService(selectedService.id)
      toast({ type: 'success', title: 'Service started' })
      setSelectedService(null)
      load()
    } catch (e: any) {
      toast({ type: 'error', title: 'Start failed', description: e.message })
    }
  }

  const handleStopService = async () => {
    if (!selectedService) return
    try {
      await stopService(selectedService.id)
      toast({ type: 'success', title: 'Service stopped' })
      setSelectedService(null)
      load()
    } catch (e: any) {
      toast({ type: 'error', title: 'Stop failed', description: e.message })
    }
  }

  if (loading) {
    return <div className="flex items-center justify-center h-full text-gray-500">Loading...</div>
  }
  if (!project) return null

  /* Status helpers */
  const isOnline = ['healthy', 'running', 'active'].includes(project.status)
  const isStopped = project.status === 'stopped'

  return (
    <div className="flex flex-col h-[calc(100vh-0px)] bg-gray-50">
      {/* ─── Top Bar ──────────────────────────────────────── */}
      <div className="shrink-0 bg-white border-b border-gray-200 px-4 py-2.5 flex items-center gap-3">
        {/* Back + breadcrumb */}
        <button onClick={() => navigate('/')}
          className="flex items-center gap-1 text-sm text-gray-400 hover:text-gray-600 transition-colors">
          <ArrowLeft className="w-4 h-4" />
        </button>
        <div className="flex items-center gap-1.5 text-sm">
          <span className="text-gray-400 cursor-pointer hover:text-gray-600" onClick={() => navigate('/')}>Projects</span>
          <ChevronRight className="w-3.5 h-3.5 text-gray-300" />
          <span className="font-semibold text-gray-900">{project.name}</span>
        </div>

        {/* Status indicator */}
        <div className="flex items-center gap-1.5 ml-2">
          <span className={`w-2 h-2 rounded-full ${
            isOnline ? 'bg-emerald-500' : project.status === 'failed' ? 'bg-red-500' :
            project.status === 'building' ? 'bg-blue-400 animate-pulse' : 'bg-gray-400'
          }`} />
          <span className="text-xs text-gray-500 capitalize">{project.status}</span>
        </div>

        {project.branch && (
          <div className="flex items-center gap-1 text-xs text-gray-400 ml-2">
            <GitBranch className="w-3 h-3" /> {project.branch}
          </div>
        )}

        <div className="flex-1" />

        {/* Actions */}
        <button onClick={() => { setShowAppDrawer(true); setSelectedService(null) }}
          className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
          title="Settings">
          <Settings className="w-4 h-4" />
        </button>

        {isStopped ? (
          <button onClick={async () => { await startProject(id!); await load() }}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-emerald-700 bg-emerald-50 border border-emerald-200 rounded-lg hover:bg-emerald-100 transition-colors">
            <Play className="w-3.5 h-3.5" /> Start
          </button>
        ) : (
          <>
            <button onClick={async () => { await restartProject(id!); await load() }}
              className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" title="Restart">
              <RefreshCw className="w-4 h-4" />
            </button>
            <button onClick={async () => { await stopProject(id!); await load() }}
              className="p-2 text-amber-500 hover:text-amber-600 hover:bg-amber-50 rounded-lg transition-colors" title="Stop">
              <Square className="w-3.5 h-3.5" />
            </button>
          </>
        )}

        <button onClick={handleDeploy} disabled={deploying}
          className="flex items-center gap-1.5 px-3.5 py-1.5 bg-gray-900 text-white rounded-lg text-sm font-medium
                     hover:bg-gray-800 disabled:opacity-60 transition-colors">
          {deploying ? <><Loader2 className="w-3.5 h-3.5 animate-spin" /> Deploying...</>
            : <><Rocket className="w-3.5 h-3.5" /> Deploy</>}
        </button>

        <button onClick={() => setShowDeleteModal(true)}
          className="p-2 text-gray-300 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors" title="Delete project">
          <Trash2 className="w-4 h-4" />
        </button>
      </div>

      {/* ─── Main Content: Canvas + Activity ─────────────── */}
      <div className="flex-1 flex min-h-0">
        {/* React Flow Canvas */}
        <div className="flex-1 relative">
          <ProjectCanvas
            project={project}
            onSelectService={handleServiceSelect}
            onUpdate={load}
          />
        </div>

        {/* Activity Sidebar (right) */}
        <div className="w-[280px] bg-white border-l border-gray-200 flex flex-col shrink-0">
          <div className="px-4 py-3 border-b border-gray-100">
            <h3 className="text-sm font-semibold text-gray-900">Activity</h3>
          </div>
          <div className="flex-1 overflow-y-auto">
            <ActivitySidebar
              deployments={deployments}
              onViewLogs={(depId) => setLogDeploymentId(depId)}
            />
          </div>
        </div>
      </div>

      {/* ─── Drawers & Modals ─────────────────────────────── */}

      {selectedService && (
        <ServiceDrawer
          service={selectedService}
          project={project}
          onClose={() => setSelectedService(null)}
          onStart={handleStartService}
          onStop={handleStopService}
          onUnlink={handleUnlinkService}
        />
      )}

      {showAppDrawer && (
        <AppDrawer
          project={project}
          deployments={deployments}
          envVars={envVars}
          stats={projectStats}
          domains={domains}
          volumes={volumes}
          onClose={() => setShowAppDrawer(false)}
          onUpdate={load}
          onRestart={async () => { await restartProject(id!); await load() }}
          onStop={async () => { await stopProject(id!); await load() }}
          onStart={async () => { await startProject(id!); await load() }}
          onDeploy={handleDeploy}
          onViewLogs={(depId) => { setShowAppDrawer(false); setLogDeploymentId(depId) }}
        />
      )}

      {logDeploymentId && (
        <LogModal
          deploymentId={logDeploymentId}
          projectId={project.id}
          onClose={() => setLogDeploymentId(null)}
        />
      )}

      {showDeleteModal && (
        <DeleteModal
          projectName={project.name}
          onConfirm={handleDelete}
          onCancel={() => setShowDeleteModal(false)}
        />
      )}
    </div>
  )
}
