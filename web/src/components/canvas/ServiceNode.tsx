import { SERVICE_DEFS } from '../../lib/catalog'
import type { CanvasNode } from '../../store/canvasStore'

interface Props {
  node: CanvasNode
  selected: boolean
  nodeWidth: number
  isConnectTarget: boolean
  onSelect: () => void
  onNodeMouseDown: (e: React.MouseEvent, nodeId: string) => void
  onPortMouseDown: (e: React.MouseEvent, nodeId: string) => void
  onPortMouseUp: (e: React.MouseEvent, nodeId: string) => void
}

export function ServiceNode({
  node,
  selected,
  nodeWidth,
  isConnectTarget,
  onSelect,
  onNodeMouseDown,
  onPortMouseDown,
  onPortMouseUp,
}: Props) {
  const def = SERVICE_DEFS[node.serviceType]
  const color = def?.color ?? '#6366f1'
  const icon = def?.icon ?? '📦'
  const typeName = def?.name ?? node.serviceType

  return (
    <div
      className={`
        absolute rounded-xl border transition-all duration-150 select-none group
        ${selected
          ? 'border-accent shadow-lg shadow-accent/20 ring-1 ring-accent/30'
          : isConnectTarget
            ? 'border-accent/50 shadow-md shadow-accent/10 ring-1 ring-accent/20'
            : 'border-surface-300 hover:border-surface-200 hover:shadow-md hover:shadow-black/20'
        }
      `}
      style={{
        left: node.x,
        top: node.y,
        width: nodeWidth,
        background: 'linear-gradient(180deg, #1e1e1e 0%, #171717 100%)',
      }}
      onClick={(e) => { e.stopPropagation(); onSelect() }}
      onMouseDown={(e) => {
        if ((e.target as HTMLElement).dataset.port) return
        e.stopPropagation()
        onNodeMouseDown(e, node.id)
      }}
    >
      {/* Top accent bar */}
      <div
        className="h-[3px] rounded-t-xl"
        style={{ background: `linear-gradient(90deg, ${color}, ${color}88)` }}
      />

      {/* Content */}
      <div className="p-3.5">
        <div className="flex items-center gap-2.5">
          <div
            className="w-9 h-9 rounded-lg flex items-center justify-center text-lg shrink-0"
            style={{ background: `${color}18` }}
          >
            {icon}
          </div>
          <div className="flex-1 min-w-0">
            <div className="font-semibold text-sm text-gray-100 truncate">
              {node.name}
            </div>
            <div className="text-[11px] text-gray-500 mt-0.5">
              {typeName}
            </div>
          </div>
        </div>

        {/* Status row */}
        <div className="mt-2.5 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-success/80" />
            <span className="text-[11px] text-gray-500">Ready</span>
          </div>
          {node.gitUrl && (
            <span className="text-[10px] text-gray-600 truncate max-w-[100px]">
              {node.branch || 'main'}
            </span>
          )}
        </div>
      </div>

      {/* Input port (left) */}
      <div
        data-port="input"
        className={`
          absolute left-0 top-1/2 -translate-x-1/2 -translate-y-1/2
          w-3 h-3 rounded-full border-2 transition-all duration-150 cursor-crosshair z-10
          ${isConnectTarget
            ? 'bg-accent border-accent scale-125'
            : 'bg-surface-200 border-surface-300 hover:border-accent hover:bg-accent/30 hover:scale-125'
          }
        `}
        onMouseUp={(e) => { e.stopPropagation(); onPortMouseUp(e, node.id) }}
      />

      {/* Output port (right) */}
      <div
        data-port="output"
        className={`
          absolute right-0 top-1/2 translate-x-1/2 -translate-y-1/2
          w-3 h-3 rounded-full border-2 transition-all duration-150 cursor-crosshair z-10
          bg-surface-200 border-surface-300 hover:border-accent hover:bg-accent/30 hover:scale-125
        `}
        onMouseDown={(e) => { e.stopPropagation(); onPortMouseDown(e, node.id) }}
      />
    </div>
  )
}
