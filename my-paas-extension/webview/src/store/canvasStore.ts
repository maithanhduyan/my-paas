import { create } from 'zustand';
import type { Node, Edge } from '@xyflow/react';
import type { ServiceNodeData, ServiceTemplate, ContainerStatus, CanvasState, VolumeNodeData } from '../types';

interface Toast {
  id: string;
  type: 'info' | 'error';
  message: string;
}

interface ContextMenu {
  x: number;
  y: number;
  nodeId: string;
  nodeType: 'serviceNode' | 'volumeNode';
}

interface CanvasStore {
  // Templates
  templates: ServiceTemplate[];
  setTemplates: (templates: ServiceTemplate[]) => void;

  // Nodes & Edges
  nodes: Node<ServiceNodeData>[];
  edges: Edge[];
  setNodes: (nodes: Node<ServiceNodeData>[]) => void;
  setEdges: (edges: Edge[]) => void;
  addNode: (node: Node<ServiceNodeData>) => void;
  updateNodeData: (nodeId: string, data: Partial<ServiceNodeData>) => void;
  removeNode: (nodeId: string) => void;
  duplicateNode: (nodeId: string) => Node<ServiceNodeData> | null;

  // Volume nodes
  volumeNodes: Node<VolumeNodeData>[];
  addVolumeNode: (node: Node<VolumeNodeData>) => void;
  removeVolumeNode: (nodeId: string) => void;

  // Context menu
  contextMenu: ContextMenu | null;
  setContextMenu: (menu: ContextMenu | null) => void;

  // Selection
  selectedNodeId: string | null;
  setSelectedNodeId: (id: string | null) => void;

  // Docker statuses
  statuses: ContainerStatus[];
  setStatuses: (statuses: ContainerStatus[]) => void;

  // Logs
  logs: Record<string, string[]>;
  addLogLine: (serviceId: string, line: string) => void;
  activeLogService: string | null;
  setActiveLogService: (serviceId: string | null) => void;

  // Toasts
  toasts: Toast[];
  addToast: (type: 'info' | 'error', message: string) => void;
  removeToast: (id: string) => void;

  // Project
  projectName: string;
  setProjectName: (name: string) => void;

  // Helpers
  getCanvasState: () => CanvasState;
  loadCanvasState: (state: CanvasState) => void;
}

let toastCounter = 0;

export const useCanvasStore = create<CanvasStore>((set, get) => ({
  templates: [],
  setTemplates: (templates) => set({ templates }),

  nodes: [],
  edges: [],
  setNodes: (nodes) => set({ nodes }),
  setEdges: (edges) => set({ edges }),

  addNode: (node) => set((s) => ({ nodes: [...s.nodes, node] })),

  updateNodeData: (nodeId, data) => set((s) => ({
    nodes: s.nodes.map(n =>
      n.id === nodeId ? { ...n, data: { ...n.data, ...data } } : n
    ),
  })),

  removeNode: (nodeId) => set((s) => ({
    nodes: s.nodes.filter(n => n.id !== nodeId),
    edges: s.edges.filter(e => e.source !== nodeId && e.target !== nodeId),
    selectedNodeId: s.selectedNodeId === nodeId ? null : s.selectedNodeId,
    // Note: volume nodes connected to this service are NOT removed (volumes persist)
  })),

  duplicateNode: (nodeId) => {
    const state = get();
    const node = state.nodes.find(n => n.id === nodeId);
    if (!node) { return null; }
    const newId = `${node.id}-copy-${Date.now()}`;
    const newNode: Node<ServiceNodeData> = {
      ...node,
      id: newId,
      position: { x: node.position.x + 40, y: node.position.y + 40 },
      data: {
        ...node.data,
        name: node.data.name + '-copy',
        ports: [...node.data.ports],
        environment: { ...node.data.environment },
        volumes: [...node.data.volumes],
      },
      selected: false,
    };
    set((s) => ({ nodes: [...s.nodes, newNode] }));
    return newNode;
  },

  // Volume nodes
  volumeNodes: [],
  addVolumeNode: (node) => set((s) => ({ volumeNodes: [...s.volumeNodes, node] })),
  removeVolumeNode: (nodeId) => set((s) => ({
    volumeNodes: s.volumeNodes.filter(n => n.id !== nodeId),
    edges: s.edges.filter(e => e.source !== nodeId && e.target !== nodeId),
  })),

  // Context menu
  contextMenu: null,
  setContextMenu: (menu) => set({ contextMenu: menu }),

  selectedNodeId: null,
  setSelectedNodeId: (id) => set({ selectedNodeId: id }),

  statuses: [],
  setStatuses: (statuses) => set({ statuses }),

  logs: {},
  addLogLine: (serviceId, line) => set((s) => ({
    logs: {
      ...s.logs,
      [serviceId]: [...(s.logs[serviceId] || []), line].slice(-500),
    },
  })),
  activeLogService: null,
  setActiveLogService: (serviceId) => set({ activeLogService: serviceId }),

  toasts: [],
  addToast: (type, message) => {
    const id = String(++toastCounter);
    set((s) => ({ toasts: [...s.toasts, { id, type, message }] }));
    setTimeout(() => {
      set((s) => ({ toasts: s.toasts.filter(t => t.id !== id) }));
    }, 4000);
  },
  removeToast: (id) => set((s) => ({ toasts: s.toasts.filter(t => t.id !== id) })),

  projectName: 'my-project',
  setProjectName: (name) => set({ projectName: name }),

  getCanvasState: () => {
    const { nodes, edges, projectName, volumeNodes } = get();
    return {
      version: '1.0',
      name: projectName,
      nodes: nodes.map(n => ({
        id: n.id,
        templateId: (n as any).templateId || '',
        position: n.position,
        data: n.data,
      })),
      edges: edges.map(e => ({
        id: e.id,
        source: e.source,
        target: e.target,
        type: (e.data?.type as any) || 'depends_on',
      })),
      volumeNodes: volumeNodes.map(v => ({
        id: v.id,
        name: v.data.name,
        position: v.position,
        mountPath: v.data.mountPath,
      })),
    };
  },

  loadCanvasState: (state) => {
    const loadedVolumeNodes: Node<VolumeNodeData>[] = (state.volumeNodes || []).map(v => ({
      id: v.id,
      type: 'volumeNode',
      position: v.position,
      data: {
        name: v.name,
        mountPath: v.mountPath,
        persistent: true,
      },
    }));

    set({
      projectName: state.name,
      nodes: state.nodes.map(n => ({
        id: n.id,
        type: 'serviceNode',
        position: n.position,
        data: n.data,
        templateId: n.templateId,
      } as any)),
      edges: state.edges.map(e => ({
        id: e.id,
        source: e.source,
        target: e.target,
        type: 'default',
        animated: true,
        data: { type: e.type },
        style: e.type === 'volume' ? { stroke: '#cca700', strokeDasharray: '5,5' } : undefined,
      })),
      volumeNodes: loadedVolumeNodes,
    });
  },
}));
