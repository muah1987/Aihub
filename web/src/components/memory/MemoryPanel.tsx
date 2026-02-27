import { useState, useEffect, FormEvent } from 'react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Card } from '../ui/Card';
import { Modal } from '../ui/Modal';
import { Plus, Search, Pin, Trash2 } from 'lucide-react';
import { memoryApi, type TeamMemory } from '../../api/memory';

interface MemoryPanelProps {
  projectId: string;
}

export function MemoryPanel({ projectId }: MemoryPanelProps) {
  const [memories, setMemories] = useState<TeamMemory[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('');
  const [showAdd, setShowAdd] = useState(false);
  const [loading, setLoading] = useState(false);
  const [category, setCategory] = useState('general');
  const [key, setKey] = useState('');
  const [content, setContent] = useState('');
  const [pinned, setPinned] = useState(false);

  const loadMemories = async () => {
    const { data } = await memoryApi.list(projectId, selectedCategory);
    setMemories(data.memories || []);
  };

  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      loadMemories();
      return;
    }
    const { data } = await memoryApi.search(projectId, searchQuery);
    setMemories(data.memories || []);
  };

  useEffect(() => { loadMemories(); }, [projectId, selectedCategory]);

  const handleAdd = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await memoryApi.set(projectId, { category, key, content, pinned });
      setShowAdd(false);
      setKey(''); setContent(''); setPinned(false);
      loadMemories();
    } finally { setLoading(false); }
  };

  const handleDelete = async (memoryId: string) => {
    await memoryApi.delete(projectId, memoryId);
    loadMemories();
  };

  const categories = [...new Set(memories.map((m) => m.category))];

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <div className="flex-1 flex gap-2">
          <div className="relative flex-1">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-secondary)]" />
            <input
              type="text"
              placeholder="Search memory..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              className="w-full pl-9 pr-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-sm text-[var(--color-text)]"
            />
          </div>
          <Button size="sm" onClick={handleSearch}>Search</Button>
        </div>
        <Button size="sm" onClick={() => setShowAdd(true)}>
          <Plus size={14} className="mr-1" /> Add
        </Button>
      </div>

      {categories.length > 0 && (
        <div className="flex gap-1 px-4 py-2 overflow-x-auto border-b border-[var(--color-border)]">
          <button
            onClick={() => setSelectedCategory('')}
            className={`px-3 py-1 rounded-full text-xs whitespace-nowrap ${!selectedCategory ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)]'}`}
          >
            All
          </button>
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setSelectedCategory(cat)}
              className={`px-3 py-1 rounded-full text-xs whitespace-nowrap ${selectedCategory === cat ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)]'}`}
            >
              {cat}
            </button>
          ))}
        </div>
      )}

      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {memories.length === 0 ? (
          <div className="text-center text-sm text-[var(--color-text-secondary)] py-12">
            No memories yet. Add knowledge that persists across sessions.
          </div>
        ) : (
          memories.map((mem) => (
            <Card key={mem.id} className="group">
              <div className="flex items-start justify-between gap-2">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    {mem.pinned && <Pin size={12} className="text-[var(--color-primary)]" />}
                    <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)]">
                      {mem.category}
                    </span>
                    <span className="font-medium text-sm truncate">{mem.key}</span>
                  </div>
                  <p className="text-sm text-[var(--color-text-secondary)] line-clamp-2">{mem.content}</p>
                </div>
                <button
                  onClick={() => handleDelete(mem.id)}
                  className="opacity-0 group-hover:opacity-100 p-1 hover:bg-[var(--color-bg-tertiary)] rounded text-[var(--color-error)]"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </Card>
          ))
        )}
      </div>

      <Modal isOpen={showAdd} onClose={() => setShowAdd(false)} title="Add Memory">
        <form onSubmit={handleAdd} className="space-y-3">
          <Input label="Category" value={category} onChange={(e) => setCategory(e.target.value)} placeholder="general" />
          <Input label="Key" value={key} onChange={(e) => setKey(e.target.value)} placeholder="coding-style" required />
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">Content</label>
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Enter the memory content..."
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-sm text-[var(--color-text)] min-h-[100px] resize-y"
              required
            />
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={pinned} onChange={(e) => setPinned(e.target.checked)} className="rounded" />
            Pin (always included in agent context)
          </label>
          <Button type="submit" loading={loading} className="w-full">Save Memory</Button>
        </form>
      </Modal>
    </div>
  );
}
