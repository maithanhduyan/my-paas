import { useCanvasStore } from '../store/canvasStore';

interface ContextMenuProps {
  postMessage: (msg: any) => void;
}

export default function ContextMenu({ postMessage }: ContextMenuProps) {
  const { contextMenu, setContextMenu, nodes, removeNode, duplicateNode,
          addVolumeNode, edges, setEdges } = useCanvasStore();

  if (!contextMenu) return null;

  const node = nodes.find(n => n.id === contextMenu.nodeId);
  const serviceName = node?.data?.name;
  const isService = contextMenu.nodeType === 'serviceNode';

  const close = () => setContextMenu(null);

  const handleStart = () => {
    if (serviceName) postMessage({ type: 'docker:start', serviceId: serviceName });
    close();
  };

  const handleStop = () => {
    if (serviceName) postMessage({ type: 'docker:stop', serviceId: serviceName });
    close();
  };

  const handleRemove = () => {
    if (serviceName) postMessage({ type: 'docker:remove', serviceId: serviceName });
    removeNode(contextMenu.nodeId);
    close();
  };

  const handleDuplicate = () => {
    duplicateNode(contextMenu.nodeId);
    close();
  };

  const handleAddVolume = () => {
    if (!node) { close(); return; }
    const volId = `vol-${Date.now()}`;
    const volNode = {
      id: volId,
      type: 'volumeNode' as const,
      position: { x: node.position.x, y: node.position.y + 180 },
      data: {
        name: `${serviceName}-data`,
        mountPath: '/data',
        persistent: true,
      },
    };
    addVolumeNode(volNode);

    // Add an edge connecting volume to service
    const newEdge = {
      id: `e-${volId}-${contextMenu.nodeId}`,
      source: volId,
      target: contextMenu.nodeId,
      type: 'default',
      animated: false,
      data: { type: 'volume' },
      style: { stroke: '#cca700', strokeDasharray: '5,5' },
    };
    setEdges([...edges, newEdge]);
    close();
  };

  const handleViewLogs = () => {
    if (serviceName) {
      postMessage({ type: 'docker:logs', serviceId: serviceName });
      useCanvasStore.getState().setActiveLogService(serviceName);
    }
    close();
  };

  const handleOpenDir = () => {
    if (serviceName) postMessage({ type: 'service:openDir', serviceId: serviceName });
    close();
  };

  return (
    <>
      {/* Backdrop */}
      <div className="context-menu-backdrop" onClick={close} />

      <div
        className="context-menu"
        style={{ left: contextMenu.x, top: contextMenu.y }}
      >
        {isService && (
          <>
            <button className="context-menu-item" onClick={handleStart}>
              ▶️ Start
            </button>
            <button className="context-menu-item" onClick={handleStop}>
              ⏹️ Stop
            </button>
            <div className="context-menu-separator" />
            <button className="context-menu-item" onClick={handleRemove}>
              🗑️ Remove
            </button>
            <button className="context-menu-item" onClick={handleDuplicate}>
              📋 Duplicate
            </button>
            <div className="context-menu-separator" />
            <button className="context-menu-item" onClick={handleAddVolume}>
              💾 Add Volume
            </button>
            <button className="context-menu-item" onClick={handleViewLogs}>
              📜 View Logs
            </button>
            <button className="context-menu-item" onClick={handleOpenDir}>
              📂 Open Directory
            </button>
          </>
        )}
      </div>
    </>
  );
}
