import { useState, useEffect } from 'react';
import { Star, GripVertical } from 'lucide-react';
import { platformApi, type PinnedProject } from '../../api/platform';

interface PinnedProjectsProps {
  onSelectProject?: (projectId: string) => void;
}

export function PinnedProjects({ onSelectProject }: PinnedProjectsProps) {
  const [pins, setPins] = useState<PinnedProject[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchPins = async () => {
    try {
      const res = await platformApi.listPinned();
      setPins(res.data.pinned || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPins();
  }, []);

  const handleUnpin = async (projectId: string) => {
    if (!window.confirm('Unpin this project?')) return;
    try {
      await platformApi.unpinProject(projectId);
      setPins((prev) => prev.filter((p) => p.project_id !== projectId));
    } catch (err) {
      console.error(err);
    }
  };

  if (loading) {
    return (
      <div className="mb-6">
        <h3 className="text-sm font-semibold text-[var(--color-text-secondary)] mb-2 flex items-center gap-1.5">
          <Star size={14} /> Pinned Projects
        </h3>
        <div className="text-xs text-[var(--color-text-secondary)]">Loading...</div>
      </div>
    );
  }

  if (pins.length === 0) return null;

  return (
    <div className="mb-6">
      <h3 className="text-sm font-semibold text-[var(--color-text-secondary)] mb-2 flex items-center gap-1.5">
        <Star size={14} /> Pinned Projects
      </h3>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
        {pins.map((pin) => (
          <div
            key={pin.id}
            className="flex items-center gap-2 px-3 py-2 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] hover:border-[var(--color-primary)] transition-colors cursor-pointer group"
            onClick={() => onSelectProject?.(pin.project_id)}
          >
            <GripVertical size={14} className="text-[var(--color-text-secondary)] opacity-0 group-hover:opacity-50 shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="text-sm font-medium truncate">{pin.project?.name || 'Project'}</div>
              {pin.project && (
                <div className="text-[11px] text-[var(--color-text-secondary)] truncate">
                  {pin.project.repo_owner}/{pin.project.repo_name}
                </div>
              )}
            </div>
            <button
              onClick={(e) => {
                e.stopPropagation();
                handleUnpin(pin.project_id);
              }}
              className="p-1 hover:bg-[var(--color-bg-tertiary)] rounded opacity-0 group-hover:opacity-100 transition-opacity"
              title="Unpin"
            >
              <Star size={14} className="text-yellow-500 fill-yellow-500" />
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
