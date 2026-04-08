import { useState } from 'react'
import type { EnvVar } from '../types'
import { updateEnvVars, deleteEnvVar } from '../api'
import { Plus, Trash2, Eye, EyeOff, Save } from 'lucide-react'

interface Props {
  projectId: string
  envVars: EnvVar[]
  onUpdate: () => void
}

interface EditableVar {
  key: string
  value: string
  is_secret: boolean
  isNew?: boolean
}

export function EnvEditor({ projectId, envVars, onUpdate }: Props) {
  const [vars, setVars] = useState<EditableVar[]>(
    envVars.map((v) => ({ key: v.key, value: v.value, is_secret: v.is_secret }))
  )
  const [revealed, setRevealed] = useState<Set<string>>(new Set())
  const [saving, setSaving] = useState(false)

  const addVar = () => {
    setVars([...vars, { key: '', value: '', is_secret: false, isNew: true }])
  }

  const removeVar = async (index: number) => {
    const v = vars[index]
    if (v && !v.isNew && v.key) {
      await deleteEnvVar(projectId, v.key)
    }
    setVars(vars.filter((_, i) => i !== index))
    onUpdate()
  }

  const updateVar = (index: number, field: keyof EditableVar, value: string | boolean) => {
    setVars(vars.map((v, i) => (i === index ? { ...v, [field]: value } : v)))
  }

  const toggleReveal = (key: string) => {
    const next = new Set(revealed)
    if (next.has(key)) next.delete(key)
    else next.add(key)
    setRevealed(next)
  }

  const save = async () => {
    const validVars = vars.filter((v) => v.key.trim())
    if (validVars.length === 0) return
    setSaving(true)
    try {
      await updateEnvVars(projectId, validVars)
      onUpdate()
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-3">
      <div className="space-y-2">
        {vars.map((v, i) => (
          <div key={i} className="flex items-center gap-2">
            <input
              type="text"
              placeholder="KEY"
              value={v.key}
              onChange={(e) => updateVar(i, 'key', e.target.value)}
              className="w-48 px-2 py-1.5 bg-surface-100 border border-surface-300 rounded text-sm font-mono
                         focus:outline-none focus:border-accent"
            />
            <span className="text-gray-500">=</span>
            <input
              type={v.is_secret && !revealed.has(v.key) ? 'password' : 'text'}
              placeholder="value"
              value={v.value}
              onChange={(e) => updateVar(i, 'value', e.target.value)}
              className="flex-1 px-2 py-1.5 bg-surface-100 border border-surface-300 rounded text-sm font-mono
                         focus:outline-none focus:border-accent"
            />
            <label className="flex items-center gap-1 text-xs text-gray-400 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={v.is_secret}
                onChange={(e) => updateVar(i, 'is_secret', e.target.checked)}
                className="rounded border-surface-300"
              />
              secret
            </label>
            {v.is_secret && (
              <button onClick={() => toggleReveal(v.key)} className="p-1 text-gray-400 hover:text-gray-200">
                {revealed.has(v.key) ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            )}
            <button onClick={() => removeVar(i)} className="p-1 text-gray-400 hover:text-danger">
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        ))}
      </div>

      <div className="flex gap-2">
        <button onClick={addVar} className="flex items-center gap-1 px-3 py-1.5 text-sm text-gray-300 border border-surface-300 rounded-md hover:bg-surface-200">
          <Plus className="w-3.5 h-3.5" /> Add Variable
        </button>
        <button
          onClick={save}
          disabled={saving}
          className="flex items-center gap-1 px-3 py-1.5 text-sm bg-accent text-white rounded-md hover:bg-accent-hover disabled:opacity-50"
        >
          <Save className="w-3.5 h-3.5" /> {saving ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </div>
  )
}
