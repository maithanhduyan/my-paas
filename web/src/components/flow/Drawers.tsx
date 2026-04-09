import { useState } from 'react'
import {
  X, Play, Square, Unlink, Globe,
  Cpu, HardDrive, ExternalLink, Copy, Check,
} from 'lucide-react'
import type { Project, Service, ContainerStats, Domain, Volume } from '../../types'
import { EnvEditor } from '../EnvEditor'
import type { EnvVar, Deployment } from '../../types'

/* ─── Props ───────────────────────────────────────────────────── */

interface ServiceDrawerProps {
  service: Service
  project: Project
  onClose: () => void
  onStart: () => void
  onStop: () => void
  onUnlink: () => void
}

interface AppDrawerProps {
  project: Project
  deployments: Deployment[]
  envVars: EnvVar[]
  stats: ContainerStats | null
  domains: Domain[]
  volumes: Volume[]
  onClose: () => void
  onUpdate: () => void
  onRestart: () => void
  onStop: () => void
  onStart: () => void
  onDeploy: () => void
  onViewLogs: (id: string) => void
}

/* ─── Shared wrapper ──────────────────────────────────────────── */

function DrawerShell({ title, subtitle, onClose, children }: {
  title: string; subtitle?: string; onClose: () => void; children: React.ReactNode
}) {
  return (
    <div className="fixed inset-0 z-50 flex justify-end" onClick={onClose}>
      <div className="absolute inset-0 bg-black/20" />
      <div
        className="relative w-full max-w-lg bg-white h-full shadow-2xl overflow-y-auto
                   animate-in slide-in-from-right-4"
        onClick={e => e.stopPropagation()}
      >
        <div className="sticky top-0 bg-white border-b border-gray-100 px-6 py-4 flex items-center justify-between z-10">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
            {subtitle && <p className="text-sm text-gray-400 mt-0.5">{subtitle}</p>}
          </div>
          <button onClick={onClose} className="p-1.5 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-100">
            <X className="w-5 h-5" />
          </button>
        </div>
        <div className="p-6 space-y-6">
          {children}
        </div>
      </div>
    </div>
  )
}

/* ─── Service Drawer ──────────────────────────────────────────── */

const SERVICE_ICONS: Record<string, string> = {
  postgres: '🐘', mysql: '🐬', redis: '⚡', mongo: '🍃', minio: '🪣',
}

export function ServiceDrawer({ service, onClose, onStart, onStop, onUnlink }: ServiceDrawerProps) {
  const icon = SERVICE_ICONS[service.type] || '📦'
  const isRunning = service.status === 'running'

  return (
    <DrawerShell title={service.name} subtitle={`${service.type} · ${service.image}`} onClose={onClose}>
      {/* Status card */}
      <div className="bg-gray-50 rounded-xl p-4 border border-gray-100">
        <div className="flex items-center gap-3">
          <span className="text-2xl">{icon}</span>
          <div>
            <div className="font-semibold text-gray-900">{service.name}</div>
            <div className="text-sm text-gray-500">{service.image}</div>
          </div>
          <div className="ml-auto flex items-center gap-1.5">
            <span className={`w-2 h-2 rounded-full ${isRunning ? 'bg-emerald-500' : 'bg-gray-400'}`} />
            <span className="text-sm text-gray-600 capitalize">{service.status}</span>
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-2">
        {isRunning ? (
          <button onClick={onStop}
            className="flex items-center gap-2 px-4 py-2 text-sm text-amber-700 bg-amber-50 border border-amber-200 rounded-lg hover:bg-amber-100 transition-colors">
            <Square className="w-3.5 h-3.5" /> Stop
          </button>
        ) : (
          <button onClick={onStart}
            className="flex items-center gap-2 px-4 py-2 text-sm text-emerald-700 bg-emerald-50 border border-emerald-200 rounded-lg hover:bg-emerald-100 transition-colors">
            <Play className="w-3.5 h-3.5" /> Start
          </button>
        )}
        <button onClick={onUnlink}
          className="flex items-center gap-2 px-4 py-2 text-sm text-gray-600 bg-gray-50 border border-gray-200 rounded-lg hover:bg-gray-100 transition-colors">
          <Unlink className="w-3.5 h-3.5" /> Unlink
        </button>
      </div>

      {/* Info */}
      <div>
        <h3 className="text-sm font-medium text-gray-700 mb-2">Details</h3>
        <div className="bg-gray-50 rounded-xl p-4 border border-gray-100 text-sm space-y-2">
          <div className="flex justify-between"><span className="text-gray-500">ID</span><span className="font-mono text-gray-700">{service.id}</span></div>
          <div className="flex justify-between"><span className="text-gray-500">Type</span><span className="text-gray-700">{service.type}</span></div>
          <div className="flex justify-between"><span className="text-gray-500">Image</span><span className="font-mono text-gray-700">{service.image}</span></div>
          <div className="flex justify-between"><span className="text-gray-500">Container</span><span className="font-mono text-gray-700 truncate ml-4">{service.container_id || '—'}</span></div>
          <div className="flex justify-between"><span className="text-gray-500">Created</span><span className="text-gray-700">{new Date(service.created_at).toLocaleString()}</span></div>
        </div>
      </div>
    </DrawerShell>
  )
}

/* ─── App Drawer (project details/settings) ───────────────────── */

type AppTab = 'deployments' | 'variables' | 'settings'

