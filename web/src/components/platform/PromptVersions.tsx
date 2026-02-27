import { useState, useEffect } from 'react';
import { Plus, Check, Trash2, Eye } from 'lucide-react';
import { platformApi, type PromptVersion } from '../../api/platform';

interface PromptVersionsProps {
  projectId: string;
  agentId: string;
}

export function PromptVersions({ projectId, agentId }: PromptVersionsProps) {
  const [versions, setVersions] = useState<PromptVersion[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [label, setLabel] = useState('');
  const [prompt, setPrompt] = useState('');
  const [viewPrompt, setViewPrompt] = useState<string | null>(null);

  const fetchVersions = async () => {
    try {
      const res = await platformApi.listPromptVersions(projectId, agentId);
      setVersions(res.data.versions || []);
    } catch {
      // silent
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchVersions();
  }, [projectId, agentId]);

  const handleCreate = async () => {
    if (!label.trim() || !prompt.trim()) return;
    try {
      await platformApi.createPromptVersion(projectId, agentId, {
        label: label.trim(),
        system_prompt: prompt.trim(),
      });
      setLabel('');
      setPrompt('');
      setShowCreate(false);
      fetchVersions();
    } catch {
      // silent
    }
  };

  const handleActivate = async (versionId: string) => {
    try {
      await platformApi.activatePromptVersion(projectId, agentId, versionId);
      fetchVersions();
    } catch {
      // silent
    }
  };

  const handleDelete = async (versionId: string) => {
    try {
      await platformApi.deletePromptVersion(projectId, agentId, versionId);
      setVersions((prev) => prev.filter((v) => v.id !== versionId));
    } catch {
      // silent
    }
  };

  if (loading) {
    return <div className="text-sm text-[var(--color-text-secondary)] p-4">Loading prompt versions...</div>;
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium">Prompt Versions</h4>
        <button
          onClick={() => setShowCreate(!showCreate)}
          className="flex items-center gap-1 text-xs text-[var(--color-primary)] hover:underline"
        >
          <Plus size={14} /> New Version
        </button>
      </div>

      {showCreate && (
        <div className="space-y-2 p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
          <input
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            placeholder="Version label (e.g. v2.1-concise)"
            className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded px-3 py-1.5 text-sm"
          />
          <textarea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="System prompt..."
            rows={4}
            className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded px-3 py-1.5 text-sm resize-none"
          />
          <div className="flex gap-2">
            <button
              onClick={handleCreate}
              className="px-3 py-1 bg-[var(--color-primary)] text-white rounded text-xs"
            >
              Create
            </button>
            <button
              onClick={() => setShowCreate(false)}
              className="px-3 py-1 bg-[var(--color-bg-tertiary)] rounded text-xs"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {versions.length === 0 && !showCreate && (
        <p className="text-xs text-[var(--color-text-secondary)]">No prompt versions yet. Create one for A/B testing.</p>
      )}

      <div className="space-y-2">
        {versions.map((v) => (
          <div
            key={v.id}
            className={`flex items-center gap-3 p-3 rounded-lg border transition-colors ${
              v.is_active
                ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/5'
                : 'border-[var(--color-border)] bg-[var(--color-bg-secondary)]'
            }`}
          >
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">{v.version_label}</span>
                {v.is_active && (
                  <span className="text-[10px] bg-[var(--color-primary)] text-white px-1.5 py-0.5 rounded-full">
                    Active
                  </span>
                )}
              </div>
              <div className="text-[11px] text-[var(--color-text-secondary)] mt-0.5">
                {v.total_invocations} invocations · Rating: {v.avg_rating.toFixed(1)} · {new Date(v.created_at).toLocaleDateString()}
              </div>
            </div>

            <div className="flex items-center gap-1">
              <button
                onClick={() => setViewPrompt(viewPrompt === v.id ? null : v.id)}
                className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded"
                title="View prompt"
              >
                <Eye size={14} />
              </button>
              {!v.is_active && (
                <button
                  onClick={() => handleActivate(v.id)}
                  className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-green-500"
                  title="Activate"
                >
                  <Check size={14} />
                </button>
              )}
              <button
                onClick={() => handleDelete(v.id)}
                className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded text-red-400"
                title="Delete"
              >
                <Trash2 size={14} />
              </button>
            </div>
          </div>
        ))}

        {viewPrompt && versions.find((v) => v.id === viewPrompt) && (
          <div className="p-3 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg">
            <div className="text-xs text-[var(--color-text-secondary)] mb-1">System Prompt</div>
            <pre className="text-xs whitespace-pre-wrap font-mono max-h-48 overflow-y-auto">
              {versions.find((v) => v.id === viewPrompt)!.system_prompt}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}
