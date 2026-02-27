import { useState } from 'react';
import type { Project } from '../../api/projects';
import { AppShell } from '../layout/AppShell';
import { ChatTimeline } from '../chat/ChatTimeline';
import { TerminalView } from '../terminal/Terminal';
import { AgentPanel } from '../agent/AgentPanel';
import { ArrowLeft } from 'lucide-react';

interface ProjectDashboardProps {
  project: Project;
  onBack: () => void;
}

export function ProjectDashboard({ project, onBack }: ProjectDashboardProps) {
  const [activeTab, setActiveTab] = useState('chat');

  const renderTabContent = () => {
    switch (activeTab) {
      case 'chat':
        return <ChatTimeline projectId={project.id} />;
      case 'cli':
        return <TerminalView projectId={project.id} />;
      case 'agents':
        return <AgentPanel projectId={project.id} />;
      case 'devops':
        return (
          <div className="flex items-center justify-center h-full text-[var(--color-text-secondary)]">
            DevOps Space - Coming in Phase 3
          </div>
        );
      case 'logs':
        return (
          <div className="flex items-center justify-center h-full text-[var(--color-text-secondary)]">
            Logs - Coming in Phase 3
          </div>
        );
      case 'settings':
        return (
          <div className="flex items-center justify-center h-full text-[var(--color-text-secondary)]">
            Project Settings - Coming Soon
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <AppShell activeTab={activeTab} onTabChange={setActiveTab} showTabs>
      <div className="flex flex-col h-full">
        <div className="flex items-center gap-3 px-4 py-3 border-b border-[var(--color-border)] shrink-0">
          <button onClick={onBack} className="p-1 hover:bg-[var(--color-bg-tertiary)] rounded-lg">
            <ArrowLeft size={20} />
          </button>
          <div>
            <h2 className="font-semibold text-sm">{project.name}</h2>
            <span className="text-xs text-[var(--color-text-secondary)]">
              {project.repo_owner}/{project.repo_name}
            </span>
          </div>
        </div>

        {/* Desktop tab bar */}
        <div className="hidden sm:flex border-b border-[var(--color-border)] px-4">
          {['chat', 'cli', 'agents', 'devops', 'logs', 'settings'].map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-2 text-sm capitalize border-b-2 transition-colors ${
                activeTab === tab
                  ? 'border-[var(--color-primary)] text-[var(--color-primary)]'
                  : 'border-transparent text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'
              }`}
            >
              {tab}
            </button>
          ))}
        </div>

        <div className="flex-1 overflow-hidden">{renderTabContent()}</div>
      </div>
    </AppShell>
  );
}
