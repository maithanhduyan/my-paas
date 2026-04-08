import type { DeploymentStatus } from '../types'

const statusConfig: Record<string, { label: string; color: string; dot: string }> = {
  active: { label: 'Active', color: 'text-success', dot: 'bg-success' },
  healthy: { label: 'Healthy', color: 'text-success', dot: 'bg-success' },
  ready_to_deploy: { label: 'Ready to Deploy', color: 'text-warning', dot: 'bg-warning' },
  queued: { label: 'Queued', color: 'text-gray-400', dot: 'bg-gray-400' },
  cloning: { label: 'Cloning', color: 'text-blue-400', dot: 'bg-blue-400 animate-pulse' },
  detecting: { label: 'Detecting', color: 'text-blue-400', dot: 'bg-blue-400 animate-pulse' },
  building: { label: 'Building', color: 'text-yellow-400', dot: 'bg-yellow-400 animate-pulse' },
  deploying: { label: 'Deploying', color: 'text-purple-400', dot: 'bg-purple-400 animate-pulse' },
  failed: { label: 'Failed', color: 'text-danger', dot: 'bg-danger' },
  rolled_back: { label: 'Rolled Back', color: 'text-orange-400', dot: 'bg-orange-400' },
  cancelled: { label: 'Cancelled', color: 'text-gray-500', dot: 'bg-gray-500' },
}

export function StatusBadge({ status }: { status: string | DeploymentStatus }) {
  const cfg = statusConfig[status] ?? { label: status, color: 'text-gray-400', dot: 'bg-gray-400' }
  return (
    <span className={`inline-flex items-center gap-1.5 text-xs font-medium ${cfg.color}`}>
      <span className={`w-2 h-2 rounded-full ${cfg.dot}`} />
      {cfg.label}
    </span>
  )
}
