import { useCallback, useEffect, useState } from 'react'
import { listTemplates, deployTemplate } from '../api'
import type { Template } from '../types'
import { Package, Rocket, Loader2, Server, Database, Zap, HardDrive, Globe } from 'lucide-react'

const iconMap: Record<string, React.ReactNode> = {
  globe: <Globe className="w-6 h-6" />,
  server: <Server className="w-6 h-6" />,
  database: <Database className="w-6 h-6" />,
  zap: <Zap className="w-6 h-6" />,
  'hard-drive': <HardDrive className="w-6 h-6" />,
}

export function Marketplace() {
  const [templates, setTemplates] = useState<Template[]>([])
  const [loading, setLoading] = useState(true)
  const [deploying, setDeploying] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setTemplates(await listTemplates())
    } catch {
      setTemplates([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const handleDeploy = async (template: Template) => {
    const name = prompt('Project name (leave blank for default):', template.name)
    if (name === null) return

    setDeploying(template.id)
    try {
      await deployTemplate(template.id, name || undefined)
      alert(`Template "${template.name}" deployed successfully!`)
    } catch (e: any) {
      alert('Deploy failed: ' + e.message)
    } finally {
      setDeploying(null)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-500">Loading...</div>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-5xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Marketplace</h1>
        <p className="text-sm text-gray-500 mt-1">Deploy pre-configured templates with one click</p>
      </div>

      {templates.length === 0 ? (
        <div className="text-center py-16 text-gray-500">
          <Package className="w-12 h-12 mx-auto mb-3 opacity-40" />
          <p>No templates available</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {templates.map((t) => (
            <div
              key={t.id}
              className="bg-surface-50 rounded-xl border border-surface-300 p-5 flex flex-col gap-3 hover:border-accent/50 transition-colors"
            >
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-accent/10 flex items-center justify-center text-accent">
                  {iconMap[t.icon] || <Package className="w-6 h-6" />}
                </div>
                <div>
                  <h3 className="font-medium text-sm">{t.name}</h3>
                  <p className="text-xs text-gray-500">{t.services.length} service{t.services.length !== 1 ? 's' : ''}</p>
                </div>
              </div>

              <p className="text-sm text-gray-400 flex-1">{t.description}</p>

              <div className="flex flex-wrap gap-1.5">
                {t.services.map((s) => (
                  <span
                    key={s.name}
                    className="text-xs bg-surface px-2 py-0.5 rounded border border-surface-300 text-gray-400"
                  >
                    {s.type === 'app' ? '🚀 App' : s.type}
                  </span>
                ))}
              </div>

              <button
                onClick={() => handleDeploy(t)}
                disabled={deploying === t.id}
                className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-accent text-white rounded-lg text-sm font-medium hover:bg-accent-hover transition-colors disabled:opacity-50"
              >
                {deploying === t.id ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Rocket className="w-4 h-4" />
                )}
                Deploy
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
