import { useEffect, useState } from 'react'
import { getSwarmStatus, initSwarm, getSwarmToken, getSwarmServices } from '../api'
import type { SwarmStatus, SwarmService } from '../types'
import { Server, Wifi, WifiOff, Copy, Check, Box, RefreshCw } from 'lucide-react'

export function Swarm() {
  const [status, setStatus] = useState<SwarmStatus | null>(null)
  const [services, setServices] = useState<SwarmService[]>([])
  const [loading, setLoading] = useState(true)
  const [initAddr, setInitAddr] = useState('')
  const [joinToken, setJoinToken] = useState('')
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState('')

  const load = async () => {
    try {
      const s = await getSwarmStatus()
      setStatus(s)
      if (s.active) {
        const svc = await getSwarmServices()
        setServices(svc || [])
      }
    } catch {
      setStatus({ active: false, nodes: [] })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleInit = async () => {
    setError('')
    try {
      await initSwarm(initAddr || undefined)
      await load()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to init swarm')
    }
  }

  const handleGetToken = async () => {
    try {
      const res = await getSwarmToken()
      setJoinToken(res.token)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to get token')
    }
  }

  const copyToken = () => {
    navigator.clipboard.writeText(joinToken)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  if (loading) {
    return <div className="flex items-center justify-center h-full text-gray-500">Loading...</div>
  }

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Swarm Cluster</h1>
        <div className="flex items-center gap-2">
          {status?.active && (
            <button
              onClick={load}
              className="p-1.5 text-gray-400 hover:text-gray-200 border border-surface-300 rounded transition-colors"
              title="Refresh"
            >
              <RefreshCw className="w-4 h-4" />
            </button>
          )}
          {status?.active ? (
            <span className="flex items-center gap-1.5 text-sm text-success">
              <Wifi className="w-4 h-4" /> Active
            </span>
          ) : (
            <span className="flex items-center gap-1.5 text-sm text-gray-500">
              <WifiOff className="w-4 h-4" /> Inactive
            </span>
          )}
        </div>
      </div>

      {error && (
        <div className="p-3 bg-danger/10 border border-danger/30 rounded-lg text-sm text-danger">
          {error}
        </div>
      )}

      {!status?.active ? (
        <div className="bg-surface-50 border border-surface-300 rounded-lg p-6 text-center space-y-4">
          <Server className="w-12 h-12 text-gray-500 mx-auto" />
          <div>
            <h2 className="text-lg font-semibold mb-1">Initialize Docker Swarm</h2>
            <p className="text-sm text-gray-500">
              Enable Swarm mode to deploy services across multiple nodes with built-in load balancing.
            </p>
          </div>
          <div className="flex items-center gap-2 justify-center max-w-md mx-auto">
            <input
              value={initAddr}
              onChange={(e) => setInitAddr(e.target.value)}
              placeholder="Advertise address (e.g. 192.168.56.10:2377)"
              className="flex-1 px-3 py-2 bg-surface border border-surface-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-accent"
            />
            <button
              onClick={handleInit}
              className="px-4 py-2 bg-accent text-white rounded text-sm font-medium hover:bg-accent-hover transition-colors"
            >
              Init Swarm
            </button>
          </div>
        </div>
      ) : (
        <>
          {/* Nodes */}
          <div>
            <h2 className="text-sm font-medium text-gray-300 mb-2">
              Nodes ({status.nodes.length})
            </h2>
            <div className="space-y-2">
              {status.nodes.map((node) => (
                <div
                  key={node.id}
                  className="flex items-center justify-between bg-surface-50 border border-surface-300 rounded-lg p-3"
                >
                  <div className="flex items-center gap-3">
                    <Server className="w-5 h-5 text-gray-400" />
                    <div>
                      <div className="text-sm font-medium">{node.hostname}</div>
                      <div className="text-xs text-gray-500">{node.addr}</div>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className={`text-xs px-2 py-0.5 rounded ${
                      node.role === 'manager'
                        ? 'bg-accent/15 text-accent-hover'
                        : 'bg-surface-200 text-gray-400'
                    }`}>
                      {node.role}
                    </span>
                    <span className={`text-xs ${
                      node.status === 'ready' ? 'text-success' : 'text-warning'
                    }`}>
                      {node.status}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Swarm Services */}
          <div>
            <h2 className="text-sm font-medium text-gray-300 mb-2">
              Services ({services.length})
            </h2>
            {services.length === 0 ? (
              <div className="bg-surface-50 border border-surface-300 rounded-lg p-4 text-center text-sm text-gray-500">
                No Swarm services running. Deploy a project to see it here.
              </div>
            ) : (
              <div className="space-y-2">
                {services.map((svc) => (
                  <div
                    key={svc.id}
                    className="bg-surface-50 border border-surface-300 rounded-lg p-3"
                  >
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <Box className="w-4 h-4 text-gray-400" />
                        <span className="text-sm font-medium">{svc.name}</span>
                      </div>
                      <span className="text-xs text-gray-500">
                        {svc.tasks.filter(t => t.state === 'running').length}/{svc.replicas} replicas
                      </span>
                    </div>
                    <div className="text-xs text-gray-500 mb-2 truncate">
                      {svc.image}
                    </div>
                    {svc.tasks.length > 0 && (
                      <div className="space-y-1">
                        {svc.tasks.map((task) => (
                          <div
                            key={task.id}
                            className="flex items-center justify-between text-xs bg-surface border border-surface-300 rounded px-2 py-1"
                          >
                            <span className="text-gray-400 font-mono">
                              {task.id.substring(0, 12)}
                            </span>
                            <span className="text-gray-300">
                              {task.node_name || task.node_id.substring(0, 12)}
                            </span>
                            <span className={
                              task.state === 'running' ? 'text-success' :
                              task.state === 'pending' || task.state === 'preparing' || task.state === 'starting' ? 'text-warning' :
                              'text-danger'
                            }>
                              {task.state}
                            </span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Join Token */}
          <div>
            <h2 className="text-sm font-medium text-gray-300 mb-2">Worker Join Token</h2>
            <div className="bg-surface-50 border border-surface-300 rounded-lg p-4 space-y-3">
              <p className="text-xs text-gray-500">
                Use this token on other machines to join them as worker nodes.
              </p>
              {joinToken ? (
                <div className="flex items-center gap-2">
                  <code className="flex-1 px-3 py-2 bg-surface border border-surface-300 rounded text-xs font-mono break-all">
                    {joinToken}
                  </code>
                  <button
                    onClick={copyToken}
                    className="p-2 text-gray-400 hover:text-gray-200 border border-surface-300 rounded transition-colors shrink-0"
                  >
                    {copied ? <Check className="w-4 h-4 text-success" /> : <Copy className="w-4 h-4" />}
                  </button>
                </div>
              ) : (
                <button
                  onClick={handleGetToken}
                  className="px-3 py-1.5 bg-accent text-white rounded text-sm hover:bg-accent-hover transition-colors"
                >
                  Show Token
                </button>
              )}
              {status.manager_addr && (
                <p className="text-xs text-gray-500">
                  Manager: <code className="text-gray-400">{status.manager_addr}</code>
                </p>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
