import { useState, useEffect } from 'react';
import { Bot, Plus, Play, Trash2 } from 'lucide-react';
import { agentsApi, type Agent } from '../../api/agents';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Modal } from '../ui/Modal';
import { Card } from '../ui/Card';
import client from '../../api/client';

interface AgentPanelProps {
  projectId: string;
}

interface ProviderConnection {
  id: string;
  provider_type: string;
  provider_name: string;
}

export function AgentPanel({ projectId }: AgentPanelProps) {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [invoking, setInvoking] = useState<string | null>(null);
  const [prompt, setPrompt] = useState('');
  const [showInvoke, setShowInvoke] = useState<string | null>(null);

  // Create form state
  const [name, setName] = useState('');
  const [role, setRole] = useState('');
  const [model, setModel] = useState('gpt-4');
  const [systemPrompt, setSystemPrompt] = useState('');
  const [providers, setProviders] = useState<ProviderConnection[]>([]);
  const [selectedProvider, setSelectedProvider] = useState('');

  useEffect(() => {
    loadAgents();
    client.get<{ providers: ProviderConnection[] }>('/providers').then(({ data }) => {
      const aiProviders = (data.providers || []).filter(
        (p) => ['openai', 'anthropic'].includes(p.provider_type)
      );
      setProviders(aiProviders);
    });
  }, [projectId]);

  const loadAgents = async () => {
    const { data } = await agentsApi.list(projectId);
    setAgents(data.agents || []);
  };

  const handleCreate = async () => {
    if (!name || !role || !model || !selectedProvider) return;
    await agentsApi.create(projectId, {
      name,
      role,
      model,
      system_prompt: systemPrompt,
      provider_connection_id: selectedProvider,
    });
    setShowCreate(false);
    setName(''); setRole(''); setSystemPrompt('');
    loadAgents();
  };

  const handleInvoke = async (agentId: string) => {
    if (!prompt.trim()) return;
    setInvoking(agentId);
    try {
      await agentsApi.invoke(projectId, agentId, prompt);
      setPrompt('');
      setShowInvoke(null);
    } catch {
      // error
    } finally {
      setInvoking(null);
      loadAgents();
    }
  };

  const handleDelete = async (agentId: string) => {
    await agentsApi.delete(projectId, agentId);
    loadAgents();
  };

  const statusColors = {
    idle: 'bg-gray-500',
    running: 'bg-green-500 animate-pulse',
    error: 'bg-red-500',
  };

  return (
    <div className="p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-semibold">Agent Profiles</h3>
        <Button size="sm" onClick={() => setShowCreate(true)}>
          <Plus size={14} className="mr-1" /> Add Agent
        </Button>
      </div>

      {agents.length === 0 ? (
        <div className="text-center py-8 text-[var(--color-text-secondary)] text-sm">
          No agents configured. Create one to get started.
        </div>
      ) : (
        <div className="grid gap-3">
          {agents.map((agent) => (
            <Card key={agent.id}>
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-purple-500/20 flex items-center justify-center">
                    <Bot size={20} className="text-purple-400" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm">{agent.name}</span>
                      <span className={`w-2 h-2 rounded-full ${statusColors[agent.status]}`} />
                    </div>
                    <span className="text-xs text-[var(--color-text-secondary)]">
                      {agent.role} &middot; {agent.model}
                    </span>
                  </div>
                </div>
                <div className="flex gap-1">
                  <button
                    onClick={() => setShowInvoke(agent.id)}
                    className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded-lg text-[var(--color-success)]"
                    title="Invoke"
                  >
                    <Play size={16} />
                  </button>
                  <button
                    onClick={() => handleDelete(agent.id)}
                    className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded-lg text-[var(--color-error)]"
                    title="Delete"
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
              </div>
              <div className="mt-2 text-xs text-[var(--color-text-secondary)]">
                Tokens used: {agent.token_usage_total.toLocaleString()}
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Create Agent Modal */}
      <Modal open={showCreate} onClose={() => setShowCreate(false)} title="Create Agent">
        <div className="space-y-3">
          <Input label="Name" value={name} onChange={(e) => setName(e.target.value)} placeholder="Code Assistant" />
          <Input label="Role" value={role} onChange={(e) => setRole(e.target.value)} placeholder="developer" />
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Model</label>
            <select
              value={model}
              onChange={(e) => setModel(e.target.value)}
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)]"
            >
              <option value="gpt-4">GPT-4</option>
              <option value="gpt-4o">GPT-4o</option>
              <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
              <option value="claude-sonnet-4-20250514">Claude Sonnet</option>
              <option value="claude-opus-4-20250514">Claude Opus</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">AI Provider</label>
            <select
              value={selectedProvider}
              onChange={(e) => setSelectedProvider(e.target.value)}
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)]"
            >
              <option value="">Select provider...</option>
              {providers.map((p) => (
                <option key={p.id} value={p.id}>{p.provider_name} ({p.provider_type})</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">System Prompt</label>
            <textarea
              value={systemPrompt}
              onChange={(e) => setSystemPrompt(e.target.value)}
              placeholder="You are a helpful coding assistant..."
              rows={3}
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] placeholder-[var(--color-text-secondary)] text-sm"
            />
          </div>
          <Button onClick={handleCreate} className="w-full" disabled={!name || !role || !selectedProvider}>
            Create Agent
          </Button>
        </div>
      </Modal>

      {/* Invoke Modal */}
      <Modal open={!!showInvoke} onClose={() => setShowInvoke(null)} title="Invoke Agent">
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Prompt</label>
            <textarea
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="Enter your prompt..."
              rows={4}
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] placeholder-[var(--color-text-secondary)] text-sm"
            />
          </div>
          <Button
            onClick={() => showInvoke && handleInvoke(showInvoke)}
            loading={!!invoking}
            className="w-full"
            disabled={!prompt.trim()}
          >
            Send to Agent
          </Button>
          <p className="text-xs text-[var(--color-text-secondary)]">
            Response will appear in the project chat.
          </p>
        </div>
      </Modal>
    </div>
  );
}
