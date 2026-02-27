import { useState, useEffect } from 'react';
import { workflowsApi, type Workflow, type WorkflowRun } from '../../api/workflows';
import { Plus, Play, Trash2, ChevronRight, ToggleLeft, ToggleRight, X } from 'lucide-react';

interface WorkflowPanelProps {
  projectId: string;
}

const triggerLabels: Record<string, string> = {
  manual: 'Manual',
  webhook: 'Webhook',
  schedule: 'Schedule',
  deploy_success: 'Deploy Success',
  deploy_failure: 'Deploy Failure',
};

const stepTypes = [
  'invoke_agent', 'invoke_team', 'run_tool', 'http_request',
  'shell_command', 'deploy', 'notify', 'condition',
];

export function WorkflowPanel({ projectId }: WorkflowPanelProps) {
  const [subTab, setSubTab] = useState<'list' | 'runs'>('list');
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [selectedWorkflow, setSelectedWorkflow] = useState<Workflow | null>(null);
  const [runs, setRuns] = useState<WorkflowRun[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [showAddStep, setShowAddStep] = useState(false);
  const [loading, setLoading] = useState(true);

  // Create form
  const [newName, setNewName] = useState('');
  const [newDesc, setNewDesc] = useState('');
  const [newTrigger, setNewTrigger] = useState('manual');

  // Step form
  const [stepName, setStepName] = useState('');
  const [stepType, setStepType] = useState('shell_command');
  const [stepOnFailure, setStepOnFailure] = useState('abort');
  const [stepTimeout, setStepTimeout] = useState(60);
  const [stepConfig, setStepConfig] = useState('{}');

  useEffect(() => {
    loadWorkflows();
  }, [projectId]);

  const loadWorkflows = async () => {
    setLoading(true);
    try {
      const res = await workflowsApi.list(projectId);
      setWorkflows(res.data.workflows || []);
    } catch { /* empty */ }
    setLoading(false);
  };

  const loadRuns = async (workflowId: string) => {
    try {
      const res = await workflowsApi.listRuns(projectId, workflowId);
      setRuns(res.data.runs || []);
    } catch { /* empty */ }
  };

  const handleCreate = async () => {
    if (!newName.trim()) return;
    try {
      await workflowsApi.create(projectId, {
        name: newName, description: newDesc, trigger_type: newTrigger,
      });
      setShowCreate(false);
      setNewName(''); setNewDesc(''); setNewTrigger('manual');
      loadWorkflows();
    } catch { /* empty */ }
  };

  const handleDelete = async (id: string) => {
    try {
      await workflowsApi.delete(projectId, id);
      if (selectedWorkflow?.id === id) setSelectedWorkflow(null);
      loadWorkflows();
    } catch { /* empty */ }
  };

  const handleToggle = async (w: Workflow) => {
    try {
      await workflowsApi.setEnabled(projectId, w.id, !w.enabled);
      loadWorkflows();
    } catch { /* empty */ }
  };

  const handleTrigger = async (workflowId: string) => {
    try {
      await workflowsApi.trigger(projectId, workflowId);
      loadRuns(workflowId);
    } catch { /* empty */ }
  };

  const selectWorkflow = async (w: Workflow) => {
    try {
      const res = await workflowsApi.get(projectId, w.id);
      setSelectedWorkflow(res.data.workflow);
      loadRuns(w.id);
    } catch { /* empty */ }
  };

  const handleAddStep = async () => {
    if (!selectedWorkflow || !stepName.trim()) return;
    let parsedConfig = {};
    try { parsedConfig = JSON.parse(stepConfig); } catch { /* empty */ }
    try {
      await workflowsApi.addStep(projectId, selectedWorkflow.id, {
        name: stepName, step_type: stepType, step_order: (selectedWorkflow.steps?.length || 0) + 1,
        config: parsedConfig, on_failure: stepOnFailure, timeout_seconds: stepTimeout,
      });
      setShowAddStep(false);
      setStepName(''); setStepConfig('{}');
      selectWorkflow(selectedWorkflow);
    } catch { /* empty */ }
  };

  const handleDeleteStep = async (stepId: string) => {
    if (!selectedWorkflow) return;
    try {
      await workflowsApi.deleteStep(projectId, selectedWorkflow.id, stepId);
      selectWorkflow(selectedWorkflow);
    } catch { /* empty */ }
  };

  const statusColor = (s: string) => {
    switch (s) {
      case 'completed': return 'text-green-400';
      case 'failed': return 'text-red-400';
      case 'running': return 'text-yellow-400';
      case 'skipped': return 'text-gray-400';
      default: return 'text-[var(--color-text-secondary)]';
    }
  };

  if (selectedWorkflow) {
    return (
      <div className="flex flex-col h-full">
        <div className="flex items-center gap-2 px-4 py-3 border-b border-[var(--color-border)] shrink-0">
          <button onClick={() => setSelectedWorkflow(null)} className="text-sm text-[var(--color-primary)]">&larr; Back</button>
          <h3 className="font-semibold text-sm flex-1">{selectedWorkflow.name}</h3>
          <button onClick={() => handleTrigger(selectedWorkflow.id)} className="flex items-center gap-1 px-3 py-1.5 bg-green-600 text-white rounded-lg text-xs hover:bg-green-700">
            <Play size={14} /> Run
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          <div className="text-xs text-[var(--color-text-secondary)]">
            Trigger: {triggerLabels[selectedWorkflow.trigger_type] || selectedWorkflow.trigger_type} &middot; {selectedWorkflow.enabled ? 'Enabled' : 'Disabled'}
          </div>
          {selectedWorkflow.description && (
            <p className="text-sm text-[var(--color-text-secondary)]">{selectedWorkflow.description}</p>
          )}

          {/* Steps */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <h4 className="text-sm font-semibold">Steps</h4>
              <button onClick={() => setShowAddStep(true)} className="flex items-center gap-1 text-xs text-[var(--color-primary)]">
                <Plus size={14} /> Add Step
              </button>
            </div>
            {(selectedWorkflow.steps?.length || 0) === 0 ? (
              <p className="text-xs text-[var(--color-text-secondary)]">No steps defined yet.</p>
            ) : (
              <div className="space-y-2">
                {selectedWorkflow.steps.map((step, i) => (
                  <div key={step.id} className="flex items-center gap-2 p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                    <span className="text-xs font-mono text-[var(--color-text-secondary)] w-6">{i + 1}.</span>
                    <div className="flex-1">
                      <div className="text-sm font-medium">{step.name}</div>
                      <div className="text-xs text-[var(--color-text-secondary)]">
                        {step.step_type} &middot; on_failure: {step.on_failure} &middot; timeout: {step.timeout_seconds}s
                      </div>
                    </div>
                    <button onClick={() => handleDeleteStep(step.id)} className="text-red-400 hover:text-red-300">
                      <Trash2 size={14} />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Runs */}
          <div>
            <h4 className="text-sm font-semibold mb-2">Recent Runs</h4>
            {runs.length === 0 ? (
              <p className="text-xs text-[var(--color-text-secondary)]">No runs yet.</p>
            ) : (
              <div className="space-y-2">
                {runs.map((run) => (
                  <div key={run.id} className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                    <div className="flex items-center justify-between mb-1">
                      <span className={`text-xs font-semibold uppercase ${statusColor(run.status)}`}>{run.status}</span>
                      <span className="text-xs text-[var(--color-text-secondary)]">{new Date(run.started_at).toLocaleString()}</span>
                    </div>
                    {run.step_runs && run.step_runs.length > 0 && (
                      <div className="flex gap-1 mt-1">
                        {run.step_runs.map((sr) => (
                          <span key={sr.id} className={`text-xs px-1.5 py-0.5 rounded ${statusColor(sr.status)} bg-[var(--color-bg-tertiary)]`}>
                            {sr.step_name}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Add step modal */}
        {showAddStep && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
            <div className="bg-[var(--color-bg)] rounded-xl border border-[var(--color-border)] p-6 w-full max-w-md space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold">Add Step</h3>
                <button onClick={() => setShowAddStep(false)}><X size={18} /></button>
              </div>
              <input value={stepName} onChange={e => setStepName(e.target.value)} placeholder="Step name" className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
              <select value={stepType} onChange={e => setStepType(e.target.value)} className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm">
                {stepTypes.map(t => <option key={t} value={t}>{t}</option>)}
              </select>
              <div className="flex gap-3">
                <select value={stepOnFailure} onChange={e => setStepOnFailure(e.target.value)} className="flex-1 px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm">
                  <option value="abort">Abort on failure</option>
                  <option value="continue">Continue on failure</option>
                </select>
                <input type="number" value={stepTimeout} onChange={e => setStepTimeout(Number(e.target.value))} placeholder="Timeout (s)" className="w-24 px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
              </div>
              <textarea value={stepConfig} onChange={e => setStepConfig(e.target.value)} placeholder='Config JSON, e.g. {"command":"npm test"}' rows={3} className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm font-mono" />
              <button onClick={handleAddStep} className="w-full py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm hover:opacity-90">Add Step</button>
            </div>
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <div className="flex gap-2">
          {(['list', 'runs'] as const).map(t => (
            <button key={t} onClick={() => setSubTab(t)} className={`px-3 py-1.5 text-xs rounded-lg capitalize ${subTab === t ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg-secondary)] text-[var(--color-text-secondary)]'}`}>
              {t === 'list' ? 'Workflows' : 'All Runs'}
            </button>
          ))}
        </div>
        <button onClick={() => setShowCreate(true)} className="flex items-center gap-1 px-3 py-1.5 bg-[var(--color-primary)] text-white rounded-lg text-xs hover:opacity-90">
          <Plus size={14} /> New Workflow
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : workflows.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-sm text-[var(--color-text-secondary)] mb-3">No workflows yet</p>
            <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm">
              Create Your First Workflow
            </button>
          </div>
        ) : (
          <div className="space-y-2">
            {workflows.map(w => (
              <div key={w.id} className="flex items-center gap-3 p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] hover:border-[var(--color-primary)] transition-colors">
                <button onClick={() => handleToggle(w)} className="shrink-0">
                  {w.enabled ? <ToggleRight size={20} className="text-green-400" /> : <ToggleLeft size={20} className="text-gray-500" />}
                </button>
                <button onClick={() => selectWorkflow(w)} className="flex-1 text-left">
                  <div className="text-sm font-medium">{w.name}</div>
                  <div className="text-xs text-[var(--color-text-secondary)]">
                    {triggerLabels[w.trigger_type] || w.trigger_type} &middot; {w.steps?.length || 0} steps
                  </div>
                </button>
                <button onClick={() => handleTrigger(w.id)} className="text-green-400 hover:text-green-300 p-1" title="Run now">
                  <Play size={16} />
                </button>
                <button onClick={() => handleDelete(w.id)} className="text-red-400 hover:text-red-300 p-1">
                  <Trash2 size={16} />
                </button>
                <button onClick={() => selectWorkflow(w)} className="text-[var(--color-text-secondary)]">
                  <ChevronRight size={16} />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-[var(--color-bg)] rounded-xl border border-[var(--color-border)] p-6 w-full max-w-md space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold">Create Workflow</h3>
              <button onClick={() => setShowCreate(false)}><X size={18} /></button>
            </div>
            <input value={newName} onChange={e => setNewName(e.target.value)} placeholder="Workflow name" className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
            <textarea value={newDesc} onChange={e => setNewDesc(e.target.value)} placeholder="Description (optional)" rows={2} className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
            <select value={newTrigger} onChange={e => setNewTrigger(e.target.value)} className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm">
              {Object.entries(triggerLabels).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
            </select>
            <button onClick={handleCreate} className="w-full py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm hover:opacity-90">Create</button>
          </div>
        </div>
      )}
    </div>
  );
}
