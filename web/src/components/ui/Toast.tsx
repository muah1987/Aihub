import { useEffect, useState } from 'react';
import { X, CheckCircle, AlertCircle, Info } from 'lucide-react';

interface Toast {
  id: string;
  type: 'success' | 'error' | 'info';
  message: string;
}

let addToast: (toast: Omit<Toast, 'id'>) => void;

export function useToast() {
  return {
    success: (message: string) => addToast?.({ type: 'success', message }),
    error: (message: string) => addToast?.({ type: 'error', message }),
    info: (message: string) => addToast?.({ type: 'info', message }),
  };
}

export function ToastContainer() {
  const [toasts, setToasts] = useState<Toast[]>([]);

  useEffect(() => {
    addToast = (toast) => {
      const id = Math.random().toString(36).slice(2);
      setToasts((prev) => [...prev, { ...toast, id }]);
      setTimeout(() => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
      }, 4000);
    };
  }, []);

  const icons = {
    success: <CheckCircle size={18} className="text-[var(--color-success)]" />,
    error: <AlertCircle size={18} className="text-[var(--color-error)]" />,
    info: <Info size={18} className="text-[var(--color-primary)]" />,
  };

  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-2">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className="bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg p-3 flex items-center gap-2 shadow-lg animate-in slide-in-from-right min-w-[280px]"
        >
          {icons[toast.type]}
          <span className="text-sm flex-1">{toast.message}</span>
          <button
            onClick={() => setToasts((prev) => prev.filter((t) => t.id !== toast.id))}
            className="p-0.5 hover:bg-[var(--color-bg-tertiary)] rounded"
          >
            <X size={14} />
          </button>
        </div>
      ))}
    </div>
  );
}