export function AppDrawer({
  project, deployments, envVars, stats, domains,
  onClose, onUpdate, onViewLogs,
}: AppDrawerProps) {
  const [tab, setTab] = useState<AppTab>('deployments')
  const [copied, setCopied] = useState(false)
  const webhookUrl = `${window.location.origin}/api/webhooks/github`
  const copyUrl = () => { navigator.clipboard.writeText(webhookUrl); setCopied(true); setTimeout(() => setCopied(false), 2000) }

  const tabStyle = (t: AppTab) =>
    `px-3 py-1.5 text-sm font-medium rounded-lg transition-colors ${
      tab === t ? 'bg-gray-900 text-white' : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
    }`

  return (
    <DrawerShell title={project.name} subtitle={project.provider ? `${project.provider}${project.framework ? ` / ${project.framework}` : ''}` : undefined} onClose={onClose}>
      {/* Tabs */}
      <div className="flex gap-1 p-1 bg-gray-100 rounded-xl">
        <button onClick={() => setTab('deployments')} className={tabStyle('deployments')}>Deployments</button>
        <button onClick={() => setTab('variables')} className={tabStyle('variables')}>Variables</button>
        <button onClick={() => setTab('settings')} className={tabStyle('settings')}>Settings</button>
      </div>

      {tab === 'deployments' && (
        <div className="space-y-3">
          {deployments.length === 0 ? (
            <div className="text-center py-8 text-gray-400 text-sm">No deployments yet</div>
          ) : (
            deployments.slice(0, 10).map(d => (
              <button
                key={d.id}
                onClick={() => onViewLogs(d.id)}
                className="w-full flex items-center gap-3 p-3 bg-gray-50 border border-gray-100 rounded-xl
                           hover:bg-gray-100 transition-colors text-left"
              >
                <div className={`px-2 py-0.5 text-[10px] font-bold uppercase rounded ${
                  d.status === 'healthy' ? 'bg-emerald-100 text-emerald-700' :
                  d.status === 'failed' ? 'bg-red-100 text-red-700' :
                  'bg-blue-100 text-blue-700'
                }`}>
                  {d.status === 'healthy' ? 'Active' : d.status}
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-sm text-gray-700 truncate">
                    {d.commit_msg ? d.commit_msg.split('\n')[0] : d.commit_hash?.slice(0, 7) || 'Deploy'}
                  </div>
                  <div className="text-[11px] text-gray-400">{new Date(d.created_at).toLocaleString()}</div>
                </div>
              </button>
            ))
          )}
        </div>
      )}

      {tab === 'variables' && (
        <EnvEditor projectId={project.id} envVars={envVars} onUpdate={onUpdate} />
      )}

      {tab === 'settings' && (
        <div className="space-y-5">
          {stats && (
            <div>
              <h3 className="text-sm font-medium text-gray-700 mb-2">Container Stats</h3>
              <div className="grid grid-cols-2 gap-3">
                <div className="bg-gray-50 border border-gray-100 rounded-xl p-3">
                  <div className="flex items-center gap-1.5 text-xs text-gray-500 mb-1"><Cpu className="w-3 h-3" /> CPU</div>
                  <div className="text-lg font-bold text-gray-900">{stats.cpu_percent.toFixed(1)}%</div>
                </div>
                <div className="bg-gray-50 border border-gray-100 rounded-xl p-3">
                  <div className="flex items-center gap-1.5 text-xs text-gray-500 mb-1"><HardDrive className="w-3 h-3" /> Memory</div>
                  <div className="text-lg font-bold text-gray-900">{(stats.mem_usage / 1024 / 1024).toFixed(1)} MB</div>
                </div>
              </div>
            </div>
          )}

          {project.git_url && (
            <div>
              <h3 className="text-sm font-medium text-gray-700 mb-2">Webhook</h3>
              <div className="flex items-center gap-2">
                <code className="flex-1 px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-xs font-mono truncate text-gray-600">{webhookUrl}</code>
                <button onClick={copyUrl} className="p-2 text-gray-400 hover:text-gray-600 border border-gray-200 rounded-lg hover:bg-gray-50">
                  {copied ? <Check className="w-4 h-4 text-emerald-500" /> : <Copy className="w-4 h-4" />}
                </button>
              </div>
            </div>
          )}

          <div>
            <h3 className="text-sm font-medium text-gray-700 mb-2">Project Info</h3>
            <div className="bg-gray-50 border border-gray-100 rounded-xl p-4 text-sm space-y-2">
              <div className="flex justify-between"><span className="text-gray-500">ID</span><span className="font-mono text-gray-700">{project.id}</span></div>
              <div className="flex justify-between"><span className="text-gray-500">Git URL</span><span className="font-mono text-gray-700 truncate ml-4">{project.git_url || '—'}</span></div>
              <div className="flex justify-between"><span className="text-gray-500">Branch</span><span className="text-gray-700">{project.branch || 'main'}</span></div>
              <div className="flex justify-between"><span className="text-gray-500">Auto Deploy</span><span className="text-gray-700">{project.auto_deploy ? 'Yes' : 'No'}</span></div>
              <div className="flex justify-between"><span className="text-gray-500">Created</span><span className="text-gray-700">{new Date(project.created_at).toLocaleString()}</span></div>
            </div>
          </div>

          {domains.length > 0 && (
            <div>
              <h3 className="text-sm font-medium text-gray-700 mb-2">Domains</h3>
              <div className="space-y-1.5">
                {domains.map(d => (
                  <a key={d.id} href={`https://${d.domain}`} target="_blank" rel="noopener noreferrer"
                    className="flex items-center gap-2 px-3 py-2 bg-gray-50 border border-gray-100 rounded-lg text-sm hover:bg-gray-100 transition-colors">
                    <Globe className="w-3.5 h-3.5 text-emerald-500" />
                    <span className="font-mono text-gray-700">{d.domain}</span>
                    <ExternalLink className="w-3 h-3 text-gray-400 ml-auto" />
                  </a>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </DrawerShell>
  )
}
