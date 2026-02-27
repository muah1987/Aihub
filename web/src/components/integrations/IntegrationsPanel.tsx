import { useState, useEffect, type FormEvent } from 'react';
import {
  integrationApi,
  type IntegrationConnection,
  type NotificationRule,
  type EmailDigest,
  type OutboundEvent,
} from '../../api/integration';
import {
  Link2, Plus, Trash2, X, Check, XCircle, ToggleLeft, ToggleRight,
  MessageSquare, Bell, Mail, Webhook, Zap, Send, Filter, RefreshCw,
} from 'lucide-react';

interface IntegrationsPanelProps {
  projectId: string;
}

type SubTab = 'connections' | 'rules' | 'digests' | 'events';

const platformIcons: Record<string, typeof MessageSquare> = {
  slack: MessageSquare,
  discord: MessageSquare,
  webhook: Webhook,
  zapier: Zap,
};

const platformColors: Record<string, string> = {
  slack: 'bg-purple-500/20 text-purple-400',
  discord: 'bg-indigo-500/20 text-indigo-400',
  webhook: 'bg-green-500/20 text-green-400',
  zapier: 'bg-orange-500/20 text-orange-400',
};

const statusColors: Record<string, string> = {
  connected: 'bg-green-500/20 text-green-400',
  pending: 'bg-yellow-500/20 text-yellow-400',
  error: 'bg-red-500/20 text-red-400',
  sent: 'bg-green-500/20 text-green-400',
  failed: 'bg-red-500/20 text-red-400',
  retrying: 'bg-yellow-500/20 text-yellow-400',
};

export function IntegrationsPanel({ projectId }: IntegrationsPanelProps) {
  const [subTab, setSubTab] = useState<SubTab>('connections');

  const tabs: { id: SubTab; label: string; icon: typeof Link2 }[] = [
    { id: 'connections', label: 'Connections', icon: Link2 },
    { id: 'rules', label: 'Rules', icon: Filter },
    { id: 'digests', label: 'Digests', icon: Mail },
    { id: 'events', label: 'Events', icon: Send },
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
        {subTab === 'connections' && <ConnectionsTab projectId={projectId} />}
        {subTab === 'rules' && <RulesTab projectId={projectId} />}
        {subTab === 'digests' && <DigestsTab />}
        {subTab === 'events' && <EventsTab projectId={projectId} />}
      </div>
    </div>
  );
}

// ---- Connections Tab ----

