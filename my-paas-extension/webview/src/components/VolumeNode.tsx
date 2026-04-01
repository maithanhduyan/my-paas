import { Handle, Position, type NodeProps } from '@xyflow/react';
import type { VolumeNodeData } from '../types';

export default function VolumeNode({ id, data, selected }: NodeProps & { data: VolumeNodeData }) {
  return (
    <div className={`volume-node ${selected ? 'selected' : ''}`}>
      <Handle type="target" position={Position.Left} />

      <div className="volume-node-header">
        <span className="volume-node-icon">💾</span>
        <span className="volume-node-title">{data.name}</span>
      </div>

      <div className="volume-node-body">
        <div style={{ opacity: 0.8 }}>Mount: {data.mountPath}</div>
        {data.persistent && (
          <div style={{ color: 'var(--warning)', fontSize: 10, marginTop: 2 }}>
            🔒 Persistent
          </div>
        )}
      </div>

      <Handle type="source" position={Position.Right} />
    </div>
  );
}
