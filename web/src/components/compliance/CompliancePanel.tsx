import { useState, useEffect, type FormEvent } from 'react';
import {
  complianceApi,
  type AuditLog,
  type DataExport,
  type RetentionPolicy,
} from '../../api/compliance';
import {
  Shield, FileDown, Clock, Search, Plus, Trash2, X,
  Download, AlertTriangle, Info, AlertCircle, RefreshCw,
} from 'lucide-react';

interface CompliancePanelProps {
  projectId: string;
}

type SubTab = 'audit' | 'exports' | 'retention';

const severityIcons: Record<string, typeof Info> = {
  info: Info,
  warning: AlertTriangle,
  critical: AlertCircle,
};

const severityColors: Record<string, string> = {
  info: 'text-blue-400',
  warning: 'text-yellow-400',
  critical: 'text-red-400',
};

const statusColors: Record<string, string> = {
  pending: 'bg-yellow-500/20 text-yellow-400',
  processing: 'bg-blue-500/20 text-blue-400',
  ready: 'bg-green-500/20 text-green-400',
  expired: 'bg-gray-500/20 text-gray-400',
  error: 'bg-red-500/20 text-red-400',
};

export function CompliancePanel({ projectId }: CompliancePanelProps) {
  const [subTab, setSubTab] = useState<SubTab>('audit');

  const tabs: { id: SubTab; label: string; icon: typeof Shield }[] = [
    { id: 'audit', label: 'Audit Log', icon: Shield },
    { id: 'exports', label: 'Data Exports', icon: FileDown },
    { id: 'retention', label: 'Retention', icon: Clock },
  ];

  return (
    <div className="flex flex-col h-full">
      <div className="flex gap-2 px-4 py-2 border-b border-[var(--color-border)] shrink-0 overflow-x-auto">
        {tabs.map(t => {
          const Icon = t.icon;
          return (
            <button key={t.id} onClick={() => setSubTab(t.id)}
              className={`flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg whitespace-nowrap ${subTab === t.id ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg-secondary)] text-[var(--color-text-secondary)]'}`}>
              <Icon size={13} /> {t.label}
            </button>
          );
        })}
      </div>
      <div className="flex-1 overflow-hidden">
        {subTab === 'audit' && <AuditTab projectId={projectId} />}
        {subTab === 'exports' && <ExportsTab projectId={projectId} />}
        {subTab === 'retention' && <RetentionTab projectId={projectId} />}
      </div>
    </div>
  );
}

// ---- Audit Tab ----

