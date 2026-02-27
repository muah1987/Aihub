import { useState, useEffect } from 'react';
import { activityApi, type ActivityLog } from '../../api/activity';
import { Activity, Filter, ChevronLeft, ChevronRight } from 'lucide-react';

interface ActivityPanelProps {
  projectId: string;
}

const actionIcons: Record<string, string> = {
  'create': '+',
  'update': '~',
  'delete': '-',
  'invoke': '>',
  'deploy': '^',
  'trigger': '!',
  'login': '@',
};

const actionColors: Record<string, string> = {
  'create': 'bg-green-500/20 text-green-400',
  'update': 'bg-blue-500/20 text-blue-400',
  'delete': 'bg-red-500/20 text-red-400',
  'invoke': 'bg-purple-500/20 text-purple-400',
  'deploy': 'bg-yellow-500/20 text-yellow-400',
  'trigger': 'bg-orange-500/20 text-orange-400',
};

const PAGE_SIZE = 25;

export function ActivityPanel({ projectId }: ActivityPanelProps) {
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [filterAction, setFilterAction] = useState('');
  const [filterResource, setFilterResource] = useState('');
  const [showFilters, setShowFilters] = useState(false);

  useEffect(() => {
    loadActivity();
  }, [projectId, page, filterAction, filterResource]);

  const loadActivity = async () => {
    setLoading(true);
    try {
      const res = await activityApi.list(projectId, {
        limit: PAGE_SIZE,
        offset: page * PAGE_SIZE,
        action: filterAction || undefined,
        resource_type: filterResource || undefined,
      });
      setLogs(res.data.activity || []);
      setTotal(res.data.total || 0);
    } catch { /* empty */ }
    setLoading(false);
  };

  const getActionColor = (action: string) => {
    const base = action.split('_')[0];
    return actionColors[base] || 'bg-gray-500/20 text-gray-400';
  };

  const getActionIcon = (action: string) => {
    const base = action.split('_')[0];
    return actionIcons[base] || '?';
  };

  const formatDetails = (details: Record<string, unknown>) => {
    const entries = Object.entries(details).filter(([, v]) => v !== null && v !== undefined);
    if (entries.length === 0) return null;
    return entries.map(([k, v]) => `${k}: ${typeof v === 'object' ? JSON.stringify(v) : String(v)}`).join(', ');
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <h3 className="font-semibold text-sm">Activity Log</h3>
        <div className="flex items-center gap-2">
          <span className="text-xs text-[var(--color-text-secondary)]">{total} events</span>
          <button onClick={() => setShowFilters(!showFilters)} className={`p-1.5 rounded-lg text-xs ${showFilters ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg-secondary)] text-[var(--color-text-secondary)]'}`}>
            <Filter size={14} />
          </button>
        </div>
      </div>

      {showFilters && (
        <div className="flex gap-2 px-4 py-2 border-b border-[var(--color-border)] shrink-0">
          <input
            value={filterAction} onChange={e => { setFilterAction(e.target.value); setPage(0); }}
            placeholder="Filter by action..."
            className="flex-1 px-3 py-1.5 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-xs"
          />
          <input
            value={filterResource} onChange={e => { setFilterResource(e.target.value); setPage(0); }}
            placeholder="Filter by resource type..."
            className="flex-1 px-3 py-1.5 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-xs"
          />
          {(filterAction || filterResource) && (
            <button onClick={() => { setFilterAction(''); setFilterResource(''); setPage(0); }} className="text-xs text-red-400 hover:text-red-300 px-2">
              Clear
            </button>
          )}
        </div>
      )}

      <div className="flex-1 overflow-y-auto">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : logs.length === 0 ? (
          <div className="text-center py-12">
            <Activity size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)]">No activity recorded yet</p>
          </div>
        ) : (
          <div className="divide-y divide-[var(--color-border)]">
            {logs.map(log => (
              <div key={log.id} className="flex items-start gap-3 px-4 py-3 hover:bg-[var(--color-bg-secondary)] transition-colors">
                <div className={`flex items-center justify-center w-7 h-7 rounded-full text-xs font-bold shrink-0 mt-0.5 ${getActionColor(log.action)}`}>
                  {getActionIcon(log.action)}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-sm font-medium">{log.action}</span>
                    <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)]">
                      {log.resource_type}
                    </span>
                    {log.resource_id && (
                      <span className="text-xs font-mono text-[var(--color-text-secondary)] truncate max-w-[120px]">
                        {log.resource_id.substring(0, 8)}...
                      </span>
                    )}
                  </div>
                  {log.details && Object.keys(log.details).length > 0 && (
                    <p className="text-xs text-[var(--color-text-secondary)] mt-0.5 truncate">
                      {formatDetails(log.details)}
                    </p>
                  )}
                </div>
                <span className="text-xs text-[var(--color-text-secondary)] shrink-0 whitespace-nowrap">
                  {new Date(log.created_at).toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between px-4 py-2 border-t border-[var(--color-border)] shrink-0">
          <button
            onClick={() => setPage(Math.max(0, page - 1))} disabled={page === 0}
            className="flex items-center gap-1 text-xs text-[var(--color-text-secondary)] disabled:opacity-30"
          >
            <ChevronLeft size={14} /> Previous
          </button>
          <span className="text-xs text-[var(--color-text-secondary)]">
            Page {page + 1} of {totalPages}
          </span>
          <button
            onClick={() => setPage(Math.min(totalPages - 1, page + 1))} disabled={page >= totalPages - 1}
            className="flex items-center gap-1 text-xs text-[var(--color-text-secondary)] disabled:opacity-30"
          >
            Next <ChevronRight size={14} />
          </button>
        </div>
      )}
    </div>
  );
}
