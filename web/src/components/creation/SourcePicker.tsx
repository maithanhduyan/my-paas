import { useState } from 'react'
import { useCanvasStore } from '../../store/canvasStore'
import {
  SERVICE_DEFS, STACK_TEMPLATES, getAutoPosition,
  type StackTemplate,
} from '../../lib/catalog'
import {
  GitBranch, Container, Database, HardDrive, Layers,
  ChevronLeft, ArrowRight, Sparkles, Search,
} from 'lucide-react'

type View = 'main' | 'databases' | 'templates'

const CATEGORIES = [
  { id: 'github', icon: GitBranch, name: 'GitHub Repository', desc: 'Import and deploy from Git', color: '#6366f1' },
  { id: 'docker-image', icon: Container, name: 'Docker Image', desc: 'Deploy any container image', color: '#0db7ed' },
  { id: 'database', icon: Database, name: 'Database', desc: 'PostgreSQL, MySQL, MongoDB, Redis', color: '#336791', hasArrow: true },
  { id: 'template', icon: Layers, name: 'Template', desc: 'Pre-configured full stacks', color: '#a855f7', hasArrow: true },
  { id: 'empty', icon: HardDrive, name: 'Empty Canvas', desc: 'Start from scratch', color: '#6b7280' },
]

const DB_OPTIONS = [
  { id: 'postgres', name: 'PostgreSQL', icon: '🐘', desc: 'Relational SQL database', color: '#336791' },
  { id: 'mysql', name: 'MySQL', icon: '🐬', desc: 'Relational SQL database', color: '#4479A1' },
  { id: 'mongo', name: 'MongoDB', icon: '🍃', desc: 'Document database', color: '#47A248' },
  { id: 'redis', name: 'Redis', icon: '⚡', desc: 'In-memory cache & queue', color: '#DC382D' },
  { id: 'minio', name: 'MinIO', icon: '🪣', desc: 'S3-compatible storage', color: '#C72C48' },
]