function ConnectionsTab({ projectId }: { projectId: string }) {
  const [conns, setConns] = useState<IntegrationConnection[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [loading, setLoading] = useState(true);
  const [testingId, setTestingId] = useState<string | null>(null);

  // Create form
  const [platform, setPlatform] = useState('slack');
  const [name, setName] = useState('');
  const [webhookUrl, setWebhookUrl] = useState('');
  const [channelId, setChannelId] = useState('');

  useEffect(() => { load(); }, [projectId]);

  const load = async () => {
    setLoading(true);
    try {
      const res = await integrationApi.listConnections(projectId);
      setConns(res.data.connections || []);
    } catch { /* empty */ }
    setLoading(false);
  };

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    const configKey = platform === 'webhook' || platform === 'zapier' ? 'url' : 'webhook_url';
    try {
      await integrationApi.createConnection(projectId, {
        platform, name, config: { [configKey]: webhookUrl }, channel_id: channelId,
      });
      setShowCreate(false);
      setName(''); setWebhookUrl(''); setChannelId('');
      load();
    } catch { /* empty */ }
  };

  const handleDelete = async (id: string) => {
    try { await integrationApi.deleteConnection(projectId, id); load(); } catch { /* empty */ }
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    try {
      const res = await integrationApi.testConnection(projectId, id);
      alert(res.data.result === 'ok' ? 'Connection test successful!' : `Test failed: ${res.data.error}`);
    } catch { alert('Test failed'); }
    setTestingId(null);
  };

  const toggleEnabled = async (conn: IntegrationConnection) => {
    try {
      await integrationApi.updateConnection(projectId, conn.id, { enabled: !conn.enabled });
      load();
    } catch { /* empty */ }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <div>
          <h3 className="font-semibold text-sm">Integration Connections</h3>
          <p className="text-xs text-[var(--color-text-secondary)]">Connect Slack, Discord, webhooks and more</p>
        </div>
        <button onClick={() => setShowCreate(true)} className="flex items-center gap-1 px-3 py-2 bg-[var(--color-primary)] text-white rounded-lg text-xs">
          <Plus size={14} /> Add
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : conns.length === 0 ? (
          <div className="text-center py-12">
            <Link2 size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)] mb-3">No integrations connected</p>
            <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm">
              Connect Your First Integration
            </button>
          </div>
        ) : (
          <div className="space-y-2">
            {conns.map(conn => {
              const PlatformIcon = platformIcons[conn.platform] || Webhook;
              return (
                <div key={conn.id} className="flex items-center gap-3 p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                  <div className={`p-2 rounded-lg ${platformColors[conn.platform] || 'bg-gray-500/20 text-gray-400'}`}>
                    <PlatformIcon size={18} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium truncate">{conn.name}</span>
                      <span className={`text-xs px-1.5 py-0.5 rounded ${statusColors[conn.status] || ''}`}>{conn.status}</span>
                    </div>
                    <div className="text-xs text-[var(--color-text-secondary)]">
                      {conn.platform} {conn.channel_id && `#${conn.channel_id}`}
                      {conn.last_used_at && ` · last used ${new Date(conn.last_used_at).toLocaleDateString()}`}
                    </div>
                  </div>
                  <button onClick={() => toggleEnabled(conn)} className="text-[var(--color-text-secondary)]" title={conn.enabled ? 'Disable' : 'Enable'}>
                    {conn.enabled ? <ToggleRight size={20} className="text-green-400" /> : <ToggleLeft size={20} />}
                  </button>
                  <button onClick={() => handleTest(conn.id)} disabled={testingId === conn.id}
                    className="text-xs px-2 py-1 bg-[var(--color-bg-tertiary)] rounded text-[var(--color-text-secondary)] hover:text-[var(--color-text)]">
                    {testingId === conn.id ? <RefreshCw size={12} className="animate-spin" /> : 'Test'}
                  </button>
                  <button onClick={() => handleDelete(conn.id)} className="text-red-400 hover:text-red-300 p-1">
                    <Trash2 size={16} />
                  </button>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-[var(--color-bg)] rounded-xl border border-[var(--color-border)] p-6 w-full max-w-md space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold">Add Integration</h3>
              <button onClick={() => setShowCreate(false)}><X size={18} /></button>
            </div>
            <form onSubmit={handleCreate} className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">Platform</label>
                <div className="grid grid-cols-4 gap-2">
                  {['slack', 'discord', 'webhook', 'zapier'].map(p => (
                    <button key={p} type="button" onClick={() => setPlatform(p)}
                      className={`p-2 rounded-lg text-xs text-center capitalize border ${platform === p ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/10' : 'border-[var(--color-border)]'}`}>
                      {p}
                    </button>
                  ))}
                </div>
              </div>
              <input value={name} onChange={e => setName(e.target.value)} placeholder="Connection name" required
                className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
              <input value={webhookUrl} onChange={e => setWebhookUrl(e.target.value)}
                placeholder={platform === 'slack' ? 'Slack webhook URL' : platform === 'discord' ? 'Discord webhook URL' : 'Webhook URL'}
                className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
              <input value={channelId} onChange={e => setChannelId(e.target.value)} placeholder="Channel ID (optional)"
                className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
              <button type="submit" className="w-full py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm">Connect</button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

// ---- Rules Tab ----

function RulesTab({ projectId }: { projectId: string }) {
  const [rules, setRules] = useState<NotificationRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);

  const [eventType, setEventType] = useState('*');
  const [channel, setChannel] = useState('in_app');
  const [severity, setSeverity] = useState('info');

  useEffect(() => { load(); }, [projectId]);

  const load = async () => {
    setLoading(true);
    try {
      const res = await integrationApi.listRules(projectId);
      setRules(res.data.rules || []);
    } catch { /* empty */ }
    setLoading(false);
  };

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    try {
      await integrationApi.createRule({ project_id: projectId, event_type: eventType, channel, min_severity: severity });
      setShowCreate(false);
      load();
    } catch { /* empty */ }
  };

  const toggleRule = async (rule: NotificationRule) => {
    try {
      await integrationApi.updateRule(rule.id, { enabled: !rule.enabled });
      load();
    } catch { /* empty */ }
  };

  const handleDelete = async (id: string) => {
    try { await integrationApi.deleteRule(id); load(); } catch { /* empty */ }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <div>
          <h3 className="font-semibold text-sm">Notification Rules</h3>
          <p className="text-xs text-[var(--color-text-secondary)]">Control which events trigger notifications on each channel</p>
        </div>
        <button onClick={() => setShowCreate(true)} className="flex items-center gap-1 px-3 py-2 bg-[var(--color-primary)] text-white rounded-lg text-xs">
          <Plus size={14} /> Add Rule
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : rules.length === 0 ? (
          <div className="text-center py-12">
            <Bell size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)]">No notification rules configured</p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">All in-app notifications are enabled by default</p>
          </div>
        ) : (
          <div className="space-y-2">
            {rules.map(rule => (
              <div key={rule.id} className="flex items-center gap-3 p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-0.5">
                    <span className="text-sm font-medium">{rule.event_type === '*' ? 'All Events' : rule.event_type}</span>
                    <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)]">{rule.channel}</span>
                    <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)]">{rule.min_severity}+</span>
                  </div>
                  {rule.quiet_hours_start !== undefined && rule.quiet_hours_end !== undefined && (
                    <span className="text-xs text-[var(--color-text-secondary)]">Quiet: {rule.quiet_hours_start}:00 - {rule.quiet_hours_end}:00</span>
                  )}
                </div>
                <button onClick={() => toggleRule(rule)}>
                  {rule.enabled ? <ToggleRight size={20} className="text-green-400" /> : <ToggleLeft size={20} className="text-[var(--color-text-secondary)]" />}
                </button>
                <button onClick={() => handleDelete(rule.id)} className="text-red-400 p-1"><Trash2 size={14} /></button>
              </div>
            ))}
          </div>
        )}
      </div>

      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-[var(--color-bg)] rounded-xl border border-[var(--color-border)] p-6 w-full max-w-md space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold">Add Notification Rule</h3>
              <button onClick={() => setShowCreate(false)}><X size={18} /></button>
            </div>
            <form onSubmit={handleCreate} className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">Event Type</label>
                <select value={eventType} onChange={e => setEventType(e.target.value)}
                  className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm">
                  <option value="*">All Events</option>
                  <option value="deployment">Deployments</option>
                  <option value="chat">Chat Messages</option>
                  <option value="agent">Agent Activity</option>
                  <option value="webhook">Webhooks</option>
                  <option value="monitoring">Monitoring Alerts</option>
                  <option value="workflow">Workflows</option>
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">Channel</label>
                <select value={channel} onChange={e => setChannel(e.target.value)}
                  className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm">
                  <option value="in_app">In-App</option>
                  <option value="email">Email</option>
                  <option value="slack">Slack</option>
                  <option value="discord">Discord</option>
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">Minimum Severity</label>
                <select value={severity} onChange={e => setSeverity(e.target.value)}
                  className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm">
                  <option value="info">Info</option>
                  <option value="warning">Warning</option>
                  <option value="critical">Critical</option>
                </select>
              </div>
              <button type="submit" className="w-full py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm">Save Rule</button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

