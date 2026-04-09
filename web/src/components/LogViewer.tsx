import { useEffect, useRef, useState } from 'react'
import { streamDeploymentLogs, streamProjectLogs } from '../api'

type LogTab = 'build' | 'runtime'

interface LogEntry {
  step?: string
  level?: string
  message?: string
  status?: string
  done?: boolean
  time?: string
}

function formatTime(iso?: string) {
  if (!iso) return ''
  try {
    const d = new Date(iso)
    return d.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
  } catch { return '' }
}

function BuildLogs({ deploymentId }: { deploymentId: string }) {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [done, setDone] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    setLogs([])
    setDone(false)
    const close = streamDeploymentLogs(deploymentId, (entry) => {
      if (entry.done) { setDone(true); return }
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

  const lineBg = (level?: string) => {
    if (level === 'error') return 'bg-danger/10 border-l-2 border-danger'
    if (level === 'warn') return 'bg-warning/10 border-l-2 border-warning'
    return ''
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

  const lastStatus = logs.length > 0 ? logs[logs.length - 1]?.status : undefined

  return (
    <>
      <div className="px-3 py-1.5 border-b border-surface-300 flex items-center justify-end">
        {!done && <span className="text-xs text-blue-400 animate-pulse">● streaming...</span>}
        {done && lastStatus === 'success' && <span className="text-xs text-success font-medium">✓ Success</span>}
        {done && lastStatus === 'failed' && <span className="text-xs text-danger font-medium">✗ Failed</span>}
      </div>
      <div className="p-3 font-mono text-xs max-h-80 overflow-y-auto space-y-0.5">
        {logs.length === 0 && !done && <div className="text-gray-500">Waiting for logs...</div>}
        {logs.map((log, i) => (
          <div key={i} className={`flex gap-2 px-1 py-0.5 rounded ${lineBg(log.level)}`}>
            <span className="w-16 shrink-0 text-gray-600">{formatTime(log.time)}</span>
            <span className={`w-20 shrink-0 ${stepColor(log.step)}`}>[{log.step}]</span>
            <span className={levelColor(log.level)}>{log.message}</span>
          </div>
        ))}
        {done && <div className="text-gray-500 mt-2">— end of logs —</div>}
        <div ref={bottomRef} />
      </div>
    </>
  )
}

function RuntimeLogs({ projectId }: { projectId: string }) {
  const [lines, setLines] = useState<string[]>([])
  const [connected, setConnected] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    setLines([])
    setConnected(true)
    const close = streamProjectLogs(projectId, (line) => {
      setLines((prev) => [...prev.slice(-500), line])
    })
    return () => { close(); setConnected(false) }
  }, [projectId])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [lines])

  // Parse timestamp from Docker log line (format: "2026-04-09T02:15:51.123456789Z message")
  const parseLine = (raw: string) => {
    const match = raw.match(/^(\d{4}-\d{2}-\d{2}T[\d:.]+Z?)\s*(.*)$/)
    if (match) {
      return { time: formatTime(match[1]), msg: match[2] }
    }
    return { time: '', msg: raw }
  }

  return (
    <>
      <div className="px-3 py-1.5 border-b border-surface-300 flex items-center justify-end">
        {connected && <span className="text-xs text-blue-400 animate-pulse">● streaming...</span>}
      </div>
      <div className="p-3 font-mono text-xs max-h-80 overflow-y-auto space-y-0.5">
        {lines.length === 0 && <div className="text-gray-500">Waiting for runtime logs...</div>}
        {lines.map((raw, i) => {
          const { time, msg } = parseLine(raw)
          return (
            <div key={i} className="flex gap-2 px-1 py-0.5">
              <span className="w-16 shrink-0 text-gray-600">{time}</span>
              <span className="text-gray-300">{msg}</span>
            </div>
          )
        })}
        <div ref={bottomRef} />
      </div>
    </>
  )
}

export function LogViewer({ deploymentId, projectId }: { deploymentId: string; projectId?: string }) {
  const [logTab, setLogTab] = useState<LogTab>('build')

  return (
    <div className="bg-surface-50 border border-surface-300 rounded-lg overflow-hidden">
      <div className="flex border-b border-surface-300">
        <button
          onClick={() => setLogTab('build')}
          className={`px-3 py-1.5 text-xs font-medium transition-colors ${
            logTab === 'build' ? 'text-accent border-b-2 border-accent bg-surface-100' : 'text-gray-500 hover:text-gray-300'
          }`}
        >Build Logs</button>
        {projectId && (
          <button
            onClick={() => setLogTab('runtime')}
            className={`px-3 py-1.5 text-xs font-medium transition-colors ${
              logTab === 'runtime' ? 'text-accent border-b-2 border-accent bg-surface-100' : 'text-gray-500 hover:text-gray-300'
            }`}
          >Runtime Logs</button>
        )}
      </div>
      {logTab === 'build' && <BuildLogs deploymentId={deploymentId} />}
      {logTab === 'runtime' && projectId && <RuntimeLogs projectId={projectId} />}
    </div>
  )
}
