import { useState } from 'react';
import { Plus } from 'lucide-react';
import { AppShell } from '../components/layout/AppShell';
import { ProjectList } from '../components/project/ProjectList';
import { CreateProject } from '../components/project/CreateProject';
import { ProjectDashboard } from '../components/project/ProjectDashboard';
import { Button } from '../components/ui/Button';
import type { Project } from '../api/projects';

export function DashboardPage() {
  const [activeProject, setActiveProject] = useState<Project | null>(null);
  const [showCreate, setShowCreate] = useState(false);

  if (activeProject) {
    return (
      <ProjectDashboard
        project={activeProject}
        onBack={() => setActiveProject(null)}
      />
    );
  }

  return (
    <AppShell activeTab="" onTabChange={() => {}}>
      <div className="p-4 max-w-5xl mx-auto">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-bold">Projects</h1>
            <p className="text-sm text-[var(--color-text-secondary)]">
              Each repository is a self-contained project
            </p>
          </div>
          <Button onClick={() => setShowCreate(true)}>
            <Plus size={16} className="mr-1" /> New Project
          </Button>
        </div>

        <ProjectList onSelectProject={setActiveProject} />
        <CreateProject open={showCreate} onClose={() => setShowCreate(false)} />
      </div>
    </AppShell>
  );
}