export function SourcePicker() {
  const [view, setView] = useState<View>('main')
  const [search, setSearch] = useState('')
  const { addNode, addConnection, setPhase, nodes } = useCanvasStore()

  const addServiceNode = (serviceType: string) => {
    const def = SERVICE_DEFS[serviceType]
    if (!def) return
    const pos = getAutoPosition(nodes, { x: 400, y: 300 })
    addNode({
      category: def.category,
      serviceType,
      name: def.defaultName,
      x: pos.x,
      y: pos.y,
      config: { ...def.defaultConfig },
      gitUrl: '',
      branch: 'main',
      dockerImage: '',
    })
    setPhase('compose')
  }

  const applyTemplate = (tpl: StackTemplate) => {
    const cx = 300
    const cy = 250
    const ids: string[] = []
    for (const n of tpl.nodes) {
      const def = SERVICE_DEFS[n.serviceType]
      if (!def) continue
      const id = addNode({
        category: def.category,
        serviceType: n.serviceType,
        name: n.name,
        x: cx + n.relX,
        y: cy + n.relY,
        config: { ...def.defaultConfig },
        gitUrl: n.gitUrl ?? '',
        branch: 'main',
        dockerImage: '',
      })
      ids.push(id)
    }
    for (const [fi, ti] of tpl.connections) {
      if (ids[fi] && ids[ti]) addConnection(ids[fi], ids[ti])
    }
    setPhase('compose')
  }

  const handleCategory = (id: string) => {
    if (id === 'database') return setView('databases')
    if (id === 'template') return setView('templates')
    if (id === 'empty') return setPhase('compose')
    addServiceNode(id)
  }

  const filteredTemplates = STACK_TEMPLATES.filter(
    (t) => !search || t.name.toLowerCase().includes(search.toLowerCase()),
  )

  return (
    <div className="absolute inset-0 z-20 flex items-center justify-center bg-surface/80 backdrop-blur-sm">
      <div className="w-full max-w-lg mx-4 animate-in fade-in zoom-in-95">
        {/* Main view */}
        {view === 'main' && (
          <div className="bg-surface-50 border border-surface-300 rounded-2xl shadow-2xl shadow-black/40 overflow-hidden">
            {/* Header */}
            <div className="p-6 pb-4">
              <h2 className="text-xl font-bold text-center mb-1">What would you like to deploy?</h2>
              <p className="text-sm text-gray-500 text-center">Choose a starting point for your infrastructure</p>
            </div>

            {/* Quick-start templates */}
            <div className="px-6 pb-4">
              <div className="flex items-center gap-2 mb-3">
                <Sparkles className="w-3.5 h-3.5 text-accent" />
                <span className="text-xs font-medium text-gray-400">Quick Start</span>
              </div>
              <div className="grid grid-cols-2 gap-2">
                {STACK_TEMPLATES.slice(0, 4).map((tpl) => (
                  <button
                    key={tpl.id}
                    onClick={() => applyTemplate(tpl)}
                    className="flex items-center gap-3 p-3 rounded-xl border border-surface-300 bg-surface-100
                               hover:border-accent/40 hover:bg-surface-200 transition-all text-left group"
                  >
                    <span className="text-xl shrink-0">{tpl.icon}</span>
                    <div className="min-w-0">
                      <div className="text-sm font-medium text-gray-200 truncate">{tpl.name}</div>
                      <div className="text-[11px] text-gray-500 truncate">{tpl.description}</div>
                    </div>
                  </button>
                ))}
              </div>
            </div>

            <div className="flex items-center gap-3 px-6 pb-3">
              <div className="h-px flex-1 bg-surface-300" />
              <span className="text-[11px] text-gray-600">or choose a resource type</span>
              <div className="h-px flex-1 bg-surface-300" />
            </div>

            {/* Categories */}
            <div className="px-6 pb-6 space-y-1">
              {CATEGORIES.map((cat) => {
                const Icon = cat.icon
                return (
                  <button
                    key={cat.id}
                    onClick={() => handleCategory(cat.id)}
                    className="w-full flex items-center gap-3 px-4 py-3 rounded-xl
                               hover:bg-surface-200 transition-colors group text-left"
                  >
                    <div
                      className="w-9 h-9 rounded-lg flex items-center justify-center shrink-0"
                      style={{ background: `${cat.color}18` }}
                    >
                      <Icon className="w-4.5 h-4.5" style={{ color: cat.color }} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-medium text-gray-200">{cat.name}</div>
                      <div className="text-[11px] text-gray-500">{cat.desc}</div>
                    </div>
                    {cat.hasArrow && (
                      <ArrowRight className="w-4 h-4 text-gray-600 group-hover:text-gray-400 transition-colors" />
                    )}
                  </button>
                )
              })}
            </div>
          </div>
        )}

        {/* Database sub-view */}
        {view === 'databases' && (
          <div className="bg-surface-50 border border-surface-300 rounded-2xl shadow-2xl shadow-black/40 overflow-hidden">
            <div className="p-6 pb-4 flex items-center gap-3">
              <button
                onClick={() => setView('main')}
                className="p-1.5 rounded-lg hover:bg-surface-200 text-gray-400 hover:text-white transition-colors"
              >
                <ChevronLeft className="w-5 h-5" />
              </button>
              <div>
                <h2 className="text-lg font-bold">Choose Database</h2>
                <p className="text-xs text-gray-500">Select a managed database service</p>
              </div>
            </div>
            <div className="px-6 pb-6 space-y-1">
              {DB_OPTIONS.map((db) => (
                <button
                  key={db.id}
                  onClick={() => addServiceNode(db.id)}
                  className="w-full flex items-center gap-3 px-4 py-3 rounded-xl
                             hover:bg-surface-200 transition-colors text-left"
                >
                  <div
                    className="w-9 h-9 rounded-lg flex items-center justify-center text-lg shrink-0"
                    style={{ background: `${db.color}18` }}
                  >
                    {db.icon}
                  </div>
                  <div>
                    <div className="text-sm font-medium text-gray-200">{db.name}</div>
                    <div className="text-[11px] text-gray-500">{db.desc}</div>
                  </div>
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Template gallery view */}
        {view === 'templates' && (
          <div className="bg-surface-50 border border-surface-300 rounded-2xl shadow-2xl shadow-black/40 overflow-hidden">
            <div className="p-6 pb-4 flex items-center gap-3">
              <button
                onClick={() => setView('main')}
                className="p-1.5 rounded-lg hover:bg-surface-200 text-gray-400 hover:text-white transition-colors"
              >
                <ChevronLeft className="w-5 h-5" />
              </button>
              <div className="flex-1">
                <h2 className="text-lg font-bold">Stack Templates</h2>
                <p className="text-xs text-gray-500">Pre-configured infrastructure stacks</p>
              </div>
            </div>
            <div className="px-6 pb-2">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                <input
                  type="text"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  placeholder="Search templates..."
                  className="w-full pl-9 pr-3 py-2 bg-surface-100 border border-surface-300 rounded-lg text-sm
                             placeholder-gray-600 focus:outline-none focus:border-accent"
                />
              </div>
            </div>
            <div className="px-6 pb-6 pt-3 space-y-2 max-h-80 overflow-y-auto">
              {filteredTemplates.map((tpl) => (
                <button
                  key={tpl.id}
                  onClick={() => applyTemplate(tpl)}
                  className="w-full flex items-center gap-4 p-4 rounded-xl border border-surface-300
                             hover:border-accent/40 hover:bg-surface-200 transition-all text-left group"
                >
                  <div
                    className="w-12 h-12 rounded-xl flex items-center justify-center text-2xl shrink-0"
                    style={{ background: `${tpl.color}15` }}
                  >
                    {tpl.icon}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium text-gray-200">{tpl.name}</div>
                    <div className="text-xs text-gray-500 mt-0.5">{tpl.description}</div>
                    <div className="flex items-center gap-1.5 mt-2">
                      {tpl.nodes.map((n, i) => {
                        const d = SERVICE_DEFS[n.serviceType]
                        return (
                          <span
                            key={i}
                            className="text-[10px] px-1.5 py-0.5 rounded-md bg-surface-100 text-gray-400"
                          >
                            {d?.icon} {d?.name ?? n.serviceType}
                          </span>
                        )
                      })}
                    </div>
                  </div>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