function AuditTab({ projectId }: { projectId: string }) {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [actionFilter, setActionFilter] = useState('');
  const [severityFilter, setSeverityFilter] = useState('');
  const [page, setPage] = useState(0);
  const limit = 30;

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setLoading(true);
      try {
        const res = await complianceApi.listAuditLogs(projectId, {
          action: actionFilter || undefined,
          severity: severityFilter || undefined,
          limit,
          offset: page * limit,
        });
        if (!cancelled) {
          setLogs(res.data.audit_logs || []);
          setTotal(res.data.total || 0);
        }
      } catch (err) { console.error(err); }
      if (!cancelled) setLoading(false);
    };
    load();
    return () => { cancelled = true; };
  }, [projectId, actionFilter, severityFilter, page]);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <div className="flex-1 flex gap-2">
          <input value={actionFilter} onChange={e => { setActionFilter(e.target.value); setPage(0); }}
            placeholder="Filter by action..."
            className="flex-1 px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
          <select value={severityFilter} onChange={e => { setSeverityFilter(e.target.value); setPage(0); }}
            className="w-32 px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm">
            <option value="">All</option>
            <option value="info">Info</option>
            <option value="warning">Warning</option>
            <option value="critical">Critical</option>
          </select>
        </div>
        <span className="text-xs text-[var(--color-text-secondary)]">{total} entries</span>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : logs.length === 0 ? (
          <div className="text-center py-12">
            <Shield size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)]">No audit logs found</p>
          </div>
        ) : (
          <div className="space-y-1">
            {logs.map(log => {
              const SevIcon = severityIcons[log.severity] || Info;
              return (
                <div key={log.id} className="flex items-start gap-3 p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                  <SevIcon size={16} className={`mt-0.5 shrink-0 ${severityColors[log.severity] || ''}`} />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-0.5">
                      <span className="text-sm font-medium">{log.action}</span>
                      <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)]">{log.resource_type}</span>
                    </div>
                    <div className="text-xs text-[var(--color-text-secondary)]">
                      {new Date(log.created_at).toLocaleString()}
                      {log.ip_address && ` · ${log.ip_address}`}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {total > limit && (
        <div className="flex items-center justify-center gap-2 px-4 py-2 border-t border-[var(--color-border)] shrink-0">
          <button onClick={() => setPage(p => Math.max(0, p - 1))} disabled={page === 0}
            className="px-3 py-1 text-xs bg-[var(--color-bg-secondary)] rounded disabled:opacity-50">Prev</button>
          <span className="text-xs text-[var(--color-text-secondary)]">Page {page + 1} of {Math.ceil(total / limit)}</span>
          <button onClick={() => setPage(p => p + 1)} disabled={(page + 1) * limit >= total}
            className="px-3 py-1 text-xs bg-[var(--color-bg-secondary)] rounded disabled:opacity-50">Next</button>
        </div>
      )}
    </div>
  );
}

// ---- Exports Tab ----

function ExportsTab({ projectId }: { projectId: string }) {
  const [exports, setExports] = useState<DataExport[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);

  const [includeChat, setIncludeChat] = useState(true);
  const [includeMemories, setIncludeMemories] = useState(true);
  const [includeAgents, setIncludeAgents] = useState(true);
  const [includeWorkflows, setIncludeWorkflows] = useState(true);
  const [includeSettings, setIncludeSettings] = useState(true);

  const loadRef = { current: () => {} };

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setLoading(true);
      try {
        const res = await complianceApi.listExports(projectId);
        if (!cancelled) setExports(res.data.exports || []);
      } catch (err) { console.error(err); }
      if (!cancelled) setLoading(false);
    };
    loadRef.current = load;
    load();
    return () => { cancelled = true; };
  }, [projectId]);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    setCreating(true);
    try {
      await complianceApi.createExport(projectId, {
        include_chat: includeChat,
        include_memories: includeMemories,
        include_agents: includeAgents,
        include_workflows: includeWorkflows,
        include_settings: includeSettings,
      });
      setShowCreate(false);
      loadRef.current();
    } catch (err) { console.error(err); alert('An error occurred. Please try again.'); }
    setCreating(false);
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Delete this export?')) return;
    try { await complianceApi.deleteExport(projectId, id); loadRef.current(); } catch (err) { console.error(err); alert('An error occurred. Please try again.'); }
  };

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <div>
          <h3 className="font-semibold text-sm">Data Exports</h3>
          <p className="text-xs text-[var(--color-text-secondary)]">Export project data for backup or migration</p>
        </div>
        <button onClick={() => setShowCreate(true)} className="flex items-center gap-1 px-3 py-2 bg-[var(--color-primary)] text-white rounded-lg text-xs">
          <Plus size={14} /> New Export
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : exports.length === 0 ? (
          <div className="text-center py-12">
            <FileDown size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)] mb-3">No exports created yet</p>
            <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm">
              Create Your First Export
            </button>
          </div>
        ) : (
          <div className="space-y-2">
            {exports.map(exp => (
              <div key={exp.id} className="flex items-center gap-3 p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                <FileDown size={18} className="text-[var(--color-text-secondary)] shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-0.5">
                    <span className="text-sm font-medium capitalize">{exp.export_type} Export</span>
                    <span className={`text-xs px-1.5 py-0.5 rounded ${statusColors[exp.status] || ''}`}>{exp.status}</span>
                  </div>
                  <div className="text-xs text-[var(--color-text-secondary)]">
                    {formatSize(exp.file_size)} · {exp.format.toUpperCase()} · {new Date(exp.created_at).toLocaleString()}
                    {exp.expires_at && ` · Expires ${new Date(exp.expires_at).toLocaleDateString()}`}
                  </div>
                </div>
                {exp.status === 'ready' && (
                  <button onClick={async () => {
                    try {
                      const res = await complianceApi.downloadExport(projectId, exp.id);
                      const url = URL.createObjectURL(res.data);
                      const a = document.createElement('a');
                      a.href = url;
                      a.download = `export-${exp.id}.json`;
                      a.click();
                      URL.revokeObjectURL(url);
                    } catch (err) { console.error(err); alert('Failed to download export.'); }
                  }}
                    className="flex items-center gap-1 px-2 py-1 text-xs bg-[var(--color-primary)] text-white rounded">
                    <Download size={12} /> Download
                  </button>
                )}
                <button onClick={() => handleDelete(exp.id)} className="text-red-400 p-1"><Trash2 size={14} /></button>
              </div>
            ))}
          </div>
        )}
      </div>

      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-[var(--color-bg)] rounded-xl border border-[var(--color-border)] p-6 w-full max-w-md space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold">Create Data Export</h3>
              <button onClick={() => setShowCreate(false)}><X size={18} /></button>
            </div>
            <form onSubmit={handleCreate} className="space-y-3">
              <div className="space-y-2">
                <label className="block text-xs font-medium text-[var(--color-text-secondary)]">Include</label>
                {[
                  { label: 'Chat messages', val: includeChat, set: setIncludeChat },
                  { label: 'Memories', val: includeMemories, set: setIncludeMemories },
                  { label: 'Agents', val: includeAgents, set: setIncludeAgents },
                  { label: 'Workflows', val: includeWorkflows, set: setIncludeWorkflows },
                  { label: 'Settings & tools', val: includeSettings, set: setIncludeSettings },
                ].map(item => (
                  <label key={item.label} className="flex items-center gap-2 text-sm">
                    <input type="checkbox" checked={item.val} onChange={e => item.set(e.target.checked)} className="rounded" />
                    {item.label}
                  </label>
                ))}
              </div>
              <button type="submit" disabled={creating} className="w-full py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm disabled:opacity-50">
                {creating ? <RefreshCw size={14} className="animate-spin mx-auto" /> : 'Start Export'}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

