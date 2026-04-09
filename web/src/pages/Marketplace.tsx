import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { listTemplates, deployTemplate } from '../api'
import type { Template } from '../types'
import { Package, Rocket, Loader2, Server, Database, Zap, HardDrive, Globe } from 'lucide-react'
import { Modal } from '../components/ui/Modal'
import { useToast } from '../components/ui/Toast'

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
  const [deployModal, setDeployModal] = useState<Template | null>(null)
  const [projectName, setProjectName] = useState('')
  const navigate = useNavigate()
  const { toast } = useToast()

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

  const openDeployModal = (template: Template) => {
    setProjectName(template.name)
    setDeployModal(template)
  }

  const handleDeploy = async () => {
    if (!deployModal) return
    const template = deployModal
    setDeployModal(null)

    setDeploying(template.id)
    try {
      await deployTemplate(template.id, projectName || undefined)
      toast({
        type: 'success',
        title: `"${template.name}" deployed successfully!`,
        description: 'Redirecting to dashboard...',
        duration: 3000,
      })
      // Navigate to dashboard after short delay
      setTimeout(() => navigate('/'), 1500)
    } catch (e: any) {
      toast({
        type: 'error',
        title: 'Deploy failed',
        description: e.message,
        duration: 6000,
      })
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
                onClick={() => openDeployModal(t)}
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

      {/* Deploy Modal */}
      <Modal
        open={!!deployModal}
        onClose={() => setDeployModal(null)}
        title={`Deploy ${deployModal?.name ?? ''}`}
        description="Choose a name for your project. Environment variables and services will be auto-configured."
      >
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-gray-400 mb-1.5">Project Name</label>
            <input
              type="text"
              value={projectName}
              onChange={(e) => setProjectName(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') handleDeploy() }}
              className="input-field"
              placeholder={deployModal?.name ?? 'my-project'}
              autoFocus
            />
          </div>
          {deployModal && (
            <div className="text-xs text-gray-500">
              <p className="mb-1.5">This template includes:</p>
              <div className="flex flex-wrap gap-1.5">
                {deployModal.services.map((s) => (
                  <span
                    key={s.name}
                    className="px-2 py-0.5 rounded border border-surface-300 bg-surface text-gray-400"
                  >
                    {s.type === 'app' ? '🚀 ' + s.name : '🗄️ ' + s.type}
                  </span>
                ))}
              </div>
            </div>
          )}
          <div className="flex gap-2 justify-end pt-2">
            <button
              onClick={() => setDeployModal(null)}
              className="px-4 py-2 text-sm text-gray-400 hover:text-gray-200 rounded-lg hover:bg-surface-200 transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleDeploy}
              className="flex items-center gap-2 px-5 py-2 bg-accent text-white rounded-lg text-sm font-medium hover:bg-accent-hover transition-colors"
            >
              <Rocket className="w-4 h-4" />
              Deploy
            </button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
