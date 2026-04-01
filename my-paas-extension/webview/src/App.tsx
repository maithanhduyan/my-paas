import { useCallback, useRef, DragEvent } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  addEdge,
  useNodesState,
  useEdgesState,
  type Connection,
  type Node,
  type Edge,
  type NodeTypes,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { useVSCode } from './hooks/useVSCode';
import { useCanvasStore } from './store/canvasStore';
import Sidebar from './components/Sidebar';
import Toolbar from './components/Toolbar';
import ServiceNode from './components/ServiceNode';
import VolumeNode from './components/VolumeNode';
import ConfigPanel from './components/ConfigPanel';
import LogPanel from './components/LogPanel';
import ContextMenu from './components/ContextMenu';
import ToastContainer from './components/ToastContainer';
import type { ServiceNodeData, VolumeNodeData, ServiceTemplate } from './types';

const nodeTypes: NodeTypes = {
  serviceNode: ServiceNode as any,
  volumeNode: VolumeNode as any,
};

let nodeIdCounter = 0;

export default function App() {
  const { postMessage } = useVSCode();
  const {
    nodes, edges, setNodes, setEdges,
    addNode, selectedNodeId, setSelectedNodeId,
    templates, statuses,
    activeLogService,
    volumeNodes, addVolumeNode,
    setContextMenu,
  } = useCanvasStore();

  const reactFlowWrapper = useRef<HTMLDivElement>(null);

  // Merge service nodes + volume nodes for ReactFlow
  const allNodes = [...nodes, ...volumeNodes] as any[];
  const [rfNodes, setRfNodes, onNodesChange] = useNodesState(allNodes);
  const [rfEdges, setRfEdges, onEdgesChange] = useEdgesState(edges);

  // Sync store nodes to local react-flow state
  const syncedRef = useRef(false);
  if (!syncedRef.current && nodes.length > 0) {
    setRfNodes([...nodes, ...volumeNodes] as any);
    setRfEdges(edges);
    syncedRef.current = true;
  }

  const onConnect = useCallback((params: Connection) => {
    const newEdge: Edge = {
      ...params,
      id: `e-${params.source}-${params.target}`,
      animated: true,
      data: { type: 'depends_on' },
    } as Edge;
    setRfEdges((eds) => addEdge(newEdge, eds));
    setEdges([...rfEdges, newEdge]);
  }, [rfEdges, setEdges, setRfEdges]);

  const onNodeClick = useCallback((_: any, node: Node) => {
    setSelectedNodeId(node.id);
  }, [setSelectedNodeId]);

  const onPaneClick = useCallback(() => {
    setSelectedNodeId(null);
    setContextMenu(null);
  }, [setSelectedNodeId, setContextMenu]);

  const onNodeContextMenu = useCallback((event: React.MouseEvent, node: Node) => {
    event.preventDefault();
    const nodeType = node.type === 'volumeNode' ? 'volumeNode' : 'serviceNode';
    setContextMenu({
      x: event.clientX,
      y: event.clientY,
      nodeId: node.id,
      nodeType: nodeType as 'serviceNode' | 'volumeNode',
    });
  }, [setContextMenu]);

  const onDragOver = useCallback((event: DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback((event: DragEvent) => {
    event.preventDefault();

    const bounds = reactFlowWrapper.current?.getBoundingClientRect();
    if (!bounds) return;

    const position = {
      x: event.clientX - bounds.left - 90,
      y: event.clientY - bounds.top - 30,
    };

    // Handle volume drop
    const volumeData = event.dataTransfer.getData('application/mypaas-volume');
    if (volumeData) {
      const volId = `volume-${++nodeIdCounter}`;
      const newVolumeNode: Node<VolumeNodeData> = {
        id: volId,
        type: 'volumeNode',
        position,
        data: {
          name: `volume-${nodeIdCounter}`,
          mountPath: '/data',
          persistent: true,
        },
      };
      setRfNodes((nds) => [...nds, newVolumeNode as any]);
      addVolumeNode(newVolumeNode);
      return;
    }

    // Handle service template drop
    const templateId = event.dataTransfer.getData('application/mypaas-template');
    if (!templateId) return;

    const template = templates.find((t: ServiceTemplate) => t.id === templateId);
    if (!template) return;

    const id = `${templateId}-${++nodeIdCounter}`;
    const newNode: Node<ServiceNodeData> = {
      id,
      type: 'serviceNode',
      position,
      data: {
        name: template.defaults.name + (nodeIdCounter > 1 ? `-${nodeIdCounter}` : ''),
        image: template.defaults.image,
        build: template.defaults.build,
        ports: [...template.defaults.ports],
        environment: { ...template.defaults.environment },
        volumes: [...template.defaults.volumes],
        healthcheck: template.defaults.healthcheck
          ? { ...template.defaults.healthcheck }
          : undefined,
        command: template.defaults.command,
        restart: template.defaults.restart,
      },
    };

    setRfNodes((nds) => [...nds, newNode as any]);
    addNode(newNode);
  }, [templates, addNode, addVolumeNode, setRfNodes]);

  const handleNodesChange = useCallback((changes: any) => {
    onNodesChange(changes);
    // Sync position changes back to store
    setRfNodes((current) => {
      setNodes(current as any);
      return current;
    });
  }, [onNodesChange, setNodes, setRfNodes]);

  const handleEdgesChange = useCallback((changes: any) => {
    onEdgesChange(changes);
    setRfEdges((current) => {
      setEdges(current);
      return current;
    });
  }, [onEdgesChange, setEdges, setRfEdges]);

  const selectedNode = rfNodes.find(n => n.id === selectedNodeId);

  return (
    <div className="app-layout">
      <Sidebar />

      <div className="canvas-area">
        <Toolbar postMessage={postMessage} />

        <div ref={reactFlowWrapper} style={{ flex: 1 }}>
          <ReactFlow
            nodes={rfNodes}
            edges={rfEdges}
            onNodesChange={handleNodesChange}
            onEdgesChange={handleEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onPaneClick={onPaneClick}
            onNodeContextMenu={onNodeContextMenu}
            onDragOver={onDragOver}
            onDrop={onDrop}
            nodeTypes={nodeTypes}
            fitView
            deleteKeyCode="Delete"
          >
            <Background />
            <Controls />
            <MiniMap
              style={{ background: 'var(--bg-secondary)' }}
              maskColor="rgba(0,0,0,0.3)"
            />
          </ReactFlow>
        </div>

        {activeLogService && <LogPanel />}
      </div>

      {selectedNode && (
        <ConfigPanel
          node={selectedNode as any}
          postMessage={postMessage}
        />
      )}

      <ToastContainer />
      <ContextMenu postMessage={postMessage} />
    </div>
  );
}
