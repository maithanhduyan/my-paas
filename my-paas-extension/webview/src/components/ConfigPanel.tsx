import { useState, useCallback } from 'react';
import type { Node } from '@xyflow/react';
import { useCanvasStore } from '../store/canvasStore';
import type { ServiceNodeData } from '../types';

interface ConfigPanelProps {
  node: Node<ServiceNodeData>;
  postMessage: (msg: any) => void;
}

export default function ConfigPanel({ node, postMessage }: ConfigPanelProps) {
  const { updateNodeData, removeNode, setSelectedNodeId } = useCanvasStore();
  const data = node.data;

  const update = useCallback(
    (patch: Partial<ServiceNodeData>) => updateNodeData(node.id, patch),
    [node.id, updateNodeData]
  );

  // Environment variables editing
  const envEntries = Object.entries(data.environment || {});

  const setEnv = (key: string, val: string, oldKey?: string) => {
    const env = { ...data.environment };
    if (oldKey && oldKey !== key) {
      delete env[oldKey];
    }
    env[key] = val;
    update({ environment: env });
  };

  const removeEnv = (key: string) => {
    const env = { ...data.environment };
    delete env[key];
    update({ environment: env });
  };

  const addEnv = () => {
    const env = { ...data.environment, '': '' };
    update({ environment: env });
  };

  // Ports editing
  const setPorts = (index: number, value: string) => {
    const ports = [...data.ports];
    ports[index] = value;
    update({ ports });
  };

  const removePort = (index: number) => {
    const ports = data.ports.filter((_, i) => i !== index);
    update({ ports });
  };

  const addPort = () => {
    update({ ports: [...data.ports, ''] });
  };

  // Volumes editing
  const setVolume = (index: number, value: string) => {
    const volumes = [...data.volumes];
    volumes[index] = value;
    update({ volumes });
  };

  const removeVolume = (index: number) => {
    const volumes = data.volumes.filter((_, i) => i !== index);
    update({ volumes });
  };

  const addVolume = () => {
    update({ volumes: [...data.volumes, ''] });
  };

  const handleDelete = () => {
    removeNode(node.id);
    setSelectedNodeId(null);
  };

  const handleStartService = () => {
    postMessage({ type: 'docker:start', serviceId: data.name });
  };

  const handleStopService = () => {
    postMessage({ type: 'docker:stop', serviceId: data.name });
  };

  const handleRestartService = () => {
    postMessage({ type: 'docker:restart', serviceId: data.name });
  };

  const handleViewLogs = () => {
    postMessage({ type: 'docker:logs', serviceId: data.name });
    useCanvasStore.getState().setActiveLogService(data.name);
  };

  return (
    <div className="config-panel">
      <div className="config-panel-header">
        <span>Configure: {data.name}</span>
        <button className="config-panel-close" onClick={() => setSelectedNodeId(null)}>✕</button>
      </div>

      {/* General */}
      <div className="config-section">
        <div className="config-section-title">General</div>

        <div className="config-field">
          <label>Service Name</label>
          <input
            value={data.name}
            onChange={(e) => update({ name: e.target.value })}
          />
        </div>

        {data.image !== undefined && (
          <div className="config-field">
            <label>Image</label>
            <input
              value={data.image || ''}
              onChange={(e) => update({ image: e.target.value })}
            />
          </div>
        )}

        {data.build && (
          <>
            <div className="config-field">
              <label>Build Context</label>
              <input
                value={data.build.context}
                onChange={(e) => update({ build: { ...data.build!, context: e.target.value } })}
              />
            </div>
            <div className="config-field">
              <label>Dockerfile</label>
              <input
                value={data.build.dockerfile}
                onChange={(e) => update({ build: { ...data.build!, dockerfile: e.target.value } })}
              />
            </div>
          </>
        )}

        <div className="config-field">
          <label>Restart Policy</label>
          <select
            value={data.restart || 'no'}
            onChange={(e) => update({ restart: e.target.value })}
          >
            <option value="no">no</option>
            <option value="always">always</option>
            <option value="unless-stopped">unless-stopped</option>
            <option value="on-failure">on-failure</option>
          </select>
        </div>

        {data.command !== undefined && (
          <div className="config-field">
            <label>Command</label>
            <input
              value={data.command || ''}
              onChange={(e) => update({ command: e.target.value })}
            />
          </div>
        )}
      </div>

      {/* Ports */}
      <div className="config-section">
        <div className="config-section-title">Ports</div>
        {data.ports.map((port, i) => (
          <div key={i} className="env-row">
            <input
              value={port}
              placeholder="host:container"
              onChange={(e) => setPorts(i, e.target.value)}
            />
            <button className="env-remove-btn" onClick={() => removePort(i)}>✕</button>
          </div>
        ))}
        <button className="add-btn" onClick={addPort}>+ Add Port</button>
      </div>

      {/* Environment Variables */}
      <div className="config-section">
        <div className="config-section-title">Environment Variables</div>
        {envEntries.map(([key, val], i) => (
          <div key={i} className="env-row">
            <input
              value={key}
              placeholder="KEY"
              onChange={(e) => setEnv(e.target.value, val, key)}
              style={{ flex: 1 }}
            />
            <input
              value={val}
              placeholder="value"
              onChange={(e) => setEnv(key, e.target.value)}
              style={{ flex: 1.5 }}
            />
            <button className="env-remove-btn" onClick={() => removeEnv(key)}>✕</button>
          </div>
        ))}
        <button className="add-btn" onClick={addEnv}>+ Add Variable</button>
      </div>

      {/* Volumes */}
      <div className="config-section">
        <div className="config-section-title">Volumes</div>
        {data.volumes.map((vol, i) => (
          <div key={i} className="env-row">
            <input
              value={vol}
              placeholder="name:/path or ./host:/container"
              onChange={(e) => setVolume(i, e.target.value)}
            />
            <button className="env-remove-btn" onClick={() => removeVolume(i)}>✕</button>
          </div>
        ))}
        <button className="add-btn" onClick={addVolume}>+ Add Volume</button>
      </div>

      {/* Docker Actions */}
      <div className="config-section">
        <div className="config-section-title">Actions</div>
        <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
          <button className="toolbar-btn" onClick={handleStartService}>▶ Start</button>
          <button className="toolbar-btn secondary" onClick={handleStopService}>⏹ Stop</button>
          <button className="toolbar-btn secondary" onClick={handleRestartService}>🔄 Restart</button>
          <button className="toolbar-btn secondary" onClick={handleViewLogs}>📋 Logs</button>
        </div>
        <div style={{ marginTop: 12 }}>
          <button className="toolbar-btn danger" onClick={handleDelete}>🗑️ Remove Service</button>
        </div>
      </div>
    </div>
  );
}
