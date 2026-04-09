import type { Project, CreateProjectInput, UpdateProjectInput, Deployment, EnvVar, HealthResponse, Service, ServiceType, ContainerStats, Domain, AuthStatus, Backup, User, AuditLog, Invitation, Volume, Template, SwarmStatus, SwarmService, Sample } from './types'

const BASE = '/api'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = localStorage.getItem('mypaas_token')
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: { ...headers, ...(init?.headers as Record<string, string> | undefined) },
  })
  if (!res.ok) {
    if (res.status === 401 && !path.startsWith('/auth/')) {
      localStorage.removeItem('mypaas_token')
      window.location.href = '/login'
      throw new Error('Session expired')
    }
    const body = await res.json().catch(() => ({}))
    throw new Error((body as Record<string, string>).error ?? `HTTP ${res.status}`)
  }
  return res.json() as Promise<T>
}

// Health
export const getHealth = () => request<HealthResponse>('/health')

// Projects
export const listProjects = () => request<Project[]>('/projects')
export const getProject = (id: string) => request<Project>(`/projects/${id}`)
export const createProject = (input: CreateProjectInput) =>
  request<Project>('/projects', { method: 'POST', body: JSON.stringify(input) })
export const updateProject = (id: string, input: UpdateProjectInput) =>
  request<Project>(`/projects/${id}`, { method: 'PUT', body: JSON.stringify(input) })
export const deleteProject = (id: string) =>
  request<{ message: string }>(`/projects/${id}`, { method: 'DELETE' })

// Deployments
export const triggerDeploy = (projectId: string) =>
  request<Deployment>(`/projects/${projectId}/deploy`, { method: 'POST' })
export const listDeployments = (projectId: string) =>
  request<Deployment[]>(`/projects/${projectId}/deployments`)
export const getDeployment = (id: string) => request<Deployment>(`/deployments/${id}`)

// Environment
export const getEnvVars = (projectId: string) =>
  request<EnvVar[]>(`/projects/${projectId}/env`)
export const updateEnvVars = (projectId: string, vars: { key: string; value: string; is_secret?: boolean }[]) =>
  request<{ message: string; status: string }>(`/projects/${projectId}/env`, {
    method: 'PUT',
    body: JSON.stringify({ vars }),
  })
export const deleteEnvVar = (projectId: string, key: string) =>
  request<{ message: string }>(`/projects/${projectId}/env/${key}`, { method: 'DELETE' })

// Rollback
export const rollbackDeployment = (deploymentId: string) =>
  request<{ message: string }>(`/deployments/${deploymentId}/rollback`, { method: 'POST' })

// Project container actions
export const restartProject = (projectId: string) =>
  request<{ message: string }>(`/projects/${projectId}/restart`, { method: 'POST' })
export const stopProject = (projectId: string) =>
  request<{ message: string }>(`/projects/${projectId}/stop`, { method: 'POST' })
export const startProject = (projectId: string) =>
  request<{ message: string }>(`/projects/${projectId}/start`, { method: 'POST' })

// Services
export const listServices = () => request<Service[]>('/services')
export const createService = (name: string, type: ServiceType, image?: string) =>
  request<Service>('/services', { method: 'POST', body: JSON.stringify({ name, type, image }) })
export const deleteService = (id: string) =>
  request<{ message: string }>(`/services/${id}`, { method: 'DELETE' })
export const startService = (id: string) =>
  request<{ message: string; container_id?: string }>(`/services/${id}/start`, { method: 'POST' })
export const stopService = (id: string) =>
  request<{ message: string }>(`/services/${id}/stop`, { method: 'POST' })
export const linkService = (serviceId: string, projectId: string, envPrefix?: string) =>
  request<{ message: string }>(`/services/${serviceId}/link/${projectId}`, {
    method: 'POST',
    body: JSON.stringify({ env_prefix: envPrefix }),
  })
export const unlinkService = (serviceId: string, projectId: string) =>
  request<{ message: string }>(`/services/${serviceId}/link/${projectId}`, { method: 'DELETE' })

// Stats
export const getSystemStats = () => request<ContainerStats[]>('/stats')
export const getProjectStats = (projectId: string) =>
  request<ContainerStats>(`/projects/${projectId}/stats`)

// Auth
export const getAuthStatus = () => request<AuthStatus>('/auth/status')
export const logout = () => request<{ message: string }>('/auth/logout', { method: 'POST' })

