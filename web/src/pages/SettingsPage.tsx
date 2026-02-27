import { useState, useEffect, FormEvent } from 'react';
import { AppShell } from '../components/layout/AppShell';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Card } from '../components/ui/Card';
import { Plus, Trash2, CheckCircle, XCircle } from 'lucide-react';
import client from '../api/client';

interface ProviderConnection {
  id: string;
  provider_type: string;
  provider_name: string;
  status: string;
  created_at: string;
}

export function SettingsPage() {
  const [providers, setProviders] = useState<ProviderConnection[]>([]);
  const [showAdd, setShowAdd] = useState(false);
  const [providerType, setProviderType] = useState('github');
  const [providerName, setProviderName] = useState('');
  const [accessToken, setAccessToken] = useState('');
  const [loading, setLoading] = useState(false);

  const loadProviders = async () => {
    const { data } = await client.get<{ providers: ProviderConnection[] }>('/providers');
    setProviders(data.providers || []);
  };

  useEffect(() => { loadProviders(); }, []);

  const handleAdd = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await client.post('/providers', {
        provider_type: providerType,
        provider_name: providerName,
        access_token: accessToken,
      });
      setShowAdd(false);
      setProviderName('');
      setAccessToken('');
      loadProviders();
    } catch {
      // error
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    await client.delete(`/providers/${id}`);
    loadProviders();
  };

  const handleValidate = async (id: string) => {
    try {
      await client.post(`/providers/${id}/validate`);
    } catch {
      // error
    }
  };

  return (
    <AppShell activeTab="" onTabChange={() => {}}>
      <div className="p-4 max-w-2xl mx-auto space-y-6">
        <h1 className="text-xl font-bold">Settings</h1>

        <section>
          <div className="flex items-center justify-between mb-3">
            <h2 className="font-semibold">Provider Connections</h2>
            <Button size="sm" onClick={() => setShowAdd(!showAdd)}>
              <Plus size={14} className="mr-1" /> Add Provider
            </Button>
          </div>

          {showAdd && (
            <Card className="mb-4">
              <form onSubmit={handleAdd} className="space-y-3">
                <div>
                  <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Type</label>
                  <select
                    value={providerType}
                    onChange={(e) => setProviderType(e.target.value)}
                    className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)]"
                  >
                    <option value="github">GitHub</option>
                    <option value="openai">OpenAI</option>
                    <option value="anthropic">Anthropic</option>
                    <option value="google">Google</option>
                    <option value="mistral">Mistral</option>
                    <option value="cohere">Cohere</option>
                  </select>
                </div>
                <Input
                  label="Name"
                  value={providerName}
                  onChange={(e) => setProviderName(e.target.value)}
                  placeholder="My GitHub Account"
                  required
                />
                <Input
                  label="Access Token"
                  type="password"
                  value={accessToken}
                  onChange={(e) => setAccessToken(e.target.value)}
                  placeholder="ghp_... or sk-..."
                  required
                />
                <div className="flex gap-2">
                  <Button type="submit" loading={loading} size="sm">Connect</Button>
                  <Button type="button" variant="ghost" size="sm" onClick={() => setShowAdd(false)}>Cancel</Button>
                </div>
              </form>
            </Card>
          )}

          <div className="space-y-2">
            {providers.length === 0 ? (
              <p className="text-sm text-[var(--color-text-secondary)] text-center py-4">
                No providers connected yet
              </p>
            ) : (
              providers.map((p) => (
                <Card key={p.id} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    {p.status === 'active' ? (
                      <CheckCircle size={18} className="text-[var(--color-success)]" />
                    ) : (
                      <XCircle size={18} className="text-[var(--color-error)]" />
                    )}
                    <div>
                      <div className="font-medium text-sm">{p.provider_name}</div>
                      <div className="text-xs text-[var(--color-text-secondary)]">{p.provider_type}</div>
                    </div>
                  </div>
                  <div className="flex gap-1">
                    <Button size="sm" variant="ghost" onClick={() => handleValidate(p.id)}>
                      Validate
                    </Button>
                    <button
                      onClick={() => handleDelete(p.id)}
                      className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded-lg text-[var(--color-error)]"
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                </Card>
              ))
            )}
          </div>
        </section>
      </div>
    </AppShell>
  );
}
