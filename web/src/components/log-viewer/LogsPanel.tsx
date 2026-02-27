import { useState, useEffect, useRef } from 'react';
import { RefreshCw, Download, Loader2, FileText } from 'lucide-react';
import { deploymentApi, type DeploymentRun, type VPSTarget } from '../../api/deployment';

interface LogsPanelProps {
  projectId: string;
}

export function LogsPanel({ projectId }: LogsPanelProps) {
  const [targets, setTargets] = useState<VPSTarget[]>([]);
  const [runs, setRuns] = useState<DeploymentRun[]>([]);
  const [selectedRun, setSelectedRun] = useState<DeploymentRun | null>(null);
  const [selectedTarget, setSelectedTarget] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const logRef = useRef<HTMLPreElement>(null);

  const load = async () => {
    setLoading(true);
    try {
      const { data: tData } = await deploymentApi.listTargets(projectId);
      setTargets(tData.targets || []);
      const { data: rData } = await deploymentApi.listRuns(
        projectId,
        selectedTarget || undefined,
        30,
      );
      const sorted = rData.runs || [];
      setRuns(sorted);
      if (sorted.length > 0 && !selectedRun) {
        setSelectedRun(sorted[0]);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [projectId, selectedTarget]);

  useEffect(() => {
    if (logRef.current) {
      logRef.current.scrollTop = logRef.current.scrollHeight;
    }
  }, [selectedRun]);

  const handleDownload = () => {
    if (!selectedRun) return;
    const blob = new Blob([selectedRun.log_output], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `deploy-${selectedRun.id.slice(0, 8)}.log`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const statusColor = (status: string) => {
    switch (status) {
      case 'success': return 'text-[var(--color-success)]';
      case 'failed': return 'text-[var(--color-error)]';
      case 'running': return 'text-[var(--color-primary)]';
      default: return 'text-[var(--color-text-secondary)]';
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full text-[var(--color-text-secondary)]">
        <Loader2 className="animate-spin mr-2" size={20} /> Loading logs...
      </div>
    );
  }

  if (runs.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center text-[var(--color-text-secondary)] gap-3">
        <FileText size={40} className="opacity-30" />
        <div>
          <p className="font-medium">No deployment logs yet</p>
          <p className="text-sm mt-1">Deploy to a VPS target to see logs here.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full overflow-hidden">
      {/* Sidebar - run list */}
      <div className="w-56 shrink-0 border-r border-[var(--color-border)] flex flex-col">
        <div className="p-3 border-b border-[var(--color-border)]">
          <select
            value={selectedTarget}
            onChange={(e) => setSelectedTarget(e.target.value)}
            className="w-full px-2 py-1.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded text-xs text-[var(--color-text)]"
          >
            <option value="">All Targets</option>
            {targets.map((t) => (
              <option key={t.id} value={t.id}>{t.name}</option>
            ))}
          </select>
        </div>
        <div className="flex-1 overflow-y-auto">
          {runs.map((run) => {
            const target = targets.find((t) => t.id === run.vps_target_id);
            return (
              <button
                key={run.id}
                onClick={() => setSelectedRun(run)}
                className={`w-full text-left px-3 py-2.5 border-b border-[var(--color-border)] transition-colors ${
                  selectedRun?.id === run.id
                    ? 'bg-[var(--color-primary)]/10'
                    : 'hover:bg-[var(--color-bg-tertiary)]'
                }`}
              >
                <div className="flex items-center gap-1.5">
                  <span className={`text-xs font-medium ${statusColor(run.status)}`}>
                    {run.status}
                  </span>
                </div>
                <div className="text-xs text-[var(--color-text-secondary)] truncate mt-0.5">
                  {target?.name || 'Unknown'}
                </div>
                <div className="text-xs text-[var(--color-text-secondary)] mt-0.5">
                  {new Date(run.started_at).toLocaleDateString()} {new Date(run.started_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Log viewer */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {selectedRun ? (
          <>
            <div className="flex items-center justify-between px-4 py-2 border-b border-[var(--color-border)] shrink-0">
              <div className="flex items-center gap-3">
                <span className={`text-sm font-medium ${statusColor(selectedRun.status)}`}>
                  {selectedRun.status.toUpperCase()}
                </span>
                <span className="text-xs text-[var(--color-text-secondary)]">
                  {targets.find((t) => t.id === selectedRun.vps_target_id)?.name}
                </span>
                <span className="text-xs text-[var(--color-text-secondary)]">
                  {new Date(selectedRun.started_at).toLocaleString()}
                </span>
                {selectedRun.finished_at && (
                  <span className="text-xs text-[var(--color-text-secondary)]">
                    · {Math.round((new Date(selectedRun.finished_at).getTime() - new Date(selectedRun.started_at).getTime()) / 1000)}s
                  </span>
                )}
              </div>
              <div className="flex items-center gap-1">
                <button
                  onClick={load}
                  className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-text-secondary)]"
                  title="Refresh"
                >
                  <RefreshCw size={14} />
                </button>
                <button
                  onClick={handleDownload}
                  className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-text-secondary)]"
                  title="Download log"
                >
                  <Download size={14} />
                </button>
              </div>
            </div>
            <pre
              ref={logRef}
              className="flex-1 overflow-y-auto p-4 text-xs font-mono text-[var(--color-text-secondary)] bg-[var(--color-bg)] whitespace-pre-wrap break-all leading-relaxed"
            >
              {selectedRun.log_output || '(no output recorded)'}
            </pre>
          </>
        ) : (
          <div className="flex items-center justify-center h-full text-[var(--color-text-secondary)] text-sm">
            Select a deployment run to view its logs
          </div>
        )}
      </div>
    </div>
  );
}
