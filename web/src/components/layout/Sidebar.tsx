import { MessageSquare, Terminal, Bot, Wrench, FileText, Settings, Brain, Users } from 'lucide-react';

interface SidebarProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

const tabs = [
  { id: 'chat', label: 'Chat', icon: MessageSquare },
  { id: 'cli', label: 'CLI', icon: Terminal },
  { id: 'agents', label: 'Agents', icon: Bot },
  { id: 'teams', label: 'Teams', icon: Users },
  { id: 'memory', label: 'Memory', icon: Brain },
  { id: 'devops', label: 'DevOps', icon: Wrench },
  { id: 'logs', label: 'Logs', icon: FileText },
  { id: 'settings', label: 'Settings', icon: Settings },
];

export function Sidebar({ activeTab, onTabChange }: SidebarProps) {
  return (
    <aside className="hidden sm:flex w-56 bg-[var(--color-bg-secondary)] border-r border-[var(--color-border)] flex-col py-4 shrink-0">
      <nav className="flex flex-col gap-1 px-3">
        {tabs.map((tab) => {
          const Icon = tab.icon;
          const active = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              onClick={() => onTabChange(tab.id)}
              className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                active
                  ? 'bg-[var(--color-primary)] text-white'
                  : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-text)]'
              }`}
            >
              <Icon size={18} />
              {tab.label}
            </button>
          );
        })}
      </nav>
    </aside>
  );
}
