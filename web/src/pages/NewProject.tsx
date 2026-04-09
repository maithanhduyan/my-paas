import { useEffect, useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Link } from 'react-router-dom'
import { ArrowLeft, Rocket, RotateCcw, Loader2 } from 'lucide-react'
import { useCanvasStore } from '../store/canvasStore'
import { InfraCanvas } from '../components/canvas/InfraCanvas'
import { SourcePicker } from '../components/creation/SourcePicker'
import { ServiceCatalog } from '../components/creation/ServiceCatalog'
import { ConfigPanel } from '../components/creation/ConfigPanel'
import {
  createProject, triggerDeploy,
  createService, startService, linkService,
} from '../api'
import type { ServiceType } from '../types'

export function NewProject() {
  const navigate = useNavigate()
  const { nodes, connections, phase, selectedNodeId, reset } = useCanvasStore()
  const [deploying, setDeploying] = useState(false)
  const [deployLog, setDeployLog] = useState<string[]>([])
  const [error, setError] = useState('')

  // Reset canvas when mounting
  useEffect(() => {
    reset()
    return () => { reset() }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // Keyboard shortcuts
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Delete' || e.key === 'Backspace') {
        const { selectedNodeId, removeNode, selectNode } = useCanvasStore.getState()
        const active = document.activeElement
        if (active && (active.tagName === 'INPUT' || active.tagName === 'TEXTAREA')) return
        if (selectedNodeId) { removeNode(selectedNodeId); selectNode(null) }
      }
      if (e.key === 'Escape') {
        useCanvasStore.getState().selectNode(null)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])

  // ─── Deploy: create all resources via API ──────────────────────
  const handleDeploy = useCallback(async () => {
    // Validate
    const appNodes = nodes.filter((n) => n.category === 'app')
    const svcNodes = nodes.filter((n) => n.category !== 'app')

    if (nodes.length === 0) { setError('Add at least one service to deploy'); return }

    for (const n of appNodes) {
      if (n.serviceType === 'github' && !n.gitUrl) {
        useCanvasStore.getState().selectNode(n.id)
        setError(`"${n.name}" needs a Git repository URL`)
        return
      }
      if (n.serviceType === 'docker-image' && !n.dockerImage) {
        useCanvasStore.getState().selectNode(n.id)
        setError(`"${n.name}" needs a Docker image`)
        return
      }
    }

    setDeploying(true)
    setError('')
    setDeployLog([])
    const log = (msg: string) => setDeployLog((prev) => [...prev, msg])

    try {
      // Map node ID → created resource ID
      const projectIds = new Map<string, string>()
      const serviceIds = new Map<string, string>()

      // 1. Create projects (app nodes)
      for (const n of appNodes) {
        log(`Creating project "${n.name}"...`)
        const p = await createProject({
          name: n.name,
          git_url: n.gitUrl,
          branch: n.branch || 'main',
        })
        projectIds.set(n.id, p.id)
        log(`✓ Project "${n.name}" created`)
      }

      // 2. Create backing services
      for (const n of svcNodes) {
        log(`Creating service "${n.name}" (${n.serviceType})...`)
        const svc = await createService(n.name, n.serviceType as ServiceType)
        serviceIds.set(n.id, svc.id)
        log(`✓ Service "${n.name}" created`)

        log(`Starting service "${n.name}"...`)
        await startService(svc.id)
        log(`✓ Service "${n.name}" started`)
      }

      // 3. Link services to projects based on connections
      for (const conn of connections) {
        const fromProject = projectIds.has(conn.from)
        const fromService = serviceIds.has(conn.from)
        const toProject = projectIds.has(conn.to)
        const toService = serviceIds.has(conn.to)

        let projectId: string | undefined
        let serviceId: string | undefined

        if (fromProject && toService) {
          projectId = projectIds.get(conn.from)
          serviceId = serviceIds.get(conn.to)
        } else if (fromService && toProject) {
          projectId = projectIds.get(conn.to)
          serviceId = serviceIds.get(conn.from)
        }

        if (projectId && serviceId) {
          log(`Linking service → project...`)
          await linkService(serviceId, projectId)
          log(`✓ Linked`)
        }
      }

      // 4. Deploy projects
      for (const n of appNodes) {
        const pid = projectIds.get(n.id)
        if (!pid) continue
        log(`Deploying "${n.name}"...`)
        await triggerDeploy(pid)
        log(`✓ Deployment triggered for "${n.name}"`)
      }

      log('🎉 All resources deployed!')

      // Navigate to first project or dashboard
      const firstProjectId = [...projectIds.values()][0]
      setTimeout(() => {
        navigate(firstProjectId ? `/projects/${firstProjectId}` : '/')
      }, 1200)
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Deployment failed'
      log(`✗ Error: ${msg}`)
      setError(msg)
      setDeploying(false)
    }
  }, [nodes, connections, navigate])

  const appCount = nodes.filter((n) => n.category === 'app').length
  const svcCount = nodes.filter((n) => n.category !== 'app').length

  return (
    <div className="flex flex-col h-full -mt-0 lg:-mt-0">
      {/* ─── Toolbar ─────────────────────────────────────────────── */}
      {phase === 'compose' && (
        <div className="flex items-center justify-between px-4 py-2 bg-surface-50 border-b border-surface-300 shrink-0 z-10">
          <div className="flex items-center gap-3">
            <Link
              to="/"
              className="flex items-center gap-1.5 text-xs text-gray-400 hover:text-gray-200 transition-colors"
            >
              <ArrowLeft className="w-3.5 h-3.5" />
              Dashboard
            </Link>
            <div className="w-px h-4 bg-surface-300" />
            <h1 className="text-sm font-semibold">Infrastructure Composer</h1>
            {nodes.length > 0 && (
              <div className="flex items-center gap-2 text-[11px] text-gray-500">
                <span>{appCount} app{appCount !== 1 ? 's' : ''}</span>
                <span>·</span>
                <span>{svcCount} service{svcCount !== 1 ? 's' : ''}</span>
                <span>·</span>
                <span>{connections.length} link{connections.length !== 1 ? 's' : ''}</span>
              </div>
            )}
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={reset}
              disabled={deploying}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-gray-400
                         hover:text-white hover:bg-surface-200 rounded-lg transition-colors
                         disabled:opacity-40"
            >
              <RotateCcw className="w-3.5 h-3.5" />
              Reset
            </button>
            <button
              onClick={handleDeploy}
              disabled={deploying || nodes.length === 0}
              className="flex items-center gap-2 px-4 py-1.5 bg-accent text-white text-xs font-medium
                         rounded-lg hover:bg-accent-hover disabled:opacity-40 transition-colors"
            >
              {deploying ? (
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
              ) : (
                <Rocket className="w-3.5 h-3.5" />
              )}
              {deploying ? 'Deploying...' : 'Deploy'}
            </button>
          </div>
        </div>
      )}

      {/* ─── Main area ───────────────────────────────────────────── */}
      <div className="flex flex-1 min-h-0 relative">
        {/* Source picker overlay */}
        {phase === 'pick' && <SourcePicker />}

        {/* Compose phase: 3-column layout */}
        {phase === 'compose' && <ServiceCatalog />}

        {/* Canvas (always rendered, visible in compose phase) */}
        <div className={`flex-1 relative ${phase === 'pick' ? 'opacity-30 pointer-events-none' : ''}`}>
          <InfraCanvas />
        </div>

        {/* Config panel (when node selected) */}
        {phase === 'compose' && selectedNodeId && <ConfigPanel />}
      </div>

      {/* ─── Deploy log overlay ──────────────────────────────────── */}
      {deploying && deployLog.length > 0 && (
        <div className="absolute bottom-4 right-4 w-80 max-h-64 bg-surface-50/95 backdrop-blur-sm
                        border border-surface-300 rounded-xl shadow-2xl shadow-black/40 z-30 overflow-hidden">
          <div className="px-3 py-2 border-b border-surface-300 text-xs font-semibold text-gray-300">
            Deploy Progress
          </div>
          <div className="p-3 overflow-y-auto max-h-48 space-y-1">
            {deployLog.map((line, i) => (
              <div key={i} className="text-[11px] font-mono text-gray-400">
                {line}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ─── Error toast ─────────────────────────────────────────── */}
      {error && !deploying && (
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 z-30">
          <div className="flex items-center gap-2 px-4 py-2 bg-danger/15 border border-danger/30 rounded-lg text-sm text-danger">
            <span>{error}</span>
            <button onClick={() => setError('')} className="text-danger/60 hover:text-danger">✕</button>
          </div>
        </div>
      )}
    </div>
  )
}