// ---- Digests Tab ----

function DigestsTab() {
  const [digests, setDigests] = useState<EmailDigest[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);

  const [frequency, setFrequency] = useState('daily');
  const [includeDeployments, setIncludeDeployments] = useState(true);
  const [includeChat, setIncludeChat] = useState(true);
  const [includeAgent, setIncludeAgent] = useState(true);
  const [includeMonitoring, setIncludeMonitoring] = useState(true);

  useEffect(() => { load(); }, []);

  const load = async () => {
    setLoading(true);
    try {
      const res = await integrationApi.listDigests();
      setDigests(res.data.digests || []);
    } catch { /* empty */ }
    setLoading(false);
  };

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    try {
      await integrationApi.upsertDigest({
        input: {
          frequency,
          include_deployments: includeDeployments,
          include_chat_summary: includeChat,
          include_agent_activity: includeAgent,
          include_monitoring: includeMonitoring,
        },
      });
      setShowCreate(false);
      load();
    } catch { /* empty */ }
  };

  const handleDelete = async (id: string) => {
    try { await integrationApi.deleteDigest(id); load(); } catch { /* empty */ }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <div>
          <h3 className="font-semibold text-sm">Email Digests</h3>
          <p className="text-xs text-[var(--color-text-secondary)]">Receive daily or weekly summaries of project activity</p>
        </div>
        <button onClick={() => setShowCreate(true)} className="flex items-center gap-1 px-3 py-2 bg-[var(--color-primary)] text-white rounded-lg text-xs">
          <Plus size={14} /> Configure
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : digests.length === 0 ? (
          <div className="text-center py-12">
            <Mail size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)] mb-3">No email digests configured</p>
            <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm">
              Set Up Email Digest
            </button>
          </div>
        ) : (
          <div className="space-y-2">
            {digests.map(d => (
              <div key={d.id} className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <Mail size={16} className="text-[var(--color-primary)]" />
                    <span className="text-sm font-medium capitalize">{d.frequency} Digest</span>
                    {d.enabled ? (
                      <span className="text-xs px-1.5 py-0.5 rounded bg-green-500/20 text-green-400">Active</span>
                    ) : (
                      <span className="text-xs px-1.5 py-0.5 rounded bg-gray-500/20 text-gray-400">Paused</span>
                    )}
                  </div>
                  <button onClick={() => handleDelete(d.id)} className="text-red-400 p-1"><Trash2 size={14} /></button>
                </div>
                <div className="flex gap-2 flex-wrap">
                  {d.include_deployments && <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-bg-tertiary)]">Deployments</span>}
                  {d.include_chat_summary && <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-bg-tertiary)]">Chat</span>}
                  {d.include_agent_activity && <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-bg-tertiary)]">Agents</span>}
                  {d.include_monitoring && <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-bg-tertiary)]">Monitoring</span>}
                </div>
                {d.next_send_at && (
                  <p className="text-xs text-[var(--color-text-secondary)] mt-1">Next: {new Date(d.next_send_at).toLocaleString()}</p>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-[var(--color-bg)] rounded-xl border border-[var(--color-border)] p-6 w-full max-w-md space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold">Configure Email Digest</h3>
              <button onClick={() => setShowCreate(false)}><X size={18} /></button>
            </div>
            <form onSubmit={handleCreate} className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">Frequency</label>
                <select value={frequency} onChange={e => setFrequency(e.target.value)}
                  className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm">
                  <option value="daily">Daily (8:00 AM)</option>
                  <option value="weekly">Weekly (Monday 8:00 AM)</option>
                  <option value="none">Disabled</option>
                </select>
              </div>
              <div className="space-y-2">
                <label className="block text-xs font-medium text-[var(--color-text-secondary)]">Include</label>
                {[
                  { label: 'Deployments', val: includeDeployments, set: setIncludeDeployments },
                  { label: 'Chat summary', val: includeChat, set: setIncludeChat },
                  { label: 'Agent activity', val: includeAgent, set: setIncludeAgent },
                  { label: 'Monitoring alerts', val: includeMonitoring, set: setIncludeMonitoring },
                ].map(item => (
                  <label key={item.label} className="flex items-center gap-2 text-sm">
                    <input type="checkbox" checked={item.val} onChange={e => item.set(e.target.checked)} className="rounded" />
                    {item.label}
                  </label>
                ))}
              </div>
              <button type="submit" className="w-full py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm">Save Digest</button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

// ---- Events Tab ----

function EventsTab({ projectId }: { projectId: string }) {
  const [events, setEvents] = useState<OutboundEvent[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => { load(); }, [projectId]);

  const load = async () => {
    setLoading(true);
    try {
      const res = await integrationApi.listEvents(projectId);
      setEvents(res.data.events || []);
    } catch { /* empty */ }
    setLoading(false);
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <div>
          <h3 className="font-semibold text-sm">Outbound Events</h3>
          <p className="text-xs text-[var(--color-text-secondary)]">Recent webhook dispatches to external integrations</p>
        </div>
        <button onClick={load} className="text-xs text-[var(--color-primary)] flex items-center gap-1">
          <RefreshCw size={12} /> Refresh
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : events.length === 0 ? (
          <div className="text-center py-12">
            <Send size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)]">No outbound events yet</p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">Events appear here when notifications are dispatched to connected integrations</p>
          </div>
        ) : (
          <div className="space-y-2">
            {events.map(ev => (
              <div key={ev.id} className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                <div className="flex items-center justify-between mb-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{ev.event_type}</span>
                    <span className={`text-xs px-1.5 py-0.5 rounded ${statusColors[ev.status] || ''}`}>{ev.status}</span>
                  </div>
                  <span className="text-xs text-[var(--color-text-secondary)]">
                    {new Date(ev.created_at).toLocaleString()}
                  </span>
                </div>
                <div className="text-xs text-[var(--color-text-secondary)]">
                  Attempts: {ev.attempts}/{ev.max_attempts}
                  {ev.http_status ? ` · HTTP ${ev.http_status}` : ''}
                  {ev.sent_at && ` · Sent ${new Date(ev.sent_at).toLocaleString()}`}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
