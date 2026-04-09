export interface Project {
  id: string
  name: string
  git_url: string
  branch: string
  provider: string
  framework: string
  auto_deploy: boolean
  status: string
  cpu_limit: number
  mem_limit: number
  replicas: number
  created_at: string
  updated_at: string
}

export interface CreateProjectInput {
  name: string
  git_url: string
  branch?: string
}

export interface Deployment {
  id: string
  project_id: string
  commit_hash: string
  commit_msg: string
  status: DeploymentStatus
  image_tag: string
  trigger: string
  started_at: string
  finished_at: string
  created_at: string
}

export type DeploymentStatus =
  | 'queued'
  | 'cloning'
  | 'detecting'
  | 'building'
  | 'deploying'
  | 'healthy'
  | 'failed'
  | 'rolled_back'
  | 'cancelled'

export interface EnvVar {
  id: string
  project_id: string
  key: string
  value: string
  is_secret: boolean
  created_at: string
}

export interface DeploymentLog {
  step: string
  level: string
  message: string
  time: string
}

export interface HealthResponse {
  status: string
  docker: string
  go: string
}

export type ServiceType = 'postgres' | 'redis' | 'mysql' | 'mongo' | 'minio'

export interface Service {
  id: string
  name: string
  type: ServiceType
  image: string
  status: string
  container_id: string
  config: string
  created_at: string
}

export interface ContainerStats {
  name: string
  id: string
  cpu_percent: number
  mem_usage: number
  mem_limit: number
  mem_percent: number
  net_input: number
  net_output: number
  status: string
}

export interface Domain {
  id: string
  project_id: string
  domain: string
  ssl_auto: boolean
  created_at: string
}

export interface User {
  id: string
  username: string
  role: string
  created_at: string
}

export interface AuthStatus {
  authenticated: boolean
  setup_required: boolean
  user?: User
}

export interface Backup {
  id: string
  type: string
  service_id: string
  filename: string
  size: number
  created_at: string
}

export interface AuditLog {
  id: string
  user_id: string
  username: string
  action: string
  resource: string
  resource_id: string
  details: string
  created_at: string
}

export interface Invitation {
  id: string
  email: string
  role: string
  token: string
  used: boolean
  created_by: string
  created_at: string
  expires_at: string
}

export interface Volume {
  id: string
  name: string
  mount_path: string
  project_id: string
  created_at: string
}

export interface Template {
  id: string
  name: string
  description: string
  icon: string
  services: TemplateService[]
}

export interface TemplateService {
  name: string
  type: string
  git_url: string
  image: string
  env: Record<string, string>
  volumes: string[]
}

export interface UpdateProjectInput {
  name?: string
  branch?: string
  auto_deploy?: boolean
  cpu_limit?: number
  mem_limit?: number
  replicas?: number
}

export interface Sample {
  id: string
  name: string
  description: string
  language: string
  icon: string
  git_url: string
}

export interface SwarmStatus {
  active: boolean
  nodes: SwarmNode[]
  manager_addr?: string
}

export interface SwarmNode {
  id: string
  hostname: string
  status: string
  role: string
  addr: string
}

export interface SwarmService {
  id: string
  name: string
  image: string
  replicas: number
  tasks: SwarmTask[]
  labels: Record<string, string>
}

export interface SwarmTask {
  id: string
  node_id: string
  node_name: string
  state: string
  message: string
}
