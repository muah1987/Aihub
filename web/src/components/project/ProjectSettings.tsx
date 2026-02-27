import { useState, useEffect, FormEvent } from 'react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Card } from '../ui/Card';
import {
  Plus, Trash2, Eye, EyeOff, Server, Play, CheckCircle,
  XCircle, Loader2, Clock, ChevronDown, ChevronRight, RefreshCw,
} from 'lucide-react';
import {
  deploymentApi,
  type EnvVar,
  type VPSTarget,
  type DeploymentRun,
  type TargetInput,
} from '../../api/deployment';

interface ProjectSettingsProps {
  projectId: string;
}

type Tab = 'env' | 'vps';

export function ProjectSettings({ projectId }: ProjectSettingsProps) {
  const [activeTab, setActiveTab] = useState<Tab>('env');

  return (
    <div className="flex flex-col h-full overflow-y-auto p-4 space-y-4 max-w-3xl mx-auto w-full">
      {/* Sub-tab nav */}
      <div className="flex border-b border-[var(--color-border)]">
        {(['env', 'vps'] as Tab[]).map((t) => (
          <button
            key={t}
            onClick={() => setActiveTab(t)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === t
                ? 'border-[var(--color-primary)] text-[var(--color-primary)]'
                : 'border-transparent text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'
            }`}
          >
            {t === 'env' ? 'Environment Variables' : 'VPS Deployment'}
          </button>
        ))}
      </div>

      {activeTab === 'env' ? (
        <EnvVarsSection projectId={projectId} />
      ) : (
        <VPSSection projectId={projectId} />
      )}
    </div>
  );
}

// ─── Env Variables Section ────────────────────────────────────────────────────

function EnvVarsSection({ projectId }: { projectId: string }) {
  const [vars, setVars] = useState<EnvVar[]>([]);
  const [showAdd, setShowAdd] = useState(false);
  const [key, setKey] = useState('');
  const [value, setValue] = useState('');
  const [isSecret, setIsSecret] = useState(false);
  const [revealed, setRevealed] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);

  const load = async () => {
    const { data } = await deploymentApi.listEnvVars(projectId);
    setVars(data.env_vars || []);
  };
  useEffect(() => { load(); }, [projectId]);

  const handleAdd = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await deploymentApi.setEnvVar(projectId, key, value, isSecret);
      setShowAdd(false);
      setKey(''); setValue(''); setIsSecret(false);
      load();
    } finally { setLoading(false); }
  };

  const handleDelete = async (id: string) => {
    await deploymentApi.deleteEnvVar(projectId, id);
    load();
  };

  const toggleReveal = (id: string) => {
    setRevealed((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-semibold">Environment Variables</h2>
          <p className="text-xs text-[var(--color-text-secondary)] mt-0.5">
            Encrypted at rest, injected during deployment
          </p>
        </div>
        <Button size="sm" onClick={() => setShowAdd(!showAdd)}>
          <Plus size={14} className="mr-1" /> Add Variable
        </Button>
      </div>

      {showAdd && (
        <Card>
          <form onSubmit={handleAdd} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <Input
                label="Key"
                value={key}
                onChange={(e) => setKey(e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g, '_'))}
                placeholder="DATABASE_URL"
                required
              />
              <Input
                label="Value"
                type={isSecret ? 'password' : 'text'}
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder="Enter value..."
                required
              />
            </div>
            <label className="flex items-center gap-2 text-sm cursor-pointer">
              <input
                type="checkbox"
                checked={isSecret}
                onChange={(e) => setIsSecret(e.target.checked)}
                className="rounded"
              />
              Mark as secret (mask value in UI)
            </label>
            <div className="flex gap-2">
              <Button type="submit" size="sm" loading={loading}>Save</Button>
              <Button type="button" variant="ghost" size="sm" onClick={() => setShowAdd(false)}>Cancel</Button>
            </div>
          </form>
        </Card>
      )}

      <div className="space-y-2">
        {vars.length === 0 ? (
          <div className="text-center text-sm text-[var(--color-text-secondary)] py-8">
            No environment variables yet. Add your first variable above.
          </div>
        ) : (
          vars.map((v) => {
            const show = revealed.has(v.id);
            return (
              <Card key={v.id} className="flex items-center gap-3">
                <div className="flex-1 min-w-0 grid grid-cols-2 gap-3">
                  <div>
                    <div className="text-xs text-[var(--color-text-secondary)]">Key</div>
                    <code className="text-sm font-mono">{v.key}</code>
                  </div>
                  <div>
                    <div className="text-xs text-[var(--color-text-secondary)]">Value</div>
                    <code className="text-sm font-mono text-[var(--color-text-secondary)] break-all">
                      {v.is_secret && !show ? '••••••••' : v.value}
                    </code>
                  </div>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  {v.is_secret && (
                    <button
                      onClick={() => toggleReveal(v.id)}
                      className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-text-secondary)]"
                      title={show ? 'Hide' : 'Show'}
                    >
                      {show ? <EyeOff size={14} /> : <Eye size={14} />}
                    </button>
                  )}
                  <button
                    onClick={() => handleDelete(v.id)}
                    className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-error)]"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </Card>
            );
          })
        )}
      </div>
    </div>
  );
}

// ─── VPS Deployment Section ───────────────────────────────────────────────────

function VPSSection({ projectId }: { projectId: string }) {
  const [targets, setTargets] = useState<VPSTarget[]>([]);
  const [runs, setRuns] = useState<DeploymentRun[]>([]);
  const [showAdd, setShowAdd] = useState(false);
  const [selectedTarget, setSelectedTarget] = useState<string | null>(null);
  const [expandedRun, setExpandedRun] = useState<string | null>(null);
  const [deploying, setDeploying] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // Form state
  const [form, setForm] = useState<TargetInput>({
    name: '',
    host: '',
    port: 22,
    username: 'root',
    auth_type: 'key',
    ssh_key: '',
    ssh_password: '',
    deploy_path: '/app',
    pre_deploy_cmd: '',
    deploy_cmd: 'git pull && npm ci && pm2 restart all',
    post_deploy_cmd: '',
  });

  const loadTargets = async () => {
    const { data } = await deploymentApi.listTargets(projectId);
    setTargets(data.targets || []);
  };

  const loadRuns = async (targetId?: string) => {
    const { data } = await deploymentApi.listRuns(projectId, targetId, 10);
    setRuns(data.runs || []);
  };

  useEffect(() => {
    loadTargets();
    loadRuns();
  }, [projectId]);

  useEffect(() => {
    if (selectedTarget) {
      loadRuns(selectedTarget);
    } else {
      loadRuns();
    }
  }, [selectedTarget]);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await deploymentApi.createTarget(projectId, form);
      setShowAdd(false);
      resetForm();
      loadTargets();
    } finally { setLoading(false); }
  };

  const handleDelete = async (id: string) => {
    await deploymentApi.deleteTarget(projectId, id);
    if (selectedTarget === id) setSelectedTarget(null);
    loadTargets();
    loadRuns();
  };

  const handleDeploy = async (targetId: string) => {
    setDeploying(targetId);
    try {
      await deploymentApi.deploy(projectId, targetId);
      // Poll for completion
      setTimeout(() => {
        loadTargets();
        loadRuns(selectedTarget || undefined);
        setDeploying(null);
      }, 2000);
    } catch {
      setDeploying(null);
    }
  };

  const resetForm = () => {
    setForm({
      name: '', host: '', port: 22, username: 'root',
      auth_type: 'key', ssh_key: '', ssh_password: '',
      deploy_path: '/app', pre_deploy_cmd: '',
      deploy_cmd: 'git pull && npm ci && pm2 restart all',
      post_deploy_cmd: '',
    });
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
      case 'success': return 'text-[var(--color-success)]';
      case 'failed': return 'text-[var(--color-error)]';
      case 'deploying': case 'running': return 'text-[var(--color-primary)]';
      default: return 'text-[var(--color-text-secondary)]';
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-semibold">VPS Deployment</h2>
          <p className="text-xs text-[var(--color-text-secondary)] mt-0.5">
            Deploy to your server via SSH. Env vars are injected automatically.
          </p>
        </div>
        <Button size="sm" onClick={() => setShowAdd(!showAdd)}>
          <Plus size={14} className="mr-1" /> Add Server
        </Button>
      </div>

      {/* Add target form */}
      {showAdd && (
        <Card>
          <h3 className="font-medium text-sm mb-3">New VPS Target</h3>
          <form onSubmit={handleCreate} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <Input label="Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="Production" required />
              <Input label="Host / IP" value={form.host} onChange={(e) => setForm({ ...form, host: e.target.value })} placeholder="192.168.1.1" required />
            </div>
            <div className="grid grid-cols-3 gap-3">
              <Input label="Port" type="number" value={form.port} onChange={(e) => setForm({ ...form, port: Number(e.target.value) })} />
              <Input label="Username" value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} required />
              <div>
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Auth Type</label>
                <select
                  value={form.auth_type}
                  onChange={(e) => setForm({ ...form, auth_type: e.target.value as 'key' | 'password' })}
                  className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-sm text-[var(--color-text)]"
                >
                  <option value="key">SSH Key</option>
                  <option value="password">Password</option>
                </select>
              </div>
            </div>

            {form.auth_type === 'key' ? (
              <div>
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                  Private Key (PEM)
                </label>
                <textarea
                  value={form.ssh_key}
                  onChange={(e) => setForm({ ...form, ssh_key: e.target.value })}
                  placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
                  className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-xs font-mono text-[var(--color-text)] min-h-[80px] resize-y"
                />
              </div>
            ) : (
              <Input label="SSH Password" type="password" value={form.ssh_password} onChange={(e) => setForm({ ...form, ssh_password: e.target.value })} />
            )}

            <Input label="Deploy Path" value={form.deploy_path} onChange={(e) => setForm({ ...form, deploy_path: e.target.value })} placeholder="/var/www/app" required />

            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Pre-Deploy Command</label>
              <Input value={form.pre_deploy_cmd} onChange={(e) => setForm({ ...form, pre_deploy_cmd: e.target.value })} placeholder="Optional: commands before deploy" />
            </div>
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Deploy Command <span className="text-[var(--color-error)]">*</span></label>
              <Input value={form.deploy_cmd} onChange={(e) => setForm({ ...form, deploy_cmd: e.target.value })} placeholder="git pull && npm ci && pm2 restart all" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Post-Deploy Command</label>
              <Input value={form.post_deploy_cmd} onChange={(e) => setForm({ ...form, post_deploy_cmd: e.target.value })} placeholder="Optional: commands after deploy" />
            </div>

            <div className="flex gap-2">
              <Button type="submit" loading={loading} size="sm"><Server size={14} className="mr-1" /> Save Server</Button>
              <Button type="button" variant="ghost" size="sm" onClick={() => { setShowAdd(false); resetForm(); }}>Cancel</Button>
            </div>
          </form>
        </Card>
      )}

      {/* Target list */}
      <div className="space-y-2">
        {targets.length === 0 && !showAdd ? (
          <div className="text-center text-sm text-[var(--color-text-secondary)] py-8">
            No VPS targets configured. Add a server to start deploying.
          </div>
        ) : (
          targets.map((t) => (
            <Card key={t.id}>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  {statusIcon(t.status)}
                  <div>
                    <div className="font-medium text-sm flex items-center gap-2">
                      {t.name}
                      <span className={`text-xs font-normal ${statusColor(t.status)}`}>{t.status}</span>
                    </div>
                    <div className="text-xs text-[var(--color-text-secondary)]">
                      {t.username}@{t.host}:{t.port} · {t.deploy_path}
                    </div>
                    {t.last_deployed_at && (
                      <div className="text-xs text-[var(--color-text-secondary)]">
                        Last deployed: {new Date(t.last_deployed_at).toLocaleString()}
                      </div>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    loading={deploying === t.id}
                    onClick={() => { setSelectedTarget(t.id); handleDeploy(t.id); }}
                    disabled={t.status === 'deploying'}
                  >
                    <Play size={13} className="mr-1" />
                    {t.status === 'deploying' ? 'Deploying...' : 'Deploy'}
                  </Button>
                  <button
                    onClick={() => setSelectedTarget(selectedTarget === t.id ? null : t.id)}
                    className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-text-secondary)]"
                    title="View runs"
                  >
                    {selectedTarget === t.id ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                  </button>
                  <button
                    onClick={() => handleDelete(t.id)}
                    className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-error)]"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>

              {/* Deployment runs for this target */}
              {selectedTarget === t.id && (
                <div className="mt-3 pt-3 border-t border-[var(--color-border)]">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs font-medium text-[var(--color-text-secondary)]">Deployment History</span>
                    <button onClick={() => loadRuns(t.id)} className="p-1 hover:bg-[var(--color-bg-tertiary)] rounded">
                      <RefreshCw size={12} />
                    </button>
                  </div>
                  <div className="space-y-1">
                    {runs.filter((r) => r.vps_target_id === t.id).length === 0 ? (
                      <p className="text-xs text-[var(--color-text-secondary)]">No deployments yet.</p>
                    ) : (
                      runs.filter((r) => r.vps_target_id === t.id).map((run) => (
                        <div key={run.id} className="rounded-lg bg-[var(--color-bg)] p-2">
                          <div
                            className="flex items-center gap-2 cursor-pointer"
                            onClick={() => setExpandedRun(expandedRun === run.id ? null : run.id)}
                          >
                            {statusIcon(run.status, 13)}
                            <span className={`text-xs font-medium ${statusColor(run.status)}`}>{run.status}</span>
                            <span className="text-xs text-[var(--color-text-secondary)] ml-auto">
                              {new Date(run.started_at).toLocaleString()}
                            </span>
                            {run.finished_at && (
                              <span className="text-xs text-[var(--color-text-secondary)]">
                                · {Math.round((new Date(run.finished_at).getTime() - new Date(run.started_at).getTime()) / 1000)}s
                              </span>
                            )}
                            {expandedRun === run.id ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
                          </div>
                          {expandedRun === run.id && run.log_output && (
                            <pre className="mt-2 text-xs font-mono bg-[var(--color-bg-secondary)] p-2 rounded overflow-x-auto whitespace-pre-wrap text-[var(--color-text-secondary)] max-h-64 overflow-y-auto">
                              {run.log_output}
                            </pre>
                          )}
                        </div>
                      ))
                    )}
                  </div>
                </div>
              )}
            </Card>
          ))
        )}
      </div>

      {/* Global recent runs (when no target selected) */}
      {!selectedTarget && runs.length > 0 && (
        <div>
          <h3 className="font-medium text-sm mb-2 text-[var(--color-text-secondary)]">Recent Deployments</h3>
          <div className="space-y-1">
            {runs.slice(0, 5).map((run) => (
              <div key={run.id} className="rounded-lg bg-[var(--color-bg-secondary)] p-2 flex items-center gap-2">
                {statusIcon(run.status, 13)}
                <span className={`text-xs font-medium ${statusColor(run.status)}`}>{run.status}</span>
                <span className="text-xs text-[var(--color-text-secondary)]">
                  {run.target?.name || run.vps_target_id.slice(0, 8)}
                </span>
                <span className="text-xs text-[var(--color-text-secondary)] ml-auto">
                  {new Date(run.started_at).toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
