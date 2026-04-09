import { useRef, useState, useCallback } from 'react'
import { useCanvasStore, getOutputPort, getInputPort } from '../../store/canvasStore'
import { SERVICE_DEFS, NODE_W, NODE_H } from '../../lib/catalog'
import { ServiceNode } from './ServiceNode'
import { ConnectionLine } from './ConnectionLine'
import { Minus, Plus, Maximize2 } from 'lucide-react'

export function InfraCanvas() {
  const containerRef = useRef<HTMLDivElement>(null)
  const {
    nodes, connections, zoom, panX, panY, selectedNodeId,
    moveNode, selectNode, setZoom, setPan, addNode, addConnection,
  } = useCanvasStore()

  // ─── Local interaction state ─────────────────────────────────
  const [dragState, setDragState] = useState<{
    nodeId: string; offsetX: number; offsetY: number
  } | null>(null)
  const [panState, setPanState] = useState<{
    startX: number; startY: number
  } | null>(null)
  const [connectState, setConnectState] = useState<{
    fromNodeId: string; mouseX: number; mouseY: number
  } | null>(null)

  // ─── Coordinate helpers ──────────────────────────────────────
  const screenToCanvas = useCallback(
    (clientX: number, clientY: number) => {
      const rect = containerRef.current?.getBoundingClientRect()
      if (!rect) return { x: 0, y: 0 }
      return {
        x: (clientX - rect.left - panX) / zoom,
        y: (clientY - rect.top - panY) / zoom,
      }
    },
    [zoom, panX, panY],
  )

  // ─── Mouse handlers ──────────────────────────────────────────
  const handleWheel = useCallback(
    (e: React.WheelEvent) => {
      e.preventDefault()
      const rect = containerRef.current?.getBoundingClientRect()
      if (!rect) return
      const delta = e.deltaY > 0 ? -0.08 : 0.08
      const newZoom = Math.max(0.3, Math.min(2, zoom + delta))
      const mx = e.clientX - rect.left
      const my = e.clientY - rect.top
      setPan(
        mx - (mx - panX) * (newZoom / zoom),
        my - (my - panY) * (newZoom / zoom),
      )
      setZoom(newZoom)
    },
    [zoom, panX, panY, setZoom, setPan],
  )

  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      // Only pan on direct canvas click
      const el = e.target as HTMLElement
      if (el === containerRef.current || el.classList.contains('canvas-grid')) {
        selectNode(null)
        setPanState({ startX: e.clientX - panX, startY: e.clientY - panY })
      }
    },
    [panX, panY, selectNode],
  )

  const handleMouseMove = useCallback(
    (e: React.MouseEvent) => {
      if (panState) {
        setPan(e.clientX - panState.startX, e.clientY - panState.startY)
      }
      if (dragState) {
        const c = screenToCanvas(e.clientX, e.clientY)
        moveNode(dragState.nodeId, c.x - dragState.offsetX, c.y - dragState.offsetY)
      }
      if (connectState) {
        const c = screenToCanvas(e.clientX, e.clientY)
        setConnectState((s) => s ? { ...s, mouseX: c.x, mouseY: c.y } : null)
      }
    },
    [panState, dragState, connectState, screenToCanvas, setPan, moveNode],
  )

  const handleMouseUp = useCallback(() => {
    setPanState(null)
    setDragState(null)
    setConnectState(null)
  }, [])

  // ─── Node interaction callbacks ──────────────────────────────
  const handleNodeMouseDown = useCallback(
    (e: React.MouseEvent, nodeId: string) => {
      const c = screenToCanvas(e.clientX, e.clientY)
      const node = nodes.find((n) => n.id === nodeId)
      if (!node) return
      setDragState({
        nodeId,
        offsetX: c.x - node.x,
        offsetY: c.y - node.y,
      })
    },
    [nodes, screenToCanvas],
  )

  const handlePortMouseDown = useCallback(
    (e: React.MouseEvent, nodeId: string) => {
      const c = screenToCanvas(e.clientX, e.clientY)
      setConnectState({ fromNodeId: nodeId, mouseX: c.x, mouseY: c.y })
    },
    [screenToCanvas],
  )

  const handlePortMouseUp = useCallback(
    (_e: React.MouseEvent, nodeId: string) => {
      if (connectState && connectState.fromNodeId !== nodeId) {
        addConnection(connectState.fromNodeId, nodeId)
      }
      setConnectState(null)
    },
    [connectState, addConnection],
  )

  // ─── Drop from catalog ───────────────────────────────────────
  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'copy'
  }, [])

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      const serviceType = e.dataTransfer.getData('application/x-service-type')
      if (!serviceType) return
      const def = SERVICE_DEFS[serviceType]
      if (!def) return
      const c = screenToCanvas(e.clientX, e.clientY)
      addNode({
        category: def.category,
        serviceType,
        name: def.defaultName,
        x: c.x - NODE_W / 2,
        y: c.y - NODE_H / 2,
        config: { ...def.defaultConfig },
        gitUrl: '',
        branch: 'main',
        dockerImage: '',
      })
    },
    [screenToCanvas, addNode],
  )

  // ─── Zoom controls ───────────────────────────────────────────
  const resetView = () => { setZoom(1); setPan(0, 0) }

  // ─── Grid bg ─────────────────────────────────────────────────
  const gs = 24 * zoom
  const gridBg = {
    backgroundImage: `radial-gradient(circle, rgba(255,255,255,0.04) 1px, transparent 1px)`,
    backgroundSize: `${gs}px ${gs}px`,
    backgroundPosition: `${panX % gs}px ${panY % gs}px`,
  }

  const cursor = panState
    ? 'cursor-grabbing'
    : dragState
      ? 'cursor-grabbing'
      : connectState
        ? 'cursor-crosshair'
        : 'cursor-grab'

  return (
    <div
      ref={containerRef}
      className={`canvas-grid relative w-full h-full overflow-hidden select-none ${cursor}`}
      style={gridBg}
      onWheel={handleWheel}
      onMouseDown={handleMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
      onDragOver={handleDragOver}
      onDrop={handleDrop}
    >
      {/* SVG connections layer */}
      <svg
        className="absolute inset-0 w-full h-full pointer-events-none"
        style={{ zIndex: 1 }}
      >
        <g transform={`translate(${panX}, ${panY}) scale(${zoom})`}>
          {connections.map((conn) => {
            const from = nodes.find((n) => n.id === conn.from)
            const to = nodes.find((n) => n.id === conn.to)
            if (!from || !to) return null
            const p1 = getOutputPort(from)
            const p2 = getInputPort(to)
            const active = selectedNodeId === conn.from || selectedNodeId === conn.to
            return (
              <ConnectionLine
                key={conn.id}
                x1={p1.x} y1={p1.y}
                x2={p2.x} y2={p2.y}
                active={active}
              />
            )
          })}
          {connectState && (() => {
            const from = nodes.find((n) => n.id === connectState.fromNodeId)
            if (!from) return null
            const p1 = getOutputPort(from)
            return (
              <ConnectionLine
                x1={p1.x} y1={p1.y}
                x2={connectState.mouseX} y2={connectState.mouseY}
                preview
              />
            )
          })()}
        </g>
      </svg>

      {/* Nodes layer */}
      <div
        className="absolute"
        style={{
          transform: `translate(${panX}px, ${panY}px) scale(${zoom})`,
          transformOrigin: '0 0',
          zIndex: 2,
        }}
      >
        {nodes.map((node) => (
          <ServiceNode
            key={node.id}
            node={node}
            selected={selectedNodeId === node.id}
            nodeWidth={NODE_W}
            isConnectTarget={!!connectState && connectState.fromNodeId !== node.id}
            onSelect={() => selectNode(node.id)}
            onNodeMouseDown={handleNodeMouseDown}
            onPortMouseDown={handlePortMouseDown}
            onPortMouseUp={handlePortMouseUp}
          />
        ))}
      </div>

      {/* Empty state */}
      {nodes.length === 0 && (
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none z-0">
          <div className="text-center">
            <div className="text-5xl mb-3 opacity-20">🏗️</div>
            <p className="text-gray-500 text-sm">Drag services from the sidebar to start building</p>
            <p className="text-gray-600 text-xs mt-1">Or use a quick-start template</p>
          </div>
        </div>
      )}

      {/* Zoom controls */}
      <div className="absolute bottom-4 left-4 flex items-center gap-1 bg-surface-100/90 backdrop-blur-sm border border-surface-300 rounded-lg p-1 z-10">
        <button
          onClick={() => setZoom(zoom - 0.1)}
          className="p-1.5 text-gray-500 hover:text-white rounded transition-colors"
        >
          <Minus className="w-3.5 h-3.5" />
        </button>
        <span className="text-[11px] text-gray-400 w-10 text-center tabular-nums">
          {Math.round(zoom * 100)}%
        </span>
        <button
          onClick={() => setZoom(zoom + 0.1)}
          className="p-1.5 text-gray-500 hover:text-white rounded transition-colors"
        >
          <Plus className="w-3.5 h-3.5" />
        </button>
        <div className="w-px h-4 bg-surface-300 mx-0.5" />
        <button
          onClick={resetView}
          className="p-1.5 text-gray-500 hover:text-white rounded transition-colors"
        >
          <Maximize2 className="w-3.5 h-3.5" />
        </button>
      </div>
    </div>
  )
}
