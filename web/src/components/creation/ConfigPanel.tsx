import { useCanvasStore, type CanvasNode } from '../../store/canvasStore'
import { SERVICE_DEFS } from '../../lib/catalog'
import { X, Trash2, Link2 } from 'lucide-react'

export function ConfigPanel() {
  const { nodes, connections, selectedNodeId, updateNode, removeNode, removeConnection, selectNode } =
    useCanvasStore()

  const node = nodes.find((n) => n.id === selectedNodeId)
  if (!node) return null

  const def = SERVICE_DEFS[node.serviceType]
  const isApp = node.category === 'app'

  const nodeConnections = connections.filter(
    (c) => c.from === node.id || c.to === node.id,
  )
  const connectedNodes = nodeConnections.map((c) => {
    const otherId = c.from === node.id ? c.to : c.from
    return { conn: c, node: nodes.find((n) => n.id === otherId)! }
  }).filter((x) => x.node)

  const update = (patch: Partial<CanvasNode>) => updateNode(node.id, patch)

  return (
    <div className="w-72 bg-surface-50 border-l border-surface-300 flex flex-col h-full overflow-hidden shrink-0 animate-in slide-in-from-right-4">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-surface-300">
        <div className="flex items-center gap-2 min-w-0">
          <div
            className="w-7 h-7 rounded-md flex items-center justify-center text-sm shrink-0"
            style={{ background: `${def?.color ?? '#6366f1'}18` }}
          >
            {def?.icon ?? '📦'}
          </div>
          <span className="text-sm font-semibold truncate">{def?.name ?? node.serviceType}</span>
        </div>
        <button
          onClick={() => selectNode(null)}
          className="p-1 text-gray-500 hover:text-white rounded transition-colors"
        >
          <X className="w-4 h-4" />
        </button>
      </div>

      {/* Form */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Name */}
        <Field label="Service Name">
          <input
            type="text"
            value={node.name}
            onChange={(e) => update({ name: e.target.value })}
            className="input-field"
            placeholder="my-service"
          />
        </Field>

        {/* App-specific fields */}
        {isApp && node.serviceType === 'github' && (
          <>
            <Field label="Git Repository URL">
              <input
                type="text"
                value={node.gitUrl}
                onChange={(e) => update({ gitUrl: e.target.value })}
                className="input-field"
                placeholder="https://github.com/user/repo.git"
              />
            </Field>
            <Field label="Branch">
              <input
                type="text"
                value={node.branch}
                onChange={(e) => update({ branch: e.target.value })}
                className="input-field"
                placeholder="main"
              />
            </Field>
          </>
        )}

        {isApp && node.serviceType === 'docker-image' && (
          <Field label="Docker Image">
            <input
              type="text"
              value={node.dockerImage}
              onChange={(e) => update({ dockerImage: e.target.value })}
              className="input-field"
              placeholder="nginx:latest"
            />
          </Field>
        )}

        {/* Database/service config */}
        {!isApp && node.config.version !== undefined && (
          <Field label="Version">
            <input
              type="text"
              value={node.config.version ?? ''}
              onChange={(e) =>
                update({ config: { ...node.config, version: e.target.value } })
              }
              className="input-field"
              placeholder="latest"
            />
          </Field>
        )}

        {/* Connections */}
        {connectedNodes.length > 0 && (
          <div>
            <label className="text-[11px] font-medium text-gray-400 uppercase tracking-wider block mb-2">
              Connections
            </label>
            <div className="space-y-1.5">
              {connectedNodes.map(({ conn, node: other }) => {
                const otherDef = SERVICE_DEFS[other.serviceType]
                return (
                  <div
                    key={conn.id}
                    className="flex items-center gap-2 px-3 py-2 rounded-lg bg-surface-100 border border-surface-300"
                  >
                    <Link2 className="w-3 h-3 text-gray-500 shrink-0" />
                    <span className="text-xs text-gray-300 truncate flex-1">
                      {otherDef?.icon} {other.name}
                    </span>
                    <button
                      onClick={() => removeConnection(conn.id)}
                      className="text-gray-600 hover:text-danger transition-colors"
                      title="Remove connection"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </div>
                )
              })}
            </div>
          </div>
        )}

        {/* Environment hints */}
        {isApp && connectedNodes.length > 0 && (
          <div>
            <label className="text-[11px] font-medium text-gray-400 uppercase tracking-wider block mb-2">
              Auto-injected Variables
            </label>
            <div className="bg-surface-100 rounded-lg border border-surface-300 p-3 space-y-1">
              {connectedNodes.map(({ node: other }) => {
                const prefix = other.name.toUpperCase().replace(/[^A-Z0-9]/g, '_')
                return (
                  <div key={other.id} className="text-[11px] font-mono text-gray-500">
                    <span className="text-accent">{prefix}_URL</span>=...
                  </div>
                )
              })}
            </div>
          </div>
        )}
      </div>

      {/* Delete */}
      <div className="p-4 border-t border-surface-300">
        <button
          onClick={() => { removeNode(node.id); selectNode(null) }}
          className="w-full flex items-center justify-center gap-2 px-3 py-2 text-xs text-danger
                     hover:bg-danger/10 rounded-lg border border-danger/20 transition-colors"
        >
          <Trash2 className="w-3.5 h-3.5" />
          Remove Service
        </button>
      </div>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="text-[11px] font-medium text-gray-400 uppercase tracking-wider block mb-1.5">
        {label}
      </label>
      {children}
    </div>
  )
}
