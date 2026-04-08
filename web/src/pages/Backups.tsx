import { useCallback, useEffect, useState } from 'react'
import { listBackups, createBackup, deleteBackup, restoreBackup, downloadBackup } from '../api'
import type { Backup } from '../types'
import { timeAgo } from '../lib/utils'
import { Archive, Download, Trash2, RotateCcw, Plus, Loader2 } from 'lucide-react'

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}

export function Backups() {
  const [backups, setBackups] = useState<Backup[]>([])
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)

  const load = useCallback(async () => {
    try {
      setBackups(await listBackups())
    } catch {
      setBackups([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const handleCreate = async () => {
    setCreating(true)
    try {
      await createBackup()
      await load()
    } finally {
      setCreating(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Delete this backup?')) return
    await deleteBackup(id)
    await load()
  }

  const handleRestore = async (id: string) => {
    if (!confirm('Restore from this backup? The server will need to be restarted.')) return
    try {
      const res = await restoreBackup(id)
      alert(res.message)
    } catch (e: any) {
      alert('Restore failed: ' + e.message)
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Backups</h1>
          <p className="text-sm text-gray-500 mt-1">{backups.length} backup{backups.length !== 1 ? 's' : ''}</p>
        </div>
        <button
          onClick={handleCreate}
          disabled={creating}
          className="flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-lg text-sm font-medium hover:bg-accent-hover transition-colors disabled:opacity-50"
        >
          {creating ? <Loader2 className="w-4 h-4 animate-spin" /> : <Plus className="w-4 h-4" />}
          Create Backup
        </button>
      </div>

      {backups.length === 0 ? (
        <div className="text-center py-16 text-gray-500">
          <Archive className="w-12 h-12 mx-auto mb-3 opacity-40" />
          <p>No backups yet</p>
          <p className="text-sm mt-1">Create a backup to protect your data</p>
        </div>
      ) : (
        <div className="bg-surface-50 rounded-xl border border-surface-300 divide-y divide-surface-300">
          {backups.map((b) => (
            <div key={b.id} className="flex items-center justify-between px-5 py-4">
              <div className="flex items-center gap-3">
                <Archive className="w-5 h-5 text-accent" />
                <div>
                  <div className="font-medium text-sm">{b.filename}</div>
                  <div className="text-xs text-gray-500 mt-0.5">
                    {b.type} &middot; {formatBytes(b.size)} &middot; {timeAgo(b.created_at)}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => downloadBackup(b.id, b.filename)}
                  className="p-1.5 rounded hover:bg-surface-200 text-gray-400 hover:text-gray-200"
                  title="Download"
                >
                  <Download className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleRestore(b.id)}
                  className="p-1.5 rounded hover:bg-surface-200 text-gray-400 hover:text-yellow-400"
                  title="Restore"
                >
                  <RotateCcw className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleDelete(b.id)}
                  className="p-1.5 rounded hover:bg-surface-200 text-gray-400 hover:text-red-400"
                  title="Delete"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
