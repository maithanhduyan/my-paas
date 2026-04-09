import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  useNodesState,
  useEdgesState,
  MarkerType,
  BackgroundVariant,
  type Node,
  type Edge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { nodeTypes, type ServiceNodeData } from './FlowNodes'
import type { Project, Service, Volume } from '../../types'
import {
  listServices, listVolumes,
  linkService,
} from '../../api'
import { useToast } from '../ui/Toast'
import { Plus } from 'lucide-react'

/* ─── Props ───────────────────────────────────────────────────── */

interface Props {
  project: Project
  onSelectService?: (service: Service | null) => void
  onUpdate: () => void
}

/* ─── Layout constants ────────────────────────────────────────── */

const APP_X = 100
const APP_Y = 160
const SERVICE_X = 480
const SERVICE_START_Y = 60
const SERVICE_GAP_Y = 140
const VOLUME_OFFSET_Y = 100

/* ─── Component ───────────────────────────────────────────────── */

export function ProjectCanvas({ project, onSelectService, onUpdate }: Props) {
  const [services, setServices] = useState<Service[]>([])
  const [volumes, setVolumes] = useState<Volume[]>([])
  const [linkedIds, setLinkedIds] = useState<Set<string>>(new Set())
  const { toast } = useToast()

  // Load services & volumes
  const load = useCallback(async () => {
    try {
      const [allServices, vols] = await Promise.all([
        listServices(),
        listVolumes(project.id).catch(() => [] as Volume[]),
      ])
      setServices(allServices)
      setVolumes(vols)
      // Heuristic: linked if name starts with project prefix
      const projectBase = project.name.replace(/-app$/, '').replace(/-wordpress-app$/, '')
      const linked = new Set<string>(
        allServices
          .filter((s: Service) => s.name.startsWith(projectBase))
          .map((s: Service) => s.id)
      )
      setLinkedIds(linked)
    } catch {
      setServices([])
    }
  }, [project.id, project.name])

  useEffect(() => { load() }, [load])

  const linkedServices = useMemo(
    () => services.filter(s => linkedIds.has(s.id)),
    [services, linkedIds]
  )

  // ─── Build React Flow nodes & edges ─────────────────────────

  const { initialNodes, initialEdges } = useMemo(() => {
    const nodes: Node[] = []
    const edges: Edge[] = []

    // App node
    nodes.push({
      id: 'app',
      type: 'app',
      position: { x: APP_X, y: APP_Y },
      data: {
        label: project.name,
        subtitle: project.git_url
          ? new URL(project.git_url).pathname.replace(/^\//, '').replace(/\.git$/, '')
          : project.provider || undefined,
        status: project.status,
        nodeType: 'app',
        onClick: () => onSelectService?.(null),
      } satisfies ServiceNodeData,
    })

    // Linked service nodes
    linkedServices.forEach((svc, i) => {
      const nodeId = `svc-${svc.id}`
      nodes.push({
        id: nodeId,
        type: 'service',
        position: { x: SERVICE_X, y: SERVICE_START_Y + i * SERVICE_GAP_Y },
        data: {
          label: svc.name,
          subtitle: `${svc.type} · ${svc.image}`,
          status: svc.status,
          nodeType: 'service',
          serviceType: svc.type,
          image: svc.image,
          onClick: () => onSelectService?.(svc),
        } satisfies ServiceNodeData,
      })

      // Edge from app → service
      edges.push({
        id: `e-app-${svc.id}`,
        source: 'app',
        target: nodeId,
        type: 'smoothstep',
        animated: svc.status === 'running',
        style: { stroke: '#d1d5db', strokeWidth: 2 },
        markerEnd: { type: MarkerType.ArrowClosed, color: '#d1d5db', width: 16, height: 16 },
      })

      // Volume nodes under service (if service has volumes)
      const svcVolumes = volumes.filter(v =>
        v.name.toLowerCase().includes(svc.type) ||
        v.name.toLowerCase().includes(svc.name.split('-').pop() || '')
      )
      svcVolumes.forEach((vol, vi) => {
        const volId = `vol-${vol.id}`
        nodes.push({
          id: volId,
          type: 'volume',
          position: {
            x: SERVICE_X + 20,
            y: SERVICE_START_Y + i * SERVICE_GAP_Y + VOLUME_OFFSET_Y + vi * 80,
          },
          data: {
            label: vol.name,
            subtitle: vol.mount_path,
            status: 'empty',
            nodeType: 'volume',
          } satisfies ServiceNodeData,
        })
        edges.push({
          id: `e-vol-${vol.id}`,
          source: nodeId,
          target: volId,
          type: 'smoothstep',
          style: { stroke: '#e5e7eb', strokeWidth: 1.5, strokeDasharray: '4 4' },
        })
      })
    })

    return { initialNodes: nodes, initialEdges: edges }
  }, [project, linkedServices, volumes, onSelectService])

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)

  // Sync when data changes
  useEffect(() => {
    setNodes(initialNodes)
    setEdges(initialEdges)
  }, [initialNodes, initialEdges, setNodes, setEdges])

  // ─── Add service button (+ Add) ────────────────────────────

  const unlinkedServices = useMemo(
    () => services.filter(s => !linkedIds.has(s.id)),
    [services, linkedIds]
  )

  const [showAddMenu, setShowAddMenu] = useState(false)

  const handleLinkService = async (serviceId: string) => {
    try {
      await linkService(serviceId, project.id)
      setLinkedIds(prev => new Set([...prev, serviceId]))
      toast({ type: 'success', title: 'Service linked' })
      await load()
      onUpdate()
    } catch (e: any) {
      toast({ type: 'error', title: 'Link failed', description: e.message })
    }
    setShowAddMenu(false)
  }

  // ─── Drop handler for drag-and-drop ─────────────────────────

  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'link'
  }, [])

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      const serviceId = e.dataTransfer.getData('text/plain')
      if (serviceId && !linkedIds.has(serviceId)) {
        handleLinkService(serviceId)
      }
    },
    [linkedIds],
  )

  return (
    <div className="relative w-full h-full">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        onDragOver={onDragOver}
        onDrop={onDrop}
        fitView
        fitViewOptions={{ padding: 0.3 }}
        proOptions={{ hideAttribution: true }}
        minZoom={0.3}
        maxZoom={2}
        defaultEdgeOptions={{
          type: 'smoothstep',
          style: { stroke: '#d1d5db', strokeWidth: 2 },
        }}
      >
        <Background variant={BackgroundVariant.Dots} gap={24} size={1} color="rgba(0,0,0,0.06)" />
        <Controls
          showInteractive={false}
          className="!bg-white !border-gray-200 !shadow-sm !rounded-lg [&>button]:!bg-white [&>button]:!border-gray-200 [&>button]:!text-gray-500 [&>button:hover]:!bg-gray-50"
        />
      </ReactFlow>

      {/* + Add button (top-right of canvas) */}
      <div className="absolute top-4 right-4 z-10">
        <div className="relative">
          <button
            onClick={() => setShowAddMenu(!showAddMenu)}
            className="flex items-center gap-1.5 px-3 py-2 bg-white text-gray-700 border border-gray-200
                       rounded-lg shadow-sm text-sm font-medium hover:bg-gray-50 transition-colors"
          >
            <Plus className="w-4 h-4" /> Add
          </button>

          {showAddMenu && (
            <>
              <div className="fixed inset-0 z-10" onClick={() => setShowAddMenu(false)} />
              <div className="absolute right-0 mt-2 w-56 bg-white border border-gray-200 rounded-xl
                              shadow-lg py-2 z-20 max-h-72 overflow-y-auto">
                {unlinkedServices.length > 0 ? (
                  <>
                    <div className="px-3 py-1.5 text-[11px] font-semibold text-gray-400 uppercase tracking-wider">
                      Available Services
                    </div>
                    {unlinkedServices.map(svc => (
                      <button
                        key={svc.id}
                        onClick={() => handleLinkService(svc.id)}
                        className="w-full flex items-center gap-2.5 px-3 py-2 text-sm text-gray-700
                                   hover:bg-gray-50 transition-colors text-left"
                      >
                        <span className="text-base">
                          {({ postgres: '🐘', mysql: '🐬', redis: '⚡', mongo: '🍃', minio: '🪣' } as Record<string, string>)[svc.type] || '📦'}
                        </span>
                        <div className="min-w-0">
                          <div className="font-medium truncate">{svc.name}</div>
                          <div className="text-[11px] text-gray-400">{svc.type} · {svc.status}</div>
                        </div>
                      </button>
                    ))}
                  </>
                ) : (
                  <div className="px-3 py-4 text-sm text-gray-400 text-center">
                    No unlinked services available
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