// Domains
export const listDomains = (projectId: string) => request<Domain[]>(`/projects/${projectId}/domains`)
export const addDomain = (projectId: string, domain: string, sslAuto = true) =>
  request<Domain>(`/projects/${projectId}/domains`, {
    method: 'POST', body: JSON.stringify({ domain, ssl_auto: sslAuto })
  })
export const deleteDomain = (domainId: string) =>
  request<{ message: string }>(`/domains/${domainId}`, { method: 'DELETE' })

// Backups
export const listBackups = () => request<Backup[]>('/backups')
export const createBackup = () => request<Backup>('/backups', { method: 'POST' })
export const downloadBackup = async (id: string, filename: string) => {
  const res = await fetch(`${BASE}/backups/${id}/download`, {
    headers: { Authorization: `Bearer ${localStorage.getItem('mypaas_token')}` },
  })
  if (!res.ok) throw new Error('Download failed')
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}
export const restoreBackup = (id: string) =>
  request<{ message: string }>(`/backups/${id}/restore`, { method: 'POST' })
export const deleteBackup = (id: string) =>
  request<{ message: string }>(`/backups/${id}`, { method: 'DELETE' })

// Users & RBAC
export const listUsers = () => request<User[]>('/users')
export const updateUserRole = (id: string, role: string) =>
  request<{ message: string }>(`/users/${id}/role`, { method: 'PUT', body: JSON.stringify({ role }) })
export const deleteUser = (id: string) =>
  request<{ message: string }>(`/users/${id}`, { method: 'DELETE' })

// Invitations
export const inviteUser = (email: string, role: string) =>
  request<Invitation>('/invitations', { method: 'POST', body: JSON.stringify({ email, role }) })
export const listInvitations = () => request<Invitation[]>('/invitations')
export const registerWithInvite = (token: string, username: string, password: string) =>
  request<{ user: User; token: string; expires: string }>('/auth/register', {
    method: 'POST', body: JSON.stringify({ token, username, password })
  })

// Audit Logs
export const listAuditLogs = () => request<AuditLog[]>('/audit')

// Volumes
export const listVolumes = (projectId: string) => request<Volume[]>(`/projects/${projectId}/volumes`)
export const createVolume = (projectId: string, name: string, mountPath: string) =>
  request<Volume>(`/projects/${projectId}/volumes`, {
    method: 'POST', body: JSON.stringify({ name, mount_path: mountPath })
  })
export const deleteVolume = (projectId: string, volumeId: string) =>
  request<{ message: string }>(`/projects/${projectId}/volumes/${volumeId}`, { method: 'DELETE' })

// Marketplace
export const listTemplates = () => request<Template[]>('/marketplace')
export const deployTemplate = (id: string, name?: string) =>
  request<{ message: string; template: string }>(`/marketplace/${id}/deploy`, {
    method: 'POST', body: JSON.stringify({ name: name ?? '' })
  })

// Samples
export const listSamples = () => request<Sample[]>('/samples')

// Swarm
export const getSwarmStatus = () => request<SwarmStatus>('/swarm/status')
export const getSwarmServices = () => request<SwarmService[]>('/swarm/services')
export const initSwarm = (advertiseAddr?: string) =>
  request<{ message: string }>('/swarm/init', {
    method: 'POST', body: JSON.stringify({ advertise_addr: advertiseAddr ?? '' })
  })
export const getSwarmToken = () => request<{ token: string }>('/swarm/token')

// SSE streams
export function streamDeploymentLogs(
  deploymentId: string,
  onLog: (log: { step?: string; level?: string; message?: string; status?: string; done?: boolean }) => void,
): () => void {
  const token = localStorage.getItem('mypaas_token')
  const qs = token ? `?token=${encodeURIComponent(token)}` : ''
  const es = new EventSource(`${BASE}/deployments/${deploymentId}/logs${qs}`)
  es.onmessage = (e) => {
    try {
      onLog(JSON.parse(e.data))
    } catch { /* ignore */ }
  }
  es.onerror = () => es.close()
  return () => es.close()
}

export function streamProjectLogs(
  projectId: string,
  onLog: (line: string) => void,
): () => void {
  const token = localStorage.getItem('mypaas_token')
  const qs = token ? `?token=${encodeURIComponent(token)}` : ''
  const es = new EventSource(`${BASE}/projects/${projectId}/logs${qs}`)
  es.onmessage = (e) => onLog(e.data)
  es.onerror = () => es.close()
  return () => es.close()
}
