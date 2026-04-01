import { Handle, Position, type NodeProps } from '@xyflow/react';
import { useCanvasStore } from '../store/canvasStore';
import type { ServiceNodeData } from '../types';

export default function ServiceNode({ id, data, selected }: NodeProps & { data: ServiceNodeData }) {
  const statuses = useCanvasStore((s) => s.statuses);
  const setActiveLogService = useCanvasStore((s) => s.setActiveLogService);
  const setContextMenu = useCanvasStore((s) => s.setContextMenu);

  const status = statuses.find((s) => s.serviceId === data.name);
  const stateClass = status?.state || 'stopped';

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setContextMenu({
      x: e.clientX,
      y: e.clientY,
      nodeId: id,
      nodeType: 'serviceNode',
    });
  };

  return (
    <div
      className={`service-node ${selected ? 'selected' : ''}`}
      onContextMenu={handleContextMenu}
    >
      <Handle type="target" position={Position.Left} />

      <div className="service-node-header">
        <span className="service-node-icon">
          {data.image?.includes('postgres') ? '🐘' :
           data.image?.includes('redis') ? '🔴' :
           data.image?.includes('mongo') ? '🍃' :
           data.image?.includes('mysql') ? '🐬' :
           data.image?.includes('nginx') ? '🌐' :
           data.image?.includes('minio') ? '📁' :
           data.image?.includes('traefik') ? '🔀' :
           data.image?.includes('rabbit') ? '🐰' :
           data.image?.includes('pgadmin') ? '🛠️' :
           data.image?.includes('adminer') ? '🛠️' :
           data.build ? '📦' : '🐳'}
        </span>
        <span className="service-node-title">{data.name}</span>
        <span className={`service-node-status ${stateClass}`} title={stateClass} />
      </div>

      <div className="service-node-body">
        {data.image && (
          <div style={{ marginBottom: 2, opacity: 0.8 }}>{data.image}</div>
        )}
        {data.build && (
          <div style={{ marginBottom: 2, opacity: 0.8 }}>Build: {data.build.context}</div>
        )}
        {data.ports.length > 0 && (
          <div className="service-node-ports">
            Ports: {data.ports.join(', ')}
          </div>
        )}
      </div>

      <div className="service-node-actions">
        <button
          className="node-action-btn"
          title="View logs"
          onClick={(e) => { e.stopPropagation(); setActiveLogService(data.name); }}
        >
          Logs
        </button>
      </div>

      <Handle type="source" position={Position.Right} />
    </div>
  );
}
