import { CheckCircle2, XCircle, Clock, Loader2, RotateCcw } from 'lucide-react'
import type { Deployment } from '../../types'
import { timeAgo } from '../../lib/utils'

interface Props {
  deployments: Deployment[]
  onViewLogs?: (id: string) => void
}

function DeployIcon({ status }: { status: string }) {
  if (status === 'healthy')
    return <CheckCircle2 className="w-5 h-5 text-emerald-500 shrink-0" />
  if (status === 'failed')
    return <XCircle className="w-5 h-5 text-red-500 shrink-0" />
  if (status === 'rolled_back')
    return <RotateCcw className="w-5 h-5 text-amber-500 shrink-0" />
  if (['building', 'deploying', 'queued', 'cloning', 'detecting'].includes(status))
    return <Loader2 className="w-5 h-5 text-blue-500 animate-spin shrink-0" />
  return <Clock className="w-5 h-5 text-gray-400 shrink-0" />
}

function statusText(status: string) {
  if (status === 'healthy') return 'Deployment successful'
  if (status === 'failed') return 'Deployment failed'
  if (status === 'rolled_back') return 'Rolled back'
  if (status === 'building') return 'Building...'
  if (status === 'deploying') return 'Deploying...'
  if (status === 'queued') return 'Queued'
  return status
}

function statusColor(status: string) {
  if (status === 'healthy') return 'text-emerald-600'
  if (status === 'failed') return 'text-red-600'
  if (status === 'rolled_back') return 'text-amber-600'
  if (['building', 'deploying'].includes(status)) return 'text-blue-600'
  return 'text-gray-500'
}

export function ActivitySidebar({ deployments, onViewLogs }: Props) {
  if (deployments.length === 0) {
    return (
      <div className="p-4 text-center text-sm text-gray-400">
        No deployments yet
      </div>
    )
  }

  // Group: first one is "active" if healthy, rest is history
  const sorted = [...deployments].sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  )

  return (
    <div className="divide-y divide-gray-100">
      {sorted.map((d) => (
        <button
          key={d.id}
          onClick={() => onViewLogs?.(d.id)}
          className="w-full flex items-start gap-3 px-4 py-3.5 hover:bg-gray-50 transition-colors text-left"
        >
          <div className="mt-0.5">
            <DeployIcon status={d.status} />
          </div>
          <div className="min-w-0 flex-1">
            <div className="text-[13px] font-medium text-gray-900 truncate">
              {d.commit_msg
                ? d.commit_msg.split('\n')[0]
                : d.commit_hash
                  ? d.commit_hash.slice(0, 7)
                  : 'Manual deploy'}
            </div>
            <div className={`text-xs mt-0.5 ${statusColor(d.status)}`}>
              {statusText(d.status)}
            </div>
            <div className="text-[11px] text-gray-400 mt-0.5">
              {timeAgo(d.created_at)}
              {d.trigger && d.trigger !== 'manual' && (
                <span> via {d.trigger}</span>
              )}
            </div>
          </div>
        </button>
      ))}
    </div>
  )
}
