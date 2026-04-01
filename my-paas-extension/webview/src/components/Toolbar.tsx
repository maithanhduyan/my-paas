import { useCanvasStore } from '../store/canvasStore';

interface ToolbarProps {
  postMessage: (msg: any) => void;
}

export default function Toolbar({ postMessage }: ToolbarProps) {
  const store = useCanvasStore();

  const handleGenerate = () => {
    const state = store.getCanvasState();
    postMessage({ type: 'compose:generate', state });
  };

  const handleComposeUp = () => {
    postMessage({ type: 'compose:up' });
  };

  const handleComposeDown = () => {
    postMessage({ type: 'compose:down' });
  };

  const handleSave = () => {
    const state = store.getCanvasState();
    postMessage({ type: 'canvas:save', state });
  };

  const handleRefreshStatus = () => {
    postMessage({ type: 'docker:status' });
  };

  return (
    <div className="toolbar">
      <button className="toolbar-btn" onClick={handleGenerate}>
        ⚙️ Generate Compose
      </button>
      <button className="toolbar-btn" onClick={handleComposeUp}>
        ▶️ Compose Up
      </button>
      <button className="toolbar-btn danger" onClick={handleComposeDown}>
        ⏹️ Compose Down
      </button>

      <div className="toolbar-spacer" />

      <button className="toolbar-btn secondary" onClick={handleRefreshStatus}>
        🔄 Refresh
      </button>
      <button className="toolbar-btn secondary" onClick={handleSave}>
        💾 Save
      </button>
    </div>
  );
}
