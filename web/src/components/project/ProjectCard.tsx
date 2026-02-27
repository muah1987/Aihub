import { GitBranch, Clock } from 'lucide-react';
import type { Project } from '../../api/projects';
import { Card } from '../ui/Card';

interface ProjectCardProps {
  project: Project;
  onClick: () => void;
}

export function ProjectCard({ project, onClick }: ProjectCardProps) {
  const timeAgo = new Date(project.updated_at).toLocaleDateString();

  return (
    <Card onClick={onClick} className="flex flex-col gap-2">
      <div className="flex items-start justify-between">
        <h3 className="font-semibold text-base truncate">{project.name}</h3>
        <span className={`text-xs px-2 py-0.5 rounded-full ${
          project.status === 'active'
            ? 'bg-green-500/20 text-green-400'
            : 'bg-gray-500/20 text-gray-400'
        }`}>
          {project.status}
        </span>
      </div>
      {project.description && (
        <p className="text-sm text-[var(--color-text-secondary)] line-clamp-2">
          {project.description}
        </p>
      )}
      <div className="flex items-center gap-3 text-xs text-[var(--color-text-secondary)] mt-auto pt-2">
        <span className="flex items-center gap-1">
          <GitBranch size={12} />
          {project.repo_owner}/{project.repo_name}
        </span>
        <span className="flex items-center gap-1">
          <Clock size={12} />
          {timeAgo}
        </span>
      </div>
    </Card>
  );
}
