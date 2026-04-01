import { useCanvasStore } from '../store/canvasStore';

export default function ToastContainer() {
  const toasts = useCanvasStore((s) => s.toasts);
  const removeToast = useCanvasStore((s) => s.removeToast);

  if (toasts.length === 0) return null;

  return (
    <div className="toast-container">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={`toast ${t.type}`}
          onClick={() => removeToast(t.id)}
          style={{ cursor: 'pointer' }}
        >
          {t.message}
        </div>
      ))}
    </div>
  );
}
