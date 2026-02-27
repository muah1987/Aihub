import { useState, useEffect, FormEvent } from 'react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Modal } from '../ui/Modal';
import { useProjectStore } from '../../store/projectStore';
import { projectsApi, type GitHubRepo } from '../../api/projects';
import client from '../../api/client';

interface CreateProjectProps {
  open: boolean;
  onClose: () => void;
}

interface ProviderConnection {
  id: string;
  provider_type: string;
  provider_name: string;
  status: string;
}

export function CreateProject({ open, onClose }: CreateProjectProps) {
  const { createProject } = useProjectStore();
  const [providers, setProviders] = useState<ProviderConnection[]>([]);
  const [selectedProvider, setSelectedProvider] = useState('');
  const [repos, setRepos] = useState<GitHubRepo[]>([]);
  const [selectedRepo, setSelectedRepo] = useState<GitHubRepo | null>(null);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingRepos, setLoadingRepos] = useState(false);

  useEffect(() => {
    if (open) {
      client.get<{ providers: ProviderConnection[] }>('/providers').then(({ data }) => {
        const githubProviders = (data.providers || []).filter(
          (p) => p.provider_type === 'github' && p.status === 'active'
        );
        setProviders(githubProviders);
        if (githubProviders.length === 1) {
          setSelectedProvider(githubProviders[0].id);
        }
      });
    }
  }, [open]);

  useEffect(() => {
    if (selectedProvider) {
      setLoadingRepos(true);
      projectsApi.listRepos(selectedProvider).then(({ data }) => {
        setRepos(data.repos || []);
        setLoadingRepos(false);
      }).catch(() => setLoadingRepos(false));
    }
  }, [selectedProvider]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!selectedRepo || !selectedProvider) return;

    setLoading(true);
    try {
      await createProject({
        name: name || selectedRepo.name,
        description,
        repo_owner: selectedRepo.full_name.split('/')[0],
        repo_name: selectedRepo.name,
        provider_connection_id: selectedProvider,
      });
      onClose();
      setName('');
      setDescription('');
      setSelectedRepo(null);
    } catch {
      // Error shown via toast
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal open={open} onClose={onClose} title="Create Project">
      <form onSubmit={handleSubmit} className="space-y-4">
        {providers.length === 0 ? (
          <p className="text-sm text-[var(--color-text-secondary)]">
            No GitHub provider connected. Add one in Settings first.
          </p>
        ) : (
          <>
            {providers.length > 1 && (
              <div>
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                  GitHub Account
                </label>
                <select
                  value={selectedProvider}
                  onChange={(e) => setSelectedProvider(e.target.value)}
                  className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)]"
                >
                  <option value="">Select account...</option>
                  {providers.map((p) => (
                    <option key={p.id} value={p.id}>{p.provider_name}</option>
                  ))}
                </select>
              </div>
            )}

            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                Repository
              </label>
              {loadingRepos ? (
                <div className="flex items-center gap-2 py-2 text-sm text-[var(--color-text-secondary)]">
                  <div className="animate-spin h-4 w-4 border-2 border-[var(--color-primary)] border-t-transparent rounded-full" />
                  Loading repositories...
                </div>
              ) : (
                <div className="max-h-48 overflow-y-auto border border-[var(--color-border)] rounded-lg">
                  {repos.map((repo) => (
                    <button
                      key={repo.id}
                      type="button"
                      onClick={() => {
                        setSelectedRepo(repo);
                        if (!name) setName(repo.name);
                      }}
                      className={`w-full text-left px-3 py-2 text-sm hover:bg-[var(--color-bg-tertiary)] transition-colors ${
                        selectedRepo?.id === repo.id ? 'bg-[var(--color-primary)]/20 text-[var(--color-primary)]' : ''
                      }`}
                    >
                      <div className="font-medium">{repo.full_name}</div>
                      {repo.description && (
                        <div className="text-xs text-[var(--color-text-secondary)] truncate">{repo.description}</div>
                      )}
                    </button>
                  ))}
                </div>
              )}
            </div>

            <Input
              label="Project Name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Project"
            />

            <Input
              label="Description (optional)"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="What is this project about?"
            />

            <Button
              type="submit"
              loading={loading}
              disabled={!selectedRepo || !selectedProvider}
              className="w-full"
            >
              Create Project
            </Button>
          </>
        )}
      </form>
    </Modal>
  );
}
