import { useEffect, useRef, useState } from 'react'
import { streamDeploymentLogs } from '../api'

interface LogEntry {
  step?: string
  level?: string
  message?: string
  status?: string
  done?: boolean
}

export function LogViewer({ deploymentId }: { deploymentId: string }) {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [done, setDone] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    setLogs([])
    setDone(false)
    const close = streamDeploymentLogs(deploymentId, (entry) => {
      if (entry.done) {
        setDone(true)
        return
      }
      setLogs((prev) => [...prev, entry])
    })
    return close
  }, [deploymentId])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  const levelColor = (level?: string) => {
    if (level === 'error') return 'text-danger'
    if (level === 'warn') return 'text-warning'
    return 'text-gray-300'
  }

  const stepColor = (step?: string) => {
    const colors: Record<string, string> = {
      clone: 'text-blue-400',
      detect: 'text-purple-400',
      build: 'text-yellow-400',
      deploy: 'text-green-400',
      healthcheck: 'text-emerald-400',
    }
    return step ? colors[step] ?? 'text-gray-400' : 'text-gray-400'
  }

  return (
    <div className="bg-surface-50 border border-surface-300 rounded-lg overflow-hidden">
      <div className="px-3 py-2 border-b border-surface-300 flex items-center justify-between">
        <span className="text-xs font-medium text-gray-400">Build Logs</span>
        {!done && (
          <span className="text-xs text-blue-400 animate-pulse">● streaming...</span>
        )}
      </div>
      <div className="p-3 font-mono text-xs max-h-96 overflow-y-auto space-y-0.5">
        {logs.length === 0 && !done && (
          <div className="text-gray-500">Waiting for logs...</div>
        )}
        {logs.map((log, i) => (
          <div key={i} className="flex gap-2">
            <span className={`w-20 shrink-0 ${stepColor(log.step)}`}>[{log.step}]</span>
            <span className={levelColor(log.level)}>{log.message}</span>
          </div>
        ))}
        {done && (
          <div className="text-gray-500 mt-2">— end of logs —</div>
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  )
}
