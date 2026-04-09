import { create } from 'zustand'
import { NODE_W, NODE_H, snap } from '../lib/catalog'

// ─── Types ───────────────────────────────────────────────────────────────────

export type NodeCategory = 'app' | 'database' | 'cache' | 'storage'

export interface CanvasNode {
  id: string
  category: NodeCategory
  serviceType: string
  name: string
  x: number
  y: number
  config: Record<string, string>
  gitUrl: string
  branch: string
  dockerImage: string
}

export interface CanvasConnection {
  id: string
  from: string
  to: string
}

// ─── Store ───────────────────────────────────────────────────────────────────

interface CanvasState {
  nodes: CanvasNode[]
  connections: CanvasConnection[]
  selectedNodeId: string | null
  zoom: number
  panX: number
  panY: number
  phase: 'pick' | 'compose'
}

interface CanvasActions {
  addNode: (partial: Omit<CanvasNode, 'id'>) => string
  removeNode: (id: string) => void
  updateNode: (id: string, patch: Partial<CanvasNode>) => void
  moveNode: (id: string, x: number, y: number) => void
  selectNode: (id: string | null) => void
  addConnection: (from: string, to: string) => void
  removeConnection: (id: string) => void
  setZoom: (zoom: number) => void
  setPan: (x: number, y: number) => void
  setPhase: (phase: 'pick' | 'compose') => void
  reset: () => void
}

const genId = () => Math.random().toString(36).slice(2, 10)

const INITIAL: CanvasState = {
  nodes: [],
  connections: [],
  selectedNodeId: null,
  zoom: 1,
  panX: 0,
  panY: 0,
  phase: 'pick',
}

export const useCanvasStore = create<CanvasState & CanvasActions>((set) => ({
  ...INITIAL,

  addNode: (partial) => {
    const id = genId()
    const node: CanvasNode = {
      ...partial,
      id,
      x: snap(partial.x),
      y: snap(partial.y),
    }
    set((s) => ({ nodes: [...s.nodes, node] }))
    return id
  },

  removeNode: (id) =>
    set((s) => ({
      nodes: s.nodes.filter((n) => n.id !== id),
      connections: s.connections.filter((c) => c.from !== id && c.to !== id),
      selectedNodeId: s.selectedNodeId === id ? null : s.selectedNodeId,
    })),

  updateNode: (id, patch) =>
    set((s) => ({
      nodes: s.nodes.map((n) => (n.id === id ? { ...n, ...patch } : n)),
    })),

  moveNode: (id, x, y) =>
    set((s) => ({
      nodes: s.nodes.map((n) =>
        n.id === id ? { ...n, x: snap(x), y: snap(y) } : n,
      ),
    })),

  selectNode: (id) => set({ selectedNodeId: id }),

  addConnection: (from, to) => {
    set((s) => {
      const dup = s.connections.some(
        (c) => (c.from === from && c.to === to) || (c.from === to && c.to === from),
      )
      if (dup || from === to) return s
      return { connections: [...s.connections, { id: genId(), from, to }] }
    })
  },

  removeConnection: (id) =>
    set((s) => ({ connections: s.connections.filter((c) => c.id !== id) })),

  setZoom: (zoom) => set({ zoom: Math.max(0.3, Math.min(2, zoom)) }),
  setPan: (panX, panY) => set({ panX, panY }),
  setPhase: (phase) => set({ phase }),
  reset: () => set(INITIAL),
}))

// ─── Selectors ───────────────────────────────────────────────────────────────

export const getNodeById = (nodes: CanvasNode[], id: string) =>
  nodes.find((n) => n.id === id)

export const getNodeCenter = (node: CanvasNode) => ({
  x: node.x + NODE_W / 2,
  y: node.y + NODE_H / 2,
})

export const getOutputPort = (node: CanvasNode) => ({
  x: node.x + NODE_W,
  y: node.y + NODE_H / 2,
})

export const getInputPort = (node: CanvasNode) => ({
  x: node.x,
  y: node.y + NODE_H / 2,
})
