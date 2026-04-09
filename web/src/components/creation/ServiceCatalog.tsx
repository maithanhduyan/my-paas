import { CATALOG_SECTIONS, SERVICE_DEFS } from '../../lib/catalog'
import { useCanvasStore } from '../../store/canvasStore'
import { ChevronDown } from 'lucide-react'
import { useState } from 'react'

export function ServiceCatalog() {
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({})
  const { setPhase } = useCanvasStore()

  const toggle = (title: string) =>
    setCollapsed((s) => ({ ...s, [title]: !s[title] }))

  const handleDragStart = (e: React.DragEvent, serviceType: string) => {
    e.dataTransfer.setData('application/x-service-type', serviceType)
    e.dataTransfer.effectAllowed = 'copy'
  }

  return (
    <div className="w-56 bg-surface-50 border-r border-surface-300 flex flex-col h-full overflow-hidden shrink-0">
      {/* Header */}
      <div className="px-4 py-3 border-b border-surface-300">
        <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">
          Services
        </h3>
        <p className="text-[10px] text-gray-600 mt-0.5">Drag onto canvas</p>
      </div>

      {/* Catalog sections */}
      <div className="flex-1 overflow-y-auto p-2 space-y-1">
        {CATALOG_SECTIONS.map((section) => (
          <div key={section.title}>
            <button
              onClick={() => toggle(section.title)}
              className="w-full flex items-center gap-2 px-2 py-1.5 text-[11px] font-semibold
                         text-gray-500 uppercase tracking-wider hover:text-gray-300 transition-colors"
            >
              <ChevronDown
                className={`w-3 h-3 transition-transform ${
                  collapsed[section.title] ? '-rotate-90' : ''
                }`}
              />
              {section.title}
            </button>

            {!collapsed[section.title] && (
              <div className="space-y-0.5 ml-1">
                {section.items.map((serviceType) => {
                  const def = SERVICE_DEFS[serviceType]
                  if (!def) return null
                  return (
                    <div
                      key={serviceType}
                      draggable
                      onDragStart={(e) => handleDragStart(e, serviceType)}
                      className="flex items-center gap-2.5 px-2.5 py-2 rounded-lg
                                 cursor-grab active:cursor-grabbing
                                 hover:bg-surface-200 transition-colors group"
                    >
                      <div
                        className="w-7 h-7 rounded-md flex items-center justify-center text-sm shrink-0"
                        style={{ background: `${def.color}15` }}
                      >
                        {def.icon}
                      </div>
                      <div className="min-w-0">
                        <div className="text-xs font-medium text-gray-300 group-hover:text-gray-100 truncate">
                          {def.name}
                        </div>
                        <div className="text-[10px] text-gray-600 truncate">
                          {def.description}
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Back to picker */}
      <div className="p-3 border-t border-surface-300">
        <button
          onClick={() => setPhase('pick')}
          className="w-full text-[11px] text-gray-500 hover:text-gray-300 transition-colors py-1"
        >
          ← Change source
        </button>
      </div>
    </div>
  )
}
