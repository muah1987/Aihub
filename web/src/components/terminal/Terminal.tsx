import { useState, useCallback, useRef, useEffect } from 'react';
import { Plus, X } from 'lucide-react';
import { useTerminal } from '../../hooks/useTerminal';
import { Button } from '../ui/Button';
import client from '../../api/client';
import '@xterm/xterm/css/xterm.css';

interface TerminalViewProps {
  projectId: string;
}

interface Session {
  id: string;
  status: string;
}

export function TerminalView({ projectId }: TerminalViewProps) {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [activeSession, setActiveSession] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const tokens = localStorage.getItem('tokens');
  const accessToken = tokens ? JSON.parse(tokens).access_token : '';
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = activeSession
    ? `${wsProtocol}//${window.location.host}/api/v1/projects/${projectId}/terminal/ws/${activeSession}?token=${accessToken}`
    : '';

  const { attach } = useTerminal({
    wsUrl,
    enabled: !!activeSession && !!accessToken,
  });

  useEffect(() => {
    // Load existing sessions
    client.get<{ sessions: Session[] }>(`/projects/${projectId}/terminal/sessions`).then(({ data }) => {
      const activeSessions = (data.sessions || []).filter((s) => s.status === 'running');
      setSessions(activeSessions);
      if (activeSessions.length > 0) {
        setActiveSession(activeSessions[0].id);
      }
    });
  }, [projectId]);

  useEffect(() => {
    if (activeSession && containerRef.current) {
      const cleanup = attach(containerRef.current);
      return cleanup;
    }
  }, [activeSession, attach]);

  const createSession = async () => {
    setCreating(true);
    try {
      const { data } = await client.post<{ session: Session }>(`/projects/${projectId}/terminal/sessions`);
      setSessions((prev) => [...prev, data.session]);
      setActiveSession(data.session.id);
    } catch {
      // Error
    } finally {
      setCreating(false);
    }
  };

  const stopSession = async (sessionId: string) => {
    await client.delete(`/projects/${projectId}/terminal/sessions/${sessionId}`);
    setSessions((prev) => prev.filter((s) => s.id !== sessionId));
    if (activeSession === sessionId) {
      const remaining = sessions.filter((s) => s.id !== sessionId);
      setActiveSession(remaining.length > 0 ? remaining[0].id : null);
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 px-3 py-2 border-b border-[var(--color-border)] overflow-x-auto">
        {sessions.map((s, i) => (
          <div
            key={s.id}
            className={`flex items-center gap-1 px-3 py-1 rounded-lg text-xs cursor-pointer shrink-0 ${
              activeSession === s.id
                ? 'bg-[var(--color-primary)]/20 text-[var(--color-primary)]'
                : 'bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)]'
            }`}
            onClick={() => setActiveSession(s.id)}
          >
            <span>Terminal {i + 1}</span>
            <button
              onClick={(e) => { e.stopPropagation(); stopSession(s.id); }}
              className="hover:text-[var(--color-error)]"
            >
              <X size={12} />
            </button>
          </div>
        ))}
        <Button size="sm" variant="ghost" onClick={createSession} loading={creating}>
          <Plus size={14} className="mr-1" /> New
        </Button>
      </div>

      <div className="flex-1 bg-[#0f172a]" ref={containerRef}>
        {!activeSession && (
          <div className="flex items-center justify-center h-full text-[var(--color-text-secondary)] text-sm">
            <Button onClick={createSession} loading={creating}>
              Create Terminal Session
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
