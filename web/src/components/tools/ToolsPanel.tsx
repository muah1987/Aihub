import { useState, useEffect } from 'react';
import { Wrench, Plus, Trash2, Link2, Unlink, Play, Clock, AlertCircle, CheckCircle } from 'lucide-react';
import { toolsApi, type AgentTool, type ToolBinding, type ToolExecution } from '../../api/tools';
import { agentsApi, type Agent } from '../../api/agents';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Modal } from '../ui/Modal';
import { Card } from '../ui/Card';

interface ToolsPanelProps {
  projectId: string;
}

export function ToolsPanel({ projectId }: ToolsPanelProps) {
  const [tab, setTab] = useState<'registry' | 'bindings' | 'executions'>('registry');
  const [tools, setTools] = useState<AgentTool[]>([]);
  const [agents, setAgents] = useState<Agent[]>([]);
  const [executions, setExecutions] = useState<ToolExecution[]>([]);
  const [showCreate, setShowCreate] = useState(false);

  // Create form
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [toolType, setToolType] = useState('function');
  const [definition, setDefinition] = useState('{}');

  // Binding state
  const [selectedAgent, setSelectedAgent] = useState('');
  const [bindings, setBindings] = useState<ToolBinding[]>([]);

  useEffect(() => {
    loadTools();
    loadAgents();
  }, [projectId]);

  useEffect(() => {
    if (tab === 'executions') loadExecutions();
  }, [tab, projectId]);

  useEffect(() => {
    if (selectedAgent) loadBindings(selectedAgent);
  }, [selectedAgent]);

  const loadTools = async () => {
    try {
      const { data } = await toolsApi.list(projectId);
      setTools(data.tools || []);
    } catch { /* ignore */ }
  };

  const loadAgents = async () => {
    try {
      const { data } = await agentsApi.list(projectId);
      setAgents(data.agents || []);
    } catch { /* ignore */ }
  };

  const loadBindings = async (agentId: string) => {
    try {
      const { data } = await toolsApi.listBindings(projectId, agentId);
      setBindings(data.bindings || []);
    } catch { /* ignore */ }
  };

  const loadExecutions = async () => {
    try {
      const { data } = await toolsApi.listExecutions(projectId, 50);
      setExecutions(data.executions || []);
    } catch { /* ignore */ }
  };

  const handleCreate = async () => {
    if (!name || !toolType) return;
    let def = {};
    try { def = JSON.parse(definition); } catch { return; }
    await toolsApi.create(projectId, { name, description, tool_type: toolType, definition: def });
    setShowCreate(false);
    setName(''); setDescription(''); setDefinition('{}');
    loadTools();
  };

  const handleDelete = async (toolId: string) => {
    await toolsApi.delete(projectId, toolId);
    loadTools();
  };

  const handleToggle = async (tool: AgentTool) => {
    await toolsApi.update(projectId, tool.id, { enabled: !tool.enabled } as Partial<AgentTool>);
    loadTools();
  };

  const handleBind = async (toolId: string) => {
    if (!selectedAgent) return;
    await toolsApi.bind(projectId, selectedAgent, toolId);
    loadBindings(selectedAgent);
  };

  const handleUnbind = async (toolId: string) => {
    if (!selectedAgent) return;
    await toolsApi.unbind(projectId, selectedAgent, toolId);
    loadBindings(selectedAgent);
  };

  const tabs = [
    { id: 'registry' as const, label: 'Tool Registry' },
    { id: 'bindings' as const, label: 'Agent Bindings' },
    { id: 'executions' as const, label: 'Execution Log' },
  ];

  const typeColors = {
    function: 'bg-blue-500/20 text-blue-400',
    http: 'bg-green-500/20 text-green-400',
    shell: 'bg-orange-500/20 text-orange-400',
  };

  const boundToolIds = new Set(bindings.map((b) => b.tool_id));

  return (
    <div className="p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-semibold flex items-center gap-2">
          <Wrench size={18} /> Tools
        </h3>
        <Button size="sm" onClick={() => setShowCreate(true)}>
          <Plus size={14} className="mr-1" /> New Tool
        </Button>
      </div>

      {/* Sub-tabs */}
      <div className="flex gap-1 bg-[var(--color-bg)] rounded-lg p-1">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`flex-1 text-xs py-1.5 px-2 rounded-md transition-colors ${
              tab === t.id ? 'bg-[var(--color-bg-tertiary)] text-[var(--color-text)]' : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Registry Tab */}
      {tab === 'registry' && (
        <div className="space-y-3">
          {tools.length === 0 ? (
            <div className="text-center py-8 text-[var(--color-text-secondary)] text-sm">
              No tools defined. Create tools that your agents can use.
            </div>
          ) : (
            tools.map((tool) => (
              <Card key={tool.id}>
                <div className="flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm">{tool.name}</span>
                      <span className={`text-[10px] px-1.5 py-0.5 rounded-full ${typeColors[tool.tool_type] || 'bg-gray-500/20 text-gray-400'}`}>
                        {tool.tool_type}
                      </span>
                      {!tool.enabled && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-red-500/20 text-red-400">disabled</span>
                      )}
                    </div>
                    <p className="text-xs text-[var(--color-text-secondary)] mt-1">{tool.description || 'No description'}</p>
                  </div>
                  <div className="flex gap-1">
                    <button
                      onClick={() => handleToggle(tool)}
                      className={`p-1.5 rounded-lg text-xs ${tool.enabled ? 'text-green-400 hover:bg-green-500/10' : 'text-gray-400 hover:bg-gray-500/10'}`}
                      title={tool.enabled ? 'Disable' : 'Enable'}
                    >
                      {tool.enabled ? <CheckCircle size={16} /> : <AlertCircle size={16} />}
                    </button>
                    <button
                      onClick={() => handleDelete(tool.id)}
                      className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded-lg text-[var(--color-error)]"
                      title="Delete"
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>
      )}

      {/* Bindings Tab */}
      {tab === 'bindings' && (
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Select Agent</label>
            <select
              value={selectedAgent}
              onChange={(e) => setSelectedAgent(e.target.value)}
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] text-sm"
            >
              <option value="">Choose an agent...</option>
              {agents.map((a) => (
                <option key={a.id} value={a.id}>{a.name} ({a.role})</option>
              ))}
            </select>
          </div>

          {selectedAgent && (
            <div className="space-y-2">
              <p className="text-xs text-[var(--color-text-secondary)]">
                {bindings.length} tool{bindings.length !== 1 ? 's' : ''} bound
              </p>
              {tools.map((tool) => {
                const bound = boundToolIds.has(tool.id);
                return (
                  <div key={tool.id} className="flex items-center justify-between p-2 rounded-lg border border-[var(--color-border)]">
                    <div className="flex items-center gap-2">
                      <span className={`text-[10px] px-1.5 py-0.5 rounded-full ${typeColors[tool.tool_type] || ''}`}>
                        {tool.tool_type}
                      </span>
                      <span className="text-sm">{tool.name}</span>
                    </div>
                    <button
                      onClick={() => bound ? handleUnbind(tool.id) : handleBind(tool.id)}
                      className={`p-1.5 rounded-lg ${bound ? 'text-red-400 hover:bg-red-500/10' : 'text-green-400 hover:bg-green-500/10'}`}
                      title={bound ? 'Unbind' : 'Bind'}
                    >
                      {bound ? <Unlink size={16} /> : <Link2 size={16} />}
                    </button>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}

      {/* Execution Log Tab */}
      {tab === 'executions' && (
        <div className="space-y-2">
          {executions.length === 0 ? (
            <div className="text-center py-8 text-[var(--color-text-secondary)] text-sm">
              No tool executions yet.
            </div>
          ) : (
            executions.map((exec) => (
              <Card key={exec.id}>
                <div className="flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <Play size={12} className={exec.status === 'success' ? 'text-green-400' : 'text-red-400'} />
                      <span className="text-xs font-mono">{exec.tool_id.slice(0, 8)}</span>
                      <span className={`text-[10px] px-1.5 py-0.5 rounded-full ${exec.status === 'success' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                        {exec.status}
                      </span>
                    </div>
                    {exec.error_message && (
                      <p className="text-xs text-red-400 mt-1">{exec.error_message}</p>
                    )}
                    {exec.output && (
                      <pre className="text-xs text-[var(--color-text-secondary)] mt-1 max-h-24 overflow-auto whitespace-pre-wrap">
                        {exec.output.slice(0, 500)}
                      </pre>
                    )}
                  </div>
                  <div className="text-right text-[10px] text-[var(--color-text-secondary)]">
                    <div className="flex items-center gap-1">
                      <Clock size={10} /> {exec.duration_ms}ms
                    </div>
                    <div>{new Date(exec.created_at).toLocaleString()}</div>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>
      )}

      {/* Create Tool Modal */}
      <Modal open={showCreate} onClose={() => setShowCreate(false)} title="Create Tool">
        <div className="space-y-3">
          <Input label="Name" value={name} onChange={(e) => setName(e.target.value)} placeholder="fetch_weather" />
          <Input label="Description" value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Fetches weather for a city" />
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Type</label>
            <select
              value={toolType}
              onChange={(e) => setToolType(e.target.value)}
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)]"
            >
              <option value="function">Function (AI native)</option>
              <option value="http">HTTP Request</option>
              <option value="shell">Shell Command</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Definition (JSON)</label>
            <textarea
              value={definition}
              onChange={(e) => setDefinition(e.target.value)}
              rows={6}
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] font-mono text-xs"
              placeholder='{"url":"https://api.example.com","method":"GET"}'
            />
          </div>
          <Button onClick={handleCreate} className="w-full" disabled={!name || !toolType}>
            Create Tool
          </Button>
        </div>
      </Modal>
    </div>
  );
}
