// Shared types (duplicated from extension for webview isolation)

export interface CanvasState {
  version: string;
  name: string;
  nodes: ServiceNodeDef[];
  edges: ServiceEdgeDef[];
  volumeNodes?: VolumeNodeDef[];
}

export interface ServiceNodeDef {
  id: string;
  templateId: string;
  position: { x: number; y: number };
  data: ServiceNodeData;
}

export interface ServiceNodeData {
  [key: string]: unknown;
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

export interface ServiceEdgeDef {
  id: string;
  source: string;
  target: string;
  type: 'depends_on' | 'network' | 'volume';
}

export interface VolumeNodeDef {
  id: string;
  name: string;
  position: { x: number; y: number };
  hostPath?: string;
  connectedServiceId?: string;
  mountPath: string;
}

export interface VolumeNodeData {
  [key: string]: unknown;
  name: string;
  mountPath: string;
  /** Whether this volume's data should be kept when the service is removed */
  persistent: boolean;
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
}

export interface AutoEnvMapping {
  key: string;
  template: string;
}

export interface ContainerStatus {
  serviceId: string;
  state: 'running' | 'stopped' | 'error' | 'unknown';
  containerId?: string;
  ports?: string[];
  cpu?: number;
  memory?: number;
}
