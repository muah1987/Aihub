import { useState } from 'react';
import type { Project } from '../../api/projects';
import { AppShell } from '../layout/AppShell';
import { ChatTimeline } from '../chat/ChatTimeline';
import { TerminalView } from '../terminal/Terminal';
import { AgentPanel } from '../agent/AgentPanel';
import { MemoryPanel } from '../memory/MemoryPanel';
import { TeamPanel } from '../team/TeamPanel';
import { ProjectSettings } from './ProjectSettings';
import { DevOpsPanel } from '../devops/DevOpsPanel';
import { LogsPanel } from '../log-viewer/LogsPanel';
import { ToolsPanel } from '../tools/ToolsPanel';
import { AnalyticsPanel } from '../analytics/AnalyticsPanel';
import { WorkflowPanel } from '../workflow/WorkflowPanel';
import { SchedulerPanel } from '../scheduler/SchedulerPanel';
import { ActivityPanel } from '../activity/ActivityPanel';
import { KnowledgePanel } from '../knowledge/KnowledgePanel';
import { IntegrationsPanel } from '../integrations/IntegrationsPanel';
import { CompliancePanel } from '../compliance/CompliancePanel';
import { PreferencesPanel } from '../platform/PreferencesPanel';
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
      case 'teams':
        return <TeamPanel projectId={project.id} />;
      case 'memory':
        return <MemoryPanel projectId={project.id} />;
      case 'tools':
        return <ToolsPanel projectId={project.id} />;
      case 'analytics':
        return <AnalyticsPanel projectId={project.id} />;
      case 'workflows':
        return <WorkflowPanel projectId={project.id} />;
      case 'schedules':
        return <SchedulerPanel projectId={project.id} />;
      case 'activity':
        return <ActivityPanel projectId={project.id} />;
      case 'knowledge':
        return <KnowledgePanel projectId={project.id} />;
      case 'integrations':
        return <IntegrationsPanel projectId={project.id} />;
      case 'compliance':
        return <CompliancePanel projectId={project.id} />;
      case 'devops':
        return <DevOpsPanel projectId={project.id} />;
      case 'logs':
        return <LogsPanel projectId={project.id} />;
      case 'preferences':
        return <PreferencesPanel />;
      case 'settings':
        return <ProjectSettings projectId={project.id} />;
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
        <div className="hidden sm:flex border-b border-[var(--color-border)] px-4 overflow-x-auto">
          {['chat', 'cli', 'agents', 'teams', 'memory', 'tools', 'analytics', 'workflows', 'schedules', 'activity', 'knowledge', 'integrations', 'compliance', 'devops', 'logs', 'preferences', 'settings'].map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-2 text-sm capitalize border-b-2 transition-colors whitespace-nowrap ${
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
