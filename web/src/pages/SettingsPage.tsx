import { useState, useEffect, FormEvent } from 'react';
import { AppShell } from '../components/layout/AppShell';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Card } from '../components/ui/Card';
import { Plus, Trash2, CheckCircle, XCircle, Shield, ShieldOff, Copy } from 'lucide-react';
import { useAuthStore } from '../store/authStore';
import { authApi, type TwoFASetup } from '../api/auth';
import client from '../api/client';

interface ProviderConnection {
  id: string;
  provider_type: string;
  provider_name: string;
  status: string;
  created_at: string;
}

export function SettingsPage() {
  const { user, updateUser } = useAuthStore();
  const [providers, setProviders] = useState<ProviderConnection[]>([]);
  const [showAdd, setShowAdd] = useState(false);
  const [providerType, setProviderType] = useState('github');
  const [providerName, setProviderName] = useState('');
  const [accessToken, setAccessToken] = useState('');
  const [loading, setLoading] = useState(false);

  // 2FA state
  const [twoFASetup, setTwoFASetup] = useState<TwoFASetup | null>(null);
  const [twoFACode, setTwoFACode] = useState('');
  const [twoFALoading, setTwoFALoading] = useState(false);
  const [twoFAError, setTwoFAError] = useState('');
  const [disableCode, setDisableCode] = useState('');
  const [showDisable, setShowDisable] = useState(false);

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

  const handleSetup2FA = async () => {
    setTwoFALoading(true);
    setTwoFAError('');
    try {
      const { data } = await authApi.setup2FA();
      setTwoFASetup(data);
    } catch {
      setTwoFAError('Failed to set up 2FA');
    } finally {
      setTwoFALoading(false);
    }
  };

  const handleConfirm2FA = async (e: FormEvent) => {
    e.preventDefault();
    setTwoFALoading(true);
    setTwoFAError('');
    try {
      await authApi.confirm2FA(twoFACode);
      if (user) updateUser({ ...user, two_factor_enabled: true });
      setTwoFASetup(null);
      setTwoFACode('');
    } catch {
      setTwoFAError('Invalid code. Please try again.');
    } finally {
      setTwoFALoading(false);
    }
  };

  const handleDisable2FA = async (e: FormEvent) => {
    e.preventDefault();
    setTwoFALoading(true);
    setTwoFAError('');
    try {
      await authApi.disable2FA(disableCode);
      if (user) updateUser({ ...user, two_factor_enabled: false });
      setShowDisable(false);
      setDisableCode('');
    } catch {
      setTwoFAError('Invalid code. Please try again.');
    } finally {
      setTwoFALoading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  return (
    <AppShell activeTab="" onTabChange={() => {}}>
      <div className="p-4 max-w-2xl mx-auto space-y-6">
        <h1 className="text-xl font-bold">Settings</h1>

        {/* Security / 2FA Section */}
        <section>
          <h2 className="font-semibold mb-3">Security</h2>
          <Card>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                {user?.two_factor_enabled ? (
                  <Shield size={20} className="text-[var(--color-success)]" />
                ) : (
                  <ShieldOff size={20} className="text-[var(--color-text-secondary)]" />
                )}
                <div>
                  <div className="font-medium text-sm">Two-Factor Authentication</div>
                  <div className="text-xs text-[var(--color-text-secondary)]">
                    {user?.two_factor_enabled ? 'Enabled - Your account is protected' : 'Not enabled'}
                  </div>
                </div>
              </div>
              {user?.two_factor_enabled ? (
                <Button size="sm" variant="ghost" onClick={() => setShowDisable(!showDisable)}>
                  Disable
                </Button>
              ) : (
                <Button size="sm" onClick={handleSetup2FA} loading={twoFALoading}>
                  Enable
                </Button>
              )}
            </div>

            {/* 2FA Setup Flow */}
            {twoFASetup && (
              <div className="mt-4 pt-4 border-t border-[var(--color-border)] space-y-3">
                <p className="text-sm text-[var(--color-text-secondary)]">
                  Scan this URL with your authenticator app, or enter the secret manually:
                </p>
                <div className="bg-[var(--color-bg)] p-3 rounded-lg">
                  <div className="text-xs text-[var(--color-text-secondary)] mb-1">Secret Key</div>
                  <div className="flex items-center gap-2">
                    <code className="text-sm font-mono flex-1 break-all">{twoFASetup.secret}</code>
                    <button onClick={() => copyToClipboard(twoFASetup.secret)} className="p-1 hover:bg-[var(--color-bg-tertiary)] rounded">
                      <Copy size={14} />
                    </button>
                  </div>
                </div>
                <div className="bg-[var(--color-bg)] p-3 rounded-lg">
                  <div className="text-xs text-[var(--color-text-secondary)] mb-1">Backup Codes (save these!)</div>
                  <div className="grid grid-cols-2 gap-1">
                    {twoFASetup.backup_codes.map((code, i) => (
                      <code key={i} className="text-xs font-mono">{code}</code>
                    ))}
                  </div>
                  <button
                    onClick={() => copyToClipboard(twoFASetup.backup_codes.join('\n'))}
                    className="mt-2 text-xs text-[var(--color-primary)] hover:underline flex items-center gap-1"
                  >
                    <Copy size={12} /> Copy all
                  </button>
                </div>
                <form onSubmit={handleConfirm2FA} className="flex gap-2">
                  <Input
                    placeholder="Enter 6-digit code"
                    value={twoFACode}
                    onChange={(e) => setTwoFACode(e.target.value)}
                    maxLength={8}
                    required
                  />
                  <Button type="submit" size="sm" loading={twoFALoading}>Verify</Button>
                </form>
              </div>
            )}

            {/* Disable 2FA */}
            {showDisable && (
              <form onSubmit={handleDisable2FA} className="mt-4 pt-4 border-t border-[var(--color-border)]">
                <p className="text-sm text-[var(--color-text-secondary)] mb-2">
                  Enter your authentication code to disable 2FA:
                </p>
                <div className="flex gap-2">
                  <Input
                    placeholder="Enter code"
                    value={disableCode}
                    onChange={(e) => setDisableCode(e.target.value)}
                    maxLength={8}
                    required
                  />
                  <Button type="submit" size="sm" variant="ghost" loading={twoFALoading}>Disable</Button>
                </div>
              </form>
            )}

            {twoFAError && <p className="text-sm text-[var(--color-error)] mt-2">{twoFAError}</p>}
          </Card>
        </section>

        {/* Provider Connections */}
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