// ---- Retention Tab ----

function RetentionTab({ projectId }: { projectId: string }) {
  const [policies, setPolicies] = useState<RetentionPolicy[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);

  const [resourceType, setResourceType] = useState('messages');
  const [retentionDays, setRetentionDays] = useState(90);

  const loadRef = { current: () => {} };

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setLoading(true);
      try {
        const res = await complianceApi.listRetentionPolicies(projectId);
        if (!cancelled) setPolicies(res.data.policies || []);
      } catch (err) { console.error(err); }
      if (!cancelled) setLoading(false);
    };
    loadRef.current = load;
    load();
    return () => { cancelled = true; };
  }, [projectId]);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    try {
      await complianceApi.upsertRetentionPolicy(projectId, {
        resource_type: resourceType,
        retention_days: retentionDays,
        enabled: true,
      });
      setShowCreate(false);
      loadRef.current();
    } catch (err) { console.error(err); alert('An error occurred. Please try again.'); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Delete this retention policy?')) return;
    try { await complianceApi.deleteRetentionPolicy(projectId, id); loadRef.current(); } catch (err) { console.error(err); alert('An error occurred. Please try again.'); }
  };

  const resourceTypes = [
    { value: 'messages', label: 'Chat Messages' },
    { value: 'audit_logs', label: 'Audit Logs' },
    { value: 'notifications', label: 'Notifications' },
    { value: 'outbound_events', label: 'Outbound Events' },
    { value: 'server_metrics', label: 'Server Metrics' },
    { value: 'activity_logs', label: 'Activity Logs' },
  ];

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <div>
          <h3 className="font-semibold text-sm">Data Retention</h3>
          <p className="text-xs text-[var(--color-text-secondary)]">Automatic cleanup of old data based on retention rules</p>
        </div>
        <button onClick={() => setShowCreate(true)} className="flex items-center gap-1 px-3 py-2 bg-[var(--color-primary)] text-white rounded-lg text-xs">
          <Plus size={14} /> Add Policy
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : policies.length === 0 ? (
          <div className="text-center py-12">
            <Clock size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)] mb-3">No retention policies configured</p>
            <p className="text-xs text-[var(--color-text-secondary)]">Data is kept indefinitely by default</p>
          </div>
        ) : (
          <div className="space-y-2">
            {policies.map(p => (
              <div key={p.id} className="flex items-center gap-3 p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                <Clock size={18} className="text-[var(--color-text-secondary)] shrink-0" />
                <div className="flex-1">
                  <div className="text-sm font-medium capitalize">{p.resource_type.replace(/_/g, ' ')}</div>
                  <div className="text-xs text-[var(--color-text-secondary)]">
                    Keep for {p.retention_days} days · {p.records_deleted} records cleaned
                    {p.last_run_at && ` · Last run ${new Date(p.last_run_at).toLocaleDateString()}`}
                  </div>
                </div>
                <span className={`text-xs px-1.5 py-0.5 rounded ${p.enabled ? 'bg-green-500/20 text-green-400' : 'bg-gray-500/20 text-gray-400'}`}>
                  {p.enabled ? 'Active' : 'Paused'}
                </span>
                <button onClick={() => handleDelete(p.id)} className="text-red-400 p-1"><Trash2 size={14} /></button>
              </div>
            ))}
          </div>
        )}
      </div>

      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-[var(--color-bg)] rounded-xl border border-[var(--color-border)] p-6 w-full max-w-md space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold">Add Retention Policy</h3>
              <button onClick={() => setShowCreate(false)}><X size={18} /></button>
            </div>
            <form onSubmit={handleCreate} className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">Data Type</label>
                <select value={resourceType} onChange={e => setResourceType(e.target.value)}
                  className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm">
                  {resourceTypes.map(rt => (
                    <option key={rt.value} value={rt.value}>{rt.label}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">Retention Period (days)</label>
                <input type="number" value={retentionDays} onChange={e => setRetentionDays(Number(e.target.value))} min={1}
                  className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
                <div className="flex gap-2 mt-2">
                  {[30, 60, 90, 180, 365].map(d => (
                    <button key={d} type="button" onClick={() => setRetentionDays(d)}
                      className={`px-2 py-1 text-xs rounded ${retentionDays === d ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg-tertiary)]'}`}>
                      {d}d
                    </button>
                  ))}
                </div>
              </div>
              <button type="submit" className="w-full py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm">Save Policy</button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
