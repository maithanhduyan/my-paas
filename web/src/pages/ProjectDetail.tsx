import { useCallback, useEffect, useState, useRef } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import {
  getProject, listDeployments, getEnvVars,
  triggerDeploy, deleteProject, rollbackDeployment,
  getProjectStats, listDomains, addDomain, deleteDomain, updateProject,
  listVolumes, createVolume, deleteVolume,
  restartProject, stopProject, startProject
} from '../api'
import type { Project, Deployment, EnvVar, ContainerStats, Domain, Volume } from '../types'
import { StatusBadge } from '../components/StatusBadge'
import { LogViewer } from '../components/LogViewer'
import { EnvEditor } from '../components/EnvEditor'
import { timeAgo, duration } from '../lib/utils'
import {
  ArrowLeft, Rocket, Trash2, GitBranch, Clock,
  ChevronDown, ChevronUp, RotateCcw, Cpu, HardDrive, Copy, Check, Globe, Database, Plus,
  ExternalLink, AlertTriangle, Loader2, MoreVertical, Play, Square, RefreshCw
} from 'lucide-react'

type Tab = 'deployments' | 'env' | 'volumes' | 'settings'

/* ─── Deployment Status Badge (ACTIVE / REMOVED / building etc) ─── */
function DeployBadge({ status, isActive }: { status: string; isActive: boolean }) {
  if (isActive && status === 'healthy') {
    return <span className="px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider bg-success/20 text-success border border-success/30 rounded">Active</span>
  }
  if (status === 'healthy') {
    return <span className="px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider bg-surface-200 text-gray-500 border border-surface-300 rounded">Removed</span>
  }
  if (status === 'failed') {
    return <span className="px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider bg-danger/20 text-danger border border-danger/30 rounded">Failed</span>
  }
  if (status === 'rolled_back') {
    return <span className="px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider bg-warning/20 text-warning border border-warning/30 rounded">Rolled back</span>
  }
  // building/deploying/queued
  return <span className="px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider bg-blue-500/20 text-blue-400 border border-blue-500/30 rounded animate-pulse">{status}</span>
}

/* ─── 3-dot Action Menu ─── */
function ActionMenu({ children, onClose }: { children: React.ReactNode; onClose: () => void }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose()
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [onClose])
  return (
    <div ref={ref} className="absolute right-0 top-full mt-1 z-20 bg-surface-100 border border-surface-300 rounded-lg shadow-xl py-1 min-w-[160px]">
      {children}
    </div>
  )
}
function MenuItem({ icon: Icon, label, onClick, danger }: {
  icon: React.ComponentType<{ className?: string }>
  label: string; onClick: () => void; danger?: boolean
}) {
  return (
    <button onClick={onClick}
      className={`w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-surface-200 transition-colors ${danger ? 'text-danger' : 'text-gray-300'}`}>
      <Icon className="w-3.5 h-3.5" />{label}
    </button>
  )
}

