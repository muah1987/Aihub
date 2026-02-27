import { useState, useEffect, useRef } from 'react';
import { Bell, X, CheckCheck, Trash2, Rocket, Shield, Info } from 'lucide-react';
import { useNotificationStore } from '../../store/notificationStore';
import { useAuthStore } from '../../store/authStore';
import type { Notification } from '../../api/notifications';

const typeIcon = (type: string) => {
  switch (type) {
    case 'deploy_triggered':
    case 'deploy_success':
    case 'deploy_failed': return <Rocket size={14} />;
    case 'security': return <Shield size={14} />;
    default: return <Info size={14} />;
  }
};

const typeColor = (type: string) => {
  if (type.includes('failed')) return 'text-[var(--color-error)]';
  if (type.includes('success')) return 'text-[var(--color-success)]';
  if (type === 'deploy_triggered') return 'text-[var(--color-primary)]';
  return 'text-[var(--color-text-secondary)]';
};

function NotificationItem({ n, onRead, onRemove }: {
  n: Notification;
  onRead: (id: string) => void;
  onRemove: (id: string) => void;
}) {
  return (
    <div
      className={`flex gap-2 px-3 py-2.5 border-b border-[var(--color-border)] last:border-0 hover:bg-[var(--color-bg-tertiary)] transition-colors ${!n.read ? 'bg-[var(--color-primary)]/5' : ''}`}
      onClick={() => !n.read && onRead(n.id)}
    >
      <div className={`mt-0.5 shrink-0 ${typeColor(n.type)}`}>{typeIcon(n.type)}</div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1">
          <span className="text-xs font-medium truncate">{n.title}</span>
          {!n.read && <span className="w-1.5 h-1.5 rounded-full bg-[var(--color-primary)] shrink-0" />}
        </div>
        <p className="text-xs text-[var(--color-text-secondary)] line-clamp-2">{n.message}</p>
        <span className="text-xs text-[var(--color-text-secondary)] opacity-60">
          {new Date(n.created_at).toLocaleString()}
        </span>
      </div>
      <button
        onClick={(e) => { e.stopPropagation(); onRemove(n.id); }}
        className="p-1 hover:bg-[var(--color-bg-secondary)] rounded shrink-0 opacity-0 group-hover:opacity-100 text-[var(--color-text-secondary)]"
      >
        <Trash2 size={11} />
      </button>
    </div>
  );
}

export function NotificationBell() {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const tokens = useAuthStore((s) => s.tokens);
  const { notifications, unreadCount, load, markRead, markAllRead, remove, addLive, setUnreadCount } = useNotificationStore();

  // Load on mount
  useEffect(() => { load(); }, []);

  // WebSocket for live notifications
  useEffect(() => {
    if (!tokens?.access_token) return;
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = import.meta.env.VITE_API_URL?.replace(/^https?:\/\//, '') || window.location.host;
    const ws = new WebSocket(`${protocol}//${host}/api/v1/notifications/ws?token=${tokens.access_token}`);

    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data);
        if (msg.event === 'connected') {
          setUnreadCount(msg.unread_count || 0);
        } else if (msg.event === 'notification' && msg.notification) {
          addLive(msg.notification);
        }
      } catch {}
    };

    return () => ws.close();
  }, [tokens?.access_token]);

  // Close on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="relative p-2 hover:bg-[var(--color-bg-tertiary)] rounded-lg"
        title="Notifications"
      >
        <Bell size={18} />
        {unreadCount > 0 && (
          <span className="absolute -top-0.5 -right-0.5 min-w-[16px] h-4 px-1 bg-[var(--color-error)] text-white text-[10px] font-bold rounded-full flex items-center justify-center">
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-1 w-80 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-xl shadow-xl z-50 overflow-hidden">
          <div className="flex items-center justify-between px-3 py-2 border-b border-[var(--color-border)]">
            <span className="text-sm font-semibold">Notifications</span>
            <div className="flex items-center gap-1">
              {unreadCount > 0 && (
                <button
                  onClick={markAllRead}
                  className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-xs text-[var(--color-text-secondary)] flex items-center gap-1"
                  title="Mark all read"
                >
                  <CheckCheck size={13} /> All read
                </button>
              )}
              <button onClick={() => setOpen(false)} className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-text-secondary)]">
                <X size={13} />
              </button>
            </div>
          </div>

          <div className="max-h-96 overflow-y-auto">
            {notifications.length === 0 ? (
              <div className="text-center text-sm text-[var(--color-text-secondary)] py-8">
                No notifications yet
              </div>
            ) : (
              notifications.map((n) => (
                <NotificationItem key={n.id} n={n} onRead={markRead} onRemove={remove} />
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
