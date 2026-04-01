// Shared types between extension and webview

export interface CanvasState {
  version: string;
  name: string;
  nodes: ServiceNode[];
  edges: ServiceEdge[];
  volumeNodes?: VolumeNode[];
}

export interface ServiceNode {
  id: string;
  templateId: string;
  position: { x: number; y: number };
  data: ServiceNodeData;
}

export interface ServiceNodeData {
  name: string;
  image?: string;
  build?: { context: string; dockerfile: string };
  ports: string[];
  environment: Record<string, string>;
  volumes: string[];
  healthcheck?: HealthcheckConfig;
  labels?: Record<string, string>;
  command?: string;
  restart?: string;
}

export interface HealthcheckConfig {
  test: string[];
  interval?: string;
  timeout?: string;
  retries?: number;
  start_period?: string;
}

export interface ServiceEdge {
  id: string;
  source: string;
  target: string;
  type: 'depends_on' | 'network' | 'volume';
}

export interface VolumeNode {
  id: string;
  name: string;
  position: { x: number; y: number };
  /** Service directory path this volume is stored in (.mypaas/services/<service-name>/volumes/<volume-name>) */
  hostPath?: string;
  /** Connected service node id */
  connectedServiceId?: string;
  /** Mount path inside container */
  mountPath: string;
}

export interface TemplateFile {
  path: string;
  content: string;
}

export interface ServiceTemplate {
  id: string;
  name: string;
  icon: string;
  category: 'database' | 'app' | 'infrastructure' | 'tools';
  defaults: ServiceNodeData & {
    depends_on?: string[];
  };
  autoEnv?: AutoEnvMapping[];
  files?: TemplateFile[];
}

export interface AutoEnvMapping {
  key: string;
  template: string; // e.g. "postgresql://{POSTGRES_USER}:{POSTGRES_PASSWORD}@{name}:5432/{POSTGRES_DB}"
}

export interface ContainerStatus {
  serviceId: string;
  state: 'running' | 'stopped' | 'error' | 'unknown';
  containerId?: string;
  ports?: string[];
  cpu?: number;
  memory?: number;
}

// Messages between webview and extension
export type WebviewMessage =
  | { type: 'canvas:save'; state: CanvasState }
  | { type: 'canvas:load' }
  | { type: 'compose:generate'; state: CanvasState }
  | { type: 'compose:up' }
  | { type: 'compose:down' }
  | { type: 'docker:start'; serviceId: string }
  | { type: 'docker:stop'; serviceId: string }
  | { type: 'docker:restart'; serviceId: string }
  | { type: 'docker:remove'; serviceId: string }
  | { type: 'docker:logs'; serviceId: string }
  | { type: 'docker:status' }
  | { type: 'service:duplicate'; nodeId: string }
  | { type: 'service:openDir'; serviceId: string }
  | { type: 'templates:list' }
  | { type: 'ready' };

export type ExtensionMessage =
  | { type: 'canvas:loaded'; state: CanvasState | null }
  | { type: 'canvas:requestState' }
  | { type: 'template:add'; templateId: string }
  | { type: 'compose:generated'; path: string }
  | { type: 'docker:statuses'; data: ContainerStatus[] }
  | { type: 'docker:log'; serviceId: string; line: string }
  | { type: 'templates:data'; templates: ServiceTemplate[] }
  | { type: 'info'; message: string }
  | { type: 'error'; message: string };