/* ─── Deployment Row ─── */
function DeploymentRow({ d, isActive, expanded, onToggle, onRollback, onRedeploy, projectId }: {
  d: Deployment; isActive: boolean; expanded: boolean
  onToggle: () => void; onRollback: (id: string) => void; onRedeploy: () => void
  projectId: string
}) {
  const [menuOpen, setMenuOpen] = useState(false)
  return (
    <div className={`bg-surface-50 border rounded-lg overflow-hidden ${isActive ? 'border-success/30' : 'border-surface-300'}`}>
      <div className="flex items-center gap-3 p-3">
        <button onClick={onToggle} className="flex-1 flex items-center gap-3 text-left hover:bg-surface-100 transition-colors rounded -m-1 p-1">
          <DeployBadge status={d.status} isActive={isActive} />
          <div className="flex-1 min-w-0">
            <span className="text-sm font-mono text-gray-300 truncate block">
              {d.commit_hash ? `${d.commit_hash.slice(0, 7)} — ${d.commit_msg || 'No message'}` : d.id.slice(0, 8)}
            </span>
          </div>
          <div className="flex items-center gap-3 text-xs text-gray-500 shrink-0">
            <span className="flex items-center gap-1">
              <Clock className="w-3 h-3" />
              {d.started_at && d.finished_at ? duration(d.started_at, d.finished_at) : '—'}
            </span>
            <span>{timeAgo(d.created_at)}</span>
            <span className="px-1.5 py-0.5 bg-surface-200 rounded text-xs">{d.trigger}</span>
            {expanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </div>
        </button>
        {/* 3-dot menu */}
        <div className="relative">
          <button onClick={() => setMenuOpen(!menuOpen)}
            className="p-1.5 text-gray-500 hover:text-gray-300 hover:bg-surface-200 rounded transition-colors">
            <MoreVertical className="w-4 h-4" />
          </button>
          {menuOpen && (
            <ActionMenu onClose={() => setMenuOpen(false)}>
              <MenuItem icon={ChevronDown} label="View logs" onClick={() => { onToggle(); setMenuOpen(false) }} />
              <MenuItem icon={Rocket} label="Redeploy" onClick={() => { onRedeploy(); setMenuOpen(false) }} />
              {d.status === 'healthy' && d.image_tag && (
                <MenuItem icon={RotateCcw} label="Rollback to this" onClick={() => { onRollback(d.id); setMenuOpen(false) }} />
              )}
            </ActionMenu>
          )}
        </div>
      </div>
      {expanded && (
        <div className="border-t border-surface-300 p-3">
          <LogViewer deploymentId={d.id} projectId={projectId} />
        </div>
      )}
    </div>
  )
}

/* ─── Delete Confirmation Modal ─── */
function DeleteModal({ projectName, onConfirm, onCancel }: {
  projectName: string
  onConfirm: () => void
  onCancel: () => void
}) {
  const [confirmText, setConfirmText] = useState('')
  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4" onClick={onCancel}>
      <div className="bg-surface-50 border border-surface-300 rounded-xl p-6 max-w-md w-full space-y-4" onClick={e => e.stopPropagation()}>
        <div className="flex items-center gap-3 text-danger">
          <AlertTriangle className="w-5 h-5 shrink-0" />
          <h3 className="font-semibold text-lg">Delete Project</h3>
        </div>
        <p className="text-sm text-gray-400">
          This action <strong className="text-gray-200">cannot be undone</strong>. This will permanently delete the
          project <strong className="text-gray-200">{projectName}</strong>, its deployments, and remove the Swarm service.
        </p>
        <div>
          <label className="block text-sm text-gray-400 mb-1.5">
            Type <strong className="text-gray-200 font-mono">{projectName}</strong> to confirm
          </label>
          <input
            type="text"
            value={confirmText}
            onChange={e => setConfirmText(e.target.value)}
            placeholder={projectName}
            autoFocus
            className="w-full px-3 py-2 bg-surface border border-surface-300 rounded-lg text-sm
                       focus:outline-none focus:ring-1 focus:ring-danger"
          />
        </div>
        <div className="flex justify-end gap-3">
          <button onClick={onCancel}
            className="px-4 py-2 text-sm text-gray-400 hover:text-gray-200 rounded-lg border border-surface-300 hover:bg-surface-200 transition-colors">
            Cancel
          </button>
          <button onClick={onConfirm}
            disabled={confirmText !== projectName}
            className="px-4 py-2 text-sm text-white bg-danger rounded-lg hover:bg-danger/80
                       disabled:opacity-30 disabled:cursor-not-allowed transition-colors">
            Delete this project
          </button>
        </div>
      </div>
    </div>
  )
}

