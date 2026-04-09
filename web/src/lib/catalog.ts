// ─── Service Catalog: definitions, categories, and stack templates ────────────

export type NodeCategory = 'app' | 'database' | 'cache' | 'storage'

export interface ServiceDef {
  category: NodeCategory
  name: string
  icon: string
  color: string
  description: string
  defaultConfig: Record<string, string>
  defaultName: string
}

export const SERVICE_DEFS: Record<string, ServiceDef> = {
  github: {
    category: 'app',
    name: 'GitHub Repository',
    icon: '⬡',
    color: '#6366f1',
    description: 'Deploy from Git repository',
    defaultConfig: {},
    defaultName: 'web-app',
  },
  'docker-image': {
    category: 'app',
    name: 'Docker Image',
    icon: '🐳',
    color: '#0db7ed',
    description: 'Deploy from Docker registry',
    defaultConfig: {},
    defaultName: 'container',
  },
  postgres: {
    category: 'database',
    name: 'PostgreSQL',
    icon: '🐘',
    color: '#336791',
    description: 'Relational SQL database',
    defaultConfig: { version: '16' },
    defaultName: 'postgres',
  },
  mysql: {
    category: 'database',
    name: 'MySQL',
    icon: '🐬',
    color: '#4479A1',
    description: 'Relational SQL database',
    defaultConfig: { version: '8' },
    defaultName: 'mysql',
  },
  redis: {
    category: 'cache',
    name: 'Redis',
    icon: '⚡',
    color: '#DC382D',
    description: 'In-memory cache & queue',
    defaultConfig: { version: '7' },
    defaultName: 'redis',
  },
  mongo: {
    category: 'database',
    name: 'MongoDB',
    icon: '🍃',
    color: '#47A248',
    description: 'Document database',
    defaultConfig: { version: '7' },
    defaultName: 'mongo',
  },
  minio: {
    category: 'storage',
    name: 'MinIO',
    icon: '🪣',
    color: '#C72C48',
    description: 'S3-compatible storage',
    defaultConfig: { version: 'latest' },
    defaultName: 'minio',
  },
}

export const CATALOG_SECTIONS = [
  { title: 'Applications', items: ['github', 'docker-image'] },
  { title: 'Databases', items: ['postgres', 'mysql', 'mongo'] },
  { title: 'Cache', items: ['redis'] },
  { title: 'Storage', items: ['minio'] },
]

export interface StackTemplate {
  id: string
  name: string
  description: string
  icon: string
  color: string
  nodes: Array<{ serviceType: string; name: string; relX: number; relY: number; gitUrl?: string }>
  connections: Array<[number, number]>
}

export const STACK_TEMPLATES: StackTemplate[] = [
  {
    id: 'node-pg',
    name: 'Node.js + PostgreSQL',
    description: 'REST API with SQL database',
    icon: '🟢',
    color: '#22c55e',
    nodes: [
      { serviceType: 'github', name: 'api', relX: 0, relY: 0 },
      { serviceType: 'postgres', name: 'db', relX: 340, relY: 0 },
    ],
    connections: [[0, 1]],
  },
  {
    id: 'fullstack',
    name: 'Full Stack + DB + Cache',
    description: 'App + PostgreSQL + Redis',
    icon: '⚛️',
    color: '#6366f1',
    nodes: [
      { serviceType: 'github', name: 'app', relX: 0, relY: 0 },
      { serviceType: 'postgres', name: 'db', relX: 340, relY: -80 },
      { serviceType: 'redis', name: 'cache', relX: 340, relY: 80 },
    ],
    connections: [[0, 1], [0, 2]],
  },
  {
    id: 'python-pg',
    name: 'Python + PostgreSQL',
    description: 'Django / FastAPI + database',
    icon: '🐍',
    color: '#3572A5',
    nodes: [
      { serviceType: 'github', name: 'app', relX: 0, relY: 0 },
      { serviceType: 'postgres', name: 'db', relX: 340, relY: 0 },
    ],
    connections: [[0, 1]],
  },
  {
    id: 'microservices',
    name: 'Microservices',
    description: '2 services + shared DB + cache',
    icon: '🔷',
    color: '#0ea5e9',
    nodes: [
      { serviceType: 'github', name: 'web', relX: 0, relY: -80 },
      { serviceType: 'github', name: 'api', relX: 0, relY: 80 },
      { serviceType: 'postgres', name: 'db', relX: 340, relY: 0 },
      { serviceType: 'redis', name: 'cache', relX: 340, relY: 160 },
    ],
    connections: [[0, 2], [1, 2], [1, 3]],
  },
]

/** Compute a nice auto-position for a new node given existing nodes */
export function getAutoPosition(
  existingNodes: Array<{ x: number; y: number }>,
  canvasCenter: { x: number; y: number },
): { x: number; y: number } {
  if (existingNodes.length === 0) {
    return { x: canvasCenter.x - 110, y: canvasCenter.y - 50 }
  }
  const rightmost = existingNodes.reduce((a, b) => (a.x > b.x ? a : b))
  return { x: rightmost.x + 300, y: rightmost.y }
}

export const GRID_SIZE = 20
export const snap = (v: number): number => Math.round(v / GRID_SIZE) * GRID_SIZE
export const NODE_W = 220
export const NODE_H = 100
