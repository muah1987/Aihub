import { MessageSquare, Terminal, Bot, Users, Brain } from 'lucide-react';

interface TabBarProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

const tabs = [
  { id: 'chat', label: 'Chat', icon: MessageSquare },
  { id: 'cli', label: 'CLI', icon: Terminal },
  { id: 'agents', label: 'Agents', icon: Bot },
  { id: 'teams', label: 'Teams', icon: Users },
  { id: 'memory', label: 'Memory', icon: Brain },
];

export function TabBar({ activeTab, onTabChange }: TabBarProps) {
  return (
    <nav className="h-16 bg-[var(--color-bg-secondary)] border-t border-[var(--color-border)] flex items-center justify-around px-2 shrink-0 sm:hidden">
      {tabs.slice(0, 5).map((tab) => {
        const Icon = tab.icon;
        const active = activeTab === tab.id;
        return (
          <button
            key={tab.id}
            onClick={() => onTabChange(tab.id)}
            className={`flex flex-col items-center gap-0.5 px-2 py-1 rounded-lg transition-colors ${
              active ? 'text-[var(--color-primary)]' : 'text-[var(--color-text-secondary)]'
            }`}
          >
            <Icon size={20} />
            <span className="text-[10px]">{tab.label}</span>
          </button>
        );
      })}
    </nav>
  );
}
