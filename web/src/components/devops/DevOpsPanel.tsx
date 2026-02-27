import { useState, useEffect, FormEvent } from 'react';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Modal } from '../ui/Modal';
import {
  Server, Play, CheckCircle, XCircle, Loader2, Clock,
  RefreshCw, ChevronDown, ChevronRight, Activity, GitBranch,
  Plus, Trash2, Copy, Eye, EyeOff, ToggleLeft, ToggleRight,
} from 'lucide-react';
import { deploymentApi, type VPSTarget, type DeploymentRun } from '../../api/deployment';
import { webhooksApi, monitoringApi, type Webhook, type ServerMetric, type PipelineStage } from '../../api/webhooks';

interface DevOpsPanelProps {
  projectId: string;
}

type SubTab = 'overview' | 'metrics' | 'webhooks' | 'pipeline';

export function DevOpsPanel({ projectId }: DevOpsPanelProps) {
  const [subTab, setSubTab] = useState<SubTab>('overview');
  const [targets, setTargets] = useState<VPSTarget[]>([]);
  const [runs, setRuns] = useState<DeploymentRun[]>([]);
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [selectedTarget, setSelectedTarget] = useState<string>('');
  const [metrics, setMetrics] = useState<ServerMetric[]>([]);
  const [latestMetric, setLatestMetric] = useState<ServerMetric | null>(null);
  const [stages, setStages] = useState<PipelineStage[]>([]);
  const [expandedRun, setExpandedRun] = useState<string | null>(null);
  const [deploying, setDeploying] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  // Webhook form
  const [showAddWebhook, setShowAddWebhook] = useState(false);
  const [newSecret, setNewSecret] = useState('');
  const [showSecret, setShowSecret] = useState(false);
  const [webhookBranch, setWebhookBranch] = useState('main');
  const [webhookTargetId, setWebhookTargetId] = useState('');

  // Pipeline stage form
  const [showAddStage, setShowAddStage] = useState(false);
  const [stageName, setStageName] = useState('');
  const [stageCmd, setStageCmd] = useState('');
  const [stageFailure, setStageFailure] = useState<'abort' | 'continue'>('abort');

  const load = async () => {
    setLoading(true);
    try {
      const [tRes, rRes, wRes] = await Promise.all([
        deploymentApi.listTargets(projectId),
        deploymentApi.listRuns(projectId, undefined, 15),
        webhooksApi.list(projectId),
      ]);
      const t = tRes.data.targets || [];
      setTargets(t);
      setRuns(rRes.data.runs || []);
      setWebhooks(wRes.data.webhooks || []);
      if (t.length > 0 && !selectedTarget) setSelectedTarget(t[0].id);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [projectId]);

  useEffect(() => {
    if (!selectedTarget) return;
    monitoringApi.history(projectId, selectedTarget, 20)
      .then(({ data }) => setMetrics(data.metrics || []))
      .catch(() => {});
    monitoringApi.latest(projectId, selectedTarget)
      .then(({ data }) => setLatestMetric(data.metric))
      .catch(() => setLatestMetric(null));
    monitoringApi.listStages(projectId, selectedTarget)
      .then(({ data }) => setStages(data.stages || []))
      .catch(() => {});
  }, [selectedTarget]);

  const handleDeploy = async (targetId: string) => {
    setDeploying(targetId);
    try {
      await deploymentApi.deploy(projectId, targetId);
      setTimeout(() => { load(); setDeploying(null); }, 1500);
    } catch { setDeploying(null); }
  };

  const handleCollectMetrics = async () => {
    if (!selectedTarget) return;
    const { data } = await monitoringApi.collect(projectId, selectedTarget);
    setLatestMetric(data.metric);
    const { data: hist } = await monitoringApi.history(projectId, selectedTarget, 20);
    setMetrics(hist.metrics || []);
  };

  const handleCreateWebhook = async (e: FormEvent) => {
    e.preventDefault();
    const { data } = await webhooksApi.create(projectId, webhookTargetId || null, webhookBranch);
    setNewSecret(data.secret);
    setShowSecret(true);
    setShowAddWebhook(false);
    load();
  };

  const handleToggleWebhook = async (wh: Webhook) => {
    await webhooksApi.setActive(projectId, wh.id, !wh.active);
    load();
  };

  const handleDeleteWebhook = async (id: string) => {
    await webhooksApi.delete(projectId, id);
    load();
  };

  const handleAddStage = async (e: FormEvent) => {
    e.preventDefault();
    await monitoringApi.createStage(projectId, selectedTarget, {
      name: stageName, command: stageCmd, stage_order: stages.length, on_failure: stageFailure,
    });
    setStageName(''); setStageCmd('');
    setShowAddStage(false);
    const { data } = await monitoringApi.listStages(projectId, selectedTarget);
    setStages(data.stages || []);
  };

  const handleDeleteStage = async (stageId: string) => {
    await monitoringApi.deleteStage(projectId, selectedTarget, stageId);
    const { data } = await monitoringApi.listStages(projectId, selectedTarget);
    setStages(data.stages || []);
  };

  const statusIcon = (status: string, size = 16) => {
    switch (status) {
      case 'success': return <CheckCircle size={size} className="text-[var(--color-success)]" />;
      case 'failed': return <XCircle size={size} className="text-[var(--color-error)]" />;
      case 'running': case 'deploying': return <Loader2 size={size} className="animate-spin text-[var(--color-primary)]" />;
      default: return <Clock size={size} className="text-[var(--color-text-secondary)]" />;
    }
  };

  const statusBadge = (status: string) => {
    const map: Record<string, string> = {
      success: 'text-[var(--color-success)] bg-[var(--color-success)]/10',
      failed: 'text-[var(--color-error)] bg-[var(--color-error)]/10',
      deploying: 'text-[var(--color-primary)] bg-[var(--color-primary)]/10',
      running: 'text-[var(--color-primary)] bg-[var(--color-primary)]/10',
    };
    return map[status] || 'text-[var(--color-text-secondary)] bg-[var(--color-bg-tertiary)]';
  };

  const gaugeColor = (v: number) => v > 80 ? 'bg-[var(--color-error)]' : v > 60 ? 'bg-yellow-500' : 'bg-[var(--color-success)]';

  const webhookUrl = `${window.location.origin}/webhooks/github/`;

  if (loading) {
    return <div className="flex items-center justify-center h-full text-[var(--color-text-secondary)]"><Loader2 className="animate-spin mr-2" size={20} /> Loading...</div>;
  }

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* Sub-tabs */}
      <div className="flex border-b border-[var(--color-border)] px-4 shrink-0">
        {(['overview', 'metrics', 'webhooks', 'pipeline'] as SubTab[]).map((t) => (
          <button
            key={t}
            onClick={() => setSubTab(t)}
            className={`px-3 py-2 text-sm font-medium border-b-2 transition-colors capitalize ${subTab === t ? 'border-[var(--color-primary)] text-[var(--color-primary)]' : 'border-transparent text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'}`}
          >
            {t}
          </button>
        ))}
        <button onClick={load} className="ml-auto p-2 hover:bg-[var(--color-bg-tertiary)] rounded-lg text-[var(--color-text-secondary)]">
          <RefreshCw size={14} />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">

        {/* ─── Overview ─── */}
        {subTab === 'overview' && (
          <>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {targets.length === 0 ? (
                <div className="col-span-2 text-center py-8 text-sm text-[var(--color-text-secondary)]">
                  <Server size={32} className="mx-auto mb-2 opacity-30" />
                  No VPS targets. Go to <strong>Settings → VPS Deployment</strong>.
                </div>
              ) : targets.map((t) => (
                <Card key={t.id}>
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex items-center gap-2 min-w-0">
                      {statusIcon(t.status)}
                      <div className="min-w-0">
                        <div className="font-medium text-sm truncate">{t.name}</div>
                        <div className="text-xs text-[var(--color-text-secondary)] truncate">{t.username}@{t.host}:{t.port}</div>
                      </div>
                    </div>
                    <div className="flex items-center gap-1 shrink-0">
                      <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${statusBadge(t.status)}`}>{t.status}</span>
                      <Button size="sm" loading={deploying === t.id} disabled={t.status === 'deploying'} onClick={() => handleDeploy(t.id)}>
                        <Play size={12} className="mr-1" /> Deploy
                      </Button>
                    </div>
                  </div>
                  {t.last_deployed_at && (
                    <p className="text-xs text-[var(--color-text-secondary)] mt-2">Last: {new Date(t.last_deployed_at).toLocaleString()}</p>
                  )}
                </Card>
              ))}
            </div>

            <div>
              <h3 className="font-medium text-sm mb-2">Recent Deployments</h3>
              <div className="space-y-1">
                {runs.length === 0 ? <p className="text-sm text-[var(--color-text-secondary)]">No deployments yet.</p> : runs.map((run) => {
                  const target = targets.find((t) => t.id === run.vps_target_id);
                  const dur = run.finished_at ? Math.round((new Date(run.finished_at).getTime() - new Date(run.started_at).getTime()) / 1000) : null;
                  return (
                    <div key={run.id} className="rounded-lg bg-[var(--color-bg-secondary)] p-3">
                      <div className="flex items-center gap-2 cursor-pointer" onClick={() => setExpandedRun(expandedRun === run.id ? null : run.id)}>
                        {statusIcon(run.status, 14)}
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${statusBadge(run.status)}`}>{run.status}</span>
                        <span className="text-sm font-medium">{target?.name || 'Unknown'}</span>
                        <span className="text-xs text-[var(--color-text-secondary)] ml-auto">{new Date(run.started_at).toLocaleString()}</span>
                        {dur !== null && <span className="text-xs text-[var(--color-text-secondary)]">{dur}s</span>}
                        {expandedRun === run.id ? <ChevronDown size={13} /> : <ChevronRight size={13} />}
                      </div>
                      {expandedRun === run.id && (
                        <pre className="mt-2 text-xs font-mono bg-[var(--color-bg)] p-3 rounded-lg overflow-x-auto whitespace-pre-wrap text-[var(--color-text-secondary)] max-h-64 overflow-y-auto">
                          {run.log_output || '(no output)'}
                        </pre>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          </>
        )}

        {/* ─── Metrics ─── */}
        {subTab === 'metrics' && (
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <select
                value={selectedTarget}
                onChange={(e) => setSelectedTarget(e.target.value)}
                className="px-3 py-1.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-sm text-[var(--color-text)]"
              >
                {targets.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
              </select>
              <Button size="sm" onClick={handleCollectMetrics}>
                <Activity size={13} className="mr-1" /> Collect Now
              </Button>
            </div>

            {latestMetric ? (
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                {[
                  { label: 'CPU', value: latestMetric.cpu_percent },
                  { label: 'Memory', value: latestMetric.mem_percent },
                  { label: 'Disk', value: latestMetric.disk_percent },
                ].map(({ label, value }) => (
                  <Card key={label}>
                    <div className="text-xs text-[var(--color-text-secondary)] mb-1">{label}</div>
                    <div className="text-2xl font-bold">{value.toFixed(0)}<span className="text-sm font-normal">%</span></div>
                    <div className="mt-2 h-1.5 rounded-full bg-[var(--color-bg-tertiary)] overflow-hidden">
                      <div className={`h-full rounded-full transition-all ${gaugeColor(value)}`} style={{ width: `${Math.min(value, 100)}%` }} />
                    </div>
                  </Card>
                ))}
                <Card>
                  <div className="text-xs text-[var(--color-text-secondary)] mb-1">Load Avg</div>
                  <div className="text-lg font-bold">{latestMetric.load_avg || '—'}</div>
                  <div className="text-xs text-[var(--color-text-secondary)] mt-1">
                    Up {Math.round(latestMetric.uptime_seconds / 3600)}h
                  </div>
                </Card>
              </div>
            ) : (
              <div className="text-center py-8 text-sm text-[var(--color-text-secondary)]">
                No metrics yet. Click "Collect Now" to gather server stats.
              </div>
            )}

            {metrics.length > 1 && (
              <Card>
                <div className="text-xs text-[var(--color-text-secondary)] mb-3">CPU History (last {metrics.length} snapshots)</div>
                <div className="flex items-end gap-0.5 h-16">
                  {metrics.map((m, i) => (
                    <div key={i} className="flex-1 flex flex-col items-center gap-0.5">
                      <div
                        className={`w-full rounded-sm transition-all ${gaugeColor(m.cpu_percent)}`}
                        style={{ height: `${Math.max(2, m.cpu_percent)}%` }}
                        title={`${m.cpu_percent.toFixed(1)}% at ${new Date(m.recorded_at).toLocaleTimeString()}`}
                      />
                    </div>
                  ))}
                </div>
              </Card>
            )}
          </div>
        )}

        {/* ─── Webhooks ─── */}
        {subTab === 'webhooks' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="font-semibold">GitHub Webhooks</h2>
                <p className="text-xs text-[var(--color-text-secondary)] mt-0.5">Auto-deploy when you push to a branch</p>
              </div>
              <Button size="sm" onClick={() => setShowAddWebhook(true)}><Plus size={14} className="mr-1" /> Add Webhook</Button>
            </div>

            {/* Revealed secret after creation */}
            {newSecret && (
              <Card>
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-sm font-medium text-[var(--color-success)]">Webhook created!</span>
                  <span className="text-xs text-[var(--color-text-secondary)]">Save this secret — shown once only.</span>
                </div>
                <div className="flex items-center gap-2">
                  <code className="text-xs font-mono flex-1 bg-[var(--color-bg)] p-2 rounded break-all">
                    {showSecret ? newSecret : '••••••••••••••••••••••••••••••••'}
                  </code>
                  <button onClick={() => setShowSecret(!showSecret)} className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded"><Eye size={13} /></button>
                  <button onClick={() => navigator.clipboard.writeText(newSecret)} className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded"><Copy size={13} /></button>
                </div>
                <div className="mt-2 text-xs text-[var(--color-text-secondary)]">
                  Webhook URL: <code className="font-mono">{webhookUrl}&lt;id&gt;</code>
                </div>
              </Card>
            )}

            <div className="space-y-2">
              {webhooks.length === 0 ? (
                <p className="text-sm text-center text-[var(--color-text-secondary)] py-6">No webhooks configured.</p>
              ) : webhooks.map((wh) => (
                <Card key={wh.id}>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <GitBranch size={16} className="text-[var(--color-primary)]" />
                      <div>
                        <div className="text-sm font-medium flex items-center gap-2">
                          {wh.branch}
                          <span className={`text-xs px-1.5 py-0.5 rounded-full ${wh.active ? 'bg-[var(--color-success)]/20 text-[var(--color-success)]' : 'bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)]'}`}>
                            {wh.active ? 'active' : 'paused'}
                          </span>
                        </div>
                        <div className="text-xs text-[var(--color-text-secondary)]">
                          {targets.find((t) => t.id === wh.vps_target_id)?.name || 'No target'}
                          {wh.last_triggered_at && ` · Last: ${new Date(wh.last_triggered_at).toLocaleString()}`}
                        </div>
                        <code className="text-xs font-mono text-[var(--color-text-secondary)] mt-0.5 block">{webhookUrl}{wh.id}</code>
                      </div>
                    </div>
                    <div className="flex items-center gap-1">
                      <button onClick={() => handleToggleWebhook(wh)} className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-text-secondary)]">
                        {wh.active ? <ToggleRight size={18} className="text-[var(--color-success)]" /> : <ToggleLeft size={18} />}
                      </button>
                      <button onClick={() => navigator.clipboard.writeText(`${webhookUrl}${wh.id}`)} className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-text-secondary)]" title="Copy URL">
                        <Copy size={14} />
                      </button>
                      <button onClick={() => handleDeleteWebhook(wh.id)} className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-error)]">
                        <Trash2 size={14} />
                      </button>
                    </div>
                  </div>
                </Card>
              ))}
            </div>
          </div>
        )}

        {/* ─── Pipeline ─── */}
        {subTab === 'pipeline' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="font-semibold">Pipeline Stages</h2>
                <p className="text-xs text-[var(--color-text-secondary)] mt-0.5">Ordered steps run during each deployment</p>
              </div>
              <div className="flex items-center gap-2">
                <select
                  value={selectedTarget}
                  onChange={(e) => setSelectedTarget(e.target.value)}
                  className="px-3 py-1.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-sm text-[var(--color-text)]"
                >
                  {targets.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
                </select>
                <Button size="sm" onClick={() => setShowAddStage(true)}><Plus size={14} className="mr-1" /> Add Stage</Button>
              </div>
            </div>

            <div className="space-y-2">
              {stages.length === 0 ? (
                <p className="text-sm text-center text-[var(--color-text-secondary)] py-6">No custom stages. The deploy_cmd in your target settings is always run.</p>
              ) : stages.map((st, i) => (
                <Card key={st.id} className="flex items-center gap-3">
                  <div className="w-6 h-6 rounded-full bg-[var(--color-primary)]/20 text-[var(--color-primary)] flex items-center justify-center text-xs font-bold shrink-0">{i + 1}</div>
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium">{st.name}</div>
                    <code className="text-xs font-mono text-[var(--color-text-secondary)] truncate block">{st.command}</code>
                    <span className={`text-xs ${st.on_failure === 'abort' ? 'text-[var(--color-error)]' : 'text-[var(--color-text-secondary)]'}`}>
                      on failure: {st.on_failure}
                    </span>
                  </div>
                  <button onClick={() => handleDeleteStage(st.id)} className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-error)] shrink-0">
                    <Trash2 size={14} />
                  </button>
                </Card>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Add Webhook Modal */}
      <Modal isOpen={showAddWebhook} onClose={() => setShowAddWebhook(false)} title="Add GitHub Webhook">
        <form onSubmit={handleCreateWebhook} className="space-y-3">
          <Input label="Branch" value={webhookBranch} onChange={(e) => setWebhookBranch(e.target.value)} placeholder="main" required />
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Auto-deploy Target (optional)</label>
            <select
              value={webhookTargetId}
              onChange={(e) => setWebhookTargetId(e.target.value)}
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-sm text-[var(--color-text)]"
            >
              <option value="">None</option>
              {targets.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
            </select>
          </div>
          <div className="flex gap-2">
            <Button type="submit" size="sm">Create Webhook</Button>
            <Button type="button" variant="ghost" size="sm" onClick={() => setShowAddWebhook(false)}>Cancel</Button>
          </div>
        </form>
      </Modal>

      {/* Add Pipeline Stage Modal */}
      <Modal isOpen={showAddStage} onClose={() => setShowAddStage(false)} title="Add Pipeline Stage">
        <form onSubmit={handleAddStage} className="space-y-3">
          <Input label="Stage Name" value={stageName} onChange={(e) => setStageName(e.target.value)} placeholder="Run Tests" required />
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Command</label>
            <Input value={stageCmd} onChange={(e) => setStageCmd(e.target.value)} placeholder="npm test" required />
          </div>
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">On Failure</label>
            <select
              value={stageFailure}
              onChange={(e) => setStageFailure(e.target.value as 'abort' | 'continue')}
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-sm text-[var(--color-text)]"
            >
              <option value="abort">Abort deployment</option>
              <option value="continue">Continue anyway</option>
            </select>
          </div>
          <div className="flex gap-2">
            <Button type="submit" size="sm">Add Stage</Button>
            <Button type="button" variant="ghost" size="sm" onClick={() => setShowAddStage(false)}>Cancel</Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
