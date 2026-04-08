import { useCallback, useEffect, useState } from 'react'
import {
  listUsers, updateUserRole, deleteUser,
  inviteUser, listInvitations, listAuditLogs
} from '../api'
import type { User, Invitation, AuditLog } from '../types'
import { timeAgo } from '../lib/utils'
import {
  Users as UsersIcon, Shield, Trash2, Plus, X,
  Mail, Clock, FileText
} from 'lucide-react'

type Tab = 'users' | 'invitations' | 'audit'

export function Users() {
  const [tab, setTab] = useState<Tab>('users')
  const [users, setUsers] = useState<User[]>([])
  const [invitations, setInvitations] = useState<Invitation[]>([])
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)
  const [showInvite, setShowInvite] = useState(false)
  const [invEmail, setInvEmail] = useState('')
  const [invRole, setInvRole] = useState('member')

  const load = useCallback(async () => {
    try {
      const [u, inv, logs] = await Promise.all([
        listUsers().catch(() => []),
        listInvitations().catch(() => []),
        listAuditLogs().catch(() => []),
      ])
      setUsers(u)
      setInvitations(inv)
      setAuditLogs(logs)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const handleRoleChange = async (userId: string, role: string) => {
    try {
      await updateUserRole(userId, role)
      await load()
    } catch (e: any) {
      alert(e.message)
    }
  }

  const handleDeleteUser = async (userId: string) => {
    if (!confirm('Delete this user? This cannot be undone.')) return
    try {
      await deleteUser(userId)
      await load()
    } catch (e: any) {
      alert(e.message)
    }
  }

  const handleInvite = async () => {
    if (!invEmail.trim()) return
    try {
      await inviteUser(invEmail.trim(), invRole)
      setShowInvite(false)
      setInvEmail('')
      await load()
    } catch (e: any) {
      alert(e.message)
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
          <h1 className="text-2xl font-bold">Team Management</h1>
          <p className="text-sm text-gray-500 mt-1">{users.length} member{users.length !== 1 ? 's' : ''}</p>
        </div>
        <button
          onClick={() => setShowInvite(true)}
          className="flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-lg text-sm font-medium hover:bg-accent-hover transition-colors"
        >
          <Plus className="w-4 h-4" /> Invite User
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-surface-50 p-1 rounded-lg border border-surface-300 w-fit">
        {(['users', 'invitations', 'audit'] as Tab[]).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
              tab === t ? 'bg-accent/15 text-accent-hover' : 'text-gray-400 hover:text-gray-200'
            }`}
          >
            {t === 'users' ? 'Users' : t === 'invitations' ? 'Invitations' : 'Audit Log'}
          </button>
        ))}
      </div>

      {/* Invite modal */}
      {showInvite && (
        <div className="bg-surface-50 rounded-xl border border-surface-300 p-5 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-medium">Invite User</h3>
            <button onClick={() => setShowInvite(false)} className="text-gray-500 hover:text-gray-300">
              <X className="w-4 h-4" />
            </button>
          </div>
          <div className="flex gap-3">
            <input
              value={invEmail}
              onChange={(e) => setInvEmail(e.target.value)}
              placeholder="Email or identifier"
              className="flex-1 px-3 py-2 bg-surface border border-surface-300 rounded-lg text-sm focus:outline-none focus:border-accent"
            />
            <select
              value={invRole}
              onChange={(e) => setInvRole(e.target.value)}
              className="px-3 py-2 bg-surface border border-surface-300 rounded-lg text-sm focus:outline-none focus:border-accent"
            >
              <option value="admin">Admin</option>
              <option value="member">Member</option>
              <option value="viewer">Viewer</option>
            </select>
            <button
              onClick={handleInvite}
              className="px-4 py-2 bg-accent text-white rounded-lg text-sm font-medium hover:bg-accent-hover"
            >
              Send Invite
            </button>
          </div>
        </div>
      )}

      {/* Users tab */}
      {tab === 'users' && (
        <div className="bg-surface-50 rounded-xl border border-surface-300 divide-y divide-surface-300">
          {users.map((u) => (
            <div key={u.id} className="flex items-center justify-between px-5 py-4">
              <div className="flex items-center gap-3">
                <UsersIcon className="w-5 h-5 text-accent" />
                <div>
                  <div className="font-medium text-sm">{u.username}</div>
                  <div className="text-xs text-gray-500 mt-0.5">
                    Joined {timeAgo(u.created_at)}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="flex items-center gap-1.5">
                  <Shield className="w-3.5 h-3.5 text-gray-500" />
                  <select
                    value={u.role}
                    onChange={(e) => handleRoleChange(u.id, e.target.value)}
                    className="bg-surface border border-surface-300 rounded text-xs px-2 py-1 focus:outline-none focus:border-accent"
                  >
                    <option value="admin">Admin</option>
                    <option value="member">Member</option>
                    <option value="viewer">Viewer</option>
                  </select>
                </div>
                <button
                  onClick={() => handleDeleteUser(u.id)}
                  className="p-1.5 rounded hover:bg-surface-200 text-gray-400 hover:text-red-400"
                  title="Delete user"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Invitations tab */}
      {tab === 'invitations' && (
        <div className="bg-surface-50 rounded-xl border border-surface-300 divide-y divide-surface-300">
          {invitations.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              <Mail className="w-10 h-10 mx-auto mb-3 opacity-40" />
              <p className="text-sm">No invitations sent yet</p>
            </div>
          ) : (
            invitations.map((inv) => (
              <div key={inv.id} className="flex items-center justify-between px-5 py-4">
                <div className="flex items-center gap-3">
                  <Mail className="w-5 h-5 text-accent" />
                  <div>
                    <div className="font-medium text-sm">{inv.email}</div>
                    <div className="text-xs text-gray-500 mt-0.5">
                      Role: {inv.role} &middot; {inv.used ? 'Used' : 'Pending'} &middot; {timeAgo(inv.created_at)}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {!inv.used && (
                    <code className="text-xs bg-surface px-2 py-1 rounded border border-surface-300 max-w-[200px] truncate">
                      {inv.token}
                    </code>
                  )}
                  <span className={`text-xs px-2 py-0.5 rounded ${inv.used ? 'bg-green-500/15 text-green-400' : 'bg-yellow-500/15 text-yellow-400'}`}>
                    {inv.used ? 'Used' : 'Active'}
                  </span>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* Audit log tab */}
      {tab === 'audit' && (
        <div className="bg-surface-50 rounded-xl border border-surface-300 divide-y divide-surface-300">
          {auditLogs.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              <FileText className="w-10 h-10 mx-auto mb-3 opacity-40" />
              <p className="text-sm">No audit logs yet</p>
            </div>
          ) : (
            auditLogs.map((log) => (
              <div key={log.id} className="px-5 py-3">
                <div className="flex items-center gap-2 text-sm">
                  <Clock className="w-3.5 h-3.5 text-gray-500" />
                  <span className="font-medium text-accent">{log.username || 'system'}</span>
                  <span className="text-gray-500">{log.action}</span>
                  <span className="text-gray-300">{log.resource}</span>
                  {log.resource_id && (
                    <code className="text-xs bg-surface px-1.5 py-0.5 rounded border border-surface-300">{log.resource_id}</code>
                  )}
                </div>
                {log.details && (
                  <div className="text-xs text-gray-500 mt-1 ml-5">{log.details}</div>
                )}
                <div className="text-xs text-gray-600 mt-0.5 ml-5">{timeAgo(log.created_at)}</div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  )
}
