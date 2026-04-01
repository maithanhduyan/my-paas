import { useEffect, useRef } from 'react';
import { useCanvasStore } from '../store/canvasStore';

export default function LogPanel() {
  const activeLogService = useCanvasStore((s) => s.activeLogService);
  const logs = useCanvasStore((s) => s.logs);
  const setActiveLogService = useCanvasStore((s) => s.setActiveLogService);
  const bottomRef = useRef<HTMLDivElement>(null);

  const lines = activeLogService ? logs[activeLogService] || [] : [];

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [lines.length]);

  if (!activeLogService) return null;

  return (
    <div className="log-panel">
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: 8,
        paddingBottom: 4,
        borderBottom: '1px solid var(--border-color)',
      }}>
        <span style={{ fontWeight: 600, fontSize: 12 }}>
          Logs: {activeLogService}
        </span>
        <button
          className="config-panel-close"
          onClick={() => setActiveLogService(null)}
          style={{ fontSize: 14 }}
        >
          ✕
        </button>
      </div>
      {lines.length === 0 ? (
        <div style={{ color: 'var(--text-secondary)', fontStyle: 'italic' }}>
          No logs yet. Start the service to see logs.
        </div>
      ) : (
        lines.map((line, i) => (
          <div key={i} className="log-line">{line}</div>
        ))
      )}
      <div ref={bottomRef} />
    </div>
  );
}
