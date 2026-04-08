import { useCallback, useEffect, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import {
  getProject, listDeployments, getEnvVars,
  triggerDeploy, deleteProject, rollbackDeployment,
  getProjectStats, listDomains, addDomain, deleteDomain, updateProject
} from '../api'
import type { Project, Deployment, EnvVar, ContainerStats, Domain } from '../types'
import { StatusBadge } from '../components/StatusBadge'
import { LogViewer } from '../components/LogViewer'
import { EnvEditor } from '../components/EnvEditor'
import { timeAgo, duration } from '../lib/utils'
import {
  ArrowLeft, Rocket, Trash2, GitBranch, Clock,
  ChevronDown, ChevronUp, RotateCcw, Cpu, HardDrive, Copy, Check, Globe
} from 'lucide-react'

type Tab = 'deployments' | 'env' | 'settings'

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
      setExpandedDeploy(dep.id)
      setTab('deployments')
      await load()
    } finally {
      setDeploying(false)
    }
  }

  const handleDelete = async () => {
    if (!id || !confirm('Delete this project? This cannot be undone.')) return
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
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={handleDeploy}
            disabled={deploying}
            className="flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-lg text-sm font-medium
                       hover:bg-accent-hover disabled:opacity-50 transition-colors"
          >
            <Rocket className="w-4 h-4" /> {deploying ? 'Deploying...' : 'Deploy'}
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-surface-300">
        {(['deployments', 'env', 'settings'] as Tab[]).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              tab === t
                ? 'border-accent text-accent-hover'
                : 'border-transparent text-gray-400 hover:text-gray-200'
            }`}
          >
            {t === 'deployments' ? `Deployments (${deployments.length})` : t === 'env' ? `Env Vars (${envVars.length})` : 'Settings'}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {tab === 'deployments' && (
        <div className="space-y-3">
          {deployments.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No deployments yet. Click Deploy to get started.
            </div>
          ) : (
            deployments.map((d) => (
              <div key={d.id} className="bg-surface-50 border border-surface-300 rounded-lg overflow-hidden">
                <button
                  onClick={() => setExpandedDeploy(expandedDeploy === d.id ? null : d.id)}
                  className="w-full flex items-center gap-3 p-3 text-left hover:bg-surface-100 transition-colors"
                >
                  <StatusBadge status={d.status} />
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
                    {expandedDeploy === d.id ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                  </div>
                </button>

                {expandedDeploy === d.id && (
                  <div className="border-t border-surface-300 p-3">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-xs text-gray-500">Deployment Logs</span>
                      {d.status === 'healthy' && d.image_tag && (
                        <button
                          onClick={(e) => { e.stopPropagation(); handleRollback(d.id) }}
                          className="flex items-center gap-1 px-2 py-1 text-xs text-warning border border-warning/30 rounded hover:bg-warning/10 transition-colors"
                        >
                          <RotateCcw className="w-3 h-3" /> Rollback to this
                        </button>
                      )}
                    </div>
                    <LogViewer deploymentId={d.id} />
                  </div>
                )}
              </div>
            ))
          )}
        </div>
      )}

      {tab === 'env' && (
        <EnvEditor projectId={project.id} envVars={envVars} onUpdate={load} />
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
              onClick={handleDelete}
              className="flex items-center gap-2 px-4 py-2 text-sm text-danger border border-danger/30 rounded-lg hover:bg-danger/10 transition-colors"
            >
              <Trash2 className="w-4 h-4" /> Delete Project
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
