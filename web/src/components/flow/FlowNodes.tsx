import { memo } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import { Database, HardDrive, Github } from 'lucide-react'

/* ─── Type definitions ────────────────────────────────────────── */

export type ServiceNodeData = {
  label: string
  subtitle?: string
  status: string
  nodeType: 'app' | 'service' | 'volume'
  serviceType?: string
  image?: string
  onClick?: () => void
}

const SERVICE_ICONS: Record<string, string> = {
  postgres: '🐘', mysql: '🐬', redis: '⚡', mongo: '🍃', minio: '🪣',
}

/* ─── Status dot ──────────────────────────────────────────────── */

function StatusDot({ status }: { status: string }) {
  const isOnline = ['healthy', 'running', 'active'].includes(status)
  const isBuilding = ['building', 'deploying', 'queued'].includes(status)
  const isFailed = status === 'failed'
  const color = isOnline
    ? 'bg-emerald-500'
    : isBuilding
      ? 'bg-blue-400 animate-pulse'
      : isFailed
        ? 'bg-red-500'
        : 'bg-gray-400'
  return <span className={`w-2 h-2 rounded-full ${color} shrink-0`} />
}

function statusLabel(status: string) {
  if (['healthy', 'running', 'active'].includes(status)) return 'Online'
  if (status === 'building') return 'Building...'
  if (status === 'deploying') return 'Deploying...'
  if (status === 'stopped') return 'Stopped'
  if (status === 'failed') return 'Failed'
  if (status === 'empty') return 'empty'
  return status
}

/* ─── App Node (dark bg, GitHub-style) ────────────────────────── */

const AppNodeComponent = ({ data }: NodeProps) => {
  const d = data as unknown as ServiceNodeData
  return (
    <div
      onClick={d.onClick}
      className="
        group relative bg-white rounded-xl border border-gray-200 shadow-sm
        hover:shadow-md hover:-translate-y-0.5 transition-all duration-200
        cursor-pointer select-none w-[240px]
      "
    >
      <Handle type="source" position={Position.Right}
        className="!w-2.5 !h-2.5 !bg-gray-300 !border-[2.5px] !border-white !-right-[6px]" />
      <Handle type="target" position={Position.Left}
        className="!w-2.5 !h-2.5 !bg-gray-300 !border-[2.5px] !border-white !-left-[6px]" />

      <div className="p-4">
        <div className="flex items-start gap-3">
          <div className="w-9 h-9 rounded-lg bg-gray-900 flex items-center justify-center shrink-0">
            <Github className="w-[18px] h-[18px] text-white" />
          </div>
          <div className="min-w-0 flex-1">
            <div className="font-semibold text-[13px] text-gray-900 truncate leading-tight">
              {d.label}
            </div>
            {d.subtitle && (
              <div className="text-[11px] text-gray-400 truncate mt-0.5">{d.subtitle}</div>
            )}
          </div>
        </div>

        <div className="flex items-center gap-1.5 mt-3">
          <StatusDot status={d.status} />
          <span className="text-xs text-gray-500">{statusLabel(d.status)}</span>
        </div>
      </div>
    </div>
  )
}

/* ─── Database / Service Node ─────────────────────────────────── */

const DbNodeComponent = ({ data }: NodeProps) => {
  const d = data as unknown as ServiceNodeData
  const emoji = d.serviceType ? SERVICE_ICONS[d.serviceType] : null

  return (
    <div
      onClick={d.onClick}
      className="
        group relative bg-white rounded-xl border border-gray-200 shadow-sm
        hover:shadow-md hover:-translate-y-0.5 transition-all duration-200
        cursor-pointer select-none w-[240px]
      "
    >
      <Handle type="target" position={Position.Left}
        className="!w-2.5 !h-2.5 !bg-gray-300 !border-[2.5px] !border-white !-left-[6px]" />
      <Handle type="source" position={Position.Right}
        className="!w-2.5 !h-2.5 !bg-gray-300 !border-[2.5px] !border-white !-right-[6px]" />

      <div className="p-4">
        <div className="flex items-start gap-3">
          <div className="w-9 h-9 rounded-lg bg-gray-50 border border-gray-100 flex items-center justify-center shrink-0">
            {emoji ? (
              <span className="text-lg">{emoji}</span>
            ) : (
              <Database className="w-4 h-4 text-gray-500" />
            )}
          </div>
          <div className="min-w-0 flex-1">
            <div className="font-semibold text-[13px] text-gray-900 truncate leading-tight">
              {d.label}
            </div>
            {d.subtitle && (
              <div className="text-[11px] text-gray-400 truncate mt-0.5">{d.subtitle}</div>
            )}
          </div>
        </div>

        <div className="flex items-center gap-1.5 mt-3">
          <StatusDot status={d.status} />
          <span className="text-xs text-gray-500">{statusLabel(d.status)}</span>
        </div>

        {d.image && (
          <div className="mt-2 flex items-center gap-1.5 text-[11px] text-gray-400">
            <HardDrive className="w-3 h-3" />
            <span className="font-mono truncate">{d.image}</span>
          </div>
        )}
      </div>
    </div>
  )
}

/* ─── Volume Node ─────────────────────────────────────────────── */

const VolumeNodeComponent = ({ data }: NodeProps) => {
  const d = data as unknown as ServiceNodeData
  return (
    <div
      onClick={d.onClick}
      className="
        group relative bg-white rounded-xl border border-gray-200 shadow-sm
        hover:shadow-md hover:-translate-y-0.5 transition-all duration-200
        cursor-pointer select-none w-[200px]
      "
    >
      <Handle type="target" position={Position.Top}
        className="!w-2.5 !h-2.5 !bg-gray-300 !border-[2.5px] !border-white !-top-[6px]" />

      <div className="p-4">
        <div className="flex items-start gap-3">
          <div className="w-9 h-9 rounded-lg bg-blue-50 border border-blue-100 flex items-center justify-center shrink-0">
            <HardDrive className="w-4 h-4 text-blue-500" />
          </div>
          <div className="min-w-0 flex-1">
            <div className="font-semibold text-[13px] text-gray-900 truncate leading-tight">
              {d.label}
            </div>
            <div className="text-[11px] text-gray-400 mt-0.5">{d.subtitle || 'empty'}</div>
          </div>
        </div>
      </div>
    </div>
  )
}

/* ─── Exports ─────────────────────────────────────────────────── */

export const AppNode = memo(AppNodeComponent)
export const DbNode = memo(DbNodeComponent)
export const VolumeNode = memo(VolumeNodeComponent)

export const nodeTypes = {
  app: AppNode,
  service: DbNode,
  volume: VolumeNode,
}
