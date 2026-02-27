import { useState, useEffect, useRef } from 'react';
import { Search, X, MessageSquare, Brain, Bot, FileText, GitBranch } from 'lucide-react';
import { platformApi, type SearchResult } from '../../api/platform';

interface GlobalSearchProps {
  open: boolean;
  onClose: () => void;
  onNavigate?: (result: SearchResult) => void;
}

const typeIcons: Record<string, typeof MessageSquare> = {
  message: MessageSquare,
  memory: Brain,
  agent: Bot,
  document: FileText,
  workflow: GitBranch,
};

const typeLabels: Record<string, string> = {
  message: 'Chat',
  memory: 'Memory',
  agent: 'Agent',
  document: 'Document',
  workflow: 'Workflow',
};

export function GlobalSearch({ open, onClose, onNavigate }: GlobalSearchProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    if (open) {
      setQuery('');
      setResults([]);
      setSelectedIndex(0);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [open]);

  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      return;
    }
    clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(async () => {
      setLoading(true);
      try {
        const res = await platformApi.search(query.trim());
        setResults(res.data.results || []);
        setSelectedIndex(0);
      } catch (err) {
        console.error(err);
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, 300);
    return () => clearTimeout(debounceRef.current);
  }, [query]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((i) => Math.min(i + 1, results.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === 'Enter' && results[selectedIndex]) {
      onNavigate?.(results[selectedIndex]);
      onClose();
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]" onClick={onClose}>
      <div className="absolute inset-0 bg-black/50" />
      <div
        className="relative w-full max-w-xl bg-[var(--color-bg)] rounded-xl shadow-2xl border border-[var(--color-border)] overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-3 px-4 py-3 border-b border-[var(--color-border)]">
          <Search size={18} className="text-[var(--color-text-secondary)] shrink-0" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search messages, memories, agents, documents..."
            className="flex-1 bg-transparent outline-none text-sm"
          />
          {query && (
            <button onClick={() => setQuery('')} className="p-1 hover:bg-[var(--color-bg-tertiary)] rounded">
              <X size={14} />
            </button>
          )}
          <kbd className="hidden sm:inline text-[10px] text-[var(--color-text-secondary)] border border-[var(--color-border)] rounded px-1.5 py-0.5">
            ESC
          </kbd>
        </div>

        <div className="max-h-80 overflow-y-auto">
          {loading && (
            <div className="px-4 py-6 text-center text-sm text-[var(--color-text-secondary)]">Searching...</div>
          )}

          {!loading && query && results.length === 0 && (
            <div className="px-4 py-6 text-center text-sm text-[var(--color-text-secondary)]">No results found</div>
          )}

          {!loading && results.map((result, i) => {
            const Icon = typeIcons[result.type] || FileText;
            return (
              <button
                key={`${result.type}-${result.id}`}
                onClick={() => {
                  onNavigate?.(result);
                  onClose();
                }}
                className={`w-full flex items-start gap-3 px-4 py-3 text-left hover:bg-[var(--color-bg-tertiary)] transition-colors ${
                  i === selectedIndex ? 'bg-[var(--color-bg-tertiary)]' : ''
                }`}
              >
                <Icon size={16} className="text-[var(--color-text-secondary)] mt-0.5 shrink-0" />
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium truncate">{result.title}</span>
                    <span className="text-[10px] uppercase tracking-wide text-[var(--color-text-secondary)] bg-[var(--color-bg-secondary)] px-1.5 py-0.5 rounded">
                      {typeLabels[result.type] || result.type}
                    </span>
                  </div>
                  <p className="text-xs text-[var(--color-text-secondary)] mt-0.5 line-clamp-2">{result.snippet}</p>
                </div>
              </button>
            );
          })}

          {!query && (
            <div className="px-4 py-6 text-center text-sm text-[var(--color-text-secondary)]">
              Type to search across all projects
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
