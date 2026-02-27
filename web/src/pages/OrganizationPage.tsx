import { useState, useEffect, FormEvent } from 'react';
import { AppShell } from '../components/layout/AppShell';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Card } from '../components/ui/Card';
import { Modal } from '../components/ui/Modal';
import { Plus, Users, Crown, Shield, Eye, UserMinus, Mail } from 'lucide-react';
import { organizationsApi, type Organization, type OrgMember } from '../api/organizations';

export function OrganizationPage() {
  const [orgs, setOrgs] = useState<Organization[]>([]);
  const [selectedOrg, setSelectedOrg] = useState<Organization | null>(null);
  const [members, setMembers] = useState<OrgMember[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [showInvite, setShowInvite] = useState(false);
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [description, setDescription] = useState('');
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState('member');
  const [loading, setLoading] = useState(false);

  const loadOrgs = async () => {
    const { data } = await organizationsApi.list();
    setOrgs(data.organizations || []);
  };

  const loadMembers = async (orgId: string) => {
    const { data } = await organizationsApi.listMembers(orgId);
    setMembers(data.members || []);
  };

  useEffect(() => { loadOrgs(); }, []);

  useEffect(() => {
    if (selectedOrg) loadMembers(selectedOrg.id);
  }, [selectedOrg]);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await organizationsApi.create({ name, slug, description });
      setShowCreate(false);
      setName(''); setSlug(''); setDescription('');
      loadOrgs();
    } finally { setLoading(false); }
  };

  const handleInvite = async (e: FormEvent) => {
    e.preventDefault();
    if (!selectedOrg) return;
    setLoading(true);
    try {
      await organizationsApi.inviteMember(selectedOrg.id, inviteEmail, inviteRole);
      setShowInvite(false);
      setInviteEmail('');
      loadMembers(selectedOrg.id);
    } finally { setLoading(false); }
  };

  const handleRemoveMember = async (userId: string) => {
    if (!selectedOrg) return;
    await organizationsApi.removeMember(selectedOrg.id, userId);
    loadMembers(selectedOrg.id);
  };

  const roleIcon = (role: string) => {
    switch (role) {
      case 'owner': return <Crown size={14} className="text-yellow-400" />;
      case 'admin': return <Shield size={14} className="text-blue-400" />;
      case 'viewer': return <Eye size={14} className="text-gray-400" />;
      default: return <Users size={14} className="text-green-400" />;
    }
  };

  if (selectedOrg) {
    return (
      <AppShell activeTab="" onTabChange={() => {}}>
        <div className="p-4 max-w-2xl mx-auto space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <button onClick={() => setSelectedOrg(null)} className="text-sm text-[var(--color-primary)] hover:underline mb-1">
                Back to Organizations
              </button>
              <h1 className="text-xl font-bold">{selectedOrg.name}</h1>
              <p className="text-sm text-[var(--color-text-secondary)]">{selectedOrg.description}</p>
            </div>
            <Button size="sm" onClick={() => setShowInvite(true)}>
              <Mail size={14} className="mr-1" /> Invite
            </Button>
          </div>

          <h2 className="font-semibold mt-6">Members</h2>
          <div className="space-y-2">
            {members.map((m) => (
              <Card key={m.id} className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  {roleIcon(m.role)}
                  <div>
                    <div className="text-sm font-medium">{m.user?.display_name || m.user?.email}</div>
                    <div className="text-xs text-[var(--color-text-secondary)]">{m.role}</div>
                  </div>
                </div>
                {m.role !== 'owner' && (
                  <button onClick={() => handleRemoveMember(m.user_id)} className="p-1.5 hover:bg-[var(--color-bg-tertiary)] rounded-lg text-[var(--color-error)]">
                    <UserMinus size={16} />
                  </button>
                )}
              </Card>
            ))}
          </div>

          <Modal isOpen={showInvite} onClose={() => setShowInvite(false)} title="Invite Member">
            <form onSubmit={handleInvite} className="space-y-3">
              <Input label="Email" type="email" value={inviteEmail} onChange={(e) => setInviteEmail(e.target.value)} required />
              <div>
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Role</label>
                <select value={inviteRole} onChange={(e) => setInviteRole(e.target.value)} className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)]">
                  <option value="admin">Admin</option>
                  <option value="member">Member</option>
                  <option value="viewer">Viewer</option>
                </select>
              </div>
              <Button type="submit" loading={loading} className="w-full">Send Invitation</Button>
            </form>
          </Modal>
        </div>
      </AppShell>
    );
  }

  return (
    <AppShell activeTab="" onTabChange={() => {}}>
      <div className="p-4 max-w-2xl mx-auto space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold">Organizations</h1>
          <Button size="sm" onClick={() => setShowCreate(true)}>
            <Plus size={14} className="mr-1" /> Create
          </Button>
        </div>

        <div className="space-y-2">
          {orgs.length === 0 ? (
            <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">No organizations yet</p>
          ) : (
            orgs.map((org) => (
              <Card key={org.id} onClick={() => setSelectedOrg(org)} className="cursor-pointer">
                <div className="flex items-center gap-3">
                  <Users size={20} className="text-[var(--color-primary)]" />
                  <div>
                    <div className="font-medium text-sm">{org.name}</div>
                    <div className="text-xs text-[var(--color-text-secondary)]">/{org.slug}</div>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>

        <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title="Create Organization">
          <form onSubmit={handleCreate} className="space-y-3">
            <Input label="Name" value={name} onChange={(e) => { setName(e.target.value); setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '-')); }} required />
            <Input label="Slug" value={slug} onChange={(e) => setSlug(e.target.value)} required />
            <Input label="Description" value={description} onChange={(e) => setDescription(e.target.value)} />
            <Button type="submit" loading={loading} className="w-full">Create</Button>
          </form>
        </Modal>
      </div>
    </AppShell>
  );
}
