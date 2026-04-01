import { DragEvent, useCallback } from 'react';
import { useCanvasStore } from '../store/canvasStore';
import type { ServiceTemplate } from '../types';

const CATEGORY_LABELS: Record<string, string> = {
  database: '🗄️ Databases',
  app: '📦 Applications',
  infrastructure: '🔧 Infrastructure',
  tools: '🛠️ Tools',
};

const CATEGORY_ORDER = ['database', 'app', 'infrastructure', 'tools'];

const ICONS: Record<string, string> = {
  database: '🗄️',
  redis: '🔴',
  code: '📦',
  globe: '🌐',
  archive: '📁',
  tools: '🛠️',
  mail: '📨',
};

export default function Sidebar() {
  const templates = useCanvasStore((s) => s.templates);

  const grouped = CATEGORY_ORDER.reduce((acc, cat) => {
    acc[cat] = templates.filter((t) => t.category === cat);
    return acc;
  }, {} as Record<string, ServiceTemplate[]>);

  const onDragStart = useCallback((event: DragEvent, templateId: string) => {
    event.dataTransfer.setData('application/mypaas-template', templateId);
    event.dataTransfer.effectAllowed = 'move';
  }, []);

  const onVolumeDragStart = useCallback((event: DragEvent) => {
    event.dataTransfer.setData('application/mypaas-volume', 'volume');
    event.dataTransfer.effectAllowed = 'move';
  }, []);

  return (
    <div className="sidebar">
      <div className="sidebar-header">Services</div>

      {CATEGORY_ORDER.map((cat) => {
        const items = grouped[cat];
        if (!items || items.length === 0) return null;

        return (
          <div key={cat} className="sidebar-section">
            <div className="sidebar-section-title">{CATEGORY_LABELS[cat]}</div>
            {items.map((t) => (
              <div
                key={t.id}
                className="template-item"
                draggable
                onDragStart={(e) => onDragStart(e, t.id)}
              >
                <span className="template-icon">{ICONS[t.icon] || '📦'}</span>
                <span>{t.name}</span>
              </div>
            ))}
          </div>
        );
      })}

      <div className="sidebar-section">
        <div className="sidebar-section-title">💾 Volumes</div>
        <div
          className="template-item"
          draggable
          onDragStart={onVolumeDragStart}
        >
          <span className="template-icon">💾</span>
          <span>Volume</span>
        </div>
      </div>
    </div>
  );
}
