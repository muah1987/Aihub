import { useState, useEffect, FormEvent } from 'react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Card } from '../ui/Card';
import { Modal } from '../ui/Modal';
import { Plus, Users, Crown, Play, Trash2, UserPlus, Loader2, CheckCircle, XCircle, Clock } from 'lucide-react';
import { teamsApi, type AgentTeam, type TeamMember, type AgentTask } from '../../api/teams';
import client from '../../api/client';

interface TeamPanelProps {
  projectId: string;
}

interface ProjectAgent {
  id: string;
  name: string;
  role: string;
  model: string;
  status: string;
}

export function TeamPanel({ projectId }: TeamPanelProps) {
  const [teams, setTeams] = useState<AgentTeam[]>([]);
  const [selectedTeam, setSelectedTeam] = useState<AgentTeam | null>(null);
  const [members, setMembers] = useState<TeamMember[]>([]);
  const [tasks, setTasks] = useState<AgentTask[]>([]);
  const [agents, setAgents] = useState<ProjectAgent[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [showAddMember, setShowAddMember] = useState(false);
  const [showInvoke, setShowInvoke] = useState(false);
  const [loading, setLoading] = useState(false);
  const [invoking, setInvoking] = useState(false);
  const [teamName, setTeamName] = useState('');
  const [teamDesc, setTeamDesc] = useState('');
  const [teamStrategy, setTeamStrategy] = useState('sequential');
  const [selectedAgentId, setSelectedAgentId] = useState('');
  const [prompt, setPrompt] = useState('');

  const loadTeams = async () => {
    const { data } = await teamsApi.list(projectId);
    setTeams(data.teams || []);
  };

  const loadMembers = async (teamId: string) => {
    const { data } = await teamsApi.listMembers(projectId, teamId);
    setMembers(data.members || []);
  };

  const loadTasks = async (teamId: string) => {
    const { data } = await teamsApi.listTasks(projectId, teamId);
    setTasks(data.tasks || []);
  };

  const loadAgents = async () => {
    const { data } = await client.get<{ agents: ProjectAgent[] }>(`/projects/${projectId}/agents`);
    setAgents(data.agents || []);
  };

  useEffect(() => { loadTeams(); loadAgents(); }, [projectId]);

  useEffect(() => {
    if (selectedTeam) {
      loadMembers(selectedTeam.id);
      loadTasks(selectedTeam.id);
    }
  }, [selectedTeam]);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await teamsApi.create(projectId, { name: teamName, description: teamDesc, strategy: teamStrategy });
      setShowCreate(false);
      setTeamName(''); setTeamDesc(''); setTeamStrategy('sequential');
      loadTeams();
    } finally { setLoading(false); }
  };

  const handleAddMember = async () => {
    if (!selectedTeam || !selectedAgentId) return;
    await teamsApi.addMember(projectId, selectedTeam.id, selectedAgentId);
    setShowAddMember(false);
    setSelectedAgentId('');
    loadMembers(selectedTeam.id);
  };

  const handleRemoveMember = async (agentId: string) => {
    if (!selectedTeam) return;
    await teamsApi.removeMember(projectId, selectedTeam.id, agentId);
    loadMembers(selectedTeam.id);
  };

  const handleSetLeader = async (agentId: string) => {
    if (!selectedTeam) return;
    await teamsApi.setLeader(projectId, selectedTeam.id, agentId);
    setSelectedTeam({ ...selectedTeam, leader_agent_id: agentId });
  };

  const handleInvoke = async (e: FormEvent) => {
    e.preventDefault();
    if (!selectedTeam) return;
    setInvoking(true);
    try {
      await teamsApi.invoke(projectId, selectedTeam.id, prompt);
      setShowInvoke(false);
      setPrompt('');
      loadTasks(selectedTeam.id);
    } finally { setInvoking(false); }
  };

  const handleDeleteTeam = async (teamId: string) => {
    await teamsApi.delete(projectId, teamId);
    setSelectedTeam(null);
    loadTeams();
  };

  const taskStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return <CheckCircle size={14} className="text-[var(--color-success)]" />;
      case 'running': return <Loader2 size={14} className="animate-spin text-[var(--color-primary)]" />;
      case 'failed': return <XCircle size={14} className="text-[var(--color-error)]" />;
      default: return <Clock size={14} className="text-[var(--color-text-secondary)]" />;
    }
  };

  if (selectedTeam) {
    return (
      <div className="flex flex-col h-full">
        <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)]">
          <div>
            <button onClick={() => setSelectedTeam(null)} className="text-sm text-[var(--color-primary)] hover:underline mb-1">
              Back to Teams
            </button>
            <h2 className="font-semibold">{selectedTeam.name}</h2>
            <span className="text-xs text-[var(--color-text-secondary)]">{selectedTeam.strategy} strategy</span>
          </div>
          <div className="flex gap-2">
            <Button size="sm" onClick={() => setShowInvoke(true)}>
              <Play size={14} className="mr-1" /> Invoke
            </Button>
            <Button size="sm" onClick={() => setShowAddMember(true)}>
              <UserPlus size={14} className="mr-1" /> Add Agent
            </Button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          <div>
            <h3 className="font-medium text-sm mb-2">Team Members</h3>
            <div className="space-y-2">
              {members.length === 0 ? (
                <p className="text-sm text-[var(--color-text-secondary)]">No members yet. Add agents to the team.</p>
              ) : (
                members.map((m) => (
                  <Card key={m.id} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      {m.agent_id === selectedTeam.leader_agent_id && <Crown size={14} className="text-yellow-400" />}
                      <div>
                        <div className="text-sm font-medium">{m.agent?.name}</div>
                        <div className="text-xs text-[var(--color-text-secondary)]">{m.agent?.role} - {m.agent?.model}</div>
                      </div>
                    </div>
                    <div className="flex gap-1">
                      {m.agent_id !== selectedTeam.leader_agent_id && (
                        <Button size="sm" variant="ghost" onClick={() => handleSetLeader(m.agent_id)}>
                          <Crown size={12} /> Leader
                        </Button>
                      )}
                      <button onClick={() => handleRemoveMember(m.agent_id)} className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-error)]">
                        <Trash2 size={14} />
                      </button>
                    </div>
                  </Card>
                ))
              )}
            </div>
          </div>

          {tasks.length > 0 && (
            <div>
              <h3 className="font-medium text-sm mb-2">Recent Tasks</h3>
              <div className="space-y-2">
                {tasks.map((t) => (
                  <Card key={t.id}>
                    <div className="flex items-start gap-2">
                      {taskStatusIcon(t.status)}
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium">{t.title}</div>
                        <div className="text-xs text-[var(--color-text-secondary)]">
                          {t.assigned_agent?.name || 'Unassigned'} - {t.status}
                        </div>
                        {t.result && (
                          <pre className="text-xs mt-1 text-[var(--color-text-secondary)] whitespace-pre-wrap line-clamp-3">{t.result}</pre>
                        )}
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            </div>
          )}
        </div>

        <Modal isOpen={showAddMember} onClose={() => setShowAddMember(false)} title="Add Agent to Team">
          <div className="space-y-3">
            <select
              value={selectedAgentId}
              onChange={(e) => setSelectedAgentId(e.target.value)}
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-sm text-[var(--color-text)]"
            >
              <option value="">Select an agent...</option>
              {agents.map((a) => (
                <option key={a.id} value={a.id}>{a.name} ({a.role})</option>
              ))}
            </select>
            <Button onClick={handleAddMember} disabled={!selectedAgentId} className="w-full">Add to Team</Button>
          </div>
        </Modal>

        <Modal isOpen={showInvoke} onClose={() => setShowInvoke(false)} title="Invoke Team">
          <form onSubmit={handleInvoke} className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Prompt</label>
              <textarea
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                placeholder="Describe the task for the team..."
                className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-sm text-[var(--color-text)] min-h-[100px] resize-y"
                required
              />
            </div>
            <Button type="submit" loading={invoking} className="w-full">
              {invoking ? 'Team is working...' : 'Invoke Team'}
            </Button>
          </form>
        </Modal>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)]">
        <h2 className="font-semibold">Agent Teams</h2>
        <Button size="sm" onClick={() => setShowCreate(true)}>
          <Plus size={14} className="mr-1" /> Create Team
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {teams.length === 0 ? (
          <div className="text-center text-sm text-[var(--color-text-secondary)] py-12">
            No teams yet. Create a team and add agents to collaborate.
          </div>
        ) : (
          teams.map((t) => (
            <Card key={t.id} className="flex items-center justify-between group">
              <div className="flex items-center gap-3 cursor-pointer flex-1" onClick={() => setSelectedTeam(t)}>
                <Users size={20} className="text-[var(--color-primary)]" />
                <div>
                  <div className="font-medium text-sm">{t.name}</div>
                  <div className="text-xs text-[var(--color-text-secondary)]">{t.strategy} - {t.status}</div>
                </div>
              </div>
              <button
                onClick={() => handleDeleteTeam(t.id)}
                className="opacity-0 group-hover:opacity-100 p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-error)]"
              >
                <Trash2 size={14} />
              </button>
            </Card>
          ))
        )}
      </div>

      <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title="Create Agent Team">
        <form onSubmit={handleCreate} className="space-y-3">
          <Input label="Name" value={teamName} onChange={(e) => setTeamName(e.target.value)} placeholder="Backend Team" required />
          <Input label="Description" value={teamDesc} onChange={(e) => setTeamDesc(e.target.value)} placeholder="Handles backend development" />
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Execution Strategy</label>
            <select
              value={teamStrategy}
              onChange={(e) => setTeamStrategy(e.target.value)}
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-sm text-[var(--color-text)]"
            >
              <option value="sequential">Sequential (one by one)</option>
              <option value="parallel">Parallel (all at once)</option>
            </select>
          </div>
          <Button type="submit" loading={loading} className="w-full">Create Team</Button>
        </form>
      </Modal>
    </div>
  );
}