export function ProjectDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [project, setProject] = useState<Project | null>(null)
  const [deployments, setDeployments] = useState<Deployment[]>([])
  const [envVars, setEnvVars] = useState<EnvVar[]>([])
  const [tab, setTab] = useState<Tab>('deployments')
  const [deploying, setDeploying] = useState(false)
  const [expandedDeploy, setExpandedDeploy] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [projectStats, setProjectStats] = useState<ContainerStats | null>(null)
  const [copied, setCopied] = useState(false)
  const [domains, setDomains] = useState<Domain[]>([])
  const [newDomain, setNewDomain] = useState('')
  const [cpuLimit, setCpuLimit] = useState(0)
  const [memLimit, setMemLimit] = useState(0)
  const [replicas, setReplicas] = useState(0)
  const [volumes, setVolumes] = useState<Volume[]>([])
  const [newVolName, setNewVolName] = useState('')
  const [newVolPath, setNewVolPath] = useState('')
  const [showDeleteModal, setShowDeleteModal] = useState(false)

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
      setCpuLimit(p.cpu_limit || 0)
      setMemLimit(p.mem_limit || 0)
      setReplicas(p.replicas || 0)
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

  const handleRollback = async (deploymentId: string) => {
    if (!confirm('Rollback to this deployment?')) return
    await rollbackDeployment(deploymentId)
    await load()
  }

  const webhookUrl = `${window.location.origin}/api/webhooks/github`
  const copyWebhookUrl = () => {
    navigator.clipboard.writeText(webhookUrl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleAddDomain = async () => {
    if (!id || !newDomain.trim()) return
    await addDomain(id, newDomain.trim())
    setNewDomain('')
    listDomains(id).then(setDomains).catch(() => {})
  }

  const handleDeleteDomain = async (domainId: string) => {
    if (!id) return
    await deleteDomain(domainId)
    listDomains(id).then(setDomains).catch(() => {})
  }

  const handleAddVolume = async () => {
    if (!id || !newVolName.trim() || !newVolPath.trim()) return
    await createVolume(id, newVolName.trim(), newVolPath.trim())
    setNewVolName('')
    setNewVolPath('')
    listVolumes(id).then(setVolumes).catch(() => {})
  }

  const handleDeleteVolume = async (volumeId: string) => {
    if (!id || !confirm('Delete this volume? Data may be lost.')) return
    await deleteVolume(id, volumeId)
    listVolumes(id).then(setVolumes).catch(() => {})
  }

  const handleSaveResources = async () => {
    if (!id) return
    await updateProject(id, { cpu_limit: cpuLimit, mem_limit: memLimit, replicas })
    await load()
  }

  const handleDeploy = async () => {
    if (!id) return
    setDeploying(true)
    try {
      const dep = await triggerDeploy(id)
      // Auto-switch to deployments tab and expand the new deployment for live streaming
      setTab('deployments')
      setExpandedDeploy(dep.id)
      await load()
      // Poll for completion so status updates in real-time
      const pollInterval = setInterval(async () => {
        try {
          const deps = await listDeployments(id)
          setDeployments(deps)
          const latest = deps.find(d => d.id === dep.id)
          if (latest && latest.status !== 'building' && latest.status !== 'deploying') {
            clearInterval(pollInterval)
            setDeploying(false)
            const p = await getProject(id)
            setProject(p)
          }
        } catch { clearInterval(pollInterval); setDeploying(false) }
      }, 3000)
    } catch {
      setDeploying(false)
    }
  }

  const handleDelete = async () => {
    if (!id) return
    await deleteProject(id)
    navigate('/')
  }

  if (loading) {
    return <div className="flex items-center justify-center h-full text-gray-500">Loading...</div>
  }

  if (!project) return null

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      {/* Header */}
      <Link to="/" className="inline-flex items-center gap-1 text-sm text-gray-400 hover:text-gray-200">
        <ArrowLeft className="w-4 h-4" /> Dashboard
      </Link>

      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{project.name}</h1>
            <StatusBadge status={project.status} />
          </div>
          <div className="flex items-center gap-4 mt-1 text-sm text-gray-500">
            {project.provider && <span>{project.provider}{project.framework ? ` / ${project.framework}` : ''}</span>}
            {project.git_url && (
              <span className="flex items-center gap-1">
                <GitBranch className="w-3.5 h-3.5" /> {project.branch}
              </span>
            )}
          </div>
          {/* Project URL */}
          {project.status === 'healthy' && (
            <div className="flex items-center gap-2 mt-2">
              {domains.length > 0 ? (
                domains.map(d => (
                  <a key={d.id} href={`https://${d.domain}`} target="_blank" rel="noopener noreferrer"
                    className="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-mono bg-success/10 text-success border border-success/20 rounded-md hover:bg-success/20 transition-colors">
                    <Globe className="w-3 h-3" />{d.domain}<ExternalLink className="w-3 h-3" />
                  </a>
                ))
              ) : (
                <span className="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-mono bg-surface-200 text-gray-400 border border-surface-300 rounded-md">
                  <Globe className="w-3 h-3" />mypaas-{project.name}:3000
                </span>
              )}
            </div>
          )}
        </div>

        <div className="flex items-center gap-2">
          {project.status === 'stopped' ? (
            <button
              onClick={async () => { await startProject(id!); await load() }}
              className="flex items-center gap-2 px-4 py-2 bg-success text-white rounded-lg text-sm font-medium hover:bg-success/80 transition-colors"
            ><Play className="w-4 h-4" /> Start</button>
          ) : (
            <>
              <button
                onClick={async () => { await restartProject(id!); await load() }}
                className="flex items-center gap-2 px-3 py-2 text-gray-400 border border-surface-300 rounded-lg text-sm
                           hover:bg-surface-100 hover:text-gray-200 transition-colors"
              ><RefreshCw className="w-4 h-4" /> Restart</button>
              <button
                onClick={async () => { await stopProject(id!); await load() }}
                className="flex items-center gap-2 px-3 py-2 text-warning border border-warning/30 rounded-lg text-sm
                           hover:bg-warning/10 transition-colors"
              ><Square className="w-3.5 h-3.5" /> Stop</button>
            </>
          )}
          <button
            onClick={handleDeploy}
            disabled={deploying}
            className="flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-lg text-sm font-medium
                       hover:bg-accent-hover disabled:opacity-70 transition-colors"
          >
            {deploying ? (
              <><Loader2 className="w-4 h-4 animate-spin" /> Deploying...</>
            ) : (
              <><Rocket className="w-4 h-4" /> Deploy</>
            )}
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-surface-300">
        {(['deployments', 'env', 'volumes', 'settings'] as Tab[]).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              tab === t
                ? 'border-accent text-accent-hover'
                : 'border-transparent text-gray-400 hover:text-gray-200'
            }`}
          >
            {t === 'deployments' ? `Deployments (${deployments.length})` : t === 'env' ? `Env Vars (${envVars.length})` : t === 'volumes' ? `Volumes (${volumes.length})` : 'Settings'}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {tab === 'deployments' && (() => {
        const activeDeployment = deployments.find(d => d.status === 'healthy')
        const historyDeployments = deployments.filter(d => d !== activeDeployment)
        return (
          <div className="space-y-6">
            {deployments.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                No deployments yet. Click Deploy to get started.
              </div>
            ) : (
              <>
                {/* Active Deployment */}
                {activeDeployment && (
                  <DeploymentRow
                    d={activeDeployment}
                    isActive
                    expanded={expandedDeploy === activeDeployment.id}
                    onToggle={() => setExpandedDeploy(expandedDeploy === activeDeployment.id ? null : activeDeployment.id)}
                    onRollback={handleRollback}
                    onRedeploy={() => handleDeploy()}
                    projectId={project.id}
                  />
                )}
                {/* History */}
                {historyDeployments.length > 0 && (
                  <div>
                    <div className="flex items-center gap-2 mb-3">
                      <span className="text-[11px] font-bold uppercase tracking-wider text-gray-500">History</span>
                      <div className="flex-1 border-t border-surface-300" />
                    </div>
                    <div className="space-y-2">
                      {historyDeployments.map((d) => (
                        <DeploymentRow
                          key={d.id}
                          d={d}
                          isActive={false}
                          expanded={expandedDeploy === d.id}
                          onToggle={() => setExpandedDeploy(expandedDeploy === d.id ? null : d.id)}
                          onRollback={handleRollback}
                          onRedeploy={() => handleDeploy()}
                          projectId={project.id}
                        />
                      ))}
                    </div>
                  </div>
                )}
              </>
            )}
          </div>
        )
      })()}

      {tab === 'env' && (
        <EnvEditor projectId={project.id} envVars={envVars} onUpdate={load} />
      )}

      {tab === 'volumes' && (
        <div className="space-y-4">
          <div className="bg-surface-50 border border-surface-300 rounded-lg p-4">
            <h3 className="text-sm font-medium text-gray-300 mb-3">Add Volume</h3>
            <div className="flex gap-2">
              <input
                value={newVolName}
                onChange={(e) => setNewVolName(e.target.value)}
                placeholder="Volume name (e.g. uploads)"
                className="flex-1 px-3 py-1.5 bg-surface border border-surface-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-accent"
              />
              <input
                value={newVolPath}
                onChange={(e) => setNewVolPath(e.target.value)}
                placeholder="Mount path (e.g. /app/uploads)"
                className="flex-1 px-3 py-1.5 bg-surface border border-surface-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-accent"
              />
              <button
                onClick={handleAddVolume}
                disabled={!newVolName.trim() || !newVolPath.trim()}
                className="flex items-center gap-1 px-3 py-1.5 bg-accent text-white rounded text-sm hover:bg-accent-hover disabled:opacity-50 transition-colors"
              >
                <Plus className="w-4 h-4" /> Add
              </button>
            </div>
            <p className="text-xs text-gray-500 mt-2">
              Volumes persist data across deploys. A Docker volume <code className="text-gray-400">mypaas-{'{name}'}-{'{vol}'}</code> will be created and mounted at the specified path.
            </p>
          </div>

          {volumes.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <Database className="w-8 h-8 mx-auto mb-2 opacity-50" />
              No volumes configured. Add a volume to persist data between deployments.
            </div>
          ) : (
            <div className="space-y-2">
              {volumes.map((v) => (
                <div key={v.id} className="flex items-center justify-between bg-surface-50 border border-surface-300 rounded-lg p-3">
                  <div className="flex items-center gap-3">
                    <Database className="w-4 h-4 text-accent" />
                    <div>
                      <span className="text-sm font-medium">{v.name}</span>
                      <span className="text-xs text-gray-500 ml-2 font-mono">{v.mount_path}</span>
                    </div>
                  </div>
                  <button
                    onClick={() => handleDeleteVolume(v.id)}
                    className="text-danger hover:text-danger/80 p-1"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {tab === 'settings' && (
        <div className="space-y-6">
          {/* Container Stats */}
          {projectStats && (
            <div>
              <h3 className="text-sm font-medium text-gray-300 mb-2">Container Stats</h3>
              <div className="grid grid-cols-2 gap-3">
                <div className="bg-surface-50 border border-surface-300 rounded-lg p-3">
                  <div className="flex items-center gap-2 text-xs text-gray-500 mb-1">
                    <Cpu className="w-3 h-3" /> CPU
                  </div>
                  <div className="text-lg font-bold">{projectStats.cpu_percent.toFixed(1)}%</div>
                </div>
                <div className="bg-surface-50 border border-surface-300 rounded-lg p-3">
                  <div className="flex items-center gap-2 text-xs text-gray-500 mb-1">
                    <HardDrive className="w-3 h-3" /> Memory
                  </div>
                  <div className="text-lg font-bold">
                    {(projectStats.mem_usage / 1024 / 1024).toFixed(1)} MB
                    <span className="text-sm text-gray-500 ml-1">
                      / {(projectStats.mem_limit / 1024 / 1024).toFixed(0)} MB
                    </span>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Webhook URL */}
          {project.git_url && (
            <div>
              <h3 className="text-sm font-medium text-gray-300 mb-2">GitHub Webhook</h3>
              <div className="bg-surface-50 border border-surface-300 rounded-lg p-4">
                <p className="text-xs text-gray-500 mb-2">
                  Add this URL as a webhook in your GitHub repo settings to enable auto-deploy on push.
                </p>
                <div className="flex items-center gap-2">
                  <code className="flex-1 px-3 py-2 bg-surface border border-surface-300 rounded text-xs font-mono truncate">
                    {webhookUrl}
                  </code>
                  <button
                    onClick={copyWebhookUrl}
                    className="p-2 text-gray-400 hover:text-gray-200 border border-surface-300 rounded transition-colors"
                    title="Copy"
                  >
                    {copied ? <Check className="w-4 h-4 text-success" /> : <Copy className="w-4 h-4" />}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Resource Limits */}
          <div>
            <h3 className="text-sm font-medium text-gray-300 mb-2">Resource Limits</h3>
            <div className="bg-surface-50 border border-surface-300 rounded-lg p-4">
              <div className="grid grid-cols-3 gap-4 mb-3">
                <div>
                  <label className="block text-xs text-gray-500 mb-1">CPU (cores, 0 = unlimited)</label>
                  <input
                    type="number"
                    value={cpuLimit}
                    onChange={(e) => setCpuLimit(parseFloat(e.target.value) || 0)}
                    min="0" max="4" step="0.25"
                    className="w-full px-3 py-1.5 bg-surface border border-surface-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-accent"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Memory (MB, 0 = unlimited)</label>
                  <input
                    type="number"
                    value={memLimit}
                    onChange={(e) => setMemLimit(parseInt(e.target.value) || 0)}
                    min="0" step="64"
                    className="w-full px-3 py-1.5 bg-surface border border-surface-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-accent"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Replicas (Swarm, 0 = default)</label>
                  <input
                    type="number"
                    value={replicas}
                    onChange={(e) => setReplicas(parseInt(e.target.value) || 0)}
                    min="0" max="10" step="1"
                    className="w-full px-3 py-1.5 bg-surface border border-surface-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-accent"
                  />
                </div>
              </div>
              <button
                onClick={handleSaveResources}
                className="px-3 py-1.5 bg-accent text-white rounded text-sm hover:bg-accent-hover transition-colors"
              >
                Save Limits
              </button>
            </div>
          </div>

          {/* Custom Domains */}
          <div>
            <h3 className="text-sm font-medium text-gray-300 mb-2">Custom Domains</h3>
            <div className="bg-surface-50 border border-surface-300 rounded-lg p-4 space-y-3">
              {domains.length > 0 && domains.map((d) => (
                <div key={d.id} className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Globe className="w-4 h-4 text-accent" />
                    <span className="text-sm font-mono">{d.domain}</span>
                    {d.ssl_auto && <span className="text-xs text-success bg-success/10 px-1.5 py-0.5 rounded">SSL</span>}
                  </div>
                  <button
                    onClick={() => handleDeleteDomain(d.id)}
                    className="text-danger hover:text-danger/80 p-1"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              ))}
              <div className="flex gap-2">
                <input
                  value={newDomain}
                  onChange={(e) => setNewDomain(e.target.value)}
                  placeholder="app.example.com"
                  className="flex-1 px-3 py-1.5 bg-surface border border-surface-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-accent"
                  onKeyDown={(e) => e.key === 'Enter' && handleAddDomain()}
                />
                <button
                  onClick={handleAddDomain}
                  disabled={!newDomain.trim()}
                  className="px-3 py-1.5 bg-accent text-white rounded text-sm hover:bg-accent-hover disabled:opacity-50 transition-colors"
                >
                  Add
                </button>
              </div>
              <p className="text-xs text-gray-500">
                Point your domain's DNS A record to your server IP. SSL will be auto-provisioned via Let's Encrypt.
              </p>
            </div>
          </div>

          <div>
            <h3 className="text-sm font-medium text-gray-300 mb-2">Project Info</h3>
            <div className="bg-surface-50 border border-surface-300 rounded-lg p-4 text-sm space-y-2">
              <div className="flex justify-between">
                <span className="text-gray-500">ID</span>
                <span className="font-mono">{project.id}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Git URL</span>
                <span className="font-mono truncate ml-4">{project.git_url || '—'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Auto Deploy</span>
                <span>{project.auto_deploy ? 'Enabled' : 'Disabled'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Created</span>
                <span>{new Date(project.created_at).toLocaleString()}</span>
              </div>
            </div>
          </div>

          <div className="border-t border-surface-300 pt-6">
            <h3 className="text-sm font-medium text-danger mb-2">Danger Zone</h3>
            <button
              onClick={() => setShowDeleteModal(true)}
              className="flex items-center gap-2 px-4 py-2 text-sm text-danger border border-danger/30 rounded-lg hover:bg-danger/10 transition-colors"
            >
              <Trash2 className="w-4 h-4" /> Delete Project
            </button>
          </div>

          {showDeleteModal && (
            <DeleteModal
              projectName={project.name}
              onConfirm={handleDelete}
              onCancel={() => setShowDeleteModal(false)}
            />
          )}
        </div>
      )}
    </div>
  )
}
