import { useState, useEffect } from 'react';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import {
  Server, Play, CheckCircle, XCircle, Loader2, Clock,
  RefreshCw, ChevronDown, ChevronRight,
} from 'lucide-react';
import { deploymentApi, type VPSTarget, type DeploymentRun } from '../../api/deployment';

interface DevOpsPanelProps {
  projectId: string;
}

export function DevOpsPanel({ projectId }: DevOpsPanelProps) {
  const [targets, setTargets] = useState<VPSTarget[]>([]);
  const [runs, setRuns] = useState<DeploymentRun[]>([]);
  const [expandedRun, setExpandedRun] = useState<string | null>(null);
  const [deploying, setDeploying] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = async () => {
    setLoading(true);
    try {
      const [tRes, rRes] = await Promise.all([
        deploymentApi.listTargets(projectId),
        deploymentApi.listRuns(projectId, undefined, 20),
      ]);
      setTargets(tRes.data.targets || []);
      setRuns(rRes.data.runs || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [projectId]);

  const handleDeploy = async (targetId: string) => {
    setDeploying(targetId);
    try {
      await deploymentApi.deploy(projectId, targetId);
      setTimeout(() => { load(); setDeploying(null); }, 1500);
    } catch {
      setDeploying(null);
    }
  };

  const statusIcon = (status: string, size = 16) => {
    switch (status) {
      case 'success': return <CheckCircle size={size} className="text-[var(--color-success)]" />;
      case 'failed': return <XCircle size={size} className="text-[var(--color-error)]" />;
      case 'running': case 'deploying': return <Loader2 size={size} className="animate-spin text-[var(--color-primary)]" />;
      default: return <Clock size={size} className="text-[var(--color-text-secondary)]" />;
    }
  };

  const statusColor = (status: string) => {
    switch (status) {
      case 'success': return 'text-[var(--color-success)] bg-[var(--color-success)]/10';
      case 'failed': return 'text-[var(--color-error)] bg-[var(--color-error)]/10';
      case 'deploying': case 'running': return 'text-[var(--color-primary)] bg-[var(--color-primary)]/10';
      default: return 'text-[var(--color-text-secondary)] bg-[var(--color-bg-tertiary)]';
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full text-[var(--color-text-secondary)]">
        <Loader2 className="animate-spin mr-2" size={20} /> Loading...
      </div>
    );
  }

  if (targets.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center text-[var(--color-text-secondary)] gap-3">
        <Server size={40} className="opacity-30" />
        <div>
          <p className="font-medium">No VPS targets configured</p>
          <p className="text-sm mt-1">Go to <strong>Settings → VPS Deployment</strong> to add your server.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full overflow-y-auto p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="font-semibold">DevOps Dashboard</h2>
        <button onClick={load} className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded-lg text-[var(--color-text-secondary)]">
          <RefreshCw size={16} />
        </button>
      </div>

      {/* Target status cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {targets.map((t) => (
          <Card key={t.id}>
            <div className="flex items-start justify-between gap-2">
              <div className="flex items-center gap-2 min-w-0">
                {statusIcon(t.status)}
                <div className="min-w-0">
                  <div className="font-medium text-sm truncate">{t.name}</div>
                  <div className="text-xs text-[var(--color-text-secondary)] truncate">
                    {t.username}@{t.host}:{t.port}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-1 shrink-0">
                <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${statusColor(t.status)}`}>
                  {t.status}
                </span>
                <Button
                  size="sm"
                  loading={deploying === t.id}
                  disabled={t.status === 'deploying'}
                  onClick={() => handleDeploy(t.id)}
                >
                  <Play size={12} className="mr-1" />
                  Deploy
                </Button>
              </div>
            </div>
            {t.last_deployed_at && (
              <p className="text-xs text-[var(--color-text-secondary)] mt-2">
                Last: {new Date(t.last_deployed_at).toLocaleString()}
              </p>
            )}
          </Card>
        ))}
      </div>

      {/* Deployment history */}
      <div>
        <h3 className="font-medium text-sm mb-2">Deployment History</h3>
        <div className="space-y-1">
          {runs.length === 0 ? (
            <p className="text-sm text-[var(--color-text-secondary)]">No deployments yet.</p>
          ) : (
            runs.map((run) => {
              const target = targets.find((t) => t.id === run.vps_target_id);
              const duration = run.finished_at
                ? Math.round((new Date(run.finished_at).getTime() - new Date(run.started_at).getTime()) / 1000)
                : null;

              return (
                <div key={run.id} className="rounded-lg bg-[var(--color-bg-secondary)] p-3">
                  <div
                    className="flex items-center gap-2 cursor-pointer"
                    onClick={() => setExpandedRun(expandedRun === run.id ? null : run.id)}
                  >
                    {statusIcon(run.status, 14)}
                    <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${statusColor(run.status)}`}>
                      {run.status}
                    </span>
                    <span className="text-sm font-medium">{target?.name || 'Unknown'}</span>
                    <span className="text-xs text-[var(--color-text-secondary)] ml-auto">
                      {new Date(run.started_at).toLocaleString()}
                    </span>
                    {duration !== null && (
                      <span className="text-xs text-[var(--color-text-secondary)]">{duration}s</span>
                    )}
                    {expandedRun === run.id ? <ChevronDown size={13} /> : <ChevronRight size={13} />}
                  </div>

                  {expandedRun === run.id && (
                    <pre className="mt-2 text-xs font-mono bg-[var(--color-bg)] p-3 rounded-lg overflow-x-auto whitespace-pre-wrap text-[var(--color-text-secondary)] max-h-64 overflow-y-auto">
                      {run.log_output || '(no output)'}
                    </pre>
                  )}
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}
