import { useState, useEffect } from 'react';
import { schedulesApi, type ScheduledJob, type ScheduledJobRun } from '../../api/schedules';
import { Plus, Trash2, ToggleLeft, ToggleRight, Clock, X, ChevronRight } from 'lucide-react';

interface SchedulerPanelProps {
  projectId: string;
}

const actionTypes = [
  { value: 'invoke_agent', label: 'Invoke Agent' },
  { value: 'invoke_team', label: 'Invoke Team' },
  { value: 'trigger_workflow', label: 'Trigger Workflow' },
  { value: 'trigger_deploy', label: 'Trigger Deploy' },
  { value: 'run_tool', label: 'Run Tool' },
  { value: 'http_request', label: 'HTTP Request' },
  { value: 'shell_command', label: 'Shell Command' },
];

export function SchedulerPanel({ projectId }: SchedulerPanelProps) {
  const [jobs, setJobs] = useState<ScheduledJob[]>([]);
  const [selectedJob, setSelectedJob] = useState<ScheduledJob | null>(null);
  const [runs, setRuns] = useState<ScheduledJobRun[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [loading, setLoading] = useState(true);

  const [newName, setNewName] = useState('');
  const [newDesc, setNewDesc] = useState('');
  const [newCron, setNewCron] = useState('0 * * * *');
  const [newAction, setNewAction] = useState('shell_command');
  const [newConfig, setNewConfig] = useState('{}');

  useEffect(() => {
    loadJobs();
  }, [projectId]);

  const loadJobs = async () => {
    setLoading(true);
    try {
      const res = await schedulesApi.list(projectId);
      setJobs(res.data.jobs || []);
    } catch { /* empty */ }
    setLoading(false);
  };

  const loadRuns = async (jobId: string) => {
    try {
      const res = await schedulesApi.listRuns(projectId, jobId);
      setRuns(res.data.runs || []);
    } catch { /* empty */ }
  };

  const handleCreate = async () => {
    if (!newName.trim() || !newCron.trim()) return;
    let parsedConfig = {};
    try { parsedConfig = JSON.parse(newConfig); } catch { /* empty */ }
    try {
      await schedulesApi.create(projectId, {
        name: newName, description: newDesc, cron_expression: newCron,
        action_type: newAction, action_config: parsedConfig,
      });
      setShowCreate(false);
      setNewName(''); setNewDesc(''); setNewCron('0 * * * *'); setNewConfig('{}');
      loadJobs();
    } catch { /* empty */ }
  };

  const handleDelete = async (id: string) => {
    try {
      await schedulesApi.delete(projectId, id);
      if (selectedJob?.id === id) setSelectedJob(null);
      loadJobs();
    } catch { /* empty */ }
  };

  const handleToggle = async (job: ScheduledJob) => {
    try {
      await schedulesApi.setEnabled(projectId, job.id, !job.enabled);
      loadJobs();
    } catch { /* empty */ }
  };

  const selectJob = (job: ScheduledJob) => {
    setSelectedJob(job);
    loadRuns(job.id);
  };

  const statusColor = (s: string) => {
    switch (s) {
      case 'completed': return 'text-green-400';
      case 'failed': return 'text-red-400';
      case 'running': return 'text-yellow-400';
      default: return 'text-[var(--color-text-secondary)]';
    }
  };

  if (selectedJob) {
    return (
      <div className="flex flex-col h-full">
        <div className="flex items-center gap-2 px-4 py-3 border-b border-[var(--color-border)] shrink-0">
          <button onClick={() => setSelectedJob(null)} className="text-sm text-[var(--color-primary)]">&larr; Back</button>
          <h3 className="font-semibold text-sm flex-1">{selectedJob.name}</h3>
        </div>
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
              <div className="text-xs text-[var(--color-text-secondary)] mb-1">Cron</div>
              <div className="text-sm font-mono">{selectedJob.cron_expression}</div>
            </div>
            <div className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
              <div className="text-xs text-[var(--color-text-secondary)] mb-1">Action</div>
              <div className="text-sm">{selectedJob.action_type}</div>
            </div>
            <div className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
              <div className="text-xs text-[var(--color-text-secondary)] mb-1">Next Run</div>
              <div className="text-sm">{selectedJob.next_run_at ? new Date(selectedJob.next_run_at).toLocaleString() : '—'}</div>
            </div>
            <div className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
              <div className="text-xs text-[var(--color-text-secondary)] mb-1">Last Run</div>
              <div className="text-sm">
                {selectedJob.last_run_at ? new Date(selectedJob.last_run_at).toLocaleString() : '—'}
                {selectedJob.last_run_status && (
                  <span className={`ml-2 ${statusColor(selectedJob.last_run_status)}`}>{selectedJob.last_run_status}</span>
                )}
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-sm font-semibold mb-2">Run History</h4>
            {runs.length === 0 ? (
              <p className="text-xs text-[var(--color-text-secondary)]">No runs recorded yet.</p>
            ) : (
              <div className="space-y-2">
                {runs.map(run => (
                  <div key={run.id} className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                    <div className="flex items-center justify-between">
                      <span className={`text-xs font-semibold uppercase ${statusColor(run.status)}`}>{run.status}</span>
                      <span className="text-xs text-[var(--color-text-secondary)]">{new Date(run.started_at).toLocaleString()}</span>
                    </div>
                    {run.output && <pre className="text-xs mt-1 text-[var(--color-text-secondary)] whitespace-pre-wrap">{run.output.substring(0, 200)}</pre>}
                    {run.error_message && <p className="text-xs text-red-400 mt-1">{run.error_message}</p>}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <h3 className="font-semibold text-sm">Scheduled Jobs</h3>
        <button onClick={() => setShowCreate(true)} className="flex items-center gap-1 px-3 py-1.5 bg-[var(--color-primary)] text-white rounded-lg text-xs hover:opacity-90">
          <Plus size={14} /> New Job
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : jobs.length === 0 ? (
          <div className="text-center py-12">
            <Clock size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)] mb-3">No scheduled jobs yet</p>
            <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm">
              Create Your First Schedule
            </button>
          </div>
        ) : (
          <div className="space-y-2">
            {jobs.map(job => (
              <div key={job.id} className="flex items-center gap-3 p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] hover:border-[var(--color-primary)] transition-colors">
                <button onClick={() => handleToggle(job)} className="shrink-0">
                  {job.enabled ? <ToggleRight size={20} className="text-green-400" /> : <ToggleLeft size={20} className="text-gray-500" />}
                </button>
                <button onClick={() => selectJob(job)} className="flex-1 text-left">
                  <div className="text-sm font-medium">{job.name}</div>
                  <div className="text-xs text-[var(--color-text-secondary)]">
                    <span className="font-mono">{job.cron_expression}</span> &middot; {job.action_type}
                    {job.next_run_at && <> &middot; Next: {new Date(job.next_run_at).toLocaleString()}</>}
                  </div>
                </button>
                <button onClick={() => handleDelete(job.id)} className="text-red-400 hover:text-red-300 p-1">
                  <Trash2 size={16} />
                </button>
                <button onClick={() => selectJob(job)} className="text-[var(--color-text-secondary)]">
                  <ChevronRight size={16} />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-[var(--color-bg)] rounded-xl border border-[var(--color-border)] p-6 w-full max-w-md space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold">Create Scheduled Job</h3>
              <button onClick={() => setShowCreate(false)}><X size={18} /></button>
            </div>
            <input value={newName} onChange={e => setNewName(e.target.value)} placeholder="Job name" className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
            <textarea value={newDesc} onChange={e => setNewDesc(e.target.value)} placeholder="Description (optional)" rows={2} className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
            <div>
              <label className="text-xs text-[var(--color-text-secondary)] mb-1 block">Cron Expression (min hour dom mon dow)</label>
              <input value={newCron} onChange={e => setNewCron(e.target.value)} placeholder="0 * * * *" className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm font-mono" />
            </div>
            <select value={newAction} onChange={e => setNewAction(e.target.value)} className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm">
              {actionTypes.map(a => <option key={a.value} value={a.value}>{a.label}</option>)}
            </select>
            <textarea value={newConfig} onChange={e => setNewConfig(e.target.value)} placeholder='Config JSON, e.g. {"command":"backup.sh"}' rows={3} className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm font-mono" />
            <button onClick={handleCreate} className="w-full py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm hover:opacity-90">Create</button>
          </div>
        </div>
      )}
    </div>
  );
}
